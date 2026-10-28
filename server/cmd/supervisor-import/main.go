// supervisor-import imports the two supervisor-provided references. It never
// overwrites AMMPS medication details or creates pharmacy addresses/inventory.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	ci "kayena/server/internal/catalogimport"
)

type medicationDecision struct {
	Row                 ci.CNOPSReference
	ID                  *int32
	Status, Fingerprint string
	Existing, Conflict  bool
}
type pharmacyDecision struct {
	Row         ci.PharmacyReference
	ID          *int32
	Status, Key string
}
type counts struct {
	Medications int `json:"medications"`
	Pharmacies  int `json:"pharmacies"`
	Stock       int `json:"stock_rows"`
	References  int `json:"cnops_references"`
	Directory   int `json:"pharmacy_source_rows"`
}
type report struct {
	Mode                string        `json:"mode"`
	CreatedAt           time.Time     `json:"created_at"`
	BundleSHA256        string        `json:"bundle_sha256"`
	Backup              string        `json:"backup,omitempty"`
	MedicationSource    ci.SourceFile `json:"medication_source"`
	PharmacySource      ci.SourceFile `json:"pharmacy_source"`
	Before              counts        `json:"before"`
	After               counts        `json:"after"`
	CNOPSRows           int           `json:"cnops_source_rows"`
	NewReferences       int           `json:"cnops_references_to_insert"`
	ExistingReferences  int           `json:"cnops_references_existing"`
	MatchedReferences   int           `json:"cnops_exact_matches"`
	UnmatchedReferences int           `json:"cnops_unmatched_for_review"`
	AmbiguousReferences int           `json:"cnops_ambiguous_for_review"`
	ReferenceConflicts  int           `json:"conflicting_existing_cnops_codes"`
	PriceDifferences    int           `json:"matched_prices_different_from_ammps"`
	NewPharmacies       int           `json:"pharmacies_to_insert"`
	MatchedPharmacies   int           `json:"pharmacy_rows_matched"`
	DuplicatePharmacies int           `json:"duplicate_source_pharmacy_rows"`
	MissingCity         int           `json:"pharmacy_rows_missing_city"`
	AmbiguousPharmacies int           `json:"ambiguous_pharmacy_rows"`
	MissingDosage       int           `json:"cnops_rows_missing_dosage"`
	Issues              []string      `json:"review_items"`
	Notes               []string      `json:"notes"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	bundlePath := flag.String("bundle", "", "Validated supervisor bundle")
	envPath := flag.String("env", "server/.env", "Database environment file")
	schemaPath := flag.String("schema", "server/database/migrations/20261022000000_add_catalog_references.sql", "Reference schema migration")
	reportPath := flag.String("report", "tmp/supervisor-import-report.json", "Report destination")
	backupPath := flag.String("backup", "", "Required new pg_dump archive for -apply")
	pgDump := flag.String("pg-dump", "pg_dump", "Path to pg_dump executable")
	expected := flag.String("expected-sha256", "", "Required reviewed bundle SHA256 for -apply")
	apply := flag.Bool("apply", false, "Apply import after backup; default is read-only preview")
	flag.Parse()
	data, err := os.ReadFile(*bundlePath)
	if err != nil {
		return err
	}
	var bundle ci.SupervisorBundle
	if err = json.Unmarshal(data, &bundle); err != nil {
		return err
	}
	if err = validateBundle(bundle); err != nil {
		return err
	}
	digest := sha256.Sum256(data)
	hash := hex.EncodeToString(digest[:])
	if *apply && (*expected != hash || *backupPath == "") {
		return fmt.Errorf("-apply requires -backup and the exact previewed -expected-sha256")
	}
	_ = godotenv.Load(*envPath)
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return fmt.Errorf("DB_URL is missing")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	if *apply {
		if err = backupDatabase(ctx, *pgDump, dbURL, *backupPath); err != nil {
			return err
		}
	}
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("could not connect to configured database")
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
		if _, err = tx.Exec(ctx, `SET LOCAL lock_timeout='10s'; LOCK TABLE medications, pharmacies IN SHARE ROW EXCLUSIVE MODE`); err != nil {
			return err
		}
	}
	r := report{Mode: "preview", CreatedAt: time.Now().UTC(), BundleSHA256: hash, MedicationSource: bundle.MedicationSource, PharmacySource: bundle.PharmacySource, Notes: []string{
		"CNOPS is a separate reference, not proof of marketing authorization or current stock.",
		"Only exact name, ingredient, dosage, form and packaging matches are linked.",
		"Existing AMMPS medication values and prices are not overwritten; differences require review.",
		"New pharmacy entries contain name/city only and require an address, GPS pin and real inventory before alerts.",
		"Missing-city, ambiguous and unmatched source rows are retained for review.",
	}}
	r.Before, err = databaseCounts(ctx, tx)
	if err != nil {
		return err
	}
	meds, prices, err := loadMedicationIdentities(ctx, tx)
	if err != nil {
		return err
	}
	existingPharmacies, existingIDs, err := loadPharmacies(ctx, tx)
	if err != nil {
		return err
	}
	beforeHash, err := existingPharmacyHash(ctx, tx, existingIDs)
	if err != nil {
		return err
	}
	existingReferences, err := loadReferenceFingerprints(ctx, tx)
	if err != nil {
		return err
	}
	medicationPlan := planReferences(bundle.Medications, meds, prices, existingReferences, &r)
	pharmacyPlan := planDirectory(bundle.Pharmacies, existingPharmacies, &r)
	if *apply {
		schema, readErr := os.ReadFile(*schemaPath)
		if readErr != nil {
			return readErr
		}
		up := strings.Split(string(schema), "-- +goose Down")[0]
		if _, err = tx.Exec(ctx, up); err != nil {
			return fmt.Errorf("reference schema: %w", err)
		}
		if err = applyPlan(ctx, tx, bundle, medicationPlan, pharmacyPlan); err != nil {
			return err
		}
		r.After, err = databaseCounts(ctx, tx)
		if err != nil {
			return err
		}
		if r.After.Medications != r.Before.Medications || r.After.Stock != r.Before.Stock || r.After.Pharmacies != r.Before.Pharmacies+r.NewPharmacies || r.After.References != r.Before.References+r.NewReferences {
			return fmt.Errorf("post-import counts differ from plan; transaction rolled back")
		}
		afterHash, hashErr := existingPharmacyHash(ctx, tx, existingIDs)
		if hashErr != nil {
			return hashErr
		}
		if afterHash != beforeHash {
			return fmt.Errorf("existing pharmacy records changed; transaction rolled back")
		}
		r.Mode = "applied"
		r.Backup = *backupPath
		// Check that the report destination is writable before committing data.
		if err = writeReport(*reportPath, r); err != nil {
			return err
		}
		if err = tx.Commit(ctx); err != nil {
			r.Mode = "rolled_back"
			_ = writeReport(*reportPath, r)
			return fmt.Errorf("commit failed: %w", err)
		}
	} else {
		r.After = r.Before
		if err = writeReport(*reportPath, r); err != nil {
			return err
		}
	}
	fmt.Printf("%s: CNOPS references +%d (%d exact links); pharmacy entries +%d; review items %d; report %s\n", r.Mode, r.NewReferences, r.MatchedReferences, r.NewPharmacies, len(r.Issues), *reportPath)
	return nil
}

func validateBundle(b ci.SupervisorBundle) error {
	if len(b.Medications) == 0 || len(b.Pharmacies) == 0 {
		return fmt.Errorf("both source datasets are required")
	}
	hexHash := regexp.MustCompile(`^[a-f0-9]{64}$`)
	codePattern := regexp.MustCompile(`^\d{13}$`)
	money := regexp.MustCompile(`^\d+(?:\.\d+)?$`)
	for _, s := range []ci.SourceFile{b.MedicationSource, b.PharmacySource} {
		if s.Filename == "" || s.Sheet == "" || !hexHash.MatchString(s.SHA256) {
			return fmt.Errorf("invalid source provenance")
		}
	}
	seen := make(map[string]bool)
	for _, m := range b.Medications {
		if !codePattern.MatchString(m.Code) || seen[m.Code] || m.SourceRow < 2 || strings.TrimSpace(m.Name) == "" || strings.TrimSpace(m.Ingredient) == "" || strings.TrimSpace(m.Form) == "" || strings.TrimSpace(m.Presentation) == "" {
			return fmt.Errorf("invalid or duplicate CNOPS product at row %d", m.SourceRow)
		}
		seen[m.Code] = true
		for _, v := range []*string{m.PublicPrice, m.HospitalPrice, m.ReimbursementBase} {
			if v != nil && !money.MatchString(*v) {
				return fmt.Errorf("invalid source price at row %d", m.SourceRow)
			}
		}
	}
	seenRows := make(map[int]bool)
	for _, p := range b.Pharmacies {
		if p.SourceRow < 5 || strings.TrimSpace(p.Name) == "" || seenRows[p.SourceRow] {
			return fmt.Errorf("invalid pharmacy source row %d", p.SourceRow)
		}
		seenRows[p.SourceRow] = true
	}
	return nil
}

func referenceFingerprint(m ci.CNOPSReference) string {
	m.SourceRow = 0
	data, _ := json.Marshal(m)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func planReferences(incoming []ci.CNOPSReference, existing map[string][]int32, prices map[int32]string, fingerprints map[string]string, r *report) []medicationDecision {
	var plan []medicationDecision
	r.CNOPSRows = len(incoming)
	for _, row := range incoming {
		d := medicationDecision{Row: row, Status: "unmatched", Fingerprint: referenceFingerprint(row)}
		if old, ok := fingerprints[row.Code]; ok {
			d.Existing = true
			r.ExistingReferences++
			if old != d.Fingerprint {
				d.Conflict = true
				r.ReferenceConflicts++
				r.Issues = append(r.Issues, "Existing CNOPS code has different source values: "+row.Code)
			}
		} else {
			r.NewReferences++
		}
		if row.Dosage == "" {
			r.MissingDosage++
		}
		key := row.MatchKey()
		matches := existing[key]
		if key != "" && len(matches) == 1 {
			id := matches[0]
			d.ID = &id
			d.Status = "matched"
			r.MatchedReferences++
			if row.PublicPrice != nil && !samePrice(*row.PublicPrice, prices[id]) {
				r.PriceDifferences++
				r.Issues = append(r.Issues, fmt.Sprintf("Price retained separately for CNOPS %s / medication %d", row.Code, id))
			}
		} else if len(matches) > 1 {
			d.Status = "ambiguous"
			r.AmbiguousReferences++
			r.Issues = append(r.Issues, "Ambiguous medication match: "+row.Code+" / "+row.Name)
		} else {
			r.UnmatchedReferences++
		}
		plan = append(plan, d)
	}
	return plan
}

func samePrice(a, b string) bool {
	left, e1 := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(a), ",", "."), 64)
	right, e2 := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(b), ",", "."), 64)
	return e1 == nil && e2 == nil && left == right
}

func planDirectory(incoming []ci.PharmacyReference, existing map[string][]int32, r *report) []pharmacyDecision {
	var plan []pharmacyDecision
	newKeys := make(map[string]bool)
	for _, row := range incoming {
		d := pharmacyDecision{Row: row, Key: ci.DirectoryKey(row.Name, row.City)}
		if strings.TrimSpace(row.City) == "" {
			d.Status = "missing_city"
			r.MissingCity++
			r.Issues = append(r.Issues, fmt.Sprintf("Missing city, pharmacy source row %d: %s", row.SourceRow, row.Name))
		} else if matches := existing[d.Key]; len(matches) > 1 {
			d.Status = "ambiguous"
			r.AmbiguousPharmacies++
			r.Issues = append(r.Issues, "Ambiguous pharmacy name/city: "+row.Name+" / "+row.City)
		} else if len(matches) == 1 {
			id := matches[0]
			d.ID = &id
			d.Status = "matched"
			r.MatchedPharmacies++
		} else {
			d.Status = "inserted"
			if newKeys[d.Key] {
				r.DuplicatePharmacies++
			} else {
				newKeys[d.Key] = true
				r.NewPharmacies++
			}
		}
		plan = append(plan, d)
	}
	return plan
}

func loadMedicationIdentities(ctx context.Context, tx pgx.Tx) (map[string][]int32, map[int32]string, error) {
	rows, err := tx.Query(ctx, `SELECT id,COALESCE(speciality,''),COALESCE(active_substance,''),COALESCE(dosage,''),COALESCE(form,''),COALESCE(presentation,''),COALESCE(ppv,'') FROM medications`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	identities := make(map[string][]int32)
	prices := make(map[int32]string)
	for rows.Next() {
		var id int32
		var name, ingredient, dose, form, pack, price string
		if err = rows.Scan(&id, &name, &ingredient, &dose, &form, &pack, &price); err != nil {
			return nil, nil, err
		}
		key := ci.ReferenceKey(name, ingredient, dose, form, pack)
		if key != "" {
			identities[key] = append(identities[key], id)
		}
		prices[id] = price
	}
	return identities, prices, rows.Err()
}

func loadPharmacies(ctx context.Context, tx pgx.Tx) (map[string][]int32, []int32, error) {
	rows, err := tx.Query(ctx, `SELECT id,name,COALESCE(city,'') FROM pharmacies`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	identities := make(map[string][]int32)
	var ids []int32
	for rows.Next() {
		var id int32
		var name, city string
		if err = rows.Scan(&id, &name, &city); err != nil {
			return nil, nil, err
		}
		key := ci.DirectoryKey(name, city)
		identities[key] = append(identities[key], id)
		ids = append(ids, id)
	}
	return identities, ids, rows.Err()
}

func tableExists(ctx context.Context, tx pgx.Tx, name string) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `SELECT to_regclass($1) IS NOT NULL`, name).Scan(&exists)
	return exists, err
}

func loadReferenceFingerprints(ctx context.Context, tx pgx.Tx) (map[string]string, error) {
	result := make(map[string]string)
	exists, err := tableExists(ctx, tx, "medication_reference_entries")
	if err != nil || !exists {
		return result, err
	}
	rows, err := tx.Query(ctx, `SELECT product_code,content_sha256 FROM medication_reference_entries WHERE provider='CNOPS'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var code, hash string
		if err = rows.Scan(&code, &hash); err != nil {
			return nil, err
		}
		result[code] = hash
	}
	return result, rows.Err()
}

