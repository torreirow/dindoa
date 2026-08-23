package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/torreirow/dindoa/internal/ics"
	"github.com/torreirow/dindoa/internal/locations"
	"github.com/torreirow/dindoa/internal/scraper"
)

// Message types

type programmaMsg struct {
	programma  *scraper.Programma
	categories []scraper.Category
	err        error
}

type teamsMsg struct {
	teams []scraper.Team
}

type doneMsg struct {
	outputFile string
	matchCount int
	missing    map[string]int
	userPath   string
	err        error
}

// Commands

// fetchProgramma loads the match programme once. It is the single source for
// categories, teams and matches, so the UI never fetches twice.
func fetchProgramma(fetcher *scraper.Fetcher, parser *scraper.Parser) tea.Cmd {
	return func() tea.Msg {
		doc, err := fetcher.FetchProgrammaPage()
		if err != nil {
			return programmaMsg{err: fmt.Errorf("fetch match programme: %w", err)}
		}

		prog, err := parser.ParseProgramma(doc, time.Now())
		if err != nil {
			return programmaMsg{err: fmt.Errorf("parse match programme (%s): %w", scraper.ProgrammaURL(), err)}
		}

		return programmaMsg{programma: prog, categories: prog.Categories()}
	}
}

func generateICS(prog *scraper.Programma, teamName string) tea.Cmd {
	return func() tea.Msg {
		matches := prog.MatchesFor(teamName)
		if len(matches) == 0 {
			return doneMsg{err: fmt.Errorf("%s has no matches in the published part of the programme", teamName)}
		}

		mapping, err := locations.Load()
		if err != nil {
			return doneMsg{err: fmt.Errorf("load location mapping: %w", err)}
		}

		generator, err := ics.NewGenerator(mapping)
		if err != nil {
			return doneMsg{err: fmt.Errorf("create generator: %w", err)}
		}

		outputFile := ics.DefaultOutputFilename(teamName)
		missing, err := generator.Generate(teamName, matches, outputFile)
		if err != nil {
			return doneMsg{err: fmt.Errorf("generate ICS: %w", err)}
		}

		return doneMsg{
			outputFile: outputFile,
			matchCount: len(matches),
			missing:    missing,
			userPath:   mapping.UserPath(),
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
		b.WriteString(fmt.Sprintf("%s %s (%d teams)\n", cursor, cat.Name, len(cat.Teams)))
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
	return fmt.Sprintf("Wedstrijden verwerken voor %s...\n", m.selectedTeam)
}

func (m model) viewDone() string {
	var b strings.Builder

	b.WriteString("✓ ICS bestand aangemaakt!\n\n")
	b.WriteString(fmt.Sprintf("Bestand:     %s\n", m.outputFile))
	b.WriteString(fmt.Sprintf("Team:        %s\n", m.selectedTeam))
	b.WriteString(fmt.Sprintf("Wedstrijden: %d\n", m.matchCount))

	if len(m.missing) > 0 {
		venues := make([]string, 0, len(m.missing))
		for v := range m.missing {
			venues = append(venues, v)
		}
		sort.Slice(venues, func(i, j int) bool {
			if m.missing[venues[i]] != m.missing[venues[j]] {
				return m.missing[venues[i]] > m.missing[venues[j]]
			}
			return venues[i] < venues[j]
		})

		b.WriteString(fmt.Sprintf("\n⚠ %d locatie(s) niet in de adressenlijst; de naam van de website is gebruikt:\n",
			len(venues)))
		for _, v := range venues {
			b.WriteString(fmt.Sprintf("    %-42s (%d wedstrijd(en))\n", v, m.missing[v]))
		}
		b.WriteString(fmt.Sprintf("  Vul ze aan in %s — 'dindoa --list-locations' geeft een fragment om te plakken.\n",
			m.userPath))
	}

	b.WriteString("\nImporteer het bestand in je kalender-app.\n\n")
	b.WriteString("[enter: afsluiten]\n")

	return b.String()
}

func (m model) viewError() string {
	return fmt.Sprintf("Error: %v\n\n[enter: afsluiten]\n", m.err)
}
