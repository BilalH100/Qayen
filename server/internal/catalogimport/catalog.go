// Package catalogimport validates AMMPS catalog exports without inventing
// addresses, prices, coordinates, clinical information, or pharmacy inventory.
package catalogimport

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type Medication struct {
	Status            string `json:"statut_amm"`
	CommercialStatus  string `json:"statut_commercialisation"`
	Name              string `json:"specialite"`
	Dosage            string `json:"dosage"`
	Form              string `json:"forme"`
	Presentation      string `json:"presentation"`
	Type              string `json:"pp_gn"`
	Ingredient        string `json:"substance_active"`
	TherapeuticClass  string `json:"classe_therapeutique"`
	Manufacturer      string `json:"epi"`
	PublicPrice       string `json:"ppv"`
	HospitalPrice     string `json:"ph"`
	ManufacturerPrice string `json:"pfht"`
	Code              string `json:"code"`
	Tax               string `json:"tva"`
}

// Directory entries intentionally have no fabricated address or map pin.
// These must be supplied separately by the pharmacy before location-based use.
type Pharmacy struct {
	Name string `json:"nom"`
	City string `json:"ville"`
}

type Summary struct {
	SourceRows           int      `json:"source_rows"`
	UniqueRows           int      `json:"unique_rows"`
	DuplicateRows        int      `json:"duplicate_rows"`
	InvalidRows          int      `json:"invalid_rows"`
	MissingDosage        int      `json:"missing_dosage,omitempty"`
	MissingPrice         int      `json:"missing_price,omitempty"`
	WithdrawnOrSuspended int      `json:"withdrawn_or_suspended,omitempty"`
	ExistingRows         int      `json:"existing_rows"`
	ToInsert             int      `json:"to_insert"`
	ToUpdate             int      `json:"to_update"`
	Unchanged            int      `json:"unchanged"`
	Ambiguous            int      `json:"ambiguous_matches"`
	Issues               []string `json:"issues,omitempty"`
}

func ReadJSON[T any](file string) ([]T, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("parse %s: %w", file, err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("%s contains no records", file)
	}
	return items, nil
}

func Normalize(value string) string {
	var out strings.Builder
	for _, char := range norm.NFD.String(strings.ToUpper(value)) {
		if unicode.Is(unicode.Mn, char) {
			continue
		}
		out.WriteRune(char)
	}
	return strings.Join(strings.Fields(out.String()), " ")
}

// Use the full product identity, not a possibly reused numeric source code.
func MedicationKey(m Medication) string {
	return strings.Join([]string{Normalize(m.Name), Normalize(m.Dosage), Normalize(m.Form), Normalize(m.Presentation), Normalize(m.Manufacturer)}, "\x1f")
}

func PharmacyKey(p Pharmacy) string {
	city := strings.TrimSpace(strings.TrimSuffix(Normalize(p.City), "MAROC"))
	city = strings.TrimRight(city, " -,")
	return Normalize(p.Name) + "\x1f" + city
}

func (s *Summary) Issue(value string) {
	// Keep reports manageable while retaining accurate totals.
	if len(s.Issues) < 50 {
		s.Issues = append(s.Issues, value)
	}
}

func DeduplicateMedications(items []Medication) ([]Medication, Summary) {
	summary := Summary{SourceRows: len(items)}
	seen := make(map[string]Medication)
	conflicts := make(map[string]bool)
	var keys []string
	for i, med := range items {
		if strings.TrimSpace(med.Name) == "" || strings.TrimSpace(med.Ingredient) == "" || strings.TrimSpace(med.Form) == "" || strings.TrimSpace(med.Presentation) == "" {
			summary.InvalidRows++
			summary.Issue(fmt.Sprintf("Medication row %d is missing a name, ingredient, form or presentation", i+1))
			continue
		}
		key := MedicationKey(med)
		if existing, ok := seen[key]; ok {
			summary.DuplicateRows++
			// A conflicting price or status needs review, not an arbitrary winner.
			existing.Code, med.Code = "", ""
			if existing != med {
				conflicts[key] = true
				summary.Issue("Conflicting medication rows: " + med.Name + " / " + med.Presentation)
			}
			continue
		}
		seen[key] = med
		keys = append(keys, key)
	}
	var result []Medication
	for _, key := range keys {
		if conflicts[key] {
			summary.Ambiguous++
			continue
		}
		med := seen[key]
		result = append(result, med)
		if strings.TrimSpace(med.Dosage) == "" {
			summary.MissingDosage++
		}
		if strings.TrimSpace(med.PublicPrice) == "" || strings.Contains(Normalize(med.PublicPrice), "SANS") {
			summary.MissingPrice++
		}
		status := Normalize(med.CommercialStatus)
		if strings.Contains(status, "RETIRE") || strings.Contains(status, "SUSPENDU") {
			summary.WithdrawnOrSuspended++
		}
	}
	summary.UniqueRows = len(result)
	return result, summary
}

func DeduplicatePharmacies(items []Pharmacy) ([]Pharmacy, Summary) {
	summary := Summary{SourceRows: len(items)}
	seen := make(map[string]bool)
	var result []Pharmacy
	for i, pharmacy := range items {
		if strings.TrimSpace(pharmacy.Name) == "" || strings.TrimSpace(pharmacy.City) == "" {
			summary.InvalidRows++
			summary.Issue(fmt.Sprintf("Pharmacy row %d is missing a name or city", i+1))
			continue
		}
		key := PharmacyKey(pharmacy)
		if seen[key] {
			summary.DuplicateRows++
			continue
		}
		seen[key] = true
		result = append(result, pharmacy)
	}
	summary.UniqueRows = len(result)
	return result, summary
}

var coordinateNumber = regexp.MustCompile(`^[+-]?(?:\d+(?:\.\d*)?|\.\d+)$`)

// ValidCoordinates rejects NaN/Infinity, missing pairs and impossible map pins.
func ValidCoordinates(lat, lon float64) bool {
	return !math.IsNaN(lat) && !math.IsNaN(lon) && !math.IsInf(lat, 0) && !math.IsInf(lon, 0) && lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}

func IsCoordinateNumber(value string) bool {
	return coordinateNumber.MatchString(strings.TrimSpace(value))
}
