package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GeoLocationResponse represents the response from the IP geolocation API
type GeoLocationResponse struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Query       string  `json:"query"`
}

// GetCountryFromIP returns the country for a given IP address
// Uses the free IP-API service (http://ip-api.com/json/)
func GetCountryFromIP(ipAddress string) (string, error) {
	// Handle localhost or empty IP
	if ipAddress == "" || ipAddress == "127.0.0.1" || ipAddress == "::1" {
		return "Unknown", nil
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make request to geolocation API
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,countryCode", ipAddress)
	resp, err := client.Get(url)
	if err != nil {
		return "Unknown", fmt.Errorf("failed to query geolocation API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Unknown", fmt.Errorf("failed to read geolocation API response: %w", err)
	}

	// Parse JSON response
	var geoResp GeoLocationResponse
	if err := json.Unmarshal(body, &geoResp); err != nil {
		return "Unknown", fmt.Errorf("failed to parse geolocation API response: %w", err)
	}

	// Check if the request was successful
	if geoResp.Status != "success" {
		return "Unknown", fmt.Errorf("geolocation API returned non-success status")
	}

	// Return country code if available, otherwise country name
	if geoResp.CountryCode != "" {
		return geoResp.CountryCode, nil
	}

	return geoResp.Country, nil
}
