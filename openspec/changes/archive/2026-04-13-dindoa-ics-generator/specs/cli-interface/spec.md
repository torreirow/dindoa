## ADDED Requirements

### Requirement: List all categories
The system SHALL provide a flag to list all available team categories.

#### Scenario: Execute list categories command
- **WHEN** user runs "dindoa --list-categories"
- **THEN** system outputs all category names, one per line, without launching interactive UI

### Requirement: List teams for a category
The system SHALL provide flags to list teams within a specific category.

#### Scenario: Execute list teams for category
- **WHEN** user runs "dindoa --category rood --list-teams"
- **THEN** system outputs all team names in the Rood category, one per line

#### Scenario: Category flag is case-insensitive
- **WHEN** user runs "dindoa --category ROOD --list-teams" or "dindoa --category Rood --list-teams"
- **THEN** system processes the category correctly regardless of case

#### Scenario: Invalid category specified
- **WHEN** user runs "dindoa --category invalid --list-teams"
- **THEN** system outputs error message indicating category does not exist

### Requirement: List all teams sorted by category
The system SHALL provide a flag to list all teams grouped and sorted by category.

#### Scenario: Execute list all teams
- **WHEN** user runs "dindoa --list-all-teams"
- **THEN** system outputs all teams organized by category with category headers

### Requirement: Generate ICS for specified team
The system SHALL provide a flag to generate ICS file for a specific team.

#### Scenario: Generate ICS with team flag
- **WHEN** user runs "dindoa --team j3"
- **THEN** system generates ICS file for team "dindoa-j3" without interactive UI

#### Scenario: Team name is case-insensitive
- **WHEN** user runs "dindoa --team J3" or "dindoa --team DINDOA J3"
- **THEN** system processes team name correctly regardless of case

#### Scenario: Accept short team names
- **WHEN** user runs "dindoa --team j3"
- **THEN** system automatically converts to "dindoa-j3" slug

#### Scenario: Invalid team specified
- **WHEN** user runs "dindoa --team nonexistent"
- **THEN** system outputs error message indicating team does not exist or page could not be fetched

### Requirement: Specify custom output filename
The system SHALL provide a flag to specify a custom output filename for the ICS file.

#### Scenario: Generate with custom output filename
- **WHEN** user runs "dindoa --team j3 --output custom.ics"
- **THEN** system generates ICS file with filename "custom.ics" instead of default

#### Scenario: Output flag requires team flag
- **WHEN** user runs "dindoa --output custom.ics" without --team flag
- **THEN** system outputs error indicating --team flag is required

### Requirement: Combine category filter with interactive mode
The system SHALL allow pre-selecting a category when launching interactive mode.

#### Scenario: Launch interactive mode with category filter
- **WHEN** user runs "dindoa --category rood" without other flags
- **THEN** system launches interactive UI with Rood category pre-selected and shows team list directly

### Requirement: Display help information
The system SHALL provide help text explaining available flags and usage.

#### Scenario: Show help with --help flag
- **WHEN** user runs "dindoa --help"
- **THEN** system displays usage information and all available flags

#### Scenario: Show help with -h flag
- **WHEN** user runs "dindoa -h"
- **THEN** system displays usage information and all available flags

### Requirement: Exit with appropriate status codes
The system SHALL exit with appropriate status codes for success and failure cases.

#### Scenario: Successful execution
- **WHEN** command completes successfully
- **THEN** system exits with status code 0

#### Scenario: Error during execution
- **WHEN** command fails (invalid team, network error, etc.)
- **THEN** system exits with non-zero status code
