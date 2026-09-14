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
	GetClosestService(ctx context.Context, c schemas.Coordinates, medId int32) (*float64, *schemas.Pharmacy, error)
	GetAllService(ctx context.Context) ([]models.Pharmacy, error)
	ListStockService(ctx context.Context, pharmacyId int32) ([]models.StockItem, error)
	UpsertStockService(ctx context.Context, pharmacyId, medicationId, quantity int32) (*models.StockItem, error)
	DeleteStockService(ctx context.Context, pharmacyId, medicationId int32) error
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

func (s *pharmacyService) GetClosestService(ctx context.Context, c schemas.Coordinates, medId int32) (*float64, *schemas.Pharmacy, error) {
	return s.pharmaryRepo.GetClosest(ctx, c, medId)
}

func (s *pharmacyService) ListStockService(ctx context.Context, pharmacyId int32) ([]models.StockItem, error) {
	return s.pharmaryRepo.ListStock(ctx, pharmacyId)
}

func (s *pharmacyService) UpsertStockService(ctx context.Context, pharmacyId, medicationId, quantity int32) (*models.StockItem, error) {
	return s.pharmaryRepo.UpsertStock(ctx, pharmacyId, medicationId, quantity)
}

func (s *pharmacyService) DeleteStockService(ctx context.Context, pharmacyId, medicationId int32) error {
	return s.pharmaryRepo.DeleteStock(ctx, pharmacyId, medicationId)
}
