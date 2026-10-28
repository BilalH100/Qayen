package catalogimport

import (
	"regexp"
	"strings"
	"unicode"
)

type SourceFile struct {
	Filename string `json:"filename"`
	SHA256   string `json:"sha256"`
	Sheet    string `json:"sheet"`
}

type CNOPSReference struct {
	SourceRow         int     `json:"source_row"`
	Code              string  `json:"code"`
	Name              string  `json:"name"`
	Ingredient        string  `json:"ingredient"`
	Dosage            string  `json:"dosage"`
	DosageUnit        string  `json:"dosage_unit"`
	Form              string  `json:"form"`
	Presentation      string  `json:"presentation"`
	PublicPrice       *string `json:"public_price"`
	HospitalPrice     *string `json:"hospital_price"`
	ReimbursementBase *string `json:"reimbursement_base"`
	ProductType       string  `json:"product_type"`
	ReimbursementRate string  `json:"reimbursement_rate"`
}

type PharmacyReference struct {
	SourceRow int    `json:"source_row"`
	Name      string `json:"name"`
	City      string `json:"city"`
	Province  string `json:"province"`
}

type SupervisorBundle struct {
	MedicationSource SourceFile          `json:"medication_source"`
	PharmacySource   SourceFile          `json:"pharmacy_source"`
	Medications      []CNOPSReference    `json:"medications"`
	Pharmacies       []PharmacyReference `json:"pharmacies"`
}

func CompactIdentity(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return -1
	}, Normalize(value))
}

var decimalComma = regexp.MustCompile(`(\d),(\d)`)

func CanonicalDose(value string) string {
	value = decimalComma.ReplaceAllString(Normalize(value), "$1.$2")
	return strings.ReplaceAll(value, " ", "")
}

func CanonicalPack(value string) string {
	value = decimalComma.ReplaceAllString(Normalize(value), "$1.$2")
	words := strings.Fields(strings.NewReplacer("'", " ", "-", " ").Replace(value))
	var kept []string
	for _, word := range words {
		if word == "DE" || word == "D" || word == "DES" {
			continue
		}
		switch word {
		case "BOITES":
			word = "BOITE"
		case "COMPRIMES":
			word = "COMPRIME"
		case "GELULES":
			word = "GELULE"
		case "FLACONS":
			word = "FLACON"
		case "AMPOULES":
			word = "AMPOULE"
		case "SACHETS":
			word = "SACHET"
		}
		kept = append(kept, word)
	}
	if len(kept) > 1 && kept[0] == "1" && kept[1] == "BOITE" {
		kept = kept[1:]
	}
	return strings.Join(kept, " ")
}

// Match only all five explicit identity fields. No inferred concentrations,
// dose conversion, partial ingredient matches, or guessed packaging sizes.
func ReferenceKey(name, ingredient, dose, form, presentation string) string {
	if strings.TrimSpace(dose) == "" || strings.TrimSpace(presentation) == "" {
		return ""
	}
	return strings.Join([]string{Normalize(name), Normalize(ingredient), CanonicalDose(dose), Normalize(form), CanonicalPack(presentation)}, "\x1f")
}

func (r CNOPSReference) MatchKey() string {
	if ingredient := Normalize(r.Ingredient); ingredient == "MULTIPLE" || ingredient == "NA" || ingredient == "N/A" {
		return ""
	}
	if r.Dosage == "" || r.DosageUnit == "" || Normalize(r.DosageUnit) == "NA" {
		return ""
	}
	// Multi-component dosages require review because a single column may encode
	// total pack content rather than concentration or per-unit strength.
	if strings.Contains(r.Dosage, "/") || strings.Contains(r.DosageUnit, "/") {
		return ""
	}
	return ReferenceKey(r.Name, r.Ingredient, r.Dosage+" "+r.DosageUnit, r.Form, r.Presentation)
}

func DirectoryKey(name, city string) string {
	city = strings.TrimSpace(strings.TrimSuffix(Normalize(city), "MAROC"))
	return CompactIdentity(name) + "\x1f" + CompactIdentity(city)
}
