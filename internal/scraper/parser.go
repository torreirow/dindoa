package scraper

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Parser handles HTML parsing logic
type Parser struct{}

// NewParser creates a new Parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseCategories extracts category names from the teams page
// The teams page has sections organized by category with h1 headers
func (p *Parser) ParseCategories(doc *goquery.Document) ([]string, error) {
	categories := []string{}
	seen := make(map[string]bool)

	// Find all h1 elements with class wp-block-heading that contain category names
	doc.Find("h1.wp-block-heading").Each(func(i int, h1 *goquery.Selection) {
		categoryName := strings.TrimSpace(h1.Text())
		if categoryName == "" || seen[categoryName] {
			return
		}

		// Check if this h1 is followed by team links
		hasTeams := false

		// Check subsequent elements for team links
		h1.NextAll().EachWithBreak(func(j int, sibling *goquery.Selection) bool {
			// Stop at next h1 (next category)
			if sibling.Is("h1") {
				return false
			}

			// Check if this element or its children contain team links
			sibling.Find("a[href*='/dindoa-']").Each(func(k int, link *goquery.Selection) {
				hasTeams = true
			})

			return !hasTeams // Continue until we find teams or hit next h1
		})

		if hasTeams {
			categories = append(categories, categoryName)
			seen[categoryName] = true
		}
	})

	if len(categories) == 0 {
		return nil, fmt.Errorf("no categories found in HTML")
	}

	return categories, nil
}

// ParseTeams extracts all teams from the teams page, organized by category
func (p *Parser) ParseTeams(doc *goquery.Document) ([]Category, error) {
	categories := []Category{}

	// Iterate through all h1.wp-block-heading category headers
	doc.Find("h1.wp-block-heading").Each(func(i int, h1 *goquery.Selection) {
		categoryName := strings.TrimSpace(h1.Text())
		if categoryName == "" {
			return
		}

		var teams []Team

		// Find all team links after this h1 until the next h1
		h1.NextAll().EachWithBreak(func(j int, sibling *goquery.Selection) bool {
			// Stop at next h1 (next category)
			if sibling.Is("h1") {
				return false
			}

			// Find team links in this element and its children
			sibling.Find("a[href*='/dindoa-']").Each(func(k int, link *goquery.Selection) {
				href, exists := link.Attr("href")
				if !exists {
					return
				}

				// Extract team slug from URL (e.g., /ws/dindoa-j3/ -> dindoa-j3)
				parts := strings.Split(strings.Trim(href, "/"), "/")
				if len(parts) < 2 {
					return
				}
				slug := parts[len(parts)-1]

				// Get team display name from the h4 inside the link
				name := strings.TrimSpace(link.Find("h4").Text())
				if name == "" {
					// Fallback to link text
					name = strings.TrimSpace(link.Text())
				}
				if name == "" {
					// Last resort: derive from slug
					name = strings.ReplaceAll(slug, "-", " ")
					name = strings.Title(name)
				}

				teams = append(teams, Team{
					Name: name,
					Slug: slug,
				})
			})

			return true // Continue to next sibling
		})

		// Only add category if it has teams
		if len(teams) > 0 {
			categories = append(categories, Category{
				Name:  categoryName,
				Teams: teams,
			})
		}
	})

	if len(categories) == 0 {
		return nil, fmt.Errorf("no teams found in HTML")
	}

	return categories, nil
}

// ParseMatches extracts match data from a team page
func (p *Parser) ParseMatches(doc *goquery.Document) ([]Match, error) {
	matches := []Match{}

	// Find the match table - look for tables with match data
	// Typical structure: table with columns for Datum, Tijd, Thuis, Uit, Locatie
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		// Check if this looks like a match table by looking for expected column headers
		headers := table.Find("thead th, th")
		isMatchTable := false
		headers.Each(func(j int, th *goquery.Selection) {
			text := strings.ToLower(th.Text())
			if strings.Contains(text, "datum") || strings.Contains(text, "tijd") || strings.Contains(text, "thuis") {
				isMatchTable = true
			}
		})

		if !isMatchTable {
			return
		}

		// Parse each row in the table body
		table.Find("tbody tr, tr").Each(func(rowIdx int, row *goquery.Selection) {
			cells := row.Find("td")
			if cells.Length() < 5 {
				return // Skip rows without enough columns
			}

			// Extract data from columns
			// Expected order: Datum, Tijd, Thuis, Uit, Locatie
			dateStr := strings.TrimSpace(cells.Eq(0).Text())
			timeStr := strings.TrimSpace(cells.Eq(1).Text())
			home := strings.TrimSpace(cells.Eq(2).Text())
			away := strings.TrimSpace(cells.Eq(3).Text())
			location := strings.TrimSpace(cells.Eq(4).Text())

			// Skip empty rows
			if dateStr == "" || home == "" || away == "" {
				return
			}

			// Parse date (DD-MM-YYYY format)
			date, err := time.Parse("02-01-2006", dateStr)
			if err != nil {
				// Try alternative format if needed
				return
			}

			match := Match{
				Date:     date,
				Time:     timeStr,
				Home:     home,
				Away:     away,
				Location: location,
			}

			// Determine if this is a home or away match
			match.DetermineHomeAway()

			matches = append(matches, match)
		})
	})

	// Return empty slice instead of error if no matches found
	// (team might not have scheduled matches yet)
	return matches, nil
}
