## ADDED Requirements

### Requirement: Update the user location mapping

The system SHALL provide a flag that fills the user location mapping file with venues from the current match programme that have no usable mapping entry yet.

#### Scenario: Execute update locations command

- **WHEN** user runs "dindoa --update-locations"
- **THEN** system adds entries for missing venues to the user mapping file, reports what was added and what is still missing, and does not launch the interactive UI

#### Scenario: Programme page is unavailable during update

- **WHEN** user runs "dindoa --update-locations" and the match programme page cannot be fetched
- **THEN** system outputs an error indicating the page could not be fetched, leaves the user mapping file unchanged and exits with a non-zero status code

#### Scenario: Update flag combined with another command flag

- **WHEN** user runs "dindoa --update-locations" together with "--team" or "--list-locations"
- **THEN** system outputs an error that the flags cannot be combined and exits with a non-zero status code without changing the user mapping file

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

#### Scenario: Update locations flag is documented

- **WHEN** user runs "dindoa --help"
- **THEN** the help output describes the flag for updating the user location mapping and states that it uses the network
