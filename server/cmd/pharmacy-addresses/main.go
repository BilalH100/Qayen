// pharmacy-addresses fills blank addresses from unambiguous public evidence.
// It never creates pharmacies, coordinates, inventory, or verified-business claims.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	ci "kayena/server/internal/catalogimport"
)

type evidence struct {
	Name      string `json:"name"`
	City      string `json:"city"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	URL       string `json:"url"`
	PageURL   string `json:"page_url"`
	PageSHA   string `json:"page_sha256"`
	Retrieved string `json:"retrieved_at"`
}
type source struct {
	Provider string     `json:"provider"`
	Records  []evidence `json:"records"`
}
type pharmacy struct {
	ID                         int32
	Name, City, Address, Other string
}
type decision struct {
	ID       int32     `json:"pharmacy_id"`
	Name     string    `json:"name"`
	City     string    `json:"city"`
	Address  string    `json:"address,omitempty"`
	Status   string    `json:"status"`
	Evidence *evidence `json:"evidence,omitempty"`
}
type totals struct {
	Pharmacies  int `json:"pharmacies"`
	WithAddress int `json:"with_recorded_address"`
	Medications int `json:"medications"`
	Stock       int `json:"stock_rows"`
}
type report struct {
	Mode       string         `json:"mode"`
	SourceSHA  string         `json:"source_sha256"`
	Backup     string         `json:"backup,omitempty"`
	Before     totals         `json:"before"`
	After      totals         `json:"after"`
	SourceRows int            `json:"source_rows"`
	Updates    int            `json:"addresses_to_fill"`
	Statuses   map[string]int `json:"statuses"`
	Cities     map[string]int `json:"updated_by_city"`
	Decisions  []decision     `json:"decisions"`
	Notes      []string       `json:"notes"`
}

func digest(data []byte) string { h := sha256.Sum256(data); return hex.EncodeToString(h[:]) }
func key(name, city string) string {
	name = strings.TrimSpace(strings.TrimPrefix(ci.Normalize(name), "PHARMACIE "))
	city = strings.TrimSpace(strings.TrimRight(strings.TrimSuffix(ci.Normalize(city), "MAROC"), " -,"))
	switch ci.CompactIdentity(city) {
	case "FEZ":
		city = "FES"
	case "TANGIER", "TANGIERS":
		city = "TANGER"
	}
	return ci.CompactIdentity(name) + "\x1f" + ci.CompactIdentity(city)
}

var number = regexp.MustCompile(`[0-9]`)
var street = regexp.MustCompile(`\b(RUE|AVENUE|AV|BOULEVARD|BD|BOUL|ROUTE|LOT|LOTISSEMENT|RESIDENCE|RES|IMM|IMMEUBLE|LOCAL|MAGASIN|N[°O]|N[.]|PLACE|PL|ALLEE|BLOC|APPARTEMENT|APPT|TRANCHE)\b`)
var landmark = regexp.MustCompile(`\b(EN FACE|FACE A|A COTE|PRES DE|DEVANT|DERRIERE|ANGLE)\b`)
var numberedLocation = regexp.MustCompile(`(^[[:space:]]*[0-9]+|\b(N[°O.]|NUMERO|LOCAL|MAGASIN|IMMEUBLE|IMM|LOT|BLOC|APPARTEMENT|APPT)[[:space:].,:°-]*[A-Z]?[[:space:]]*[0-9]+)`)

func specificAddress(address string) bool {
	a := ci.Normalize(address)
	if len([]rune(a)) < 12 || strings.Contains(a, "?") || strings.Contains(a, "...") || strings.Contains(a, "…") {
		return false
	}
	if a == "N-D" || a == "N/A" || strings.Contains(a, "ADRESSE INCONNUE") {
		return false
	}
	// A district/city alone is not a usable exact address. Require a street or
	// numbered building/lot, or an explicit landmark description.
	return (number.MatchString(a) && numberedLocation.MatchString(a) && street.MatchString(a)) || (street.MatchString(a) && landmark.MatchString(a))
}

func validate(s source) error {
	if s.Provider != "Pharmacy.ma" || len(s.Records) == 0 {
		return fmt.Errorf("unsupported or empty evidence source")
	}
	hex64 := regexp.MustCompile(`^[a-f0-9]{64}$`)
	for i, r := range s.Records {
		if !hex64.MatchString(r.PageSHA) {
			return fmt.Errorf("invalid evidence record %d", i)
		}
		for _, value := range []string{r.URL, r.PageURL} {
			u, err := url.Parse(value)
			if err != nil || u.Scheme != "https" || u.Host != "pharmacy.ma" || u.User != nil {
				return fmt.Errorf("invalid source URL at record %d", i)
			}
		}
		if !strings.HasPrefix(r.URL, "https://pharmacy.ma/pharmacie/") || !strings.HasPrefix(r.PageURL, "https://pharmacy.ma/pharmacies") {
			return fmt.Errorf("unexpected evidence URL")
		}
		captured, err := time.Parse(time.RFC3339Nano, r.Retrieved)
		if err != nil || captured.After(time.Now().Add(time.Hour)) {
			return fmt.Errorf("invalid evidence date")
		}
	}
	return nil
}

func plan(pharmacies []pharmacy, records []evidence) []decision {
	existing := map[string]int{}
	sources := map[string]map[string]evidence{}
	for _, p := range pharmacies {
		existing[key(p.Name, p.City)]++
	}
	for _, r := range records {
		if strings.TrimSpace(r.Name) == "" || strings.TrimSpace(r.City) == "" {
			continue // Retain the raw evidence, but never match an unnamed listing.
		}
		k := key(r.Name, r.City)
		if sources[k] == nil {
			sources[k] = map[string]evidence{}
		}
		// Include URL and address to retain disagreements, even on a single listing.
		identity := r.URL + "\x1f" + ci.Normalize(r.Address)
		sources[k][identity] = r
	}
	result := make([]decision, 0, len(pharmacies))
	for _, p := range pharmacies {
		d := decision{ID: p.ID, Name: p.Name, City: p.City}
		k := key(p.Name, p.City)
		matches := sources[k]
		switch {
		case strings.TrimSpace(p.Address) != "":
			d.Status = "preserve_existing_address"
		case existing[k] > 1:
			d.Status = "ambiguous_database_identity"
		case len(matches) == 0:
			d.Status = "no_source_match"
		case len(matches) > 1:
			d.Status = "ambiguous_source_identity"
		default:
			for _, r := range matches {
				r := r
				d.Evidence = &r
				if specificAddress(r.Address) {
					d.Status = "fill_blank_address"
					d.Address = strings.TrimSpace(r.Address)
				} else {
					d.Status = "incomplete_source_address"
				}
			}
		}
		result = append(result, d)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func snapshot(ctx context.Context, tx pgx.Tx) ([]pharmacy, totals, string, error) {
	rows, err := tx.Query(ctx, `SELECT id,name,COALESCE(city,''),COALESCE(address,''),(to_jsonb(p)-'address')::text FROM pharmacies p ORDER BY id`)
	if err != nil {
		return nil, totals{}, "", err
	}
	var result []pharmacy
	for rows.Next() {
		var p pharmacy
		if err = rows.Scan(&p.ID, &p.Name, &p.City, &p.Address, &p.Other); err != nil {
			rows.Close()
			return nil, totals{}, "", err
		}
		result = append(result, p)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, totals{}, "", err
	}
	t := totals{Pharmacies: len(result)}
	for _, p := range result {
		if strings.TrimSpace(p.Address) != "" {
			t.WithAddress++
		}
	}
	var medHash, stockHash string
	err = tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM medications),(SELECT count(*) FROM stock),(SELECT md5(COALESCE(string_agg(to_jsonb(m)::text,'' ORDER BY id),'')) FROM medications m),(SELECT md5(COALESCE(string_agg(to_jsonb(s)::text,'' ORDER BY id),'')) FROM stock s)`).Scan(&t.Medications, &t.Stock, &medHash, &stockHash)
	return result, t, medHash + stockHash, err
}

