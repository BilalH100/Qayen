package schemas

type MedicationJson struct {
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
	Pfht             string `json:"pfht"`
}
type Pharmacy struct {
	Name      string  `json:"nom"`
	City      string  `json:"ville"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
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

type CreateMedicationRequest struct {
	Status           string   `json:"status"`
	CommercialStatus string   `json:"commercial_status"`
	Speciality       string   `json:"speciality"`
	Dosage           string   `json:"dosage"`
	Form             string   `json:"form"`
	Presentation     string   `json:"presentation"`
	Pp               string   `json:"pp"`
	ActiveSubstance  string   `json:"active_substance"`
	TherapeuticClass string   `json:"therapeutic_class"`
	Epi              string   `json:"epi"`
	Ppv              string   `json:"ppv"`
	Ph               string   `json:"ph"`
	Pfht             string   `json:"pfht"`
	Code             string   `json:"code"`
	Tva              string   `json:"tva"`
	Description      string   `json:"description"`
	CommonSd         []string `json:"common_sd"`
	SeriousSd        []string `json:"serious_sd"`
	GeneralInfo      []string `json:"general_info"`
}

type Message struct {
	Message string `json:"message" validate:"required"`
}

type GetUserResponse struct {
	Id    int32  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type Options struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type LoginResponse struct {
	Token string          `json:"token"`
	User  GetUserResponse `json:"user"`
}
