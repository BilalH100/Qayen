// estimated-alert-testing activates isolated, explicitly marked alert fixtures.
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
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type location struct {
	ID        int32   `json:"pharmacy_id"`
	Address   string  `json:"address"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
type medication struct {
	ID           int32  `json:"id"`
	Name         string `json:"name"`
	Dosage       string `json:"dosage"`
	Presentation string `json:"presentation"`
}
type report struct {
	Mode        string       `json:"mode"`
	SHA         string       `json:"source_sha256"`
	Locations   int          `json:"test_locations"`
	StockRows   int          `json:"test_stock_rows"`
	Medications []medication `json:"test_medications"`
	Backup      string       `json:"backup,omitempty"`
	Notes       []string     `json:"notes"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	input := flag.String("locations", "", "Reviewed city-area locations JSON")
	output := flag.String("report", "tmp/estimated-alert-testing-report.json", "Report path")
	envPath := flag.String("env", "server/.env", "Local backend environment")
	schema := flag.String("schema", "server/database/migrations/20261026000000_add_estimated_alert_testing.sql", "Schema migration")
	apply := flag.Bool("apply", false, "Activate test fixtures and local test flag after backup")
	expected := flag.String("expected-sha256", "", "Reviewed input SHA256")
	archive := flag.String("backup", "", "New full backup path")
	dump := flag.String("pg-dump", "pg_dump", "pg_dump executable")
	flag.Parse()
	data, err := os.ReadFile(*input)
	if err != nil {
		return err
	}
	var source struct {
		Locations []location `json:"locations"`
	}
	if err = json.Unmarshal(data, &source); err != nil {
		return err
	}
	digest := sha256.Sum256(data)
	hash := hex.EncodeToString(digest[:])
	if len(source.Locations) == 0 {
		return fmt.Errorf("no test locations supplied")
	}
	seen := map[int32]bool{}
	for _, p := range source.Locations {
		if p.ID <= 0 || seen[p.ID] || p.Address == "" || p.City == "" || math.IsNaN(p.Latitude) || math.IsNaN(p.Longitude) || p.Latitude < 20 || p.Latitude > 36.1 || p.Longitude < -18 || p.Longitude > -0.8 {
			return fmt.Errorf("invalid test location %d", p.ID)
		}
		seen[p.ID] = true
	}
	if *apply && (*expected != hash || *archive == "") {
		return fmt.Errorf("apply requires reviewed source SHA256 and a new backup path")
	}
	config, err := godotenv.Read(*envPath)
	if err != nil {
		return fmt.Errorf("cannot read database configuration")
	}
	dbURL := config["DB_URL"]
	if dbURL == "" {
		return fmt.Errorf("DB_URL missing")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	var prepared string
	if *apply {
		if err = backup(ctx, *dump, dbURL, *archive); err != nil {
			return err
		}
		original, err := os.ReadFile(*envPath)
		if err != nil {
			return err
		}
		flagLine := "QAYEN_ENABLE_APPROXIMATE_ALERTS=true"
		pattern := regexp.MustCompile(`(?m)^[ \t]*QAYEN_ENABLE_APPROXIMATE_ALERTS=.*$`)
		contents := string(original)
		if pattern.MatchString(contents) {
			contents = pattern.ReplaceAllString(contents, flagLine)
		} else {
			contents = strings.TrimRight(contents, "\r\n") + "\n" + flagLine + "\n"
		}
		file, err := os.CreateTemp(filepath.Dir(*envPath), ".estimated-alerts-env-*")
		if err != nil {
			return err
		}
		prepared = file.Name()
		defer os.Remove(prepared)
		if _, err = file.WriteString(contents); err != nil {
			file.Close()
			return err
		}
		if err = file.Close(); err != nil {
			return err
		}
	}
	db, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("cannot connect to database")
	}
	defer db.Close(context.Background())
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	if *apply {
		if _, err = tx.Exec(ctx, `SET LOCAL lock_timeout='10s'; LOCK TABLE pharmacies,stock,medications IN SHARE MODE`); err != nil {
			return err
		}
	}
	fingerprint := func() (string, error) {
		var value string
		err := tx.QueryRow(ctx, `SELECT md5((SELECT string_agg(to_jsonb(p)::text,'' ORDER BY id) FROM pharmacies p)||(SELECT string_agg(to_jsonb(m)::text,'' ORDER BY id) FROM medications m)||COALESCE((SELECT string_agg(to_jsonb(s)::text,'' ORDER BY id) FROM stock s),''))`).Scan(&value)
		return value, err
	}
	before, err := fingerprint()
	if err != nil {
		return err
	}
	for _, p := range source.Locations {
		var valid bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pharmacies WHERE id=$1 AND address=$2 AND city=$3 AND latitude IS NULL AND longitude IS NULL)`, p.ID, p.Address, p.City).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return fmt.Errorf("pharmacy %d changed since the location snapshot; regenerate the input", p.ID)
		}
	}
	rows, err := tx.Query(ctx, `WITH candidates AS (
 SELECT m.id,m.speciality,m.dosage,m.presentation,ROW_NUMBER() OVER(PARTITION BY lower(trim(m.speciality)) ORDER BY count(s.id) DESC,m.id) AS rank
 FROM medications m JOIN stock s ON s.medication_id=m.id
 WHERE lower(trim(m.speciality)) IN ('codoliprane','rhumix','rinomicine')
 GROUP BY m.id,m.speciality,m.dosage,m.presentation)
 SELECT id,speciality,COALESCE(dosage,''),COALESCE(presentation,'') FROM candidates WHERE rank=1 ORDER BY id`)
	if err != nil {
		return err
	}
	meds := []medication{}
	for rows.Next() {
		var m medication
		if err = rows.Scan(&m.ID, &m.Name, &m.Dosage, &m.Presentation); err != nil {
			rows.Close()
			return err
		}
		meds = append(meds, m)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if len(meds) != 3 {
		return fmt.Errorf("expected existing stock for all three test medications; found %d", len(meds))
	}
	r := report{Mode: "preview", SHA: hash, Locations: len(source.Locations), StockRows: len(source.Locations) * len(meds), Medications: meds, Notes: []string{"City-area estimates and stock are test fixtures, not verified pharmacy locations or real inventory.", "Both local environment flag and database test setting must be enabled.", "Original pharmacy GPS, medications, users and stock rows are preserved.", "Unmanaged pharmacies use the existing automatic response flow; managed accounts retain manual dashboard responses.", "Only the three listed medication SKUs have simulated stock; other products still require actual stock.", "Restart the backend after activation. Set QAYEN_ENABLE_APPROXIMATE_ALERTS=false and restart to disable test locations and inventory."}}
	if *apply {
		schemaData, err := os.ReadFile(*schema)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, strings.Split(string(schemaData), "-- +goose Down")[0]); err != nil {
			return err
		}
		for _, p := range source.Locations {
			if _, err = tx.Exec(ctx, `INSERT INTO catalog_test_locations(pharmacy_id,matched_address,latitude,longitude) VALUES($1,$2,$3,$4) ON CONFLICT(pharmacy_id) DO NOTHING`, p.ID, p.Address, p.Latitude, p.Longitude); err != nil {
				return err
			}
			for _, m := range meds {
				if _, err = tx.Exec(ctx, `INSERT INTO catalog_test_stock(pharmacy_id,medication_id,quantity) VALUES($1,$2,20) ON CONFLICT(pharmacy_id,medication_id) DO NOTHING`, p.ID, m.ID); err != nil {
					return err
				}
			}
		}
		if _, err = tx.Exec(ctx, `UPDATE catalog_test_settings SET enabled=true WHERE id=true`); err != nil {
			return err
		}
		after, err := fingerprint()
		if err != nil {
			return err
		}
		if before != after {
			return fmt.Errorf("original catalog or inventory unexpectedly changed")
		}
		r.Mode = "applied"
		r.Backup = *archive
	}
	if err = os.MkdirAll(filepath.Dir(*output), 0755); err != nil {
		return err
	}
	reportData, _ := json.MarshalIndent(r, "", "  ")
	if err = os.WriteFile(*output, reportData, 0600); err != nil {
		return err
	}
	if *apply {
		if err = tx.Commit(ctx); err != nil {
			return err
		}
		if err = os.Rename(prepared, *envPath); err != nil {
			return fmt.Errorf("test fixtures saved but local flag could not be installed; set QAYEN_ENABLE_APPROXIMATE_ALERTS=true manually")
		}
	}
	fmt.Printf("%s: %d estimated test locations; %d isolated stock rows; report %s\n", r.Mode, r.Locations, r.StockRows, *output)
	return nil
}
