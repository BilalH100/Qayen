package main

import "testing"

func TestEquivalentAddressSpellingNeedsPhoneAndSpecificAddress(t *testing.T) {
	p, r := target(), fixture()
	p.Address = "3 Bis, bd Mohamed V, complexe Almoggar"
	r.Address = "3 Bis, bd Mohammed V, cplx. Almoggar Fes"
	if !equivalentSpellingAddress(p, r, r.Phone) {
		t.Fatal("equivalent spelling rejected")
	}
	if equivalentSpellingAddress(p, r, "") {
		t.Fatal("new spelling rule bypassed telephone corroboration")
	}
	r.Address = "3 Ter, bd Mohammed V, cplx. Almoggar Fes"
	if equivalentSpellingAddress(p, r, r.Phone) {
		t.Fatal("different street-number suffix accepted")
	}
	p.Address = "Hay Laazib Amsernat, bloc 304, n° 22"
	r.Address = "cité Amsernat, bloc 304, n°22, hay Laazib Fes"
	if !equivalentSpellingAddress(p, r, r.Phone) {
		t.Fatal("equivalent typed building identifiers rejected")
	}
	r.Address = "cité Amsernat, bloc 22, n°304, hay Laazib Fes"
	if equivalentSpellingAddress(p, r, r.Phone) {
		t.Fatal("swapped building and unit identifiers accepted")
	}
	p.Address = "54, avenue Hassan II"
	r.Address = "Pharmacie Centrale Fes"
	if equivalentSpellingAddress(p, r, r.Phone) {
		t.Fatal("name/city alone treated as a full address")
	}
}
