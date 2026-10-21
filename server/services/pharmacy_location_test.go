package services

import (
	"errors"
	"kayena/server/schemas"
	"testing"
)

func TestPharmacyLocationValidation(t *testing.T) {
	valid := schemas.Pharmacy{Name: "Pharmacy", Address: "12 Example Street", City: "Rabat", Latitude: "34.0209", Longitude: "-6.8416"}
	if err := validatePharmacyUpdate(&valid); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*schemas.Pharmacy){
		func(p *schemas.Pharmacy) { p.Address = " " },
		func(p *schemas.Pharmacy) { p.Latitude = "NaN" },
		func(p *schemas.Pharmacy) { p.Longitude = "" },
		func(p *schemas.Pharmacy) { p.Latitude = "91" },
		func(p *schemas.Pharmacy) { p.Longitude = "181" },
	} {
		invalid := valid
		change(&invalid)
		if err := validatePharmacyUpdate(&invalid); !errors.Is(err, ErrInvalidPharmacyLocation) {
			t.Fatalf("invalid location accepted: %+v, error=%v", invalid, err)
		}
	}
	legacy := valid
	legacy.Latitude, legacy.Longitude = "", ""
	if err := validatePharmacyUpdate(&legacy); err != nil {
		t.Fatal("older clients must be able to keep existing coordinates", err)
	}
}