func applyPlan(ctx context.Context, tx pgx.Tx, decisions []decision) error {
	for _, d := range decisions {
		if d.Status != "fill_blank_address" {
			continue
		}
		if d.Evidence == nil {
			return fmt.Errorf("missing evidence for %d", d.ID)
		}
		var previous string
		if err := tx.QueryRow(ctx, `SELECT COALESCE(address,'') FROM pharmacies WHERE id=$1 FOR UPDATE`, d.ID).Scan(&previous); err != nil {
			return err
		}
		if strings.TrimSpace(previous) != "" {
			return fmt.Errorf("address changed since preview for %d", d.ID)
		}
		tag, err := tx.Exec(ctx, `UPDATE pharmacies SET address=$1 WHERE id=$2 AND COALESCE(trim(address),'')=''`, d.Address, d.ID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return fmt.Errorf("address update mismatch")
		}
		r := d.Evidence
		data, _ := json.Marshal(r)
		captured, _ := time.Parse(time.RFC3339Nano, r.Retrieved)
		_, err = tx.Exec(ctx, `INSERT INTO pharmacy_address_evidence(pharmacy_id,provider,source_url,page_url,page_sha256,record_sha256,retrieved_at,previous_address,address,match_method,business_verified) VALUES($1,'Pharmacy.ma',$2,$3,$4,$5,$6,$7,$8,'unique_name_and_city_specific_address',false) ON CONFLICT(pharmacy_id,provider,record_sha256) DO NOTHING`, d.ID, r.URL, r.PageURL, r.PageSHA, digest(data), captured, previous, d.Address)
		if err != nil {
			return err
		}
	}
	return nil
}