func databaseCounts(ctx context.Context, tx pgx.Tx) (counts, error) {
	var c counts
	err := tx.QueryRow(ctx, `SELECT (SELECT COUNT(*) FROM medications),(SELECT COUNT(*) FROM pharmacies),(SELECT COUNT(*) FROM stock)`).Scan(&c.Medications, &c.Pharmacies, &c.Stock)
	if err != nil {
		return c, err
	}
	for name, target := range map[string]*int{"medication_reference_entries": &c.References, "pharmacy_directory_entries": &c.Directory} {
		exists, e := tableExists(ctx, tx, name)
		if e != nil {
			return c, e
		}
		if exists {
			if e = tx.QueryRow(ctx, "SELECT COUNT(*) FROM "+name).Scan(target); e != nil {
				return c, e
			}
		}
	}
	return c, nil
}

func existingPharmacyHash(ctx context.Context, tx pgx.Tx, ids []int32) (string, error) {
	var hash string
	err := tx.QueryRow(ctx, `SELECT COALESCE(md5(string_agg(to_jsonb(p)::text,'|' ORDER BY id)),'') FROM pharmacies p WHERE id=ANY($1::int[])`, ids).Scan(&hash)
	return hash, err
}

func sourceID(ctx context.Context, tx pgx.Tx, provider string, s ci.SourceFile) (int64, error) {
	var id int64
	err := tx.QueryRow(ctx, `INSERT INTO catalog_import_sources(provider,filename,sheet_name,sha256) VALUES($1,$2,$3,$4) ON CONFLICT(provider,sha256) DO UPDATE SET filename=EXCLUDED.filename RETURNING id`, provider, s.Filename, s.Sheet, s.SHA256).Scan(&id)
	return id, err
}

