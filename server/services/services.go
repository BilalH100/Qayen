package services

import (
	sqlc "kayena/server/database/generated"
	"kayena/server/repository"
)

type Services struct {
	MedService      MedService
	UserService     UserService
	PharmacyService PharmacyService
}

func NewService(db sqlc.DBTX) *Services {
	return &Services{
		MedService:      NewMedService(repository.NewMedRepo(db)),
		UserService:     NewUserService(repository.NewUserRepo(db)),
		PharmacyService: NewPharmacyService(repository.NewPharmacyRepo(db)),
	}
}