func verify(before, after []pharmacy, decisions []decision) error {
	if len(before) != len(after) {
		return fmt.Errorf("pharmacy count changed")
	}
	updates := map[int32]string{}
	for _, d := range decisions {
		if d.Status == "fill_blank_address" {
			updates[d.ID] = d.Address
		}
	}
	for i, p := range before {
		q := after[i]
		expected := p.Address
		if value, ok := updates[p.ID]; ok {
			expected = value
		}
		if p.ID != q.ID || p.Other != q.Other || q.Address != expected {
			return fmt.Errorf("unexpected pharmacy change at %d", p.ID)
		}
	}
	return nil
}

func backup(ctx context.Context, executable, dbURL, path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return fmt.Errorf("backup path already exists or is inaccessible")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	config, err := pgx.ParseConfig(dbURL)
	if err != nil {
		return fmt.Errorf("invalid database configuration")
	}
	command := exec.CommandContext(ctx, executable, "--format=custom", "--no-owner", "--no-acl", "--no-password", "--file", path)
	ssl := "disable"
	if config.TLSConfig != nil {
		ssl = "require"
	}
	command.Env = append(os.Environ(), "PGHOST="+config.Host, "PGPORT="+strconv.Itoa(int(config.Port)), "PGDATABASE="+config.Database, "PGUSER="+config.User, "PGPASSWORD="+config.Password, "PGSSLMODE="+ssl, "PGCONNECT_TIMEOUT=10")
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err = command.Run(); err != nil {
		return fmt.Errorf("backup failed before writes: %s", strings.TrimSpace(stderr.String()))
	}
	restore := "pg_restore" + filepath.Ext(executable)
	if executable != filepath.Base(executable) {
		restore = filepath.Join(filepath.Dir(executable), restore)
	}
	if err = exec.CommandContext(ctx, restore, "--list", path).Run(); err != nil {
		return fmt.Errorf("backup verification failed")
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	input := flag.String("source", "", "Collected address evidence JSON")
	output := flag.String("report", "tmp/pharmacy-addresses-report.json", "Report destination")
	env := flag.String("env", "server/.env", "Database configuration")
	schema := flag.String("schema", "server/database/migrations/20261022000001_add_pharmacy_address_evidence.sql", "Audit schema")
	backupPath := flag.String("backup", "", "New backup archive required to apply")
	dump := flag.String("pg-dump", "pg_dump", "pg_dump executable")
	expected := flag.String("expected-sha256", "", "Reviewed source hash required to apply")
	apply := flag.Bool("apply", false, "Apply reviewed address additions")
	flag.Parse()
	data, err := os.ReadFile(*input)
	if err != nil {
		return err
	}
	var src source
	if err = json.Unmarshal(data, &src); err != nil {
		return err
	}
	if err = validate(src); err != nil {
		return err
	}
	hash := digest(data)
	if *apply && (*expected != hash || *backupPath == "") {
		return fmt.Errorf("apply requires backup path and exact reviewed source hash")
	}
	_ = godotenv.Load(*env)
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return fmt.Errorf("DB_URL missing")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	if *apply {
		if err = backup(ctx, *dump, dbURL, *backupPath); err != nil {
			return err
		}
	}
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("cannot connect to configured database")
	}
	defer conn.Close(context.Background())
	mode := pgx.ReadOnly
	if *apply {
		mode = pgx.ReadWrite
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable, AccessMode: mode})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	if *apply {
		if _, err = tx.Exec(ctx, `SET LOCAL lock_timeout='10s'; LOCK TABLE pharmacies IN SHARE ROW EXCLUSIVE MODE; LOCK TABLE medications,stock IN SHARE MODE`); err != nil {
			return err
		}
	}
	before, counts, fingerprint, err := snapshot(ctx, tx)
	if err != nil {
		return err
	}
	r := report{Mode: "preview", SourceSHA: hash, Before: counts, After: counts, SourceRows: len(src.Records), Statuses: map[string]int{}, Cities: map[string]int{}, Notes: []string{
		"Public directory addresses are source-matched, not independently verified current businesses.",
		"Only unambiguous name/city matches with a specific address fill blank addresses.",
		"Existing addresses, GPS coordinates, phone numbers, medication data and stock are preserved.",
		"The 2016 government workbook was examined but not used alone for address updates.",
		"Unmatched, ambiguous, incomplete and possibly closed pharmacies remain unchanged.",
	}}
	r.Decisions = plan(before, src.Records)
	for _, d := range r.Decisions {
		r.Statuses[d.Status]++
		if d.Status == "fill_blank_address" {
			r.Updates++
			r.Cities[d.City]++
		}
	}
	if *apply {
		schemaData, err := os.ReadFile(*schema)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, strings.Split(string(schemaData), "-- +goose Down")[0]); err != nil {
			return err
		}
		if err = applyPlan(ctx, tx, r.Decisions); err != nil {
			return err
		}
		after, afterCounts, afterHash, err := snapshot(ctx, tx)
		if err != nil {
			return err
		}
		if err = verify(before, after, r.Decisions); err != nil {
			return err
		}
		if afterHash != fingerprint || afterCounts.WithAddress != counts.WithAddress+r.Updates || afterCounts.Medications != counts.Medications || afterCounts.Stock != counts.Stock {
			return fmt.Errorf("catalog or inventory preservation check failed")
		}
		r.Mode = "applied"
		r.Backup = *backupPath
		r.After = afterCounts
	}
	if err = os.MkdirAll(filepath.Dir(*output), 0755); err != nil {
		return err
	}
	reportData, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(*output, append(reportData, '\n'), 0600); err != nil {
		return err
	}
	if *apply {
		if err = tx.Commit(ctx); err != nil {
			r.Mode = "rolled_back"
			failure, _ := json.MarshalIndent(r, "", "  ")
			_ = os.WriteFile(*output, failure, 0600)
			return err
		}
	}
	fmt.Printf("%s: %d specific addresses to fill across %d cities; %d source rows; report %s\n", r.Mode, r.Updates, len(r.Cities), r.SourceRows, *output)
	return nil
}
