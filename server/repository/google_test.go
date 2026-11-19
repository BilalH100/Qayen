package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"kayena/server/models"
	"os"
	"testing"
)

func TestGoogleAccountsInTemporaryTable(t *testing.T) {
	path := os.Getenv("GOOGLE_AUTH_TEST_DB_ENV")
	if path == "" {
		t.Skip("set GOOGLE_AUTH_TEST_DB_ENV for an isolated database test")
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
	_, err = conn.Exec(ctx, `CREATE TEMP TABLE users(id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,email text UNIQUE NOT NULL,name text NOT NULL DEFAULT '',phone text NOT NULL DEFAULT '',password text NOT NULL,created_at timestamp DEFAULT now(),user_role text NOT NULL DEFAULT 'regular',managed_pharmacy_id integer,google_subject text UNIQUE)`)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewUserRepo(conn)
	user, err := repo.GoogleLogin(ctx, "subject-1", "test@gmail.com", "Test User")
	if err != nil {
		t.Fatal(err)
	}
	if user.Role != models.Regular || user.Email != "test@gmail.com" || user.ManagedPharmacyID != 0 {
		t.Fatal("incorrect default privileges")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("")) == nil {
		t.Fatal("empty password accepted")
	}
	// Google subject remains stable even if the Google email/name changes.
	same, err := repo.GoogleLogin(ctx, "subject-1", "changed@gmail.com", "Changed")
	if err != nil || same.ID != user.ID {
		t.Fatal("returning identity was duplicated")
	}
	_, err = conn.Exec(ctx, `INSERT INTO users(email,name,password,user_role) VALUES('owner@gmail.com','Owner','existing-hash','admin')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = repo.GoogleLogin(ctx, "other-subject", "OWNER@gmail.com", "Attacker")
	if !errors.Is(err, ErrGoogleEmailExists) {
		t.Fatal("existing password account linked without authentication")
	}
	var count int
	if err = conn.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&count); err != nil || count != 2 {
		t.Fatal("unexpected account mutations")
	}
}
