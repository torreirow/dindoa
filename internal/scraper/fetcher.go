package scraper

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	baseURL      = "https://dindoa.nl/ws"
	programmaURL = baseURL + "/competitie-programma/"
	userAgent    = "Dindoa-ICS-Generator/1.0"
	timeout      = 30 * time.Second
)

// Fetcher handles HTTP requests and HTML parsing
type Fetcher struct {
	client *http.Client
}

// NewFetcher creates a new Fetcher with default settings
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// FetchDocument fetches a URL and returns a goquery Document
func (f *Fetcher) FetchDocument(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	return doc, nil
}

// FetchProgrammaPage fetches and returns the match programme document.
// This page is the single source for matches, teams, categories and colours.
func (f *Fetcher) FetchProgrammaPage() (*goquery.Document, error) {
	return f.FetchDocument(programmaURL)
}

// ProgrammaURL returns the URL the match programme is read from.
func ProgrammaURL() string {
	return programmaURL
}
