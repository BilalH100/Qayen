package main

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"os"
	"strings"
	"testing"
	"time"
)

func num(value float64) *float64 { return &value }
func fixture() location {
	return location{Name: "Pharmacie Centrale", City: "Fes", Address: "52, rue Hassan II", Phone: "+212535123456", Latitude: num(34.040257), Longitude: num(-5.000391), URL: "https://pharmap.ma/pharmacie/pharmacie-centrale-1", PageURL: "https://pharmap.ma/pharmacie/pharmacie-centrale-1", PageSHA: strings.Repeat("a", 64), Retrieved: time.Now().UTC().Format(time.RFC3339)}
}
func target() pharmacy {
	return pharmacy{ID: 1, Name: "PHARMACIE CENTRALE", City: "Fès", Address: "52, rue Hassan II"}
}

func TestMatchRequiresStreetAndCity(t *testing.T) {
	p, r := target(), fixture()
	if method := compatibleAddress(p, r, ""); method != "unique_name_city_same_address" {
		t.Fatal(method)
	}
	r.Address = "53, rue Hassan II"
	if compatibleAddress(p, r, r.Phone) != "" {
		t.Fatal("wrong building accepted")
	}
	r.Address = "52, rue Mohammed V"
	if compatibleAddress(p, r, r.Phone) != "" {
		t.Fatal("wrong street accepted even with matching telephone")
	}
	r = fixture()
	r.City = "Meknes"
	if plan([]pharmacy{p}, []location{r}, nil)[0].Status != "no_published_point" {
		t.Fatal("wrong city accepted")
	}
	r = fixture()
	r.Address = "52, rue Hassan III"
	if compatibleAddress(p, r, "") != "" {
		t.Fatal("partial street-name prefix accepted")
	}
	r = fixture()
	r.Address = "52, rue Hassan II, quartier Atlas, Fes"
	if compatibleAddress(p, r, "") == "" {
		t.Fatal("same numbered street with additional district rejected")
	}
}

func TestExplicitBuildingAddressMatchesAcrossArticleSpelling(t *testing.T) {
	p, r := target(), fixture()
	p.Address = "Hay Al Mohammadi, imm. A1B16 Assafa"
	r.Address = "hay Mohammadi, immeuble A1B16 Assafa Fes"
	if compatibleAddress(p, r, "") != "unique_name_city_same_address" {
		t.Fatal("identical explicit building address was not recognized")
	}
	r.Address = "hay Mohammadi, immeuble A1B17 Assafa Fes"
	if compatibleAddress(p, r, r.Phone) != "" {
		t.Fatal("different building was accepted")
	}
}

func TestUnsafeCoordinatesAndAmbiguityAreHeld(t *testing.T) {
	p, r := target(), fixture()
	for _, point := range [][2]float64{{0, 0}, {34.04, -5.00}, {-5, 34}, {50, 2}, {35, -190}} {
		r.Latitude = num(point[0])
		r.Longitude = num(point[1])
		if precisePoint(r) {
			t.Fatalf("unsafe point accepted: %v", point)
		}
	}
	r = fixture()
	r.Latitude = nil
	if precisePoint(r) {
		t.Fatal("incomplete coordinate accepted")
	}
	r = fixture()
	other := r
	other.URL = "https://pharmap.ma/pharmacie/pharmacie-centrale-2"
	other.Longitude = num(-5.002391)
	if plan([]pharmacy{p}, []location{r, other}, nil)[0].Status != "ambiguous_source_identity" {
		t.Fatal("ambiguous branch accepted")
	}
	other = fixture()
	other.Name = "Pharmacie Different"
	other.URL = "https://pharmap.ma/pharmacie/pharmacie-different-2"
	if plan([]pharmacy{p}, []location{r, other}, nil)[0].Status != "shared_point_requires_review" {
		t.Fatal("shared centre accepted")
	}
	p.Latitude = num(34.02)
	p.Longitude = num(-6.84)
	if plan([]pharmacy{p}, []location{r}, nil)[0].Status != "preserve_existing_coordinates" {
		t.Fatal("existing coordinates replaced")
	}
	p = target()
	p.Address = ""
	if plan([]pharmacy{p}, []location{r}, nil)[0].Status != "missing_address" {
		t.Fatal("unaddressed pharmacy located")
	}
}

