package repository

import (
	"context"
	"fmt"
	sqlc "kayena/server/database/generated"
	"kayena/server/models"
	"kayena/server/schemas"
	"kayena/server/utils"
	"strconv"
)

type PharmacyRepository interface {
	GetById(ctx context.Context, id int32) (*schemas.Pharmacy, error)
	Create(ctx context.Context, p *schemas.Pharmacy) error
	GetClosest(ctx context.Context, c schemas.Coordinates, medId int32) (*int32, *schemas.Pharmacy, error)
	GetAll(ctx context.Context) ([]models.Pharmacy, error)
}

type sqlcPharmacyRepo struct {
	queries *sqlc.Queries
}

func NewPharmacyRepo(db sqlc.DBTX) PharmacyRepository {
	return &sqlcPharmacyRepo{
		queries: sqlc.New(db),
	}
}

func (s *sqlcPharmacyRepo) GetById(ctx context.Context, id int32) (*schemas.Pharmacy, error) {
	res, err := s.queries.GetPharmacyByID(ctx, id)
	if err != nil {
		return nil, wrap(err, "")
	}

	return &schemas.Pharmacy{
		Name:      res.Name,
		City:      res.City,
		Latitude:  utils.Float8ToString(res.Latitude),
		Longitude: utils.Float8ToString(res.Longitude),
		Address:   res.Address,
	}, nil
}

func (s *sqlcPharmacyRepo) Create(ctx context.Context, p *schemas.Pharmacy) error {
	lat, err := utils.StringToFloat8(p.Latitude)
	if err != nil {
		return wrap(err, "")
	}
	lon, err := utils.StringToFloat8(p.Longitude)
	if err != nil {
		return wrap(err, "")
	}
	_, err = s.queries.CreatePharmacy(ctx, sqlc.CreatePharmacyParams{
		Name:      p.Name,
		Address:   p.Address,
		Latitude:  lat,
		Longitude: lon,
		City:      p.City,
		Phone:     p.Phone,
	})
	if err != nil {
		return wrap(err, "")
	}
	return nil
}

func (s *sqlcPharmacyRepo) GetClosest(ctx context.Context, c schemas.Coordinates, medId int32) (*int32, *schemas.Pharmacy, error) {
	lat, err := strconv.ParseFloat(c.Lat, 64)
	if err != nil {
		return nil, nil, wrap(err, "invalid latitude format")
	}
	lon, err := strconv.ParseFloat(c.Long, 64)
	if err != nil {
		return nil, nil, wrap(err, "invalid longitude format")
	}
	res, err := s.queries.GetClosestPharmacyWithMedication(ctx, sqlc.GetClosestPharmacyWithMedicationParams{
		Radians:      lat,
		Radians_2:    lon,
		MedicationID: utils.Int32ToPgInt4(medId),
	})
	if err != nil {
		return nil, nil, wrap(err, "failed to find closest pharmacy")
	}

	distance := res.DistanceKm
	pharmacy := &schemas.Pharmacy{
		Name:      res.Name,
		City:      res.City,
		Latitude:  utils.Float8ToString(res.Latitude),
		Longitude: utils.Float8ToString(res.Longitude),
		Address:   res.Address,
		Phone:     res.Phone,
	}

	return &distance, pharmacy, nil
}

func (s *sqlcPharmacyRepo) GetAll(ctx context.Context) ([]models.Pharmacy, error) {
	res, err := s.queries.ListPharmacies(ctx)
	if err != nil {
		return nil, wrap(err, "")
	}

	pharmacies := make([]models.Pharmacy, len(res))
	for i, pharmacy := range res {
		pharmacies[i] = models.Pharmacy{
			ID:        pharmacy.ID,
			Name:      pharmacy.Name,
			City:      pharmacy.City,
			Latitude:  utils.Float8ToString(pharmacy.Latitude),
			Longitude: utils.Float8ToString(pharmacy.Longitude),
			Address:   pharmacy.Address,
			Phone:     pharmacy.Phone,
		}
	}

	return pharmacies, nil
}

func wrap(err error, m string) error {
	if m == "" {
		return fmt.Errorf("Repo : %w", err)
	}
	return fmt.Errorf("Repo : %s %w", m, err)
}
