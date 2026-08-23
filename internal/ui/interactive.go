package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/torreirow/dindoa/internal/scraper"
)

// state represents the current screen/state of the UI
type state int

const (
	stateLoadingCategories state = iota
	stateCategorySelection
	stateTeamSelection
	stateProcessing
	stateDone
	stateError
)

// model holds the application state
type model struct {
	state      state
	err        error
	programma  *scraper.Programma
	categories []scraper.Category
	teams      []scraper.Team
	selected   int

	// preselect is the category given on the command line, if any. When it
	// matches, the category screen is skipped and the team list is shown.
	preselect string
	// notice explains why a preselected category was not used.
	notice string

	selectedCategory string
	selectedTeam     string
	outputFile       string
	matchCount       int
	missing          map[string]int
	userPath         string

	fetcher *scraper.Fetcher
	parser  *scraper.Parser
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return fetchProgramma(m.fetcher, m.parser)
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}

		case "down", "j":
			switch m.state {
			case stateCategorySelection:
				if m.selected < len(m.categories)-1 {
					m.selected++
				}
			case stateTeamSelection:
				if m.selected < len(m.teams)-1 {
					m.selected++
				}
			}

		case "enter":
			return m.handleEnter()
		}

	case programmaMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateError
			break
		}
		m.programma = msg.programma
		m.categories = msg.categories
		m.selected = 0
		m.state = stateCategorySelection

		if m.preselect != "" {
			if cat, ok := findCategory(m.categories, m.preselect); ok {
				m.selectedCategory = cat.Name
				m.teams = cat.Teams
				m.state = stateTeamSelection
			} else {
				m.notice = fmt.Sprintf("Categorie %q bestaat niet; kies hieronder.", m.preselect)
			}
		}

	case teamsMsg:
		m.teams = msg.teams
		m.state = stateTeamSelection
		m.selected = 0

	case doneMsg:
		m.outputFile = msg.outputFile
		m.matchCount = msg.matchCount
		m.missing = msg.missing
		m.userPath = msg.userPath
		if msg.err != nil {
			m.err = msg.err
			m.state = stateError
		} else {
			m.state = stateDone
		}
	}

	return m, nil
}

// View renders the UI
func (m model) View() string {
	switch m.state {
	case stateLoadingCategories:
		return "Wedstrijdprogramma ophalen...\n"

	case stateCategorySelection:
		return m.viewCategorySelection()

	case stateTeamSelection:
		return m.viewTeamSelection()

	case stateProcessing:
		return m.viewProcessing()

	case stateDone:
		return m.viewDone()

	case stateError:
		return m.viewError()
	}

	return ""
}

// handleEnter processes the Enter key based on current state
func (m model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateCategorySelection:
		if m.selected < len(m.categories) {
			m.selectedCategory = m.categories[m.selected].Name
			m.teams = m.categories[m.selected].Teams
			m.state = stateTeamSelection
			m.selected = 0
		}

	case stateTeamSelection:
		if m.selected < len(m.teams) {
			m.selectedTeam = m.teams[m.selected].Name
			m.state = stateProcessing
			return m, generateICS(m.programma, m.selectedTeam)
		}

	case stateDone, stateError:
		return m, tea.Quit
	}

	return m, nil
}

// findCategory looks up a category by name, ignoring case.
func findCategory(cats []scraper.Category, name string) (scraper.Category, bool) {
	want := strings.ToLower(strings.TrimSpace(name))
	for _, c := range cats {
		if strings.ToLower(c.Name) == want {
			return c, true
		}
	}
	return scraper.Category{}, false
}

// NewInteractiveApp creates a new Bubbletea application. When preselect names
// a category, the category screen is skipped and its teams are shown directly.
func NewInteractiveApp(preselect string) *tea.Program {
	m := model{
		state:     stateLoadingCategories,
		preselect: preselect,
		fetcher:   scraper.NewFetcher(),
		parser:    scraper.NewParser(),
	}

	return tea.NewProgram(m)
}
