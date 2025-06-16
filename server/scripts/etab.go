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

type EtabScraper struct {
	BaseURL   string
	Client    *http.Client
	UserAgent string
	Delay     time.Duration
	Etabs     []Etab
}

type Etab struct {
	Nom         string `json:"nom" csv:"nom"`
	Ville       string `json:"ville" csv:"ville"`
	Adresse     string `json:"adresse" csv:"adresse"`
	Tel         string `json:"tel" csv:"tel"`
	Fax         string `json:"fax" csv:"fax"`
	Responsable string `json:"responsable" csv:"responsable"`
}

func NewEtabScrepper() *EtabScraper {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &EtabScraper{
		BaseURL:   "https://ammps.sante.gov.ma/basesdedonnes/etablissements-pharmaceutiques-grossistes-repartiteurs",
		UserAgent: "go-etab-scraoer/1.0",
		Delay:     2 * time.Second,
		Client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: tr,
		},
		Etabs: make([]Etab, 0),
	}
}

func (s *EtabScraper) makeRequest(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", s.UserAgent)
	req.Header.Set("Accept", "text/html, application/html+xml, application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request :%w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code :%d", resp.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html : %w", err)
	}
	return doc, nil
}

func (s *EtabScraper) ScrapeAll() error {
	log.Println("Starting to scrape all etabs ...")
	page := 1
	for {
		var pageURL string
		if page == 1 {
			pageURL = s.BaseURL
		} else {
			pageURL = fmt.Sprintf("%s?page=%d", s.BaseURL, page)
		}
		log.Printf("Scraping page %d : %s", page, pageURL)
		doc, err := s.makeRequest(pageURL)
		if err != nil {
			return fmt.Errorf("failed to scrape page %d : %w", page, err)
		}
		etabs := s.parseEtabsTable(doc)
		if len(etabs) == 0 {
			log.Printf("no more etabs found on page %d", page)
			break
		}
		s.Etabs = append(s.Etabs, etabs...)
		log.Printf("Found %d medications on page %d (total : %d)", len(etabs), page, len(s.Etabs))
		// hasNextPage := doc.Find("a[rel='next']").Length() > 0
		// if !hasNextPage {
		// 	log.Printf("no next page found, stoppinnng")
		// 	// break
		// }
		time.Sleep(s.Delay)
		page++
	}
	log.Printf("scraping finished , total etabs : %d ", len(s.Etabs))
	return nil
}

func getCell(cells []string, index int) string {
	if index < len(cells) {
		return cells[index]
	}
	return ""
}

func (s *EtabScraper) parseEtabsTable(doc *goquery.Document) []Etab {
	var etabs []Etab
	doc.Find("table.table tbody tr").Each(func(i int, row *goquery.Selection) {
		var cells []string
		row.Find("td").Each(func(j int, cell *goquery.Selection) {
			text := strings.TrimSpace(cell.Text())
			cells = append(cells, text)
		})
		if len(cells) >= 5 {
			etab := Etab{
				Nom:         getCell(cells, 0),
				Ville:       getCell(cells, 1),
				Adresse:     getCell(cells, 2),
				Tel:         getCell(cells, 3),
				Fax:         getCell(cells, 4),
				Responsable: getCell(cells, 5),
			}
			etabs = append(etabs, etab)
		}
	})
	return etabs
}

func (s *EtabScraper) PrintStuff() {
	if len(s.Etabs) == 0 {
		log.Println("No etab found")
		return
	}

	log.Printf("\n=== etab Database Summary ===")
	log.Printf("Total etab: %d", len(s.Etabs))

	if len(s.Etabs) > 0 {
		sample := s.Etabs[0]
		log.Printf("\nSample Etab:")
		log.Printf("  Nom: %s", sample.Nom)
		log.Printf("  Ville: %s", sample.Ville)
		log.Printf("  Adresse: %s", sample.Adresse)
		log.Printf("  Tel: %s", sample.Tel)
		log.Printf("  Fax: %s", sample.Fax)
		log.Printf("  Responsable: %s", sample.Responsable)
	}
}

func (s *EtabScraper) SaveToJSON(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(s.Etabs); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	log.Printf("Saved %d etabs to %s", len(s.Etabs), filename)
	return nil
}

func (s *EtabScraper) SaveToCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"nom", "ville", "adresse", "tel", "fax", "responsable",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, med := range s.Etabs {
		record := []string{
			med.Nom, med.Ville, med.Adresse, med.Tel,
			med.Fax, med.Responsable,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	log.Printf("Saved %d medications to %s", len(s.Etabs), filename)
	return nil
}
func ScrapeEtab() {
	scraper := NewEtabScrepper()
	err := scraper.ScrapeAll()
	if err != nil {
		log.Fatal("Failed to scrape all the etab")
	}
	scraper.PrintStuff()

	if len(scraper.Etabs) > 0 {
		err := scraper.SaveToJSON("etablissement-pharmaceutiques-grossites-repartiteurs.json") // long aahh name
		if err != nil {
			log.Printf("Failed to save JSON %v", err)
		}

		err = scraper.SaveToCSV("etablissement-pharmaceutiques-grossites-repartiteurs.csv")
		if err != nil {
			log.Printf("Failed to save JSON %v", err)
		}
	}

}
