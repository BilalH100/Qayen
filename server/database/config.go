package database

import (
	"context"
	"fmt"
	"kayena/server/config"
	"kayena/server/services"

	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DbConn *pgxpool.Pool
var err error

func ConnectDb(c *config.Config) (*pgxpool.Pool, error) {
	DbConn, err = pgxpool.New(context.Background(), c.DbUrl)
	if err != nil {
		return nil, err
	}
	err = DbConn.Ping(context.Background())
	if err != nil {
		return nil, err
	}
	return DbConn, nil
}

func SeedDb(s *services.Services) error {
	ctx := context.Background()
	_, err := s.MedService.GetMedicationByCode(ctx, "1")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Println("Seeding meds ...")
			err = SeedMedications(ctx, s.MedService)
			if err != nil {
				return fmt.Errorf("error seeding medications : %s", err)
			}
			return nil
		}
		return fmt.Errorf("seed check error : error getting medication : %s", err)
	}
	fmt.Println("medications already seeded skipping ...")
	// _, err = Queries.GetPharmacyByID(ctx, 1)
	// if err != nil {
	// 	if errors.Is(err, pgx.ErrNoRows){
	// 		fmt.Println("Seeding Pharmacies ...")
	// 		err = SeedPharmacies()
	// 		if err != nil {
	// 			return fmt.Errorf("error seeding pharmacies : %s", err)
	// 		}
	// 	}
	// 	return fmt.Errorf("seed check error : error getting pharmacy: %s", err)
	// }
	return nil
}
