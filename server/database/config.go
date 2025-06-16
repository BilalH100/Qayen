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
			fmt.Println("Seeding meds ...")
			err = SeedMedications()
			if err != nil {
				return fmt.Errorf("error seeding medications : %s", err)
			}
		}
		return fmt.Errorf("seed check error : error getting medication : %s", err)
	}
	fmt.Println("medications already seeded skipping ...")
	_, err = Queries.GetPharmacyByID(ctx, 1)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows){
			fmt.Println("Seeding Pharmacies ...")
			err = SeedPharmacies()
			if err != nil {
				return fmt.Errorf("error seeding pharmacies : %s", err)
			}
		}
		return fmt.Errorf("seed check error : error getting pharmacy: %s", err)
	}
	return nil
}
