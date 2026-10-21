// catalog-import previews AMMPS JSON imports by default. -apply commits a single
// transaction while preserving existing IDs, stock, accounts and alert history.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"kayena/server/internal/catalogimport"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type medicationRow struct {
	ID         int32
	Medication catalogimport.Medication
}

type change[T any] struct {
	ID    int32
	Value T
}

type report struct {
	GeneratedAt               time.Time             `json:"generated_at"`
	Mode                      string                `json:"mode"`
	Medications               catalogimport.Summary `json:"medications"`
	Pharmacies                catalogimport.Summary `json:"pharmacies"`
	PharmaciesNeedingLocation int                   `json:"pharmacies_needing_address_and_coordinates"`
	Notes                     []string              `json:"notes"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	medicationFile := flag.String("medications", "", "AMMPS medication JSON export")
	pharmacyFile := flag.String("pharmacies", "", "AMMPS pharmacy name/city JSON export")
	envFile := flag.String("env", "server/.env", "Environment file containing DB_URL")
	reportFile := flag.String("report", "tmp/catalog-import-report.json", "Import report destination")
	apply := flag.Bool("apply", false, "Apply the reviewed import in one transaction (default: preview only)")
	flag.Parse()
	if *medicationFile == "" && *pharmacyFile == "" {
		return fmt.Errorf("provide -medications and/or -pharmacies; no data was changed")
	}

	result := report{GeneratedAt: time.Now().UTC(), Mode: "preview", Notes: []string{
		"Source files must be verified AMMPS exports; this command does not certify their freshness.",
		"Pharmacy names/cities are directory entries, not verified addresses or live availability.",
		"Existing pharmacy contact details, coordinates, medication IDs and stock are preserved.",
		"No inventory, clinical descriptions or missing prices are generated.",
	}}
	var medications []catalogimport.Medication
	var pharmacies []catalogimport.Pharmacy
	if *medicationFile != "" {
		items, err := catalogimport.ReadJSON[catalogimport.Medication](*medicationFile)
		if err != nil {
			return err
		}
		medications, result.Medications = catalogimport.DeduplicateMedications(items)
	}
	if *pharmacyFile != "" {
		items, err := catalogimport.ReadJSON[catalogimport.Pharmacy](*pharmacyFile)
		if err != nil {
			return err
		}
		pharmacies, result.Pharmacies = catalogimport.DeduplicatePharmacies(items)
	}

	_ = godotenv.Load(*envFile)
	if os.Getenv("DB_URL") == "" {
		return fmt.Errorf("DB_URL is missing; configure it through -env or your environment")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool, err := pgxpool.New(ctx, os.Getenv("DB_URL"))
	if err != nil {
		return fmt.Errorf("could not configure database connection")
	}
	defer pool.Close()
	accessMode := pgx.ReadOnly
	if *apply {
		accessMode = pgx.ReadWrite
		result.Mode = "applied"
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable, AccessMode: accessMode})
	if err != nil {
		return fmt.Errorf("could not connect to the configured database")
	}
	defer tx.Rollback(context.Background())
	if *apply {
		// Serialize imports and prevent concurrent writes from invalidating the plan.
		if _, err := tx.Exec(ctx, "LOCK TABLE medications, pharmacies IN SHARE ROW EXCLUSIVE MODE"); err != nil {
			return fmt.Errorf("lock import tables: %w", err)
		}
	}

	medicationChanges, err := planMedications(ctx, tx, medications, &result.Medications)
	if err != nil {
		return err
	}
	pharmacyChanges, err := planPharmacies(ctx, tx, pharmacies, &result.Pharmacies)
	if err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM pharmacies WHERE COALESCE(TRIM(address),'')='' OR latitude IS NULL OR longitude IS NULL`).Scan(&result.PharmaciesNeedingLocation); err != nil {
		return err
	}
	result.PharmaciesNeedingLocation += result.Pharmacies.ToInsert
	if err := os.MkdirAll(filepath.Dir(*reportFile), 0755); err != nil {
		return err
	}
	if *apply {
		if err := applyChanges(ctx, tx, medicationChanges, pharmacyChanges); err != nil {
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit import: %w", err)
		}
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(*reportFile, append(data, '\n'), 0644); err != nil {
		return err
	}
	fmt.Printf("%s: medications +%d / updated %d; pharmacies +%d; report: %s\n", result.Mode, result.Medications.ToInsert, result.Medications.ToUpdate, result.Pharmacies.ToInsert, *reportFile)
	return nil
}

