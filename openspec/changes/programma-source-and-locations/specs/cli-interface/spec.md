## ADDED Requirements

### Requirement: List locations with mapping status

The system SHALL provide a flag to list all venues appearing in the match programme together with their mapping status.

#### Scenario: Execute list locations command

- **WHEN** user runs "dindoa --list-locations"
- **THEN** system outputs every venue from the match programme, whether it is present in the location mapping, and the mapped address when known, without launching the interactive UI

#### Scenario: Locations are ordered by impact

- **WHEN** user runs "dindoa --list-locations"
- **THEN** system orders the output so that venues covering the most matches appear first

#### Scenario: Missing locations are actionable

- **WHEN** user runs "dindoa --list-locations" and one or more venues are absent from the mapping
- **THEN** system outputs, for those venues, a fragment the user can paste into the user mapping file, together with the path of that file

#### Scenario: Coverage summary

- **WHEN** user runs "dindoa --list-locations"
- **THEN** system outputs how many venues are mapped and what share of matches that covers

#### Scenario: Programme page is unavailable

- **WHEN** user runs "dindoa --list-locations" and the match programme page cannot be fetched
- **THEN** system outputs an error indicating the page could not be fetched and exits with a non-zero status code

## MODIFIED Requirements

### Requirement: Display help information

The system SHALL provide help text explaining available flags and usage. The usage summary SHALL list every flag that the options list also reports, so that both halves of the help stay consistent.

#### Scenario: Show help with --help flag

- **WHEN** user runs "dindoa --help"
- **THEN** system displays usage information and all available flags

#### Scenario: Show help with -h flag

- **WHEN** user runs "dindoa -h"
- **THEN** system displays usage information and all available flags

#### Scenario: New flags appear in the usage summary

- **WHEN** the system reports a flag in the options list
- **THEN** that flag also appears in the usage summary, so a flag is never documented in only one half of the help output

#### Scenario: Locations flag is documented

- **WHEN** user runs "dindoa --help"
- **THEN** the help output describes the flag for listing locations and their mapping status

### Requirement: Generate ICS for specified team

The system SHALL provide a flag to generate an ICS file for a specific team, and SHALL write a file whenever the team has matches in the published programme.

#### Scenario: Generate ICS with team flag

- **WHEN** user runs "dindoa --team j3"
- **THEN** system generates an ICS file for team "Dindoa J3" without interactive UI

#### Scenario: Team name is case-insensitive

- **WHEN** user runs "dindoa --team J3" or "dindoa --team DINDOA J3"
- **THEN** system processes team name correctly regardless of case

#### Scenario: Accept short team names

- **WHEN** user runs "dindoa --team j3"
- **THEN** system resolves the input to the team "Dindoa J3" as published in the match programme

#### Scenario: Invalid team specified

- **WHEN** user runs "dindoa --team nonexistent"
- **THEN** system outputs an error indicating the team does not appear in the match programme, lists the teams that do, and exits with a non-zero status code

#### Scenario: Team has no matches in the published programme

- **WHEN** user runs "dindoa --team j3" for a team that exists but has no matches in the published part of the programme
- **THEN** system reports that no matches were found and does not silently produce nothing

#### Scenario: Report missing locations after generating

- **WHEN** user runs "dindoa --team j3" and one or more of that team's venues are absent from the location mapping
- **THEN** system writes the ICS file, reports each missing venue with the number of affected matches, and exits with status code 0
