package services

import (
	sqlc "kayena/server/database/generated"
	"kayena/server/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Services struct {
	MedService      MedService
	UserService     UserService
	PharmacyService PharmacyService
	AlertService    *AlertService
}

func NewService(db sqlc.DBTX) *Services {
	var alertService *AlertService
	if pool, ok := db.(*pgxpool.Pool); ok {
		alertService = NewAlertService(pool)
	}

	return &Services{
		MedService:      NewMedService(repository.NewMedRepo(db)),
		UserService:     NewUserService(repository.NewUserRepo(db)),
		PharmacyService: NewPharmacyService(repository.NewPharmacyRepo(db)),
		AlertService:    alertService,
	}
}
