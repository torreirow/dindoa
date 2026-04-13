# team-scraping Specification

## Purpose
TBD - created by archiving change dindoa-ics-generator. Update Purpose after archive.
## Requirements
### Requirement: Scrape team categories from dindoa.nl
The system SHALL fetch and parse the team categories page at https://dindoa.nl/ws/teams/ to extract available categories.

#### Scenario: Successfully fetch categories
- **WHEN** the system requests the teams page
- **THEN** system extracts all category names (Senioren, Wedstrijdsport, Rood, Oranje, Geel, Groen, Blauw, etc.)

#### Scenario: Categories page is unavailable
- **WHEN** the teams page returns HTTP error or network failure
- **THEN** system returns error with message indicating the page could not be fetched

### Requirement: Scrape team list per category
The system SHALL extract team names and their URL slugs from the teams page, organized by category.

#### Scenario: Successfully extract teams for category
- **WHEN** the system parses the teams page for a specific category
- **THEN** system extracts all team names and their corresponding slugs (e.g., "J3" → "dindoa-j3")

#### Scenario: Category has no teams
- **WHEN** a category section exists but contains no team links
- **THEN** system returns empty team list for that category

### Requirement: Scrape match schedule for team
The system SHALL fetch and parse a team's page to extract all match information from the schedule table.

#### Scenario: Successfully scrape matches
- **WHEN** the system fetches a team page (e.g., https://dindoa.nl/ws/dindoa-j3/)
- **THEN** system extracts all matches with date, time, home team, away team, and location from HTML table

#### Scenario: Parse date and time correctly
- **WHEN** the system extracts match data from table row
- **THEN** system parses "DD-MM-YYYY" format dates and "HH:MM" format times correctly

#### Scenario: Identify home vs away matches
- **WHEN** the system extracts a match where "Dindoa" appears in home team column
- **THEN** system marks the match as home match
- **WHEN** the system extracts a match where "Dindoa" appears in away team column
- **THEN** system marks the match as away match

#### Scenario: Team page is unavailable
- **WHEN** the team page returns HTTP error or network failure
- **THEN** system returns error with message indicating the team page could not be fetched

#### Scenario: Team page has no matches
- **WHEN** the team page loads successfully but contains no match table or empty table
- **THEN** system returns empty match list without error

### Requirement: Normalize team names to URL slugs
The system SHALL convert user-provided team names to URL slugs for fetching team pages.

#### Scenario: Short team name normalization
- **WHEN** user provides "j3" as team name
- **THEN** system converts to "dindoa-j3" slug

#### Scenario: Full team name normalization
- **WHEN** user provides "Dindoa J3" or "dindoa j3" as team name
- **THEN** system converts to "dindoa-j3" slug

#### Scenario: Case-insensitive handling
- **WHEN** user provides team name in any case (e.g., "J3", "j3", "DINDOA J3")
- **THEN** system converts to lowercase "dindoa-j3" slug

