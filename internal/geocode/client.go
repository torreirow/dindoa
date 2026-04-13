package geocode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	nominatimURL = "https://nominatim.openstreetmap.org/search"
	userAgent    = "Dindoa-ICS-Generator/1.0 (korfbal calendar tool)"
	timeout      = 30 * time.Second
)

// Result represents a geocoding result
type Result struct {
	Query   string    `json:"query"`
	Address string    `json:"address"`
	Lat     float64   `json:"lat"`
	Lng     float64   `json:"lng"`
	CachedAt time.Time `json:"cached_at"`
}

// NominatimResponse represents the JSON response from Nominatim
type NominatimResponse struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

// Client handles geocoding requests to OpenStreetMap Nominatim
type Client struct {
	httpClient  *http.Client
	rateLimiter *RateLimiter
}

// NewClient creates a new geocoding client
func NewClient(rateLimiter *RateLimiter) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		rateLimiter: rateLimiter,
	}
}

// Geocode searches for a location and returns the geocoded result
// Falls back to the original query string if geocoding fails
func (c *Client) Geocode(query string) Result {
	// Wait for rate limiter
	c.rateLimiter.Wait()

	// Build request URL
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("limit", "1")
	params.Set("addressdetails", "1")

	reqURL := fmt.Sprintf("%s?%s", nominatimURL, params.Encode())

	// Create request
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return c.fallback(query, err)
	}

	req.Header.Set("User-Agent", userAgent)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return c.fallback(query, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.fallback(query, fmt.Errorf("status %d", resp.StatusCode))
	}

	// Parse response
	var results []NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return c.fallback(query, err)
	}

	if len(results) == 0 {
		return c.fallback(query, fmt.Errorf("no results found"))
	}

	// Extract first result
	first := results[0]
	var lat, lng float64
	fmt.Sscanf(first.Lat, "%f", &lat)
	fmt.Sscanf(first.Lon, "%f", &lng)

	return Result{
		Query:   query,
		Address: first.DisplayName,
		Lat:     lat,
		Lng:     lng,
		CachedAt: time.Now(),
	}
}

// fallback returns a result using the original query as the address
func (c *Client) fallback(query string, err error) Result {
	// Log warning could be added here if needed
	return Result{
		Query:   query,
		Address: query, // Use original query as fallback
		Lat:     0,
		Lng:     0,
		CachedAt: time.Now(),
	}
}
