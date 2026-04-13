package ics

import (
	"fmt"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/torreirow/dindoa/internal/scraper"
)

// Generator handles ICS file creation
type Generator struct {
	timezone *time.Location
}

// NewGenerator creates a new ICS generator
func NewGenerator() (*Generator, error) {
	// Load Europe/Amsterdam timezone for proper CET/CEST handling
	tz, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		return nil, fmt.Errorf("load timezone: %w", err)
	}

	return &Generator{
		timezone: tz,
	}, nil
}

// Generate creates an ICS file for the given matches
func (g *Generator) Generate(teamName string, matches []scraper.Match, outputFile string) error {
	// Create calendar
	cal := ics.NewCalendar()
	cal.SetVersion("2.0")
	cal.SetProductId("-//Dindoa//Dindoa ICS Generator//NL")

	// Add each match as an event
	for _, match := range matches {
		event := g.createEvent(teamName, match)
		cal.AddVEvent(event)
	}

	// Write to file
	if err := os.WriteFile(outputFile, []byte(cal.Serialize()), 0644); err != nil {
		return fmt.Errorf("write ICS file: %w", err)
	}

	return nil
}

// createEvent creates an ICS event for a match
func (g *Generator) createEvent(teamName string, match scraper.Match) *ics.VEvent {
	event := ics.NewEvent(g.generateUID(teamName, match))

	// Set timestamp
	event.SetDtStampTime(time.Now())

	// Parse match time in Europe/Amsterdam timezone
	dateTime := g.parseMatchDateTime(match)
	event.SetStartAt(dateTime)

	// Set summary (title)
	title := g.formatTitle(teamName, match)
	event.SetSummary(title)

	// Set location
	event.SetLocation(match.Location)

	// Set description
	desc := g.formatDescription(match)
	event.SetDescription(desc)

	return event
}

// generateUID creates a unique identifier for the event
func (g *Generator) generateUID(teamName string, match scraper.Match) string {
	slug := scraper.NormalizeTeamName(teamName)
	dateStr := match.Date.Format("2006-01-02")
	timeStr := strings.ReplaceAll(match.Time, ":", "")
	return fmt.Sprintf("%s-%s-%s@dindoa.nl", slug, dateStr, timeStr)
}

// formatTitle formats the event title based on home/away status
func (g *Generator) formatTitle(teamName string, match scraper.Match) string {
	if match.IsHome {
		// Home match: "Dindoa J3 - ASVD J1"
		return fmt.Sprintf("%s - %s", teamName, match.Away)
	}
	// Away match: "ASVD J1 - Dindoa J3"
	return fmt.Sprintf("%s - %s", match.Home, teamName)
}

// formatDescription creates the event description
func (g *Generator) formatDescription(match scraper.Match) string {
	matchType := "Uitwedstrijd"
	if match.IsHome {
		matchType = "Thuiswedstrijd"
	}
	opponent := match.Away
	if match.IsHome {
		opponent = match.Away
	} else {
		opponent = match.Home
	}
	return fmt.Sprintf("%s tegen %s", matchType, opponent)
}

// parseMatchDateTime combines date and time into a single time.Time in Europe/Amsterdam timezone
func (g *Generator) parseMatchDateTime(match scraper.Match) time.Time {
	// Parse time (HH:MM format)
	var hour, minute int
	fmt.Sscanf(match.Time, "%d:%d", &hour, &minute)

	// Create datetime in Europe/Amsterdam timezone
	return time.Date(
		match.Date.Year(),
		match.Date.Month(),
		match.Date.Day(),
		hour,
		minute,
		0,
		0,
		g.timezone,
	)
}

// DefaultOutputFilename generates a default output filename based on team name
func DefaultOutputFilename(teamName string) string {
	slug := scraper.NormalizeTeamName(teamName)
	return fmt.Sprintf("%s.ics", slug)
}
