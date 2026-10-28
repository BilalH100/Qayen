package main

import (
	"fmt"
	ci "kayena/server/internal/catalogimport"
	"math"
	"regexp"
	"sort"
	"strings"
)

type location struct {
	Name      string   `json:"name"`
	City      string   `json:"city"`
	Address   string   `json:"address"`
	Phone     string   `json:"phone"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	URL       string   `json:"source_url"`
	PageURL   string   `json:"page_url"`
	PageSHA   string   `json:"page_sha256"`
	Retrieved string   `json:"retrieved_at"`
}
type pharmacy struct {
	ID                                int32
	Name, City, Address, Phone, Other string
	Latitude, Longitude               *float64
}
type addressReference struct{ Name, City, Address, Phone string }
type decision struct {
	ID       int32     `json:"pharmacy_id"`
	Name     string    `json:"name"`
	City     string    `json:"city"`
	Address  string    `json:"address"`
	Status   string    `json:"status"`
	Method   string    `json:"match_method,omitempty"`
	Evidence *location `json:"evidence,omitempty"`
}

func key(name, city string) string {
	name = strings.TrimSpace(strings.TrimPrefix(ci.Normalize(name), "PHARMACIE "))
	city = strings.TrimSpace(strings.TrimRight(strings.TrimSuffix(ci.Normalize(city), "MAROC"), " -,"))
	switch ci.CompactIdentity(city) {
	case "FEZ":
		city = "FES"
	case "TANGIER", "TANGIERS":
		city = "TANGER"
	}
	return ci.CompactIdentity(name) + "\x1f" + ci.CompactIdentity(city)
}

var words = regexp.MustCompile(`[^\p{L}\p{N}]+`)
var firstNumber = regexp.MustCompile(`^\s*([0-9]+)(?:\s+|[,./-]|$)`)
var nondigits = regexp.MustCompile(`[^0-9]`)
var joinedArticle = regexp.MustCompile(`\b(?:El|Al)([A-Z])`)
var joinedNumberLabel = regexp.MustCompile(`\bN\s*°\s*([0-9])`)
var buildingAnchor = regexp.MustCompile(`\b(BLOC|IMMEUBLE|MAGASIN|LOCAL|LOTISSEMENT|NUMERO) ([A-Z]*[0-9][A-Z0-9]*)\b`)

// This additional comparison is used only with an equal business telephone.
// It expands explicit spelling/abbreviation variants, never edit distances.
func expandedAddress(value, city string) string {
	value = joinedArticle.ReplaceAllString(value, "$1")
	value = joinedNumberLabel.ReplaceAllString(ci.Normalize(value), " NUMERO $1")
	value = addressWords(value, city)
	var tokens []string
	for _, token := range strings.Fields(value) {
		switch token {
		case "MOHAMMED":
			token = "MOHAMED"
		case "CPLX", "COMPLEXE":
			token = "COMPLEXE"
		case "MAG", "MAGASIN":
			token = "MAGASIN"
		case "LOTIS":
			token = "LOTISSEMENT"
		case "MY":
			token = "MOULAY"
		case "CITE", "HAY", "QUARTIER":
			// District labels do not identify a different street or building.
			continue
		}
		tokens = append(tokens, token)
	}
	return strings.Join(tokens, " ")
}

func equivalentSpellingAddress(p pharmacy, r location, referencePhone string) bool {
	phone := phoneKey(referencePhone)
	if phone == "" || phone != phoneKey(r.Phone) {
		return false
	}
	a, b := expandedAddress(p.Address, p.City), expandedAddress(r.Address, r.City)
	if len(a) < 12 || len(b) < 12 {
		return false
	}
	an, bn := firstNumber.FindStringSubmatch(a), firstNumber.FindStringSubmatch(b)
	sameHouse := len(an) > 1 && len(bn) > 1 && an[1] == bn[1]
	anchors := func(value string) string {
		var result []string
		for _, match := range buildingAnchor.FindAllStringSubmatch(value, -1) {
			result = append(result, match[1]+":"+match[2])
		}
		sort.Strings(result)
		return strings.Join(result, "|")
	}
	aa, ba := anchors(a), anchors(b)
	if aa != ba || (!sameHouse && aa == "") {
		return false
	}
	// Preserve every word/code/number and its multiplicity. Typed building
	// identifiers must agree above, even if address phrases change order.
	left, right := strings.Fields(a), strings.Fields(b)
	sort.Strings(left)
	sort.Strings(right)
	return strings.Join(left, " ") == strings.Join(right, " ")
}

func phoneKey(s string) string {
	s = nondigits.ReplaceAllString(s, "")
	s = strings.TrimPrefix(s, "00")
	if strings.HasPrefix(s, "212") && len(s) == 12 {
		s = s[3:]
	}
	if strings.HasPrefix(s, "0") && len(s) == 10 {
		s = s[1:]
	}
	if len(s) != 9 {
		return ""
	}
	return s
}
func addressWords(value, city string) string {
	tokens := strings.Fields(words.ReplaceAllString(ci.Normalize(value), " "))
	var out []string
	for _, word := range tokens {
		switch word {
		case "AV", "AVENUE":
			word = "AVENUE"
		case "BD", "BOUL", "BOULEVARD":
			word = "BOULEVARD"
		case "RTE", "ROUTE":
			word = "ROUTE"
		case "IMM", "IMMEUBLE":
			word = "IMMEUBLE"
		case "RES", "RESID", "RESIDENCE":
			word = "RESIDENCE"
		case "LOT", "LOTISS", "LOTISSEMENT":
			word = "LOTISSEMENT"
		case "Q", "QT", "QRT", "QUARTIER":
			word = "QUARTIER"
		case "DE", "DU", "DES", "D", "L", "LA", "LE", "LES", "AL", "EL":
			continue
		}
		out = append(out, word)
	}
	result := strings.Join(out, " ")
	city = ci.Normalize(city)
	if strings.HasSuffix(result, " "+city) {
		result = strings.TrimSuffix(result, " "+city)
	}
	return strings.TrimSpace(result)
}
func compatibleAddress(p pharmacy, r location, referencePhone string) string {
	a, b := addressWords(p.Address, p.City), addressWords(r.Address, r.City)
	if len(a) < 12 || len(b) < 12 {
		return ""
	}
	if a == b {
		return "unique_name_city_same_address"
	}
	// Accept an address with additional neighborhood detail only at a complete
	// token boundary, with the same explicit building/street number.
	an, bn := firstNumber.FindStringSubmatch(a), firstNumber.FindStringSubmatch(b)
	sameNumber := len(an) > 1 && len(bn) > 1 && an[1] == bn[1]
	if sameNumber && len(a) >= 12 && len(b) >= 12 && (strings.HasPrefix(a, b+" ") || strings.HasPrefix(b, a+" ")) {
		return "unique_name_city_same_numbered_street"
	}
	phone := phoneKey(referencePhone)
	if equivalentSpellingAddress(p, r, referencePhone) {
		return "unique_name_city_phone_equivalent_address_spelling"
	}
	if phone == "" || phone != phoneKey(r.Phone) || !sameNumber {
		return ""
	}
	// Equal business telephone is corroboration, never a substitute for matching
	// the street. Require two meaningful common address terms as well.
	ignore := map[string]bool{"AVENUE": true, "BOULEVARD": true, "RUE": true, "ROUTE": true, "QUARTIER": true, "IMMEUBLE": true, "LOTISSEMENT": true, "RESIDENCE": true, "BLOC": true, "BIS": true, "MAROC": true}
	unique := map[string]bool{}
	for _, t := range strings.Fields(a) {
		if len(t) >= 3 && !ignore[t] && nondigits.ReplaceAllString(t, "") == "" {
			unique[t] = true
		}
	}
	common := map[string]bool{}
	for _, t := range strings.Fields(b) {
		if unique[t] {
			common[t] = true
		}
	}
	if len(common) >= 2 {
		return "unique_name_city_phone_building_and_street"
	}
	return ""
}

func precisePoint(r location) bool {
	if r.Latitude == nil || r.Longitude == nil {
		return false
	}
	lat, lon := *r.Latitude, *r.Longitude
	if math.IsNaN(lat) || math.IsNaN(lon) || math.IsInf(lat, 0) || math.IsInf(lon, 0) || lat < 20 || lat > 36.1 || lon < -18 || lon > -0.8 {
		return false
	}
	// Exclude coarse rounded points that could be neighborhood/city centres.
	return math.Abs(lat*1000-math.Round(lat*1000)) > 1e-7 || math.Abs(lon*1000-math.Round(lon*1000)) > 1e-7
}
func pointKey(r location) string {
	if r.Latitude == nil || r.Longitude == nil {
		return ""
	}
	return fmt.Sprintf("%.6f,%.6f", *r.Latitude, *r.Longitude)
}

func plan(pharmacies []pharmacy, locations []location, references []addressReference) []decision {
	counts := map[string]int{}
	sources := map[string]map[string]location{}
	points := map[string]map[string]bool{}
	referencePhones := map[string]map[string]bool{}
	for _, p := range pharmacies {
		counts[key(p.Name, p.City)]++
	}
	for _, r := range references {
		k := key(r.Name, r.City) + "\x1f" + ci.Normalize(r.Address)
		if referencePhones[k] == nil {
			referencePhones[k] = map[string]bool{}
		}
		if phone := phoneKey(r.Phone); phone != "" {
			referencePhones[k][phone] = true
		}
	}
	for _, r := range locations {
		k := key(r.Name, r.City)
		if sources[k] == nil {
			sources[k] = map[string]location{}
		}
		sources[k][r.URL] = r
		if point := pointKey(r); point != "" {
			if points[point] == nil {
				points[point] = map[string]bool{}
			}
			points[point][k] = true
		}
	}
	var result []decision
	for _, p := range pharmacies {
		d := decision{ID: p.ID, Name: p.Name, City: p.City, Address: p.Address}
		k := key(p.Name, p.City)
		matches := sources[k]
		switch {
		case p.Latitude != nil || p.Longitude != nil:
			d.Status = "preserve_existing_coordinates"
		case strings.TrimSpace(p.Address) == "":
			d.Status = "missing_address"
		case counts[k] > 1:
			d.Status = "ambiguous_database_identity"
		case len(matches) == 0:
			d.Status = "no_published_point"
		case len(matches) > 1:
			d.Status = "ambiguous_source_identity"
		default:
			for _, r := range matches {
				r := r
				d.Evidence = &r
				switch {
				case !precisePoint(r):
					d.Status = "invalid_or_coarse_point"
				case len(points[pointKey(r)]) > 1:
					d.Status = "shared_point_requires_review"
				default:
					phone := p.Phone
					phones := referencePhones[k+"\x1f"+ci.Normalize(p.Address)]
					if phoneKey(phone) == "" && len(phones) == 1 {
						for value := range phones {
							phone = value
						}
					}
					d.Method = compatibleAddress(p, r, phone)
					if d.Method == "" {
						d.Status = "address_not_corroborated"
					} else {
						d.Status = "fill_missing_coordinates"
					}
				}
			}
		}
		result = append(result, d)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
