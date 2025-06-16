package services

import (
	"encoding/json"
	"fmt"
	"io"
	"kayena/server/schemas"
	"net/http"
	"net/url"
)

func GetGeoCoordinates(phar schemas.Pharmacy) (*schemas.Coordinates, error) {
	url := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&polygon_geojson=1&format=jsonv2",
		url.QueryEscape(fmt.Sprintf("%s %s", phar.Name, phar.City)))
	fmt.Println("GEO URL : ", url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var m []map[string]interface{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}
	if len(m) == 0 {
		return nil, fmt.Errorf("no results found for query")
	}
	lat, ok := m[0]["lat"].(string)
	if !ok {
		return nil, fmt.Errorf("latitude not found in response")
	}
	long, ok := m[0]["lon"].(string)
	if !ok {
		return nil, fmt.Errorf("longitude not found in response")
	}

	c := &schemas.Coordinates{
		Lat:  lat,
		Long: long,
	}
	return c, nil
}

func GetAddressByGeo(c schemas.Coordinates) (string, error) {
	url := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?lat=%s&lon=%s&format=jsonv2", c.Lat, c.Long)
	fmt.Println("ADDRESS URL :", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	// req.Header.Set("User-Agent", "kayena") 

	resp, err := client.Do(req)
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

	var m map[string]interface{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	address, ok := m["address"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("address field not found in response")
	}
	road, ok := address["road"].(string)
	if !ok {
		return "", fmt.Errorf("road field not found in address")
	}
	return road, nil
}
