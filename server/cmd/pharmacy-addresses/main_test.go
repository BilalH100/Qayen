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

func sample() evidence {
	return evidence{Name: "Pharmacie Centrale", City: "Fès", Address: "52, rue Hassan II", URL: "https://pharmacy.ma/pharmacie/centrale-fes", PageURL: "https://pharmacy.ma/pharmacies", PageSHA: strings.Repeat("a", 64), Retrieved: time.Now().UTC().Format(time.RFC3339)}
}

func TestSpecificAddressesOnly(t *testing.T) {
	for _, a := range []string{"Hay Al Massira", "Casablanca", "CP 10000, Rabat", "Rue 10 Mars, Casablanca", "Rue Hassan II, CP 10000, Rabat", "N/A", "52 rue...?", "52 rue…"} {
		if specificAddress(a) {
			t.Errorf("vague address accepted: %s", a)
		}
	}
	for _, a := range []string{"52, rue Hassan II", "Bd. de Rabat, en face de l'ancienne Rahba", "Avenue Hassan II, immeuble 24, Casablanca"} {
		if !specificAddress(a) {
			t.Errorf("specific address rejected: %s", a)
		}
	}
}

func TestMatchingHoldsAmbiguousAndExisting(t *testing.T) {
	r := sample()
	p := pharmacy{ID: 1, Name: "PHARMACIE CENTRALE", City: "Fes"}
	if plan([]pharmacy{p}, []evidence{r})[0].Status != "fill_blank_address" {
		t.Fatal("exact name and city not matched")
	}
	p.Address = "Keep original"
	if plan([]pharmacy{p}, []evidence{r})[0].Status != "preserve_existing_address" {
		t.Fatal("existing address overwritten")
	}
	p.Address = ""
	other := r
	other.URL = "https://pharmacy.ma/pharmacie/centrale-fes-2"
	if plan([]pharmacy{p}, []evidence{r, other})[0].Status != "ambiguous_source_identity" {
		t.Fatal("multiple businesses were merged")
	}
	other = r
	other.Address = "77, rue Nouvelle"
	if plan([]pharmacy{p}, []evidence{r, other})[0].Status != "ambiguous_source_identity" {
		t.Fatal("conflicting addresses accepted")
	}
	duplicate := p
	duplicate.ID = 2
	if plan([]pharmacy{p, duplicate}, []evidence{r})[0].Status != "ambiguous_database_identity" {
		t.Fatal("ambiguous database match accepted")
	}
	r.City = "Rabat"
	if plan([]pharmacy{p}, []evidence{r})[0].Status != "no_source_match" {
		t.Fatal("wrong city matched")
	}
	r = sample()
	r.Address = "Hay Al Massira"
	if plan([]pharmacy{p}, []evidence{r})[0].Status != "incomplete_source_address" {
		t.Fatal("vague neighborhood accepted")
	}
}

func TestEvidenceValidation(t *testing.T) {
	s := source{Provider: "Pharmacy.ma", Records: []evidence{sample()}}
	if err := validate(s); err != nil {
		t.Fatal(err)
	}
	s.Records[0].URL = "https://pharmacy.ma.evil.test/pharmacie/x"
	if validate(s) == nil {
		t.Fatal("unrelated source accepted")
	}
}

func TestAddressImportPreservesEverythingElse(t *testing.T) {
	path := os.Getenv("CATALOG_TEST_DB_ENV")
	if path == "" {
		t.Skip("opt-in PostgreSQL TEMP-table integration")
	}
	env, err := godotenv.Read(path)
	if err != nil {
		t.Fatal("configuration unavailable")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, env["DB_URL"])
	if err != nil {
		t.Fatal("database unavailable")
	}
	defer conn.Close(ctx)
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `CREATE TEMP TABLE pharmacies(id SERIAL PRIMARY KEY,name TEXT,city TEXT,address TEXT,latitude DOUBLE PRECISION,longitude DOUBLE PRECISION,phone TEXT);
 CREATE TEMP TABLE medications(id SERIAL PRIMARY KEY,name TEXT,price TEXT);
 CREATE TEMP TABLE stock(id SERIAL PRIMARY KEY,pharmacy_id INT,medication_id INT,quantity INT);
 INSERT INTO pharmacies VALUES(1,'PHARMACIE CENTRALE','Fes','',NULL,NULL,''),(2,'Existing','Rabat','Keep verified address',34.02,-6.84,'0537000000');
 INSERT INTO medications VALUES(1,'Preserve medicine','10');INSERT INTO stock VALUES(1,2,1,7);`)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../database/migrations/20261022000001_add_pharmacy_address_evidence.sql")
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
	decisions := plan(before, []evidence{sample()})
	if err = applyPlan(ctx, tx, decisions); err != nil {
		t.Fatal(err)
	}
	after, newCounts, newHash, err := snapshot(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	if err = verify(before, after, decisions); err != nil {
		t.Fatal(err)
	}
	if counts.WithAddress+1 != newCounts.WithAddress || newHash != hash {
		t.Fatal("unexpected catalog/inventory change")
	}
	second := plan(after, []evidence{sample()})
	for _, d := range second {
		if d.Status == "fill_blank_address" {
			t.Fatal("repeat import would rewrite addresses")
		}
	}
	var safe bool
	if err = tx.QueryRow(ctx, `SELECT (SELECT count(*)=1 FROM pharmacy_address_evidence WHERE business_verified=false AND source_url LIKE 'https://pharmacy.ma/%') AND (SELECT latitude IS NULL AND longitude IS NULL AND phone='' FROM pharmacies WHERE id=1)`).Scan(&safe); err != nil || !safe {
		t.Fatalf("provenance/coordinates check failed: %v", err)
	}
	after[0].Other = "tampered GPS"
	if verify(before, after, decisions) == nil {
		t.Fatal("unplanned field changes undetected")
	}
}
