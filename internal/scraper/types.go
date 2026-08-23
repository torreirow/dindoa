package scraper

import (
	"regexp"
	"strings"
	"time"
)

// CategoryOther is the category for teams whose colour column is empty:
// senior, midweek and under-age teams.
const CategoryOther = "Senioren/Wedstrijdsport"

// Category represents a team category (e.g., Rood, Oranje, Senioren/Wedstrijdsport)
type Category struct {
	Name  string
	Teams []Team
}

// Team represents a Dindoa team as published in the match programme
type Team struct {
	Name     string // Display name as published (e.g., "Dindoa J3")
	Slug     string // Slug form, used for default output filenames (e.g., "dindoa-j3")
	Category string // Colour or CategoryOther
}

// Match represents a single match from the programme
type Match struct {
	Date     time.Time
	Time     string // HH:MM format
	Home     string // Home team name as published
	Away     string // Away team name as published
	Colour   string // Colour column; empty for senior/midweek/under-age teams
	Location string // Venue string as published, e.g. "De Zanderij (Dindoa) ERMELO"
	Referee  string // Referee column; often empty
	IsHome   bool   // True if the selected team is the home team
}

// Opponent returns the other team, given the selected team's display name.
func (m Match) Opponent() string {
	if m.IsHome {
		return m.Away
	}
	return m.Home
}

var (
	multiSpace  = regexp.MustCompile(`\s+`)
	teamSuffix  = regexp.MustCompile(`\s+(J\d+|U\d+-\d+|MW\d+|\d+)$`)
	slugStripRe = regexp.MustCompile(`[^a-z0-9]+`)
)

// NormalizeTeamInput converts user input to the display name used in the
// match programme.
//
//	"j3"        -> "Dindoa J3"
//	"J3"        -> "Dindoa J3"
//	"dindoa j3" -> "Dindoa J3"
//	"Dindoa J3" -> "Dindoa J3"
//	"4"         -> "Dindoa 4"
//	"u15-1"     -> "Dindoa U15-1"
//
// The result is matched exactly against the programme's team columns, so a
// caller must never use substring matching on top of this.
func NormalizeTeamInput(input string) string {
	s := multiSpace.ReplaceAllString(strings.TrimSpace(input), " ")
	if s == "" {
		return ""
	}

	// Drop a leading "dindoa" in any case; what remains is the team code.
	if lower := strings.ToLower(s); strings.HasPrefix(lower, "dindoa") {
		s = strings.TrimSpace(s[len("dindoa"):])
		s = strings.TrimLeft(s, "-_ ")
	}
	if s == "" {
		return ""
	}

	return "Dindoa " + strings.ToUpper(s)
}

// TeamSlug converts a team display name to a filename-safe slug.
//
//	"Dindoa J3"    -> "dindoa-j3"
//	"Dindoa U15-1" -> "dindoa-u15-1"
func TeamSlug(displayName string) string {
	s := slugStripRe.ReplaceAllString(strings.ToLower(displayName), "-")
	return strings.Trim(s, "-")
}

// IsDindoaTeam checks if a team name refers to a Dindoa team
func IsDindoaTeam(teamName string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(teamName)), "dindoa")
}

// ClubName strips the team number from a team name, leaving the club.
//
//	"Dindoa J3"                     -> "Dindoa"
//	"Antilopen/Bloemendal Bouw J3"  -> "Antilopen/Bloemendal Bouw"
func ClubName(teamName string) string {
	return strings.TrimSpace(teamSuffix.ReplaceAllString(strings.TrimSpace(teamName), ""))
}

// CategoryOf returns the category for a colour value.
func CategoryOf(colour string) string {
	if strings.TrimSpace(colour) == "" {
		return CategoryOther
	}
	return colour
}
