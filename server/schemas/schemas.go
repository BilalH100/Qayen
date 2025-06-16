package schemas

type Pharmacy struct {
	Name      string  `json:"nom"`
	City      string  `json:"ville"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
}

type Medication struct {
	ID               int32  `json:"id"`
	Status           string `json:"statut_amm"`
	CommercialStatus string `json:"statut_commercialisation"`
	Speciality       string `json:"specialite"`
	Dosage           string `json:"dosage"`
	Form             string `json:"forme"`
	Presentation     string `json:"presentation"`
	Pp               string `json:"pp_gn"`
	ActiveSubstance  string `json:"substance_active"`
	TherapeuticClass string `json:"classe_therapeutique"`
	Epi              string `json:"epi"`
	Ppv              string `json:"ppv"`
	Ph               string `json:"ph"`
	Code             string `json:"code"`
	Tva              string `json:"tva"`
	CreatedAt        string `json:"created_at"`
	Pfht 			 string `json:"pfht"`
}
