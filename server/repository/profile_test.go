package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"os"
	"testing"
)

func TestProfilePersistenceInTemporaryTable(t *testing.T) {
	path := os.Getenv("PROFILE_TEST_DB_ENV")
	if path == "" {
		t.Skip("set PROFILE_TEST_DB_ENV for the isolated persistence test")
	}
	config, err := godotenv.Read(path)
	if err != nil {
		t.Fatal("cannot read test database configuration")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, config["DB_URL"])
	if err != nil {
		t.Fatal("test database unavailable")
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, `CREATE TEMP TABLE users(id integer PRIMARY KEY,email text,name text,phone text,password text,user_role text,managed_pharmacy_id integer,google_subject text);INSERT INTO users VALUES(1,'owner@example.com','Old','', 'original-hash','pharmacist',95,'original-subject'),(2,'other@example.com','Other','', 'other-hash','regular',NULL,NULL)`)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewUserRepo(conn)
	user, err := repo.UpdateProfile(ctx, 1, "New Name", "+212600000000")
	if err != nil {
		t.Fatal(err)
	}
	if user.Name != "New Name" || user.Phone != "+212600000000" || user.Email != "owner@example.com" || user.Role != "pharmacist" || user.ManagedPharmacyID != 95 {
		t.Fatal("incorrect returned profile")
	}
	var name, phone, password, subject, other string
	err = conn.QueryRow(ctx, `SELECT name,phone,password,google_subject FROM users WHERE id=1`).Scan(&name, &phone, &password, &subject)
	if err != nil {
		t.Fatal(err)
	}
	if name != "New Name" || phone != "+212600000000" || password != "original-hash" || subject != "original-subject" {
		t.Fatal("edit not persisted or identity changed")
	}
	if err = conn.QueryRow(ctx, `SELECT name FROM users WHERE id=2`).Scan(&other); err != nil || other != "Other" {
		t.Fatal("another user's profile changed")
	}
}
