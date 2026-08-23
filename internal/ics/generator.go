package ics

import (
	"fmt"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/torreirow/dindoa/internal/locations"
	"github.com/torreirow/dindoa/internal/scraper"
)

// matchDuration is how long a match is assumed to last. A VEVENT with a
// DATE-TIME DTSTART and no DTEND is zero seconds long per RFC 5545, which
// calendar apps render as a moment rather than an appointment.
const matchDuration = time.Hour

// Generator handles ICS file creation
type Generator struct {
	timezone *time.Location
	mapping  *locations.Mapping
}

// NewGenerator creates a new ICS generator using the given location mapping.
// The mapping may be nil, in which case every venue falls back to the string
// published on the website.
func NewGenerator(mapping *locations.Mapping) (*Generator, error) {
	// Load Europe/Amsterdam timezone for proper CET/CEST handling
	tz, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		return nil, fmt.Errorf("load timezone: %w", err)
	}

	return &Generator{
		timezone: tz,
		mapping:  mapping,
	}, nil
}

// Generate creates an ICS file for the given matches and reports which venues
// were not present in the location mapping, with the number of matches each
// affects. A missing venue never stops generation.
func (g *Generator) Generate(teamName string, matches []scraper.Match, outputFile string) (map[string]int, error) {
	cal := ics.NewCalendar()
	cal.SetVersion("2.0")
	cal.SetProductId("-//Dindoa//Dindoa ICS Generator//NL")

	missing := map[string]int{}
	for _, match := range matches {
		entry, found := g.resolve(match.Location)
		if !found && match.Location != "" {
			missing[match.Location]++
		}
		cal.AddVEvent(g.createEvent(teamName, match, entry, found))
	}

	if err := os.WriteFile(outputFile, []byte(cal.Serialize()), 0o644); err != nil {
		return missing, fmt.Errorf("write ICS file: %w", err)
	}

	return missing, nil
}

func (g *Generator) resolve(venue string) (locations.Entry, bool) {
	if g.mapping == nil || venue == "" {
		return locations.Entry{}, false
	}
	return g.mapping.Lookup(venue)
}

// createEvent creates an ICS event for a match
func (g *Generator) createEvent(teamName string, match scraper.Match, entry locations.Entry, found bool) *ics.VEvent {
	event := ics.NewEvent(g.generateUID(teamName, match))

	event.SetDtStampTime(time.Now())

	start := g.parseMatchDateTime(match)
	event.SetStartAt(start)
	event.SetEndAt(start.Add(matchDuration))

	event.SetSummary(g.formatTitle(teamName, match))
	event.SetLocation(formatLocation(match.Location, entry, found))
	event.SetDescription(g.formatDescription(match, entry, found))

	if found && entry.HasCoordinates() {
		event.SetGeo(entry.Lat, entry.Lon)
	}
	if match.Colour != "" {
		event.AddProperty(ics.ComponentPropertyCategories, match.Colour)
	}

	return event
}

// formatLocation keeps the readable venue name from the website and appends the
// mapped address. The published name is never replaced, so information from the
// website cannot be lost.
func formatLocation(venue string, entry locations.Entry, found bool) string {
	if !found {
		return venue
	}

	name := entry.Name
	if name == "" {
		name = venue
	}
	if entry.Address == "" {
		return name
	}
	return name + ", " + entry.Address
}

// generateUID creates an identifier that stays stable when a match is moved to
// a different kick-off time, so regenerating the calendar updates the existing
// event instead of adding a second one.
func (g *Generator) generateUID(teamName string, match scraper.Match) string {
	return fmt.Sprintf("%s-%s-%s@dindoa.nl",
		scraper.TeamSlug(teamName),
		match.Date.Format("2006-01-02"),
		scraper.TeamSlug(match.Opponent()),
	)
}

// formatTitle formats the event title based on home/away status
func (g *Generator) formatTitle(teamName string, match scraper.Match) string {
	if match.IsHome {
		return fmt.Sprintf("%s - %s", teamName, match.Away)
	}
	return fmt.Sprintf("%s - %s", match.Home, teamName)
}

// formatDescription creates the event description
func (g *Generator) formatDescription(match scraper.Match, entry locations.Entry, found bool) string {
	matchType := "Uitwedstrijd"
	if match.IsHome {
		matchType = "Thuiswedstrijd"
	}

	lines := []string{fmt.Sprintf("%s tegen %s", matchType, match.Opponent())}
	if match.Colour != "" {
		lines = append(lines, "Kleur: "+match.Colour)
	}
	if match.Referee != "" {
		lines = append(lines, "Scheidsrechter: "+match.Referee)
	}
	if !found && match.Location != "" {
		lines = append(lines, "Locatie niet in de adressenlijst; alleen de naam van de website is bekend.")
	}
	return strings.Join(lines, "\n")
}

// parseMatchDateTime combines date and time into a single time.Time in
// Europe/Amsterdam timezone.
func (g *Generator) parseMatchDateTime(match scraper.Match) time.Time {
	var hour, minute int
	fmt.Sscanf(match.Time, "%d:%d", &hour, &minute)

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
	return fmt.Sprintf("%s.ics", scraper.TeamSlug(teamName))
}
