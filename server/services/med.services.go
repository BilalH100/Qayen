package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	sqlc "kayena/server/database/generated"
	"kayena/server/models"
	"kayena/server/repository"
	"kayena/server/schemas"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type MedService interface {
	CreateMedication(ctx context.Context, med models.Medication) error
	UpdateMedication(ctx context.Context, med models.Medication) (*models.Medication, error)
	DeleteMedication(ctx context.Context, id int32) error
	GetMedicationByCode(ctx context.Context, code string) (*models.Medication, error)
	GetMedicationById(ctx context.Context, id int32) (*models.Medication, error)
	GetAllMedications(ctx context.Context, opt schemas.Options) ([]models.Medication, error)
}

type medService struct {
	medRepo repository.MedRepository
}

func NewMedService(medRepo repository.MedRepository) MedService {
	return &medService{
		medRepo: medRepo,
	}
}

func (s *medService) CreateMedication(ctx context.Context, med models.Medication) error {
	m := sqlc.CreateMedicationParams{
		Code:             med.Code,
		Status:           med.Status,
		CommercialStatus: med.CommercialStatus,
		Speciality:       med.Speciality,
		Dosage:           med.Dosage,
		Form:             med.Form,
		Presentation:     med.Presentation,
		Pp:               med.Pp,
		ActiveSubstance:  med.ActiveSubstance,
		TherapeuticClass: med.TherapeuticClass,
		Epi:              med.Epi,
		Ppv:              med.Ppv,
		Ph:               med.Ph,
		Pfht:             med.Pfht,
		Tva:              med.Tva,
		Description:      med.Description,
		CommonSd:         med.CommonSd,
		SeriousSd:        med.SeriousSd,
		GeneralInfo:      med.GeneralInfo,
	}
	return s.medRepo.Create(ctx, m)
}

func (s *medService) UpdateMedication(ctx context.Context, med models.Medication) (*models.Medication, error) {
	return s.medRepo.Update(ctx, sqlc.UpdateMedicationParams{
		Code:             med.Code,
		Status:           med.Status,
		CommercialStatus: med.CommercialStatus,
		Speciality:       med.Speciality,
		Dosage:           med.Dosage,
		Form:             med.Form,
		Presentation:     med.Presentation,
		Pp:               med.Pp,
		ActiveSubstance:  med.ActiveSubstance,
		TherapeuticClass: med.TherapeuticClass,
		Epi:              med.Epi,
		Ppv:              med.Ppv,
		Ph:               med.Ph,
		Pfht:             med.Pfht,
		Tva:              med.Tva,
		Description:      med.Description,
		CommonSd:         med.CommonSd,
		SeriousSd:        med.SeriousSd,
		GeneralInfo:      med.GeneralInfo,
	})
}

func (s *medService) DeleteMedication(ctx context.Context, id int32) error {
	return s.medRepo.Delete(ctx, id)
}

func (s *medService) GetMedicationByCode(ctx context.Context, code string) (*models.Medication, error) {
	return s.medRepo.GetByCode(ctx, code)
}
func (s *medService) GetMedicationById(ctx context.Context, id int32) (*models.Medication, error) {
	return s.medRepo.GetByID(ctx, id)
}

func (s *medService) GetAllMedications(ctx context.Context, opt schemas.Options) ([]models.Medication, error) {
	return s.medRepo.GetAll(ctx, opt)
}

func GetMetaData(drugName string, e *schemas.Failed) (string, schemas.SideEffects, error) {
	d, err := GetDescription(drugName)
	if err != nil {
		e.Desc += 1
		log.Println("Error getting description for :", drugName)
	}
	sd, err := GetSideEffects(drugName)
	if err != nil {
		e.Side += 1
		log.Println("Erro getting side effects for :", drugName)
	}

	return d, sd, nil
}

