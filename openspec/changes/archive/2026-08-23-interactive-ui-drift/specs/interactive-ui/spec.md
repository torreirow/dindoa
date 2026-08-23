## MODIFIED Requirements

### Requirement: Launch interactive mode by default

The system SHALL start the interactive terminal UI when invoked with the `start` command, and SHALL show usage information when invoked without arguments.

#### Scenario: Run without flags

- **WHEN** user runs "dindoa" with no arguments
- **THEN** system displays usage information and does not launch the interactive UI, so that the tool stays predictable when used from a script

#### Scenario: Run the start command

- **WHEN** user runs "dindoa start"
- **THEN** system launches the interactive UI

#### Scenario: Run with flags

- **WHEN** user runs "dindoa" with command-line flags
- **THEN** system executes in CLI mode without launching interactive UI

#### Scenario: Run with only a category

- **WHEN** user runs "dindoa --category rood" without other flags
- **THEN** system launches the interactive UI with that category pre-selected

### Requirement: Handle errors gracefully in UI

The system SHALL display clear error messages in the interactive UI when operations fail, and SHALL offer a way back to an earlier screen whenever one is available.

#### Scenario: Network error during scraping

- **WHEN** website scraping fails due to network error
- **THEN** UI displays error message explaining the failure

#### Scenario: No matches found

- **WHEN** selected team has no matches scheduled
- **THEN** UI displays message indicating no matches were found

#### Scenario: Return to previous screen on error

- **WHEN** an error occurs after a team has been chosen
- **THEN** the user can return to the team list and choose again without restarting the tool

#### Scenario: No earlier screen to return to

- **WHEN** an error occurs before the match programme has been loaded
- **THEN** UI offers only to quit, because there is no earlier screen to return to

#### Scenario: Available keys are shown

- **WHEN** the UI displays an error
- **THEN** the screen names the keys that actually work on it

## ADDED Requirements

### Requirement: Show progress and result of generation

The system SHALL display progress while fetching and processing match data, and SHALL report the result when generation finishes.

#### Scenario: Show scraping progress

- **WHEN** system is fetching match data from website
- **THEN** UI displays a progress message indicating data is being fetched

#### Scenario: Show processing progress

- **WHEN** system is processing the matches of the selected team
- **THEN** UI displays a message naming the team being processed

#### Scenario: Show completion message

- **WHEN** ICS file generation completes successfully
- **THEN** UI displays success message with filename, team and match count

#### Scenario: Report venues without an address

- **WHEN** generation completes and one or more venues were not present in the location mapping
- **THEN** UI lists those venues with the number of affected matches and where to add them

## REMOVED Requirements

### Requirement: Show progress during operations

**Reason**: Deze requirement beschreef voortgang tijdens "scraping and geocoding", met een scenario per gegeocodeerde locatie. Geocoding tijdens het genereren is verwijderd; adressen komen uit een meegeleverde lijst en er is geen stap per locatie meer waarvan voortgang te tonen is. Ook de naam van de requirement dekt de lading niet meer, want er is geen tweede langdurige bewerking naast het ophalen van de pagina. Deze opruiming had bij de vorige change moeten gebeuren, waar "geen wijziging aan de interactieve UI" als non-goal stond terwijl het verwerkingsscherm wel veranderde.

**Migration**: Vervangen door "Show progress and result of generation", die het ophalen, het verwerken van het gekozen team en het melden van locaties zonder adres beschrijft. Er is geen actie nodig van gebruikers.
