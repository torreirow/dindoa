# team-scraping Specification

## Purpose
TBD - created by archiving change dindoa-ics-generator. Update Purpose after archive.
## Requirements
### Requirement: Scrape team categories from dindoa.nl

The system SHALL fetch and parse the match programme page at https://dindoa.nl/ws/competitie-programma/ to extract available categories from the colour column.

#### Scenario: Successfully fetch categories

- **WHEN** the system requests the match programme page
- **THEN** system extracts all category names from the colour column (Rood, Oranje, Geel, Groen, Blauw)

#### Scenario: Teams without a colour

- **WHEN** a match row has an empty colour column (senior, midweek and under-age teams)
- **THEN** system assigns those teams to a category that is not a colour rather than discarding them

#### Scenario: Categories page is unavailable

- **WHEN** the match programme page returns HTTP error or network failure
- **THEN** system returns error with message indicating the page could not be fetched

#### Scenario: Colour column is missing

- **WHEN** the match programme page loads but the expected column headers are not present
- **THEN** system returns an error identifying the unexpected table layout rather than returning an empty category list

### Requirement: Scrape team list per category

The system SHALL extract Dindoa team names from the match programme page, organized by the category derived from the colour column.

#### Scenario: Successfully extract teams for category

- **WHEN** the system parses the match programme page for a specific category
- **THEN** system extracts all Dindoa team display names in that category (e.g., "Dindoa J3", "Dindoa J4" for Rood)

#### Scenario: Team appears in multiple rows

- **WHEN** a team plays several matches in the published programme
- **THEN** system lists that team once

#### Scenario: Category has no teams

- **WHEN** no match row in the programme carries a given category
- **THEN** system returns empty team list for that category

#### Scenario: Only Dindoa teams are listed

- **WHEN** the programme contains opponent teams alongside Dindoa teams
- **THEN** system lists only the Dindoa teams

### Requirement: Normalize team names to URL slugs

The system SHALL convert user-provided team names to the display name used in the match programme, and SHALL match that display name exactly against the programme's team columns. The system SHALL also derive a slug form for use as a default output filename.

#### Scenario: Short team name normalization

- **WHEN** user provides "j3" as team name
- **THEN** system resolves it to the display name "Dindoa J3"

#### Scenario: Full team name normalization

- **WHEN** user provides "Dindoa J3" or "dindoa j3" as team name
- **THEN** system resolves it to the display name "Dindoa J3"

#### Scenario: Case-insensitive handling

- **WHEN** user provides team name in any case (e.g., "J3", "j3", "DINDOA J3")
- **THEN** system resolves it to the display name "Dindoa J3"

#### Scenario: Exact matching prevents cross-team leakage

- **WHEN** the user selects "Dindoa J1" and the programme also contains "Dindoa J10" through "Dindoa J19"
- **THEN** system returns only the matches of "Dindoa J1"

#### Scenario: Opponent with the same team code is excluded

- **WHEN** the user selects "Dindoa J4" and the programme contains opponents named "Revival J4" and "Unitas/Perspectief J4"
- **THEN** system returns only the matches in which "Dindoa J4" itself plays

#### Scenario: Senior team is distinct from a junior team

- **WHEN** the user selects "Dindoa 4"
- **THEN** system returns the matches of the senior team and not those of "Dindoa J4"

#### Scenario: Slug for the default filename

- **WHEN** the system needs a default output filename for team "Dindoa J3"
- **THEN** system derives the slug "dindoa-j3"

### Requirement: Scrape match schedule from the match programme

The system SHALL parse the match programme page to extract all matches for a given Dindoa team, taking the date from the heading above each table and the remaining fields from the table row.

#### Scenario: Successfully scrape matches

- **WHEN** the system fetches the match programme page and selects a team
- **THEN** system extracts all matches for that team with date, time, home team, away team, colour and location

#### Scenario: Date comes from the heading above the table

- **WHEN** the system reads a match row
- **THEN** system uses the date from the most recent preceding date heading, because the row itself contains no date

#### Scenario: Parse date and time correctly

- **WHEN** the system extracts a date heading such as "5 september" and a row time such as "11:50"
- **THEN** system parses the Dutch month name and the "HH:MM" time correctly

#### Scenario: Derive the year for a date without one

- **WHEN** a date heading contains a day and month but no year
- **THEN** system derives the season year so that dates from August through December fall in the season's opening year and dates from January through July fall in the following year

#### Scenario: Validate the derived year against the weekday

- **WHEN** the system has derived a year for the published dates
- **THEN** system verifies the dates fall on Saturdays and Wednesdays, and reports that the year could not be established when they do not

#### Scenario: Identify home vs away matches

- **WHEN** the selected team appears in the home column
- **THEN** system marks the match as home match
- **WHEN** the selected team appears in the away column
- **THEN** system marks the match as away match

#### Scenario: Match between two Dindoa teams

- **WHEN** both the home and away column contain a Dindoa team
- **THEN** system determines home or away by comparing against the selected team's full name

#### Scenario: Programme page is unavailable

- **WHEN** the match programme page returns HTTP error or network failure
- **THEN** system returns error with message indicating the page could not be fetched

#### Scenario: Team has no matches

- **WHEN** the programme loads successfully but contains no rows for the selected team
- **THEN** system reports that the team has no matches in the published programme and lists the teams that do

#### Scenario: Unexpected table layout

- **WHEN** the programme page loads but the table headers do not match the expected columns
- **THEN** system returns an error identifying the unexpected layout rather than returning an empty match list

