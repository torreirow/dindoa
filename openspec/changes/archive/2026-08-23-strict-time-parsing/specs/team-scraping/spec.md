## MODIFIED Requirements

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

#### Scenario: Reject an unreadable kick-off time

- **WHEN** the time column holds a value that cannot be read as HH:MM, such as "13.45", "1345" or an empty cell
- **THEN** system returns an error naming the match and the value found, rather than deriving a time from it

#### Scenario: Reject a kick-off time outside the clock

- **WHEN** the time column holds hours above 23 or minutes above 59, such as "25:99"
- **THEN** system returns an error, because normalising such a value silently moves the event to another moment

#### Scenario: Accept a single-digit hour

- **WHEN** the time column holds "9:30"
- **THEN** system reads it as half past nine