func applyChanges(ctx context.Context, tx pgx.Tx, medications []change[catalogimport.Medication], pharmacies []change[catalogimport.Pharmacy]) error {
	for _, item := range medications {
		m := item.Value
		values := []any{m.Status, m.CommercialStatus, m.Name, m.Dosage, m.Form, m.Presentation, m.Type, m.Ingredient, m.TherapeuticClass, m.Manufacturer, m.PublicPrice, m.HospitalPrice, m.ManufacturerPrice, m.Code, m.Tax}
		query := `INSERT INTO medications (status,commercial_status,speciality,dosage,form,presentation,pp,active_substance,therapeutic_class,epi,ppv,ph,pfht,code,tva) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`
		if item.ID != 0 {
			query = `UPDATE medications SET status=$1,commercial_status=$2,speciality=$3,dosage=$4,form=$5,presentation=$6,pp=$7,active_substance=$8,therapeutic_class=$9,epi=$10,ppv=$11,ph=$12,pfht=$13,code=$14,tva=$15 WHERE id=$16`
			values = append(values, item.ID)
		}
		if _, err := tx.Exec(ctx, query, values...); err != nil {
			return fmt.Errorf("import medication %q: %w", m.Name, err)
		}
	}
	for _, item := range pharmacies {
		if _, err := tx.Exec(ctx, `INSERT INTO pharmacies (name,city) VALUES ($1,$2)`, item.Value.Name, item.Value.City); err != nil {
			return fmt.Errorf("import pharmacy %q: %w", item.Value.Name, err)
		}
	}
	return nil
}

func planMedications(ctx context.Context, tx pgx.Tx, incoming []catalogimport.Medication, summary *catalogimport.Summary) ([]change[catalogimport.Medication], error) {
	rows, err := tx.Query(ctx, `SELECT id, COALESCE(status,''),COALESCE(commercial_status,''),COALESCE(speciality,''),COALESCE(dosage,''),COALESCE(form,''),COALESCE(presentation,''),COALESCE(pp,''),COALESCE(active_substance,''),COALESCE(therapeutic_class,''),COALESCE(epi,''),COALESCE(ppv,''),COALESCE(ph,''),COALESCE(pfht,''),COALESCE(code,''),COALESCE(tva,'') FROM medications`)
	if err != nil {
		return nil, err
	}
	existing := make(map[string][]medicationRow)
	existingCodes := make(map[string]bool)
	for rows.Next() {
		var row medicationRow
		m := &row.Medication
		if err := rows.Scan(&row.ID, &m.Status, &m.CommercialStatus, &m.Name, &m.Dosage, &m.Form, &m.Presentation, &m.Type, &m.Ingredient, &m.TherapeuticClass, &m.Manufacturer, &m.PublicPrice, &m.HospitalPrice, &m.ManufacturerPrice, &m.Code, &m.Tax); err != nil {
			rows.Close()
			return nil, err
		}
		key := catalogimport.MedicationKey(*m)
		existing[key] = append(existing[key], row)
		existingCodes[m.Code] = true
		summary.ExistingRows++
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	var changes []change[catalogimport.Medication]
	for _, m := range incoming {
		matches := existing[catalogimport.MedicationKey(m)]
		if len(matches) > 1 {
			summary.Ambiguous++
			summary.Issue("Existing duplicate product needs review: " + m.Name + " / " + m.Presentation)
			continue
		}
		if len(matches) == 0 {
			if m.Code != "" && existingCodes[m.Code] {
				summary.Ambiguous++
				summary.Issue("Source code identifies a different existing product: " + m.Code + " / " + m.Name)
				continue
			}
			summary.ToInsert++
			changes = append(changes, change[catalogimport.Medication]{Value: m})
			continue
		}
		current := matches[0]
		m.Code = current.Medication.Code
		// Formatting/capitalization differences are not updates.
		if equivalent(current.Medication, m) {
			summary.Unchanged++
			continue
		}
		summary.ToUpdate++
		changes = append(changes, change[catalogimport.Medication]{ID: current.ID, Value: m})
	}
	return changes, nil
}

func equivalent(a, b catalogimport.Medication) bool {
	left, _ := json.Marshal(a)
	right, _ := json.Marshal(b)
	return catalogimport.Normalize(string(left)) == catalogimport.Normalize(string(right))
}

func planPharmacies(ctx context.Context, tx pgx.Tx, incoming []catalogimport.Pharmacy, summary *catalogimport.Summary) ([]change[catalogimport.Pharmacy], error) {
	rows, err := tx.Query(ctx, `SELECT name,COALESCE(city,'') FROM pharmacies`)
	if err != nil {
		return nil, err
	}
	existing := make(map[string]int)
	for rows.Next() {
		var pharmacy catalogimport.Pharmacy
		if err := rows.Scan(&pharmacy.Name, &pharmacy.City); err != nil {
			rows.Close()
			return nil, err
		}
		existing[catalogimport.PharmacyKey(pharmacy)]++
		summary.ExistingRows++
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	var changes []change[catalogimport.Pharmacy]
	for _, p := range incoming {
		switch existing[catalogimport.PharmacyKey(p)] {
		case 0:
			summary.ToInsert++
			changes = append(changes, change[catalogimport.Pharmacy]{Value: p})
		case 1:
			summary.Unchanged++
		default:
			summary.Ambiguous++
			summary.Issue("Existing pharmacy name/city collision: " + p.Name + " / " + p.City)
		}
	}
	return changes, nil
}