func applyPlan(ctx context.Context, tx pgx.Tx, b ci.SupervisorBundle, medications []medicationDecision, pharmacies []pharmacyDecision) error {
	medSource, err := sourceID(ctx, tx, "CNOPS", b.MedicationSource)
	if err != nil {
		return err
	}
	pharmacySource, err := sourceID(ctx, tx, "AMMPS", b.PharmacySource)
	if err != nil {
		return err
	}
	batch := &pgx.Batch{}
	for _, d := range medications {
		if d.Conflict {
			continue
		}
		m := d.Row
		batch.Queue(`INSERT INTO medication_reference_entries(provider,product_code,name,active_ingredient,dosage,dosage_unit,form,presentation,public_price,hospital_price,reimbursement_base,product_type,reimbursement_rate,medication_id,match_status,source_id,source_row,content_sha256) VALUES('CNOPS',$1,$2,$3,$4,$5,$6,$7,$8::numeric,$9::numeric,$10::numeric,$11,$12,$13,$14,$15,$16,$17) ON CONFLICT(provider,product_code) DO UPDATE SET medication_id=COALESCE(medication_reference_entries.medication_id,EXCLUDED.medication_id),match_status=CASE WHEN medication_reference_entries.medication_id IS NOT NULL THEN medication_reference_entries.match_status ELSE EXCLUDED.match_status END WHERE medication_reference_entries.content_sha256=EXCLUDED.content_sha256`, m.Code, m.Name, m.Ingredient, m.Dosage, m.DosageUnit, m.Form, m.Presentation, m.PublicPrice, m.HospitalPrice, m.ReimbursementBase, m.ProductType, m.ReimbursementRate, d.ID, d.Status, medSource, m.SourceRow, d.Fingerprint)
	}
	results := tx.SendBatch(ctx, batch)
	if err = results.Close(); err != nil {
		return fmt.Errorf("import CNOPS references: %w", err)
	}
	newIDs := make(map[string]int32)
	for _, d := range pharmacies {
		id := d.ID
		if d.Status == "inserted" {
			value, ok := newIDs[d.Key]
			if !ok {
				// The existing SQLC model scans address and phone into strings.
				// Keep unknown text blank and unknown coordinates NULL.
				if err = tx.QueryRow(ctx, `INSERT INTO pharmacies(name,city,address,phone) VALUES($1,$2,'','') RETURNING id`, d.Row.Name, d.Row.City).Scan(&value); err != nil {
					return err
				}
				newIDs[d.Key] = value
			}
			id = &value
		}
		if _, err = tx.Exec(ctx, `INSERT INTO pharmacy_directory_entries(source_id,source_row,name,city,province,pharmacy_id,match_status) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(source_id,source_row) DO UPDATE SET pharmacy_id=COALESCE(pharmacy_directory_entries.pharmacy_id,EXCLUDED.pharmacy_id),match_status=EXCLUDED.match_status`, pharmacySource, d.Row.SourceRow, d.Row.Name, d.Row.City, d.Row.Province, id, d.Status); err != nil {
			return err
		}
	}
	return nil
}

