## ADDED Requirements

### Requirement: Launch interactive mode by default
The system SHALL start interactive terminal UI when invoked without command-line flags.

#### Scenario: Run without flags
- **WHEN** user runs "dindoa" with no arguments
- **THEN** system launches Bubbletea interactive UI

#### Scenario: Run with flags
- **WHEN** user runs "dindoa" with command-line flags
- **THEN** system executes in CLI mode without launching interactive UI

### Requirement: Display category selection screen
The system SHALL present a list of available team categories for user selection.

#### Scenario: Show category list
- **WHEN** interactive UI starts
- **THEN** system displays list of all categories (Senioren, Wedstrijdsport, Rood, Oranje, Geel, Groen, Blauw, etc.)

#### Scenario: Navigate categories
- **WHEN** user presses up/down arrow keys
- **THEN** selection cursor moves through category list

#### Scenario: Select category
- **WHEN** user presses Enter on a category
- **THEN** system proceeds to team selection screen for that category

### Requirement: Display team selection screen
The system SHALL present a list of teams within the selected category for user selection.

#### Scenario: Show teams for category
- **WHEN** user selects a category
- **THEN** system displays list of all teams in that category

#### Scenario: Navigate teams
- **WHEN** user presses up/down arrow keys on team selection screen
- **THEN** selection cursor moves through team list

#### Scenario: Select team
- **WHEN** user presses Enter on a team
- **THEN** system proceeds to generate ICS file for that team

### Requirement: Show progress during operations
The system SHALL display progress indicators while scraping and geocoding.

#### Scenario: Show scraping progress
- **WHEN** system is fetching match data from website
- **THEN** UI displays spinner or progress message indicating data is being scraped

#### Scenario: Show geocoding progress
- **WHEN** system is geocoding locations
- **THEN** UI displays progress for each location being geocoded

#### Scenario: Show completion message
- **WHEN** ICS file generation completes successfully
- **THEN** UI displays success message with filename and match count

### Requirement: Handle errors gracefully in UI
The system SHALL display clear error messages in the interactive UI when operations fail.

#### Scenario: Network error during scraping
- **WHEN** website scraping fails due to network error
- **THEN** UI displays error message explaining the failure

#### Scenario: No matches found
- **WHEN** selected team has no matches scheduled
- **THEN** UI displays message indicating no matches were found

#### Scenario: Return to previous screen on error
- **WHEN** an error occurs during team selection
- **THEN** user can return to category selection to try again

### Requirement: Support UI navigation controls
The system SHALL respond to standard keyboard controls for navigation and exit.

#### Scenario: Exit interactive mode
- **WHEN** user presses Ctrl+C or Esc
- **THEN** system exits cleanly without error

#### Scenario: Navigate with arrow keys
- **WHEN** user presses up/down arrows
- **THEN** selection cursor moves accordingly in lists
