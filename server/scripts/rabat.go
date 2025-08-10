package scripts

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

type Pharmacy struct {
	Name       string `json:"name"`
	ProfileURL string `json:"profile_url"`
	Activity   string `json:"activity"`
	Adress     string `json:"address"`
	City       string `json:"city"`
	Latitude   string `json:"latitude,omitempty"`
	Longitude  string `json:"longitude,omitempty"`
	Phone      string `json:"phone"`
}

func main() {
	file, err := os.ReadFile("pharmacies.json")
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	var pharmacies []Pharmacy
	if err := json.Unmarshal(file, &pharmacies); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	c := colly.NewCollector(
		colly.AllowedDomains("www.telecontact.ma", "telecontact.ma"),
		colly.UserAgent("Mozilla/5.0"),
	)

	latRegex := regexp.MustCompile(`window\.latitude\s*=\s*"([^"]+)"`)
	lonRegex := regexp.MustCompile(`window\.longitude\s*=\s*"([^"]+)"`)

	c.OnResponse(func(r *colly.Response) {
		url := r.Request.URL.String()

		for i, p := range pharmacies {
			if p.ProfileURL == url {
				body := string(r.Body)
				latMatch := latRegex.FindStringSubmatch(body)
				lonMatch := lonRegex.FindStringSubmatch(body)

				if len(latMatch) > 1 {
					pharmacies[i].Latitude = latMatch[1]
				}
				if len(lonMatch) > 1 {
					pharmacies[i].Longitude = lonMatch[1]
				}

				c.OnHTML("a[itemprop='telephone']", func(e *colly.HTMLElement) {
					url := e.Request.URL.String()
					for i, p := range pharmacies {
						if p.ProfileURL == url {
							pharmacies[i].Phone = strings.TrimSpace(e.Text)
							fmt.Printf("Got phone for %s: %s\n", p.Name, e.Text)
							break
						}
					}
				})
				fmt.Printf("✔ Got coordinates for %s\n", p.Name)
				break
			}
		}
	})

	for _, p := range pharmacies {
		if strings.HasPrefix(p.ProfileURL, "https://") {
			err := c.Visit(p.ProfileURL)
			if err != nil {
				log.Printf("Failed to visit %s: %v", p.ProfileURL, err)
			}
		}
	}
	output, err := os.Create("pharmacies_with_coords.json")
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer output.Close()

	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(pharmacies); err != nil {
		log.Fatalf("Error encoding result: %v", err)
	}

	fmt.Println("Done. Saved to pharmacies_with_coords.json")
}
