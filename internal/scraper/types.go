package scraper

import (
	"strings"
	"time"
)

// Category represents a team category (e.g., Senioren, Rood, Oranje)
type Category struct {
	Name  string
	Teams []Team
}

// Team represents a Dindoa team
type Team struct {
	Name string // Display name (e.g., "Dindoa J3")
	Slug string // URL slug (e.g., "dindoa-j3")
}

// Match represents a single match
type Match struct {
	Date     time.Time
	Time     string // HH:MM format
	Home     string // Home team name
	Away     string // Away team name
	Location string // Venue name or address
	IsHome   bool   // True if Dindoa is home team
}

// NormalizeTeamName converts user input to URL slug format
// Examples:
//   "j3" -> "dindoa-j3"
//   "J3" -> "dindoa-j3"
//   "Dindoa J3" -> "dindoa-j3"
//   "dindoa j3" -> "dindoa-j3"
func NormalizeTeamName(input string) string {
	// Convert to lowercase and trim whitespace
	slug := strings.ToLower(strings.TrimSpace(input))

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Add "dindoa-" prefix if not present
	if !strings.HasPrefix(slug, "dindoa-") {
		slug = "dindoa-" + slug
	}

	return slug
}

// IsDindoaTeam checks if a team name contains "Dindoa"
func IsDindoaTeam(teamName string) bool {
	return strings.Contains(strings.ToLower(teamName), "dindoa")
}

// DetermineHomeAway sets the IsHome field based on which column contains "Dindoa"
func (m *Match) DetermineHomeAway() {
	m.IsHome = IsDindoaTeam(m.Home)
}
