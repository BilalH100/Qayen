package main

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	ci "kayena/server/internal/catalogimport"
	"os"
	"strings"
	"testing"
)

func fixture() ci.SupervisorBundle {
	price := "12.50"
	source := ci.SourceFile{Filename: "source.xlsx", Sheet: "Sheet", SHA256: strings.Repeat("a", 64)}
	return ci.SupervisorBundle{MedicationSource: source, PharmacySource: source, Medications: []ci.CNOPSReference{{SourceRow: 2, Code: "6118000000001", Name: "Example", Ingredient: "Ingredient", Dosage: "500", DosageUnit: "MG", Form: "TABLET", Presentation: "BOX OF 10", PublicPrice: &price}}, Pharmacies: []ci.PharmacyReference{{SourceRow: 5, Name: "Existing", City: "Rabat"}, {SourceRow: 6, Name: "New", City: "Fes"}, {SourceRow: 7, Name: "New", City: "Fes"}, {SourceRow: 8, Name: "Unknown"}}}
}

func TestValidationAndConservativeMatching(t *testing.T) {
	b := fixture()
	if err := validateBundle(b); err != nil {
		t.Fatal(err)
	}
	negative := "-1"
	b.Medications[0].PublicPrice = &negative
	if validateBundle(b) == nil {
		t.Fatal("negative price accepted")
	}
	b = fixture()
	b.Medications = append(b.Medications, b.Medications[0])
	if validateBundle(b) == nil {
		t.Fatal("duplicate code accepted")
	}
	b = fixture()
	row := b.Medications[0]
	key := row.MatchKey()
	r := report{}
	p := planReferences(b.Medications, map[string][]int32{key: {5001}}, map[int32]string{5001: "10"}, nil, &r)
	if p[0].Status != "matched" || r.PriceDifferences != 1 {
		t.Fatal("exact match or price conflict not recorded")
	}
	row.Dosage = "250"
	if row.MatchKey() == key {
		t.Fatal("different strength matched")
	}
	row = b.Medications[0]
	row.Ingredient = "MULTIPLE"
	if row.MatchKey() != "" {
		t.Fatal("unknown ingredients matched")
	}
	row = b.Medications[0]
	row.DosageUnit = "MG/ML"
	if row.MatchKey() != "" {
		t.Fatal("concentration inferred")
	}
	if ci.CanonicalPack("FLACON 1.5 ML") == ci.CanonicalPack("FLACON 15 ML") {
		t.Fatal("pack size precision lost")
	}
	r = report{}
	directory := planDirectory(b.Pharmacies, map[string][]int32{ci.DirectoryKey("Existing", "Rabat"): {5001}}, &r)
	if r.NewPharmacies != 1 || r.DuplicatePharmacies != 1 || r.MissingCity != 1 || directory[3].ID != nil {
		t.Fatal("unsafe directory decisions")
	}
	r = report{}
	p = planReferences(b.Medications, nil, nil, map[string]string{b.Medications[0].Code: "different"}, &r)
	if !p[0].Conflict {
		t.Fatal("changed source record was not held")
	}
}

// Uses temporary tables and rolls back every write; permanent data is untouched.
func TestImportIntegration(t *testing.T) {
	path := os.Getenv("CATALOG_TEST_DB_ENV")
	if path == "" {
		t.Skip("database integration is opt-in")
	}
	env, err := godotenv.Read(path)
	if err != nil {
		t.Fatal("cannot read database configuration")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, env["DB_URL"])
	if err != nil {
		t.Fatal("cannot connect to test database")
	}
	defer conn.Close(ctx)
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `CREATE TEMP TABLE medications(id SERIAL PRIMARY KEY,speciality TEXT,active_substance TEXT,dosage TEXT,form TEXT,presentation TEXT,ppv TEXT);
 CREATE TEMP TABLE pharmacies(id SERIAL PRIMARY KEY,name TEXT,city TEXT,address TEXT,latitude DOUBLE PRECISION,longitude DOUBLE PRECISION,phone TEXT);
 CREATE TEMP TABLE stock(pharmacy_id INT,medication_id INT,quantity INT);
 INSERT INTO medications VALUES(5001,'Example','Ingredient','500 MG','TABLET','BOX OF 10','10');
 INSERT INTO pharmacies(id,name,city,address) VALUES(5001,'Existing','Rabat','Verified address');
 INSERT INTO stock VALUES(5001,5001,7);`)
	if err != nil {
		t.Fatal(err)
	}
	schema, err := os.ReadFile("../../database/migrations/20261022000000_add_catalog_references.sql")
	if err != nil {
		t.Fatal(err)
	}
	up := strings.Split(string(schema), "-- +goose Down")[0]
	up = strings.ReplaceAll(up, "CREATE TABLE IF NOT EXISTS", "CREATE TEMP TABLE")
	_, err = tx.Exec(ctx, up)
	if err != nil {
		t.Fatal(err)
	}
	b := fixture()
	for iteration := 0; iteration < 2; iteration++ {
		identities, prices, err := loadMedicationIdentities(ctx, tx)
		if err != nil {
			t.Fatal(err)
		}
		pharmacies, _, err := loadPharmacies(ctx, tx)
		if err != nil {
			t.Fatal(err)
		}
		fingerprints, err := loadReferenceFingerprints(ctx, tx)
		if err != nil {
			t.Fatal(err)
		}
		r := report{}
		meds := planReferences(b.Medications, identities, prices, fingerprints, &r)
		directory := planDirectory(b.Pharmacies, pharmacies, &r)
		if err = applyPlan(ctx, tx, b, meds, directory); err != nil {
			t.Fatal(err)
		}
		var count int
		if err = tx.QueryRow(ctx, `SELECT count(*) FROM pharmacies`).Scan(&count); err != nil || count != 2 {
			t.Fatalf("pharmacy count %d: %v", count, err)
		}
	}
	var safe bool
	err = tx.QueryRow(ctx, `SELECT (SELECT ppv='10' FROM medications WHERE id=5001) AND (SELECT quantity=7 FROM stock) AND (SELECT address='Verified address' FROM pharmacies WHERE id=5001) AND (SELECT address='' AND latitude IS NULL AND longitude IS NULL AND phone='' FROM pharmacies WHERE name='New') AND (SELECT count(*)=1 FROM medication_reference_entries WHERE medication_id=5001 AND public_price=12.50) AND (SELECT count(*)=1 FROM pharmacy_directory_entries WHERE match_status='missing_city' AND pharmacy_id IS NULL)`).Scan(&safe)
	if err != nil || !safe {
		t.Fatalf("preservation checks failed: %v", err)
	}
	b.Medications[0].PublicPrice = nil
	identities, prices, _ := loadMedicationIdentities(ctx, tx)
	fingerprints, _ := loadReferenceFingerprints(ctx, tx)
	r := report{}
	meds := planReferences(b.Medications, identities, prices, fingerprints, &r)
	if err = applyPlan(ctx, tx, b, meds, nil); err != nil {
		t.Fatal(err)
	}
	err = tx.QueryRow(ctx, `SELECT public_price=12.50 FROM medication_reference_entries`).Scan(&safe)
	if err != nil || !safe {
		t.Fatal("conflicting re-import overwrote price")
	}
}
