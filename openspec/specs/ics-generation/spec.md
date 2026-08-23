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

### Requirement: Handle timezone correctly
The system SHALL use Europe/Amsterdam timezone for all match events to handle CET/CEST automatically.

#### Scenario: Set timezone for match events
- **WHEN** generating event with date and time
- **THEN** system uses Europe/Amsterdam timezone which automatically handles CET (UTC+1) and CEST (UTC+2) transitions

#### Scenario: Parse website time as local time
- **WHEN** parsing match time from website (already in local CET/CEST)
- **THEN** system treats time as Europe/Amsterdam local time without conversion

### Requirement: Generate unique event UIDs

The system SHALL generate unique identifiers for each match event that remain stable when a match is rescheduled to a different kick-off time, so that regenerating the calendar updates existing entries instead of creating duplicates.

#### Scenario: Create unique UID

- **WHEN** generating event for a match
- **THEN** system creates UID based on team, match date and opponent, without the kick-off time

#### Scenario: Consistent UID generation

- **WHEN** the same match is processed multiple times
- **THEN** system generates the same UID each time

#### Scenario: Kick-off time changes

- **WHEN** a match keeps its date and opponent but is moved to a different kick-off time and the calendar is regenerated
- **THEN** system produces the same UID, so calendar applications update the existing event rather than adding a second one

#### Scenario: Two matches on the same date

- **WHEN** a team plays two matches against different opponents on the same date
- **THEN** system produces a distinct UID for each

### Requirement: Include event timestamps

The system SHALL include proper ICS timestamps for event creation, start and end.

#### Scenario: Set DTSTAMP

- **WHEN** generating event
- **THEN** event includes DTSTAMP with current UTC timestamp

#### Scenario: Set DTSTART

- **WHEN** generating event with match date and time
- **THEN** event includes DTSTART with match datetime in Europe/Amsterdam timezone

#### Scenario: Set DTEND

- **WHEN** generating event with match date and time
- **THEN** event includes DTEND one hour after DTSTART, so that calendar applications show the match as occupying time rather than as a zero-length moment

### Requirement: Include calendar metadata
The system SHALL include proper ICS calendar metadata in the VCALENDAR component.

#### Scenario: Set calendar properties
- **WHEN** generating ICS file
- **THEN** VCALENDAR includes VERSION:2.0 and PRODID identifying the Dindoa ICS Generator

### Requirement: Include venue name and address in events

The system SHALL use the LOCATION field to carry both the venue name as published on the website and the mapped address, and SHALL NOT replace the published venue name with a looked-up address.

#### Scenario: Known location

- **WHEN** a location is present in the location mapping
- **THEN** event LOCATION field contains the readable venue name followed by the address (e.g., "De Zanderij (Dindoa), Watervalweg 170, 3853 PT Ermelo")

#### Scenario: Known location without a house number

- **WHEN** a mapped address is only precise to street level
- **THEN** event LOCATION field still contains the readable venue name and the street-level address, and navigation remains possible through the coordinates

#### Scenario: Unknown location

- **WHEN** a location is not present in the location mapping
- **THEN** event LOCATION field contains the original venue name from the website unchanged

### Requirement: Include coordinates in events

The system SHALL include the geographic coordinates of a match location in the event when those coordinates are known, so that calendar applications can offer map display and route planning.

#### Scenario: Location with coordinates

- **WHEN** the location mapping provides coordinates for a match location
- **THEN** the event includes a GEO property with those coordinates

#### Scenario: Location without coordinates

- **WHEN** the location mapping provides an address but no coordinates, or the location is unknown
- **THEN** the event omits the GEO property and remains valid

### Requirement: Include team category in events

The system SHALL record the colour category of the team in the event, so that the information published alongside the match schedule is not lost.

#### Scenario: Team with a colour

- **WHEN** generating events for a team whose matches carry a colour (e.g., Rood)
- **THEN** each event includes that colour as a category

#### Scenario: Team without a colour

- **WHEN** generating events for a senior, midweek or under-age team whose matches carry no colour
- **THEN** the events omit the category and remain valid

