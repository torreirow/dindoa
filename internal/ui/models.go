package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/torreirow/dindoa/internal/geocode"
	"github.com/torreirow/dindoa/internal/ics"
	"github.com/torreirow/dindoa/internal/scraper"
)

// Message types

type categoriesMsg struct {
	categories []scraper.Category
	err        error
}

type teamsMsg struct {
	teams []scraper.Team
}

type doneMsg struct {
	outputFile string
	matchCount int
	err        error
}

// Commands

func fetchCategories(fetcher *scraper.Fetcher, parser *scraper.Parser) tea.Cmd {
	return func() tea.Msg {
		doc, err := fetcher.FetchTeamsPage()
		if err != nil {
			return categoriesMsg{err: err}
		}

		categories, err := parser.ParseTeams(doc)
		if err != nil {
			return categoriesMsg{err: err}
		}

		return categoriesMsg{categories: categories}
	}
}

func generateICS(teamName string) tea.Cmd {
	return func() tea.Msg {
		// Normalize team name
		teamSlug := scraper.NormalizeTeamName(teamName)

		// Fetch matches
		fetcher := scraper.NewFetcher()
		parser := scraper.NewParser()

		doc, err := fetcher.FetchTeamPage(teamSlug)
		if err != nil {
			return doneMsg{err: fmt.Errorf("fetch team page: %w", err)}
		}

		matches, err := parser.ParseMatches(doc)
		if err != nil {
			return doneMsg{err: fmt.Errorf("parse matches: %w", err)}
		}

		if len(matches) == 0 {
			return doneMsg{err: fmt.Errorf("no matches found for this team")}
		}

		// Initialize geocoding
		cache, _ := geocode.NewCache()
		rateLimiter := geocode.NewRateLimiter()
		geocoder := geocode.NewClient(rateLimiter)

		// Geocode locations
		for i := range matches {
			location := matches[i].Location

			// Check cache
			if cache != nil {
				if result, found := cache.Lookup(location); found {
					matches[i].Location = result.Address
					continue
				}
			}

			// Geocode
			result := geocoder.Geocode(location)
			matches[i].Location = result.Address

			// Store in cache
			if cache != nil {
				cache.Store(result)
			}
		}

		// Generate ICS
		generator, err := ics.NewGenerator()
		if err != nil {
			return doneMsg{err: fmt.Errorf("create generator: %w", err)}
		}

		outputFile := ics.DefaultOutputFilename(teamName)

		// Use display name
		displayName := teamName
		if !strings.Contains(strings.ToLower(teamName), "dindoa") {
			displayName = "Dindoa " + strings.ToUpper(teamName)
		}

		if err := generator.Generate(displayName, matches, outputFile); err != nil {
			return doneMsg{err: fmt.Errorf("generate ICS: %w", err)}
		}

		return doneMsg{
			outputFile: outputFile,
			matchCount: len(matches),
		}
	}
}

// View functions

func (m model) viewCategorySelection() string {
	var b strings.Builder

	b.WriteString("Dindoa ICS Generator\n\n")
	b.WriteString("Selecteer categorie:\n\n")

	for i, cat := range m.categories {
		cursor := " "
		if i == m.selected {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, cat.Name))
	}

	b.WriteString("\n[↑↓: navigeren] [enter: kiezen] [q: afsluiten]\n")

	return b.String()
}

func (m model) viewTeamSelection() string {
	var b strings.Builder

	b.WriteString("Dindoa ICS Generator\n\n")
	b.WriteString(fmt.Sprintf("Teams in %s:\n\n", m.selectedCategory))

	for i, team := range m.teams {
		cursor := " "
		if i == m.selected {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, team.Name))
	}

	b.WriteString("\n[↑↓: navigeren] [enter: kiezen] [q: afsluiten]\n")

	return b.String()
}

func (m model) viewProcessing() string {
	return fmt.Sprintf("Wedstrijden ophalen voor %s...\n\nGeocoding locaties...\n", m.selectedTeam)
}

func (m model) viewDone() string {
	var b strings.Builder

	b.WriteString("✓ ICS bestand aangemaakt!\n\n")
	b.WriteString(fmt.Sprintf("Bestand: %s\n", m.outputFile))
	b.WriteString(fmt.Sprintf("Wedstrijden: %d\n\n", m.matchCount))
	b.WriteString("Import in je kalender app.\n\n")
	b.WriteString("[enter: afsluiten]\n")

	return b.String()
}

func (m model) viewError() string {
	return fmt.Sprintf("Error: %v\n\n[enter: afsluiten]\n", m.err)
}
