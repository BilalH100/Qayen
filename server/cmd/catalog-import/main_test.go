package main

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"kayena/server/internal/catalogimport"
	"os"
	"testing"
	"time"
)

// Opt-in integration test: every write targets TEMP tables and is rolled back.
// It never updates the configured database's permanent catalog or stock.
func TestImportPreservesReferencesAndIsIdempotent(t *testing.T) {
	envFile := os.Getenv("CATALOG_TEST_DB_ENV")
	if envFile == "" {
		t.Skip("set CATALOG_TEST_DB_ENV to run PostgreSQL TEMP-table integration test")
	}
	config, err := godotenv.Read(envFile)
	if err != nil {
		t.Fatal("could not read test database configuration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, config["DB_URL"])
	if err != nil {
		t.Fatal("could not connect to test database")
	}
	defer conn.Close(context.Background())
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	_, err = tx.Exec(ctx, `
CREATE TEMP TABLE medications (id SERIAL PRIMARY KEY,status TEXT,commercial_status TEXT,speciality TEXT,dosage TEXT,form TEXT,presentation TEXT,pp TEXT,active_substance TEXT,therapeutic_class TEXT,epi TEXT,ppv TEXT,ph TEXT,pfht TEXT,code TEXT,tva TEXT,description TEXT) ON COMMIT DROP;
CREATE TEMP TABLE pharmacies (id SERIAL PRIMARY KEY,name TEXT,city TEXT,address TEXT,latitude DOUBLE PRECISION,longitude DOUBLE PRECISION,phone TEXT) ON COMMIT DROP;
CREATE TEMP TABLE stock (pharmacy_id INT REFERENCES pharmacies(id),medication_id INT REFERENCES medications(id),quantity INT) ON COMMIT DROP;
INSERT INTO medications (id,speciality,dosage,form,presentation,epi,active_substance,ppv,code,description) VALUES (5001,'Example','500 MG','TABLET','BOX OF 10','Maker','Ingredient','10','original-code','Preserve existing description');
INSERT INTO pharmacies (id,name,city,address,latitude,longitude) VALUES (5001,'Pharmacie Centrale','Rabat - Maroc','Verified existing address',34.02,-6.84);
INSERT INTO stock VALUES (5001,5001,7);`)
	if err != nil {
		t.Fatal(err)
	}
	med := catalogimport.Medication{Name: "EXAMPLE", Dosage: "500 MG", Form: "TABLET", Presentation: "BOX OF 10", Manufacturer: "Maker", Ingredient: "Ingredient", PublicPrice: "12", Code: "changed-source-code"}
	other := med
	other.Dosage = "1000 MG"
	other.Code = "new-code"
	collision := med
	collision.Name = "Different product"
	collision.Code = "original-code"
	var summary catalogimport.Summary
	changes, err := planMedications(ctx, tx, []catalogimport.Medication{med, other, collision}, &summary)
	if err != nil {
		t.Fatal(err)
	}
	if summary.ToInsert != 1 || summary.ToUpdate != 1 || summary.Ambiguous != 1 {
		t.Fatalf("unexpected medication plan: %+v", summary)
	}
	pharmacies := []catalogimport.Pharmacy{{Name: "PHARMACIE CENTRALE", City: "RABAT"}, {Name: "New Pharmacy", City: "FES"}}
	var pharmacySummary catalogimport.Summary
	pharmacyChanges, err := planPharmacies(ctx, tx, pharmacies, &pharmacySummary)
	if err != nil {
		t.Fatal(err)
	}
	if pharmacySummary.ToInsert != 1 || pharmacySummary.Unchanged != 1 {
		t.Fatalf("unexpected pharmacy plan: %+v", pharmacySummary)
	}
	if err := applyChanges(ctx, tx, changes, pharmacyChanges); err != nil {
		t.Fatal(err)
	}
	var code, description, price string
	if err := tx.QueryRow(ctx, `SELECT code,description,ppv FROM medications WHERE id=5001`).Scan(&code, &description, &price); err != nil {
		t.Fatal(err)
	}
	if code != "original-code" || description != "Preserve existing description" || price != "12" {
		t.Fatal("update changed references/clinical text or failed to update source price")
	}
	var quantity int
	if err := tx.QueryRow(ctx, `SELECT quantity FROM stock WHERE pharmacy_id=5001 AND medication_id=5001`).Scan(&quantity); err != nil || quantity != 7 {
		t.Fatal("existing stock reference was changed", err)
	}
	var missingLocation bool
	if err := tx.QueryRow(ctx, `SELECT address IS NULL AND latitude IS NULL AND longitude IS NULL FROM pharmacies WHERE name='New Pharmacy'`).Scan(&missingLocation); err != nil || !missingLocation {
		t.Fatal("directory import fabricated a location", err)
	}
	var repeated catalogimport.Summary
	changes, err = planMedications(ctx, tx, []catalogimport.Medication{med, other}, &repeated)
	if err != nil || len(changes) != 0 || repeated.Unchanged != 2 {
		t.Fatalf("repeat import is not idempotent: %+v; %v", repeated, err)
	}
	var repeatedPharmacy catalogimport.Summary
	pharmacyChanges, err = planPharmacies(ctx, tx, pharmacies, &repeatedPharmacy)
	if err != nil || len(pharmacyChanges) != 0 || repeatedPharmacy.Unchanged != 2 {
		t.Fatalf("repeat pharmacy import is not idempotent: %+v; %v", repeatedPharmacy, err)
	}
}