func backupDatabase(ctx context.Context, executable, dbURL, path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return fmt.Errorf("backup destination already exists or is inaccessible; choose a new path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	config, err := pgx.ParseConfig(dbURL)
	if err != nil {
		return fmt.Errorf("invalid database configuration")
	}
	cmd := exec.CommandContext(ctx, executable, "--format=custom", "--no-owner", "--no-acl", "--no-password", "--file", path)
	sslMode := "disable"
	if config.TLSConfig != nil {
		sslMode = "require"
	}
	cmd.Env = append(os.Environ(), "PGHOST="+config.Host, "PGPORT="+strconv.Itoa(int(config.Port)), "PGDATABASE="+config.Database, "PGUSER="+config.User, "PGPASSWORD="+config.Password, "PGSSLMODE="+sslMode, "PGCONNECT_TIMEOUT=10")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("database backup failed before import: %w; %s", err, strings.TrimSpace(stderr.String()))
	}
	info, err := os.Stat(path)
	if err != nil || info.Size() == 0 {
		return fmt.Errorf("database backup is empty")
	}
	restoreName := "pg_restore" + filepath.Ext(executable)
	if executable != filepath.Base(executable) {
		restoreName = filepath.Join(filepath.Dir(executable), restoreName)
	}
	if err = exec.CommandContext(ctx, restoreName, "--list", path).Run(); err != nil {
		return fmt.Errorf("backup archive verification failed before import: %w", err)
	}
	return nil
}

func writeReport(path string, r report) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0600)
}
