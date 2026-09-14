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

type PharmScraper struct {
	BaseURL   string
	Client    *http.Client
	UserAgent string
	Delay     time.Duration
	Pharms      []Pharm
}

type Pharm struct {
	Nom string `json:"nom" csv:"nom"`
	Ville string `json:"ville" csv:"ville"`
}

func NewPharmScrepper() *PharmScraper {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &PharmScraper{
		BaseURL:   "https://ammps.sante.gov.ma/basesdedonnes/pharmacies",
		UserAgent: "go-Pharm-scraoer/1.0",
		Delay:     2 * time.Second,
		Client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: tr,
		},
		Pharms: make([]Pharm, 0),
	}
}

func (s *PharmScraper) makeRequest(url string) (*goquery.Document, error) {
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

func (s *PharmScraper) ScrapeAll() error {
	log.Println("Starting to scrape all Pharms ...")
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
		Pharms := s.parsePharmsTable(doc)
		if len(Pharms) == 0 {
			log.Printf("no more Pharms found on page %d",page)
			break
		}
		s.Pharms = append(s.Pharms, Pharms...)
		log.Printf("Found %d medications on page %d (total : %d)", len(Pharms), page, len(s.Pharms))
		// hasNextPage := doc.Find("a[rel='next']").Length() > 0
		// if !hasNextPage {
		// 	log.Printf("no next page found, stoppinnng")
		// 	// break
		// }
		time.Sleep(s.Delay)
		page++
	}
	log.Printf("scraping finished , total Pharms : %d ", len(s.Pharms))
	return nil
}



func (s *PharmScraper) parsePharmsTable(doc *goquery.Document) []Pharm {
	var Pharms []Pharm
	doc.Find("table.table tbody tr").Each(func(i int, row *goquery.Selection) {
		var cells []string
		row.Find("td").Each(func(j int, cell *goquery.Selection) {
			text := strings.TrimSpace(cell.Text())
			cells = append(cells, text)
		})
		if len(cells) >= 1 {
			Pharm := Pharm{
				Nom:         getCell(cells, 0),
				Ville:       getCell(cells, 1),
			}
			Pharms = append(Pharms, Pharm)
		}
	})
	fmt.Println("Pharms", Pharms)
	return Pharms
}

func (s *PharmScraper) PrintStuff() {
	if len(s.Pharms) == 0 {
		log.Println("No Pharm found")
		return
	}

	log.Printf("\n=== Pharm Database Summary ===")
	log.Printf("Total Pharm: %d", len(s.Pharms))

	if len(s.Pharms) > 0 {
		sample := s.Pharms[0]
		log.Printf("\nSample Pharm:")
		log.Printf("  Nom: %s", sample.Nom)
		log.Printf("  Ville: %s", sample.Ville)
	}
}

func (s *PharmScraper) SaveToJSON(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(s.Pharms); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	log.Printf("Saved %d Pharms to %s", len(s.Pharms), filename)
	return nil
}

func (s *PharmScraper) SaveToCSV(filename string) error {
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

	for _, med := range s.Pharms {
		record := []string{
			med.Nom, med.Ville,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	log.Printf("Saved %d medications to %s", len(s.Pharms), filename)
	return nil
}

func ScrapePhar() {
	scraper := NewPharmScrepper()
	err := scraper.ScrapeAll()
	if err != nil {
		log.Fatal("Failed to scrape all the Pharm")
	}
	scraper.PrintStuff()

	if len(scraper.Pharms) > 0 {
		err := scraper.SaveToJSON("Pharmlissement-pharmaceutiques-grossites-repartiteurs.json") // long aahh name
		if err != nil {
			log.Printf("Failed to save JSON %v", err)
		}

		err = scraper.SaveToCSV("Pharmlissement-pharmaceutiques-grossites-repartiteurs.csv")
		if err != nil {
			log.Printf("Failed to save JSON %v", err)
		}
	}

}

