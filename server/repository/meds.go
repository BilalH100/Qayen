package repository

import (
	"context"
	sqlc "kayena/server/database/generated"
	"kayena/server/models"
	"kayena/server/schemas"
)

type MedRepository interface {
	Create(ctx context.Context, med sqlc.CreateMedicationParams) error
	Delete(ctx context.Context, id int32) error
	GetByID(ctx context.Context, id int32) (*models.Medication, error)
	GetByCode(ctx context.Context, code string) (*models.Medication, error)
	Update(ctx context.Context, med sqlc.UpdateMedicationParams) (*models.Medication, error)
	GetAll(ctx context.Context, options schemas.Options) ([]models.Medication, error)
	GetCategories(ctx context.Context) ([]models.Category, error)
}

type sqlcMedRepo struct {
	queries *sqlc.Queries
}

func NewMedRepo(db sqlc.DBTX) MedRepository {
	return &sqlcMedRepo{queries: sqlc.New(db)}
}

func (s *sqlcMedRepo) Create(ctx context.Context, med sqlc.CreateMedicationParams) error {
	return s.queries.CreateMedication(ctx, med)
}

func (s *sqlcMedRepo) Delete(ctx context.Context, id int32) error {
	return s.queries.DeleteMedication(ctx, id)
}

func (s *sqlcMedRepo) GetByID(ctx context.Context, id int32) (*models.Medication, error) {
	med, err := s.queries.GetMedicationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &models.Medication{
		ID:               med.ID,
		Code:             med.Code,
		Status:           med.Status,
		CommercialStatus: med.CommercialStatus,
		Speciality:       med.Speciality,
		Dosage:           med.Dosage,
		Form:             med.Form,
		Presentation:     med.Presentation,
		Pp:               med.Pp,
		ActiveSubstance:  med.ActiveSubstance,
		TherapeuticClass: med.TherapeuticClass,
		Epi:              med.Epi,
		Ppv:              med.Ppv,
		Ph:               med.Ph,
		Pfht:             med.Pfht,
		Tva:              med.Tva,
		CreatedAt:        med.CreatedAt,
		Description:      med.Description,
		CommonSd:         med.CommonSd,
		SeriousSd:        med.SeriousSd,
		GeneralInfo:      med.GeneralInfo,
	}, nil
}

func (s *sqlcMedRepo) GetByCode(ctx context.Context, code string) (*models.Medication, error) {
	med, err := s.queries.GetMedicationByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return &models.Medication{
		ID:               med.ID,
		Code:             med.Code,
		Status:           med.Status,
		CommercialStatus: med.CommercialStatus,
		Speciality:       med.Speciality,
		Dosage:           med.Dosage,
		Form:             med.Form,
		Presentation:     med.Presentation,
		Pp:               med.Pp,
		ActiveSubstance:  med.ActiveSubstance,
		TherapeuticClass: med.TherapeuticClass,
		Epi:              med.Epi,
		Ppv:              med.Ppv,
		Ph:               med.Ph,
		Pfht:             med.Pfht,
		Tva:              med.Tva,
		CreatedAt:        med.CreatedAt,
		Description:      med.Description,
		CommonSd:         med.CommonSd,
		SeriousSd:        med.SeriousSd,
		GeneralInfo:      med.GeneralInfo,
	}, nil
}

func (s *sqlcMedRepo) Update(ctx context.Context, med sqlc.UpdateMedicationParams) (*models.Medication, error) {
	new, err := s.queries.UpdateMedication(ctx, med)
	if err != nil {
		return nil, err
	}
	return &models.Medication{
		ID:               new.ID,
		Code:             new.Code,
		Status:           new.Status,
		CommercialStatus: new.CommercialStatus,
		Speciality:       new.Speciality,
		Dosage:           new.Dosage,
		Form:             new.Form,
		Presentation:     new.Presentation,
		Pp:               new.Pp,
		ActiveSubstance:  new.ActiveSubstance,
		TherapeuticClass: new.TherapeuticClass,
		Epi:              new.Epi,
		Ppv:              new.Ppv,
		Ph:               new.Ph,
		Pfht:             new.Pfht,
		Tva:              new.Tva,
		Description:      new.Description,
		CommonSd:         new.CommonSd,
		SeriousSd:        new.SeriousSd,
		GeneralInfo:      new.GeneralInfo,
	}, nil
}

func (s *sqlcMedRepo) GetAll(ctx context.Context, options schemas.Options) ([]models.Medication, error) {
	var meds []models.Medication
	m, err := s.queries.GetMedications(ctx, sqlc.GetMedicationsParams{
		Offset: options.Offset,
		Limit:  options.Limit,
	})
	if err != nil {
		return nil, err
	}
	for _, med := range m {
		meds = append(meds, models.Medication{
			ID:               med.ID,
			Code:             med.Code,
			Status:           med.Status,
			CommercialStatus: med.CommercialStatus,
			Speciality:       med.Speciality,
			Dosage:           med.Dosage,
			Form:             med.Form,
			Presentation:     med.Presentation,
			Pp:               med.Pp,
			ActiveSubstance:  med.ActiveSubstance,
			TherapeuticClass: med.TherapeuticClass,
			Epi:              med.Epi,
			Ppv:              med.Ppv,
			Ph:               med.Ph,
			Pfht:             med.Pfht,
			Tva:              med.Tva,
			Description:      med.Description,
			CommonSd:         med.CommonSd,
			SeriousSd:        med.SeriousSd,
			GeneralInfo:      med.GeneralInfo,
		})
	}
	return meds, nil
}

func (s *sqlcMedRepo) GetCategories(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	cat, err := s.queries.GetCategories(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range cat {
		categories = append(categories, models.Category{
			ID:   c.ID,
			Name: c.Name,
		})
	}
	return categories, nil
}
