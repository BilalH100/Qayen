package services

import (
	"context"
	"errors"
	"fmt"
	"kayena/server/internal/catalogimport"
	"kayena/server/models"
	"kayena/server/repository"
	"kayena/server/schemas"
	"strconv"
	"strings"
)

var ErrInvalidPharmacyLocation = errors.New("invalid pharmacy address or coordinates")

func validatePharmacyUpdate(p *schemas.Pharmacy) error {
	if strings.TrimSpace(p.Name) == "" || strings.TrimSpace(p.Address) == "" || strings.TrimSpace(p.City) == "" {
		return fmt.Errorf("%w: name, exact address and city are required", ErrInvalidPharmacyLocation)
	}
	if p.Latitude == "" && p.Longitude == "" {
		return nil
	}
	if !catalogimport.IsCoordinateNumber(p.Latitude) || !catalogimport.IsCoordinateNumber(p.Longitude) {
		return fmt.Errorf("%w: provide both latitude and longitude as numbers", ErrInvalidPharmacyLocation)
	}
	lat, latErr := strconv.ParseFloat(p.Latitude, 64)
	lon, lonErr := strconv.ParseFloat(p.Longitude, 64)
	if latErr != nil || lonErr != nil || !catalogimport.ValidCoordinates(lat, lon) {
		return fmt.Errorf("%w: latitude must be between -90 and 90 and longitude between -180 and 180", ErrInvalidPharmacyLocation)
	}
	return nil
}

type PharmacyService interface {
	GetByIdService(ctx context.Context, id int32) (*schemas.Pharmacy, error)
	CreateService(ctx context.Context, p *schemas.Pharmacy) error
	UpdateService(ctx context.Context, id int32, p *schemas.Pharmacy) (*schemas.Pharmacy, error)
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

func (s *pharmacyService) UpdateService(ctx context.Context, id int32, p *schemas.Pharmacy) (*schemas.Pharmacy, error) {
	if err := validatePharmacyUpdate(p); err != nil {
		return nil, err
	}
	return s.pharmaryRepo.Update(ctx, id, p)
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
