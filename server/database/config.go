package database

import (
	"context"
	"errors"
	"fmt"
	"kayena/server/config"
	"kayena/server/models"
	"kayena/server/services"

	log "github.com/charmbracelet/log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
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
	if err := SeedDemoUsers(ctx, s, logger); err != nil {
		return fmt.Errorf("error seeding demo users: %w", err)
	}
	return nil
}

func SeedDemoUsers(ctx context.Context, s *services.Services, logger *log.Logger) error {
	demoUsers := []struct {
		email             string
		password          string
		name              string
		phone             string
		role              models.Role
		managedPharmacyID int32
	}{
		{
			email:    "user@demo.com",
			password: "password123",
			name:     "Demo User (Patient)",
			phone:    "+212600000001",
			role:     models.Regular,
		},
		{
			email:             "pharmacist@demo.com",
			password:          "password123",
			name:              "Demo Pharmacist",
			phone:             "+212600000002",
			role:              models.Pharmacist,
			managedPharmacyID: 1,
		},
		{
			email:    "admin@demo.com",
			password: "password123",
			name:     "Demo Admin",
			phone:    "+212600000003",
			role:     models.Admin,
		},
	}

	for _, d := range demoUsers {
		_, err := s.UserService.GetUserProfileByEmail(ctx, d.email)
		if err == nil {
			logger.Infof("Demo user %s already exists, skipping", d.email)
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(d.password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password for %s: %w", d.email, err)
		}

		user := models.User{
			Email:             d.email,
			Name:              d.name,
			Phone:             d.phone,
			Password:          string(hashed),
			Role:              d.role,
			ManagedPharmacyID: d.managedPharmacyID,
		}

		if err := s.UserService.RegisterUser(ctx, user); err != nil {
			return fmt.Errorf("failed to register demo user %s: %w", d.email, err)
		}

		if d.managedPharmacyID > 0 {
			created, err := s.UserService.GetUserProfileByEmail(ctx, d.email)
			if err == nil {
				if _, err := s.UserService.UpdateUserPharmacy(ctx, created.ID, d.managedPharmacyID); err != nil {
					logger.Warnf("Failed to assign pharmacy to %s: %v", d.email, err)
				}
			}
		}

		logger.Infof("Created demo user: %s (role=%s)", d.email, d.role)
	}

	return nil
}
