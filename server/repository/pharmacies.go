package repository

import (
	"context"
	"fmt"
	sqlc "kayena/server/database/generated"
	"kayena/server/schemas"
	"kayena/server/utils"
	"strconv"
)

type PharmacyRepository interface {
	GetById(ctx context.Context, id int32) (*schemas.Pharmacy, error)
	Create(ctx context.Context, p *schemas.Pharmacy) error
	GetClosest(ctx context.Context, c schemas.Coordinates, medId int32) (*int32, *schemas.Pharmacy, error)
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
		return nil, nil, wrap(err, "")
	}
	lon, err := strconv.ParseFloat(c.Lat, 64)
	if err != nil {
		return nil, nil, wrap(err, "")
	}
	res, err := s.queries.GetClosestPharmacyWithMedication(ctx, sqlc.GetClosestPharmacyWithMedicationParams{
		Radians:      lat,
		Radians_2:    lon,
		MedicationID: utils.Int32ToPgInt4(medId),
	})
	return &res.DistanceKm, &schemas.Pharmacy{
		Name:      res.Name,
		City:      res.City,
		Latitude:  utils.Float8ToString(res.Latitude),
		Longitude: utils.Float8ToString(res.Longitude),
		Address:   res.Address,
		Phone:     res.Phone,
	}, nil
}

func wrap(err error, m string) error {
	if m == "" {
		return fmt.Errorf("Repo : %w", err)
	}
	return fmt.Errorf("Repo : %s %w", m, err)
}
