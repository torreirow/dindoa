# ics-generation Specification

## Purpose
TBD - created by archiving change dindoa-ics-generator. Update Purpose after archive.
## Requirements
### Requirement: Generate ICS file with match events
The system SHALL create a valid ICS (iCalendar) file containing all matches for the selected team.

#### Scenario: Create ICS file with multiple matches
- **WHEN** the system has fetched matches for a team
- **THEN** system generates ICS file with VCALENDAR and VEVENT entries for each match

#### Scenario: Default output filename
- **WHEN** no custom output filename is specified
- **THEN** system creates file named "dindoa-{team}.ics" (e.g., "dindoa-j3.ics")

#### Scenario: Custom output filename
- **WHEN** user specifies custom output filename
- **THEN** system creates file with the specified name

### Requirement: Format event titles based on home/away status
The system SHALL format match event titles with team names in the correct order based on home or away match.

#### Scenario: Home match title
- **WHEN** generating event for home match (Dindoa team is home)
- **THEN** event SUMMARY is "{Dindoa team} - {opponent}" (e.g., "Dindoa J3 - ASVD J1")

#### Scenario: Away match title
- **WHEN** generating event for away match (Dindoa team is away)
- **THEN** event SUMMARY is "{opponent} - {Dindoa team}" (e.g., "ASVD J1 - Dindoa J3")

### Requirement: Include geocoded location in events
The system SHALL use geocoded full address as the event LOCATION field.

#### Scenario: Use geocoded address
- **WHEN** location has been successfully geocoded
- **THEN** event LOCATION field contains full address from geocoding

#### Scenario: Use original location on geocoding failure
- **WHEN** geocoding fails for a location
- **THEN** event LOCATION field contains original venue name from website

### Requirement: Handle timezone correctly
The system SHALL use Europe/Amsterdam timezone for all match events to handle CET/CEST automatically.

#### Scenario: Set timezone for match events
- **WHEN** generating event with date and time
- **THEN** system uses Europe/Amsterdam timezone which automatically handles CET (UTC+1) and CEST (UTC+2) transitions

#### Scenario: Parse website time as local time
- **WHEN** parsing match time from website (already in local CET/CEST)
- **THEN** system treats time as Europe/Amsterdam local time without conversion

### Requirement: Generate unique event UIDs
The system SHALL generate unique identifiers for each match event to prevent duplicates.

#### Scenario: Create unique UID
- **WHEN** generating event for a match
- **THEN** system creates UID based on team slug, date, and time (e.g., "dindoa-j3-2026-05-09-1200@dindoa.nl")

#### Scenario: Consistent UID generation
- **WHEN** the same match is processed multiple times
- **THEN** system generates the same UID each time

### Requirement: Include event timestamps
The system SHALL include proper ICS timestamps for event creation and scheduling.

#### Scenario: Set DTSTAMP
- **WHEN** generating event
- **THEN** event includes DTSTAMP with current UTC timestamp

#### Scenario: Set DTSTART
- **WHEN** generating event with match date and time
- **THEN** event includes DTSTART with match datetime in Europe/Amsterdam timezone

### Requirement: Include calendar metadata
The system SHALL include proper ICS calendar metadata in the VCALENDAR component.

#### Scenario: Set calendar properties
- **WHEN** generating ICS file
- **THEN** VCALENDAR includes VERSION:2.0 and PRODID identifying the Dindoa ICS Generator

