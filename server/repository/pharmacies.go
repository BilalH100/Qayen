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
	GetClosest(ctx context.Context, c schemas.Coordinates, medId int32) (*float64, *schemas.Pharmacy, error)
	GetAll(ctx context.Context) ([]models.Pharmacy, error)
	ListStock(ctx context.Context, pharmacyId int32) ([]models.StockItem, error)
	UpsertStock(ctx context.Context, pharmacyId, medicationId, quantity int32) (*models.StockItem, error)
	DeleteStock(ctx context.Context, pharmacyId, medicationId int32) error
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
		Id:        res.ID,
		Name:      res.Name,
		City:      res.City,
		Latitude:  utils.Float8ToString(res.Latitude),
		Longitude: utils.Float8ToString(res.Longitude),
		Address:   res.Address,
		Phone:     res.Phone,
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

func (s *sqlcPharmacyRepo) GetClosest(ctx context.Context, c schemas.Coordinates, medId int32) (*float64, *schemas.Pharmacy, error) {
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
		Id:        res.ID,
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

// ListStock returns everything a pharmacy currently has in stock, joined with medication names.
func (s *sqlcPharmacyRepo) ListStock(ctx context.Context, pharmacyId int32) ([]models.StockItem, error) {
	res, err := s.queries.ListStockByPharmacy(ctx, utils.Int32ToPgInt4(pharmacyId))
	if err != nil {
		return nil, wrap(err, "failed to list stock")
	}

	items := make([]models.StockItem, len(res))
	for i, row := range res {
		items[i] = models.StockItem{
			ID:           row.ID,
			PharmacyID:   pharmacyId,
			MedicationID: row.MedicationID.Int32,
			Quantity:     row.Quantity,
			Speciality:   row.Speciality,
			UpdatedAt:    row.UpdatedAt.Time,
		}
	}
	return items, nil
}

// UpsertStock adds a new stock entry, or updates the quantity if one already exists for this pharmacy+medication.
func (s *sqlcPharmacyRepo) UpsertStock(ctx context.Context, pharmacyId, medicationId, quantity int32) (*models.StockItem, error) {
	existing, err := s.queries.GetStock(ctx, sqlc.GetStockParams{
		PharmacyID:   utils.Int32ToPgInt4(pharmacyId),
		MedicationID: utils.Int32ToPgInt4(medicationId),
	})
	if err == nil {
		updated, err := s.queries.UpdateStockQuantity(ctx, sqlc.UpdateStockQuantityParams{
			PharmacyID:   utils.Int32ToPgInt4(pharmacyId),
			MedicationID: utils.Int32ToPgInt4(medicationId),
			Quantity:     quantity,
		})
		if err != nil {
			return nil, wrap(err, "failed to update stock")
		}
		return &models.StockItem{
			ID:           updated.ID,
			PharmacyID:   pharmacyId,
			MedicationID: medicationId,
			Quantity:     updated.Quantity,
		}, nil
	}
	_ = existing

	created, err := s.queries.CreateStock(ctx, sqlc.CreateStockParams{
		PharmacyID:   utils.Int32ToPgInt4(pharmacyId),
		MedicationID: utils.Int32ToPgInt4(medicationId),
		Quantity:     quantity,
	})
	if err != nil {
		return nil, wrap(err, "failed to create stock")
	}
	return &models.StockItem{
		ID:           created.ID,
		PharmacyID:   pharmacyId,
		MedicationID: medicationId,
		Quantity:     created.Quantity,
	}, nil
}

// DeleteStock removes a medication from a pharmacy's stock entirely.
func (s *sqlcPharmacyRepo) DeleteStock(ctx context.Context, pharmacyId, medicationId int32) error {
	err := s.queries.DeleteStockEntry(ctx, sqlc.DeleteStockEntryParams{
		PharmacyID:   utils.Int32ToPgInt4(pharmacyId),
		MedicationID: utils.Int32ToPgInt4(medicationId),
	})
	if err != nil {
		return wrap(err, "failed to delete stock")
	}
	return nil
}
