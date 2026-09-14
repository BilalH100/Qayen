package scripts

import (
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Medication struct {
	StatutAMM               string `json:"statut_amm" csv:"statut_amm"`
	StatutCommercialisation string `json:"statut_commercialisation" csv:"statut_commercialisation"`
	Specialite              string `json:"specialite" csv:"specialite"`
	Dosage                  string `json:"dosage" csv:"dosage"`
	Forme                   string `json:"forme" csv:"forme"`
	Presentation            string `json:"presentation" csv:"presentation"`
	PPGN                    string `json:"pp_gn" csv:"pp_gn"`
	SubstanceActive         string `json:"substance_active" csv:"substance_active"`
	ClasseTherapeutique     string `json:"classe_therapeutique" csv:"classe_therapeutique"`
	EPI                     string `json:"epi" csv:"epi"`
	PPV                     string `json:"ppv" csv:"ppv"`
	PH                      string `json:"ph" csv:"ph"`
	PFHT                    string `json:"pfht" csv:"pfht"`
	Code                    string `json:"code" csv:"code"`
	TVA                     string `json:"tva" csv:"tva"`
}

type MedicationScraper struct {
	BaseURL     string
	Client      *http.Client
	UserAgent   string
	Delay       time.Duration
	Medications []Medication
}

func NewMedicationScraper() *MedicationScraper {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &MedicationScraper{
		BaseURL:   "https://ammps.sante.gov.ma/basesdedonnes/listes-medicaments",
		UserAgent: "Go-Medication-Scraper/1.0",
		Delay:     2 * time.Second,
		Client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: tr,
		},
		Medications: make([]Medication, 0),
	}
}

func (s *MedicationScraper) makeRequest(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", s.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	return doc, nil
}

func (s *MedicationScraper) parseMedicationTable(doc *goquery.Document) []Medication {
	var medications []Medication

	doc.Find("table.table tbody tr").Each(func(i int, row *goquery.Selection) {
		var cells []string

		row.Find("td").Each(func(j int, cell *goquery.Selection) {
			text := strings.TrimSpace(cell.Text())
			cells = append(cells, text)
		})

		if len(cells) >= 14 {
			medication := Medication{
				StatutAMM:               getCell(cells, 0),
				StatutCommercialisation: getCell(cells, 1),
				Specialite:              getCell(cells, 2),
				Dosage:                  getCell(cells, 3),
				Forme:                   getCell(cells, 4),
				Presentation:            getCell(cells, 5),
				PPGN:                    getCell(cells, 6),
				SubstanceActive:         getCell(cells, 7),
				ClasseTherapeutique:     getCell(cells, 8),
				EPI:                     getCell(cells, 9),
				PPV:                     getCell(cells, 10),
				PH:                      getCell(cells, 11),
				PFHT:                    getCell(cells, 12),
				Code:                    getCell(cells, 13),
				TVA:                     getCell(cells, 14),
			}
			medications = append(medications, medication)
		}
	})

	return medications
}

func (s *MedicationScraper) ScrapeAll() error {
	log.Println("Starting to scrape all medications...")

	page := 1
	for {
		var pageURL string
		if page == 1 {
			pageURL = s.BaseURL
		} else {
			pageURL = fmt.Sprintf("%s?page=%d", s.BaseURL, page)
		}
		fmt.Println("pageUrl aaaaaa", pageURL)
		log.Println(page)
		log.Printf("Scraping page %d: %s", page, pageURL)

		doc, err := s.makeRequest(pageURL)
		if err != nil {
			return fmt.Errorf("failed to scrape page %d: %w", page, err)
		}

		medications := s.parseMedicationTable(doc)
		if len(medications) == 0 {
			log.Printf("No more medications found on page %d, stopping", page)
			break
		}

		s.Medications = append(s.Medications, medications...)
		log.Printf("Found %d medications on page %d (total: %d)", len(medications), page, len(s.Medications))
		hasNextPage := doc.Find("a[rel='next']").Length() > 0 ||
			doc.Find(".pagination .next").Length() > 0

		if !hasNextPage {
			log.Println("No next page found, stopping")
			// break
		}

		time.Sleep(s.Delay)
		page++
	}

	log.Printf("Scraping completed. Total medications found: %d", len(s.Medications))
	return nil
}

func (s *MedicationScraper) SaveToJSON(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(s.Medications); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	log.Printf("Saved %d medications to %s", len(s.Medications), filename)
	return nil
}

func (s *MedicationScraper) SaveToCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"statut_amm", "statut_commercialisation", "specialite", "dosage", "forme",
		"presentation", "pp_gn", "substance_active", "classe_therapeutique", "epi",
		"ppv", "ph", "pfht", "code", "tva",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, med := range s.Medications {
		record := []string{
			med.StatutAMM, med.StatutCommercialisation, med.Specialite, med.Dosage,
			med.Forme, med.Presentation, med.PPGN, med.SubstanceActive,
			med.ClasseTherapeutique, med.EPI, med.PPV, med.PH, med.PFHT,
			med.Code, med.TVA,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	log.Printf("Saved %d medications to %s", len(s.Medications), filename)
	return nil
}

func (s *MedicationScraper) PrintSummary() {
	if len(s.Medications) == 0 {
		log.Println("No medications found")
		return
	}

	log.Printf("\n=== Medication Database Summary ===")
	log.Printf("Total medications: %d", len(s.Medications))

	if len(s.Medications) > 0 {
		sample := s.Medications[0]
		log.Printf("\nSample medication:")
		log.Printf("  Name: %s", sample.Specialite)
		log.Printf("  Active Substance: %s", sample.SubstanceActive)
		log.Printf("  Form: %s", sample.Forme)
		log.Printf("  Dosage: %s", sample.Dosage)
		log.Printf("  Status: %s", sample.StatutCommercialisation)
	}
}

func ScrapeMeds() {
	scraper := NewMedicationScraper()
	err := scraper.ScrapeAll()
	log.Println(err)
	scraper.PrintSummary()

	if len(scraper.Medications) > 0 {
		if err := scraper.SaveToJSON("medications.json"); err != nil {
			log.Printf("Failed to save JSON: %v", err)
		}

		if err := scraper.SaveToCSV("medications.csv"); err != nil {
			log.Printf("Failed to save CSV: %v", err)
		}
	}

	log.Println("3ich")
}
