package database

import (
	"context"
	"errors"
	"fmt"
	"kayena/server/config"
	"kayena/server/services"

	log "github.com/charmbracelet/log"

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

func SeedDb(s *services.Services, logger *log.Logger) error {
	ctx := context.Background()
	_, err := s.MedService.GetMedicationByCode(ctx, "1")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Info("Seeding meds ...")
			err = SeedMedications(ctx, s.MedService)
			if err != nil {
				return fmt.Errorf("error seeding medications : %w", err)
			}
		} else {
			return fmt.Errorf("seeding check failed: %w", err)
		}
	}
	logger.Info("medications already seeded skipping ...")
	_, err = s.PharmacyService.GetByIdService(ctx, 1)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Println("Seeding Pharmacies ...")
			err = SeedRabatPharmacies(ctx, s)
			if err != nil {
				return fmt.Errorf("error seeding pharmacies : %w", err)
			}
		} else {
			return fmt.Errorf("seed check error : error getting pharmacy: %s", err)
		}
	}
	return nil
}
