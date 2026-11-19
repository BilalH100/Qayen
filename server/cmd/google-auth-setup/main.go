// Applies only the additive Google subject column/index to an existing Qayen DB.
package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"os"
	"time"
)

func main() {
	_ = godotenv.Load("server/.env")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, os.Getenv("DB_URL"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot connect to the database; check server/.env.")
		os.Exit(1)
	}
	defer conn.Close(ctx)
	tx, err := conn.Begin(ctx)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS google_subject TEXT; CREATE UNIQUE INDEX IF NOT EXISTS users_google_subject_key ON users(google_subject);`)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Google schema setup failed:", err)
		os.Exit(1)
	}
	if err = tx.Commit(ctx); err != nil {
		panic(err)
	}
	fmt.Println("Google sign-in schema ready. Existing accounts and catalog preserved.")
}
