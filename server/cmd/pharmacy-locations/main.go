package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type source struct {
	Provider string     `json:"provider"`
	Records  []location `json:"records"`
	Complete bool       `json:"complete"`
}
type totals struct {
	Pharmacies  int `json:"pharmacies"`
	Addressed   int `json:"with_address"`
	Located     int `json:"with_coordinates"`
	Medications int `json:"medications"`
	Stock       int `json:"stock_rows"`
}
type report struct {
	Mode      string         `json:"mode"`
	SHA       string         `json:"evidence_sha256"`
	Complete  bool           `json:"source_collection_complete"`
	Backup    string         `json:"backup,omitempty"`
	Before    totals         `json:"before"`
	After     totals         `json:"after"`
	Updates   int            `json:"coordinates_to_fill"`
	Cities    map[string]int `json:"updated_by_city"`
	Statuses  map[string]int `json:"statuses"`
	Decisions []decision     `json:"decisions"`
	Notes     []string       `json:"notes"`
}

func digest(data []byte) string { h := sha256.Sum256(data); return hex.EncodeToString(h[:]) }
func validate(src source) error {
	if src.Provider != "pharmap.ma" || len(src.Records) == 0 {
		return fmt.Errorf("unsupported or empty location evidence")
	}
	hash := regexp.MustCompile(`^[a-f0-9]{64}$`)
	seen := map[string]bool{}
	for _, r := range src.Records {
		if !hash.MatchString(r.PageSHA) || strings.TrimSpace(r.Name) == "" || strings.TrimSpace(r.City) == "" || seen[r.URL] {
			return fmt.Errorf("invalid or duplicate source record")
		}
		seen[r.URL] = true
		for _, raw := range []string{r.URL, r.PageURL} {
			u, err := url.Parse(raw)
			if err != nil || u.Scheme != "https" || u.Host != "pharmap.ma" || u.User != nil || !strings.HasPrefix(u.Path, "/pharmacie/") {
				return fmt.Errorf("invalid source URL")
			}
		}
		t, err := time.Parse(time.RFC3339Nano, r.Retrieved)
		if err != nil || t.After(time.Now().Add(time.Hour)) {
			return fmt.Errorf("invalid retrieval time")
		}
	}
	return nil
}
func snapshot(ctx context.Context, tx pgx.Tx) ([]pharmacy, totals, string, error) {
	rows, err := tx.Query(ctx, `SELECT id,name,COALESCE(city,''),COALESCE(address,''),COALESCE(phone,''),latitude,longitude,(to_jsonb(p)-'latitude'-'longitude')::text FROM pharmacies p ORDER BY id`)
	if err != nil {
		return nil, totals{}, "", err
	}
	var result []pharmacy
	for rows.Next() {
		var p pharmacy
		if err = rows.Scan(&p.ID, &p.Name, &p.City, &p.Address, &p.Phone, &p.Latitude, &p.Longitude, &p.Other); err != nil {
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
			t.Addressed++
		}
		if p.Latitude != nil && p.Longitude != nil && !math.IsNaN(*p.Latitude) && !math.IsNaN(*p.Longitude) {
			t.Located++
		}
	}
	var a, b string
	err = tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM medications),(SELECT count(*) FROM stock),(SELECT md5(COALESCE(string_agg(to_jsonb(m)::text,'' ORDER BY id),'')) FROM medications m),(SELECT md5(COALESCE(string_agg(to_jsonb(s)::text,'' ORDER BY id),'')) FROM stock s)`).Scan(&t.Medications, &t.Stock, &a, &b)
	return result, t, a + b, err
}
func applyPlan(ctx context.Context, tx pgx.Tx, decisions []decision) error {
	for _, d := range decisions {
		if d.Status != "fill_missing_coordinates" {
			continue
		}
		r := d.Evidence
		if r == nil || !precisePoint(*r) || d.Method == "" {
			return fmt.Errorf("invalid planned location")
		}
		tag, err := tx.Exec(ctx, `UPDATE pharmacies SET latitude=$1,longitude=$2 WHERE id=$3 AND address=$4 AND latitude IS NULL AND longitude IS NULL`, *r.Latitude, *r.Longitude, d.ID, d.Address)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return fmt.Errorf("location or address changed since preview for %d", d.ID)
		}
		data, _ := json.Marshal(r)
		captured, _ := time.Parse(time.RFC3339Nano, r.Retrieved)
		_, err = tx.Exec(ctx, `INSERT INTO pharmacy_location_evidence(pharmacy_id,provider,source_url,page_sha256,record_sha256,retrieved_at,source_address,matched_address,latitude,longitude,match_method,business_verified) VALUES($1,'pharmap.ma',$2,$3,$4,$5,$6,$7,$8,$9,$10,false)`, d.ID, r.URL, r.PageSHA, digest(data), captured, r.Address, d.Address, *r.Latitude, *r.Longitude, d.Method)
		if err != nil {
			return err
		}
	}
	return nil
}
func same(a, b *float64) bool { return a == nil && b == nil || a != nil && b != nil && *a == *b }
func verify(before, after []pharmacy, decisions []decision) error {
	if len(before) != len(after) {
		return fmt.Errorf("pharmacy count changed")
	}
	updates := map[int32]*location{}
	for _, d := range decisions {
		if d.Status == "fill_missing_coordinates" {
			updates[d.ID] = d.Evidence
		}
	}
	for i, p := range before {
		q := after[i]
		lat, lon := p.Latitude, p.Longitude
		if r := updates[p.ID]; r != nil {
			lat, lon = r.Latitude, r.Longitude
		}
		if p.ID != q.ID || p.Other != q.Other || !same(lat, q.Latitude) || !same(lon, q.Longitude) {
			return fmt.Errorf("unexpected pharmacy change at %d", p.ID)
		}
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
	input := flag.String("source", "", "Published location JSON")
	reference := flag.String("address-reference", "", "Previously collected address evidence JSON")
	output := flag.String("report", "tmp/pharmacy-locations-report.json", "Report path")
	env := flag.String("env", "server/.env", "DB environment file")
	schema := flag.String("schema", "server/database/migrations/20261022000002_add_pharmacy_location_evidence.sql", "Audit schema")
	archive := flag.String("backup", "", "New backup archive required to apply")
	dump := flag.String("pg-dump", "pg_dump", "pg_dump path")
	expected := flag.String("expected-sha256", "", "Reviewed combined evidence hash")
	apply := flag.Bool("apply", false, "Apply reviewed locations")
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
	refData, err := os.ReadFile(*reference)
	if err != nil {
		return err
	}
	var refs struct {
		Records []addressReference `json:"records"`
	}
	if err = json.Unmarshal(refData, &refs); err != nil {
		return err
	}
	sha := digest(append(append([]byte{}, data...), refData...))
	if *apply && (!src.Complete || *archive == "" || *expected != sha) {
		return fmt.Errorf("apply requires completed collection, backup path and reviewed combined evidence SHA256")
	}
	_ = godotenv.Load(*env)
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return fmt.Errorf("DB_URL missing")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	if *apply {
		if err = backup(ctx, *dump, dbURL, *archive); err != nil {
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
	r := report{Mode: "preview", SHA: sha, Complete: src.Complete, Before: counts, After: counts, Cities: map[string]int{}, Statuses: map[string]int{}, Notes: []string{"Published pharmacy POI points only; no city centres or street geocoding.", "Unique pharmacy name/city plus corroborated street address are required.", "Shared, coarse, ambiguous and conflicting points are held; existing GPS is preserved.", "Source directory locations are not independently verified on site or proof of current opening/stock.", "Medication/stock data and all other pharmacy fields must remain unchanged."}}
	r.Decisions = plan(before, src.Records, refs.Records)
	for _, d := range r.Decisions {
		r.Statuses[d.Status]++
		if d.Status == "fill_missing_coordinates" {
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
		if fingerprint != afterHash || afterCounts.Located != counts.Located+r.Updates || afterCounts.Addressed != counts.Addressed || afterCounts.Medications != counts.Medications || afterCounts.Stock != counts.Stock {
			return fmt.Errorf("preservation checks failed")
		}
		r.After = afterCounts
		r.Mode = "applied"
		r.Backup = *archive
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
			failed, _ := json.MarshalIndent(r, "", "  ")
			_ = os.WriteFile(*output, failed, 0600)
			return err
		}
	}
	fmt.Printf("%s: %d pharmacy locations across %d cities; collection complete=%v; report %s\n", r.Mode, r.Updates, len(r.Cities), src.Complete, *output)
	return nil
}
