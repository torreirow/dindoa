## MODIFIED Requirements

### Requirement: Include geocoded location in events

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

## ADDED Requirements

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
