package ui

import (
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
	categories []scraper.Category
	teams      []scraper.Team
	selected   int

	selectedCategory string
	selectedTeam     string
	outputFile       string
	matchCount       int

	fetcher *scraper.Fetcher
	parser  *scraper.Parser
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return fetchCategories(m.fetcher, m.parser)
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

	case categoriesMsg:
		m.categories = msg.categories
		if msg.err != nil {
			m.err = msg.err
			m.state = stateError
		} else {
			m.state = stateCategorySelection
			m.selected = 0
		}

	case teamsMsg:
		m.teams = msg.teams
		m.state = stateTeamSelection
		m.selected = 0

	case doneMsg:
		m.outputFile = msg.outputFile
		m.matchCount = msg.matchCount
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
		return "Loading categories...\n"

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
			return m, generateICS(m.selectedTeam)
		}

	case stateDone, stateError:
		return m, tea.Quit
	}

	return m, nil
}

// NewInteractiveApp creates a new Bubbletea application
func NewInteractiveApp() *tea.Program {
	m := model{
		state:   stateLoadingCategories,
		fetcher: scraper.NewFetcher(),
		parser:  scraper.NewParser(),
	}

	return tea.NewProgram(m)
}
