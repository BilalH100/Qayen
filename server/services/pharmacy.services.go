package services

import (
	"context"
	"kayena/server/models"
	"kayena/server/repository"
	"kayena/server/schemas"
)

type PharmacyService interface {
	GetByIdService(ctx context.Context, id int32) (*schemas.Pharmacy, error)
	CreateService(ctx context.Context, p *schemas.Pharmacy) error
	GetClosestService(ctx context.Context, c schemas.Coordinates, medId int32) (*int32, *schemas.Pharmacy, error)
	GetAllService(ctx context.Context) ([]models.Pharmacy, error)
}

type pharmacyService struct {
	pharmaryRepo repository.PharmacyRepository
}

func NewPharmacyService(pharmacyRepo repository.PharmacyRepository) PharmacyService {
	return &pharmacyService{
		pharmaryRepo: pharmacyRepo,
	}
}

func (s *pharmacyService) GetByIdService(ctx context.Context, id int32) (*schemas.Pharmacy, error) {
	return s.pharmaryRepo.GetById(ctx, id)
}

func (s *pharmacyService) CreateService(ctx context.Context, p *schemas.Pharmacy) error {
	return s.pharmaryRepo.Create(ctx, p)
}

func (s *pharmacyService) GetAllService(ctx context.Context) ([]models.Pharmacy, error) {
	return s.pharmaryRepo.GetAll(ctx)
}

func (s *pharmacyService) GetClosestService(ctx context.Context, c schemas.Coordinates, medId int32) (*int32, *schemas.Pharmacy, error) {
	return s.pharmaryRepo.GetClosest(ctx, c, medId)
}