func TestTelephoneCorroboratesTheSavedAddressOnly(t *testing.T) {
	p := target()
	p.Address = "52, bd. Hassan El Fassi, résidence Atlas"
	r := fixture()
	r.Address = "52 boulevard Hassan El Fassi, Atlas Fes"
	refs := []addressReference{{Name: p.Name, City: p.City, Address: p.Address, Phone: "0535123456"}}
	if plan([]pharmacy{p}, []location{r}, refs)[0].Status != "fill_missing_coordinates" {
		t.Fatal("corroborated point rejected")
	}
	refs[0].Address = "An old different address"
	if plan([]pharmacy{p}, []location{r}, refs)[0].Status != "address_not_corroborated" {
		t.Fatal("telephone from a different address used")
	}
}

func TestPublicSourceValidation(t *testing.T) {
	src := source{Provider: "pharmap.ma", Records: []location{fixture()}, Complete: true}
	if err := validate(src); err != nil {
		t.Fatal(err)
	}
	src.Records[0].URL = "https://pharmap.ma.evil.test/pharmacie/x"
	if validate(src) == nil {
		t.Fatal("unrelated evidence accepted")
	}
}

// Opt-in: every write is to a TEMP table and the transaction rolls back.
func TestGPSImportPreservesExistingDataAndIsIdempotent(t *testing.T) {
	path := os.Getenv("CATALOG_TEST_DB_ENV")
	if path == "" {
		t.Skip("set CATALOG_TEST_DB_ENV for PostgreSQL TEMP-table integration")
	}
	env, err := godotenv.Read(path)
	if err != nil {
		t.Fatal("configuration unavailable")
	}
	ctx := context.Background()
	db, err := pgx.Connect(ctx, env["DB_URL"])
	if err != nil {
		t.Fatal("database unavailable")
	}
	defer db.Close(ctx)
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `CREATE TEMP TABLE pharmacies(id SERIAL PRIMARY KEY,name TEXT,city TEXT,address TEXT,phone TEXT,latitude DOUBLE PRECISION,longitude DOUBLE PRECISION);
 CREATE TEMP TABLE medications(id SERIAL PRIMARY KEY,name TEXT,price TEXT);
 CREATE TEMP TABLE stock(id SERIAL PRIMARY KEY,pharmacy_id INT,medication_id INT,quantity INT);
 INSERT INTO pharmacies VALUES(1,'PHARMACIE CENTRALE','Fes','52, rue Hassan II','',NULL,NULL),(2,'Existing','Rabat','Original address','0537000000',34.02,-6.84),(3,'No address','Fes','','',NULL,NULL);
 INSERT INTO medications VALUES(1,'Preserve medicine','10');INSERT INTO stock VALUES(1,2,1,7);`)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../database/migrations/20261022000002_add_pharmacy_location_evidence.sql")
	if err != nil {
		t.Fatal(err)
	}
	up := strings.ReplaceAll(strings.Split(string(data), "-- +goose Down")[0], "CREATE TABLE IF NOT EXISTS", "CREATE TEMP TABLE")
	if _, err = tx.Exec(ctx, up); err != nil {
		t.Fatal(err)
	}
	before, counts, hash, err := snapshot(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	decisions := plan(before, []location{fixture()}, nil)
	if err = applyPlan(ctx, tx, decisions); err != nil {
		t.Fatal(err)
	}
	after, afterCounts, afterHash, err := snapshot(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	if err = verify(before, after, decisions); err != nil {
		t.Fatal(err)
	}
	if hash != afterHash || afterCounts.Located != counts.Located+1 || afterCounts.Addressed != counts.Addressed {
		t.Fatal("unplanned catalog changes")
	}
	next := plan(after, []location{fixture()}, nil)
	for _, d := range next {
		if d.Status == "fill_missing_coordinates" {
			t.Fatal("repeat import changes existing GPS")
		}
	}
	var safe bool
	if err = tx.QueryRow(ctx, `SELECT (SELECT count(*)=1 FROM pharmacy_location_evidence WHERE business_verified=false AND matched_address='52, rue Hassan II') AND (SELECT latitude=34.02 AND longitude=-6.84 AND address='Original address' AND phone='0537000000' FROM pharmacies WHERE id=2)`).Scan(&safe); err != nil || !safe {
		t.Fatalf("GPS audit preservation failed: %v", err)
	}
	after[0].Other = "unexpected address change"
	if verify(before, after, decisions) == nil {
		t.Fatal("unplanned address changes undetected")
	}
}
