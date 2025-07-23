package models

import "github.com/jackc/pgx/v5/pgtype"

type Medication struct {
	ID               int32            `json:"id"`
	Status           string           `json:"status"`
	CommercialStatus string           `json:"commercial_status"`
	Speciality       string           `json:"speciality"`
	Dosage           string           `json:"dosage"`
	Form             string           `json:"form"`
	Presentation     string           `json:"presentation"`
	Pp               string           `json:"pp"`
	ActiveSubstance  string           `json:"active_substance"`
	TherapeuticClass string           `json:"therapeutic_class"`
	Epi              string           `json:"epi"`
	Ppv              string           `json:"ppv"`
	Ph               string           `json:"ph"`
	Pfht             string           `json:"pfht"`
	Code             string           `json:"code"`
	Tva              string           `json:"tva"`
	CreatedAt        pgtype.Timestamp `json:"created_at"`
	Description      string           `json:"description"`
	CommonSd         []string         `json:"common_sd"`
	SeriousSd        []string         `json:"serious_sd"`
	GeneralInfo      []string         `json:"general_info"`
}

type Pharmacy struct {
	ID        int32            `json:"id"`
	Name      string           `json:"name"`
	Address   string           `json:"address"`
	Latitude  string           `json:"latitude"`
	Longitude string           `json:"longitude"`
	City      string           `json:"city"`
	Phone     string           `json:"phone"`
	CreatedAt pgtype.Timestamp `json:"created_at"`
}

type Stock struct {
	ID           int32            `json:"id"`
	PharmacyID   pgtype.Int4      `json:"pharmacy_id"`
	MedicationID pgtype.Int4      `json:"medication_id"`
	Quantity     int32            `json:"quantity"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at"`
}

type User struct {
	ID        int32            `json:"id"`
	Password  string           `json:"assword"`
	Email     string           `json:"email"`
	Name      string           `json:"name"`
	Phone     string           `json:"phone"`
	CreatedAt pgtype.Timestamp `json:"created_at"`
}
