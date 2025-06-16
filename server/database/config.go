package database

import (
	"context"
	"fmt"
	repository "kayena/server/database/generated"
	"log"
	"os"

	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Queries *repository.Queries
var DbConn *pgxpool.Pool
var err error

func ConnectDb() error {
	DB_URL := os.Getenv("DB_URL")
	DbConn, err = pgxpool.New(context.Background(), DB_URL)
	if err != nil {
		log.Fatal("failed to connect to db :", err)
		return err
	}
	Queries = repository.New(DbConn)
	err = DbConn.Ping(context.Background())
	if err != nil {
		log.Fatal("Failed to ping DB:", err)
	}
	err := SeedDatabase()
	if err != nil {
		return err
	}
	return nil
}

func SeedDatabase() error {
	ctx := context.Background()
	_, err := Queries.GetMedicationByCode(ctx, "1")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Println("hna")
			err = SeedMedications()
			if err != nil {
				return fmt.Errorf("error seeding medications : %s", err)
			}
		}
		return fmt.Errorf("error getting medication : %s", err)
	}
	return nil
}
