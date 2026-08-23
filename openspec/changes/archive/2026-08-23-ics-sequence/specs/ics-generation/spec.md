## ADDED Requirements

### Requirement: Include a revision number in events

The system SHALL include a SEQUENCE property in every event, whose value is higher for each newly generated file, so that calendar applications tracking revisions recognise a regenerated calendar as newer than the one they already hold.

#### Scenario: Event carries a revision number

- **WHEN** generating an event
- **THEN** the event includes a SEQUENCE property with a non-negative integer value

#### Scenario: A later generation is recognised as newer

- **WHEN** the same team's calendar is generated again at a later moment
- **THEN** the SEQUENCE value in the new file is higher than the value in the earlier file

#### Scenario: One revision number per generated file

- **WHEN** generating a file containing several matches
- **THEN** every event in that file carries the same SEQUENCE value, because the value identifies the edition of the calendar rather than the individual match

#### Scenario: Value stays within the permitted range

- **WHEN** generating an event
- **THEN** the SEQUENCE value fits in a 32-bit signed integer, as required for this property

#### Scenario: Regenerating without changes remains safe

- **WHEN** the calendar is generated twice with unchanged match data and both files are imported
- **THEN** the calendar contains one entry per match, updated rather than duplicated, because the UID is unchanged
