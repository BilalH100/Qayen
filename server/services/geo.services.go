package services

import (
	"encoding/json"
	"fmt"
	"io"
	"kayena/server/schemas"
	"net/http"
	"net/url"
	"os"
)

var apiKey = os.Getenv("MAPS_API_KEY") 

func GetGeoCoordinates(phar schemas.Pharmacy) (*schemas.Coordinates, error) {
	baseURL := "https://maps.googleapis.com/maps/api/geocode/json"
	query := fmt.Sprintf("%s %s", phar.Name, phar.City)
	params := url.Values{}
	params.Add("address", query)
	params.Add("key", apiKey)

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	fmt.Println("GEO URL:", fullURL)

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var res struct {
		Results []struct {
			Geometry struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
		} `json:"results"`
		Status string `json:"status"`
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if res.Status != "OK" || len(res.Results) == 0 {
		return nil, fmt.Errorf("no results found or API returned status: %s", res.Status)
	}

	c := &schemas.Coordinates{
		Lat:  fmt.Sprintf("%f", res.Results[0].Geometry.Location.Lat),
		Long: fmt.Sprintf("%f", res.Results[0].Geometry.Location.Lng),
	}
	return c, nil
}

func GetAddressByGeo(c schemas.Coordinates) (string, error) {
	baseURL := "https://maps.googleapis.com/maps/api/geocode/json"
	params := url.Values{}
	params.Add("latlng", fmt.Sprintf("%s,%s", c.Lat, c.Long))
	params.Add("key", os.Getenv("GOOGLE_MAPS_API_KEY"))

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	fmt.Println("ADDRESS URL:", fullURL)

	resp, err := http.Get(fullURL)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var res struct {
		Results []struct {
			FormattedAddress string `json:"formatted_address"`
		} `json:"results"`
		Status string `json:"status"`
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if res.Status != "OK" || len(res.Results) == 0 {
		return "", fmt.Errorf("no results found or API returned status: %s", res.Status)
	}

	return res.Results[0].FormattedAddress, nil
}
