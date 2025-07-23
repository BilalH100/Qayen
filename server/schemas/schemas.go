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
	Pfht             string `json:"pfht"`
}

type Coordinates struct {
	Lat  string `json:"lat"`
	Long string `json:"lon"`
}

type Failed struct {
	Desc int
	Side int
}

type SideEffects struct {
	Common  []string
	Serious []string
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name"  validate:"required"`
	Phone    string `json:"phone" validate:"required,min=10"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required, email"`
	Password string `json:"password" validate:"required,min=6"`
}

type Message struct {
	Message string `json:"message" validate:"required"`
}

type GetUserResponse struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}
