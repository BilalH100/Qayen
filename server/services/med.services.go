package services

import (
	"encoding/json"
	"fmt"
	"io"
	"kayena/server/schemas"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetMetaData(drugName string, e *schemas.Failed) (string, schemas.SideEffects, error) {
	d, err := getDescription(drugName)
	if err != nil {
		e.Desc += 1
		log.Println("Error getting description for :", drugName)
	}
	sd, err := getSideEffects(drugName)
	if err != nil {
		e.Side += 1
		log.Println("Erro getting side effects for :", drugName)
	}

	return d, sd, nil
}

func getDescription(dn string) (string, error) {
	var (
		desc string
		err  error
	)

	if desc, err = getFromWikipedia(dn); err == nil && desc != "" {
		return desc, nil
	}
	if desc, err = getFromDrugs(dn); err == nil && desc != "" {
		return desc, nil
	}
	if desc, err := getFromMedicamentMA(dn); err == nil && desc != "" {
		return desc, nil
	}
	return desc, nil
}
func getSideEffects(dn string) (schemas.SideEffects, error) {
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

func getFromWikipedia(drugName string) (string, error) {
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

func getFromDrugs(drugName string) (string, error) {
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

func getFromMedicamentMA(drugName string) (string, error) {
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
