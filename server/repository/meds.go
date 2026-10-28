package repository

import (
	"context"
	"fmt"
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
	Search(ctx context.Context, query string) ([]models.Medication, error)
	SearchBySpeciality(ctx context.Context, query string) ([]models.Medication, error)
}

type sqlcMedRepo struct {
	queries *sqlc.Queries
	db      sqlc.DBTX
}

func NewMedRepo(db sqlc.DBTX) MedRepository {
	return &sqlcMedRepo{queries: sqlc.New(db), db: db}
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
	result := &models.Medication{
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
	}
	result.References, err = s.getMedicationReferences(ctx, id)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *sqlcMedRepo) getMedicationReferences(ctx context.Context, id int32) ([]models.MedicationReference, error) {
	var exists bool
	if err := s.db.QueryRow(ctx, `SELECT to_regclass('medication_reference_entries') IS NOT NULL`).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	rows, err := s.db.Query(ctx, `SELECT r.product_code,r.public_price::text,r.reimbursement_base::text,COALESCE(r.reimbursement_rate,''),s.filename,r.currency FROM medication_reference_entries r JOIN catalog_import_sources s ON s.id=r.source_id WHERE r.medication_id=$1 AND r.match_status='matched' ORDER BY r.product_code LIMIT 50`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var references []models.MedicationReference
	for rows.Next() {
		var item models.MedicationReference
		if err = rows.Scan(&item.Code, &item.PublicPrice, &item.ReimbursementBase, &item.ReimbursementRate, &item.SourceFile, &item.Currency); err != nil {
			return nil, err
		}
		references = append(references, item)
	}
	return references, rows.Err()
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
	// The `category` table is never populated by the seeder; the real
	// classification data lives on medications.therapeutic_class instead.
	classes, err := s.queries.GetDistinctTherapeuticClasses(ctx)
	if err != nil {
		return nil, err
	}
	categories := make([]models.Category, len(classes))
	for i, name := range classes {
		categories[i] = models.Category{
			ID:   int32(i + 1),
			Name: name,
		}
	}
	return categories, nil
}

func (s *sqlcMedRepo) Search(ctx context.Context, query string) ([]models.Medication, error) {
	searchQuery := "%" + query + "%"
	// Use the prioritized search that focuses on speciality first
	results, err := s.queries.SearchMedicationsWithPriority(ctx, searchQuery)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var medications []models.Medication
	for _, med := range results {
		medications = append(medications, models.Medication{
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
		})
	}
	return medications, nil
}

func (s *sqlcMedRepo) SearchBySpeciality(ctx context.Context, query string) ([]models.Medication, error) {
	searchQuery := "%" + query + "%"
	// Search only in speciality field
	results, err := s.queries.SearchMedicationsBySpeciality(ctx, searchQuery)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var medications []models.Medication
	for _, med := range results {
		medications = append(medications, models.Medication{
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
		})
	}
	return medications, nil
}