func GetDescription(dn string) (string, error) {
	var (
		desc string
		err  error
	)
	if desc, err = GetFromDrugs(dn); err == nil && desc != "" {
		return desc, nil
	}
	if desc, err := GetFromMedicamentMA(dn); err == nil && desc != "" {
		return desc, nil
	}
	if desc, err = GetFromWikipedia(dn); err == nil && desc != "" {
		return desc, nil
	}

	return desc, nil
}
func GetSideEffects(dn string) (schemas.SideEffects, error) {
	page := url.PathEscape(strings.ToLower(strings.ReplaceAll(dn, " ", "_")))
	apiURL := fmt.Sprintf("https://www.drugs.com/search.php?searchterm=%s", page)
	doc, err := goquery.NewDocument(apiURL)
	if err != nil {
		log.Fatal("Failed to fetch URL:", err)
		return schemas.SideEffects{}, err
	}

	var sde []string
	doc.Find("h2#side-effects").Each(func(i int, s *goquery.Selection) {
		p := s.NextAllFiltered("p").FilterFunction(func(_ int, p *goquery.Selection) bool {
			return strings.Contains(p.Text(), "Serious side effects of paracetamol include")
		}).First()

		if p.Length() == 0 {
			return
		}
		ul := p.NextFiltered("ul")
		ul.Find("li").Each(func(_ int, li *goquery.Selection) {
			sde = append(sde, strings.TrimSpace(li.Text()))
		})
	})

	fmt.Println("Serious side effects:")
	for _, s := range sde {
		fmt.Println("- " + s)
	}
	fmt.Println()
	var cde []string
	doc.Find("h2#side-effects").Each(func(i int, s *goquery.Selection) {
		p := s.NextAllFiltered("p").FilterFunction(func(_ int, p *goquery.Selection) bool {
			return strings.Contains(p.Text(), "Common side effect of paracetamol suppositories include")
		}).First()

		if p.Length() == 0 {
			return
		}

		ul := p.NextFiltered("ul")
		ul.Find("li").Each(func(_ int, li *goquery.Selection) {
			cde = append(cde, strings.TrimSpace(li.Text()))
		})
	})

	fmt.Println("Common side effects:")
	for _, c := range cde {
		fmt.Println("- " + c)
	}

	return schemas.SideEffects{
		Common:  cde,
		Serious: sde,
	}, nil
}

func GetFromWikipedia(drugName string) (string, error) {
	page := url.PathEscape(strings.ToLower(strings.ReplaceAll(drugName, " ", "_")))
	apiURL := fmt.Sprintf("https://en.wikipedia.org/api/rest_v1/page/summary/%s", page)
	fmt.Println("wiki :", apiURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("wikipedia returned status: %d", resp.StatusCode)
	}

	var data struct {
		Extract string `json:"extract"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	if data.Extract == "" {
		return "", fmt.Errorf("no extract")
	}
	return strings.TrimSpace(data.Extract), nil
}

func GetFromDrugs(drugName string) (string, error) {
	page := url.PathEscape(strings.ToLower(strings.ReplaceAll(drugName, " ", "_")))
	apiURL := fmt.Sprintf("https://www.drugs.com/search.php?searchterm=%s", page)
	fmt.Println("drugs.com :", apiURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal("Failed to parse HTML:", err)
	}
	description := doc.Find(".ddc-media-content p").First().Text()
	if description == "" {
		return "", fmt.Errorf("no description found")
	}
	return description, nil
}

func GetFromMedicamentMA(drugName string) (string, error) {
	slug := strings.ToLower(strings.ReplaceAll(drugName, " ", "-"))
	url := fmt.Sprintf("https://medicament.ma/medicament/%s", slug)
	fmt.Println("med.ma :", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch medicament.ma: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("medicament.ma returned status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	desc := strings.TrimSpace(doc.Find("tr.field-indication td.value").Text())
	if desc != "" {
		return desc, nil
	}

	desc = strings.TrimSpace(doc.Find("tr.field-famille td.value").Text())
	if desc != "" {
		return "Classe thérapeutique : " + desc, nil
	}

	return "", fmt.Errorf("no indication or fallback field found")
}
