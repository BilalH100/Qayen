package catalogimport

import (
	"math"
	"testing"
)

func TestMedicationIdentityKeepsStrengthAndPackagingSeparate(t *testing.T) {
	base := Medication{Name: "Example", Ingredient: "Ingredient", Dosage: "500 MG", Form: "TABLET", Presentation: "BOX OF 10", Manufacturer: "Maker"}
	copy := base
	copy.Name = "  example "
	copy.Code = "new-source-code"
	if MedicationKey(base) != MedicationKey(copy) {
		t.Fatal("name formatting and source code must not create a different product")
	}
	copy.Dosage = "1000 MG"
	if MedicationKey(base) == MedicationKey(copy) {
		t.Fatal("different strengths must remain separate")
	}
	copy = base
	copy.Presentation = "BOX OF 30"
	if MedicationKey(base) == MedicationKey(copy) {
		t.Fatal("different packaging must remain separate")
	}
}

func TestConflictingPricesAreQuarantinedRatherThanChosen(t *testing.T) {
	med := Medication{Name: "Example", Ingredient: "Ingredient", Form: "TABLET", Presentation: "BOX", PublicPrice: "10"}
	copy := med
	copy.PublicPrice = "20"
	items, summary := DeduplicateMedications([]Medication{med, copy})
	if len(items) != 0 || summary.Ambiguous != 1 {
		t.Fatalf("expected conflicting price quarantine: %+v", summary)
	}
}

func TestSourceDuplicatesAndMissingFieldsAreReported(t *testing.T) {
	med := Medication{Name: "Example", Ingredient: "Ingredient", Form: "TABLET", Presentation: "BOX", CommercialStatus: "Retiré du marché"}
	items, summary := DeduplicateMedications([]Medication{med, med, {Name: "Missing ingredient"}})
	if len(items) != 1 || summary.DuplicateRows != 1 || summary.InvalidRows != 1 || summary.MissingDosage != 1 || summary.MissingPrice != 1 || summary.WithdrawnOrSuspended != 1 {
		t.Fatalf("unexpected report: %+v", summary)
	}
	if items[0].Dosage != "" || items[0].PublicPrice != "" {
		t.Fatal("missing data must stay missing")
	}
}

func TestPharmacyMatchingUsesCityAndHandlesExistingRabatLabel(t *testing.T) {
	a := Pharmacy{Name: "PHARMACIE CENTRALE", City: "RABAT"}
	b := Pharmacy{Name: "Pharmacie Centrale", City: "Rabat - Maroc"}
	if PharmacyKey(a) != PharmacyKey(b) {
		t.Fatal("Rabat labels should match")
	}
	b.City = "CASABLANCA"
	if PharmacyKey(a) == PharmacyKey(b) {
		t.Fatal("same name in another city must be separate")
	}
	items, summary := DeduplicatePharmacies([]Pharmacy{a, a, {Name: "Unknown"}})
	if len(items) != 1 || summary.DuplicateRows != 1 || summary.InvalidRows != 1 {
		t.Fatalf("unexpected report: %+v", summary)
	}
}

func TestCoordinatesRejectImpossibleAndNonFiniteLocations(t *testing.T) {
	if !ValidCoordinates(34.0209, -6.8416) {
		t.Fatal("valid Rabat pin rejected")
	}
	for _, pair := range [][2]float64{{91, 0}, {0, -181}, {math.NaN(), 0}, {0, math.Inf(1)}} {
		if ValidCoordinates(pair[0], pair[1]) {
			t.Fatalf("invalid pin accepted: %v", pair)
		}
	}
	for _, text := range []string{"", "NaN", "Infinity", "34abc"} {
		if IsCoordinateNumber(text) {
			t.Fatalf("invalid coordinate accepted: %q", text)
		}
	}
}
