## ADDED Requirements

### Requirement: Geocode venue names using OpenStreetMap Nominatim
The system SHALL convert venue names to full addresses using the OpenStreetMap Nominatim API.

#### Scenario: Successfully geocode location
- **WHEN** the system requests geocoding for "Veld ASVD DRONTEN"
- **THEN** system queries Nominatim API and returns full address with street, city, and postal code

#### Scenario: Geocoding fails or no results found
- **WHEN** the Nominatim API returns no results or an error for a location
- **THEN** system falls back to using the original venue name as the location string

#### Scenario: Network error during geocoding
- **WHEN** the Nominatim API request fails due to network error
- **THEN** system logs warning and falls back to original venue name without stopping ICS generation

### Requirement: Rate limit geocoding requests
The system SHALL enforce rate limiting to comply with OpenStreetMap Nominatim usage policy of maximum 1 request per second.

#### Scenario: Sequential geocoding requests
- **WHEN** the system needs to geocode multiple locations
- **THEN** system waits at least 1 second between consecutive Nominatim API requests

#### Scenario: Respect rate limit with cache hits
- **WHEN** the system geocodes a location and finds it in cache
- **THEN** system returns cached result immediately without API call or rate limit delay

### Requirement: Cache geocoding results
The system SHALL cache geocoding results in a JSON file at a cross-platform location to minimize API calls.

#### Scenario: Cache location on Linux
- **WHEN** the system runs on Linux
- **THEN** cache file is stored at ~/.cache/dindoa/geocode.json

#### Scenario: Cache location on macOS
- **WHEN** the system runs on macOS
- **THEN** cache file is stored at ~/Library/Caches/dindoa/geocode.json

#### Scenario: Cache location on Windows
- **WHEN** the system runs on Windows
- **THEN** cache file is stored at %LOCALAPPDATA%\dindoa\cache\geocode.json

#### Scenario: Use cached geocoding result
- **WHEN** the system needs to geocode a location that exists in cache
- **THEN** system returns the cached address without making API call

#### Scenario: Store new geocoding result in cache
- **WHEN** the system successfully geocodes a new location
- **THEN** system stores the result in cache with normalized location name as key

#### Scenario: Normalize cache keys
- **WHEN** the system stores or looks up a location in cache
- **THEN** system uses lowercase, trimmed location string as cache key (e.g., "Veld ASVD DRONTEN" → "veld asvd dronten")

### Requirement: Geocode all match locations
The system SHALL geocode all match locations including home and away venues.

#### Scenario: Geocode home match location
- **WHEN** processing a home match at "De Zanderij (Dindoa) ERMELO"
- **THEN** system geocodes the location to full address

#### Scenario: Geocode away match location
- **WHEN** processing an away match at "Veld ASVD DRONTEN"
- **THEN** system geocodes the location to full address

### Requirement: Include User-Agent in API requests
The system SHALL include a proper User-Agent header in all Nominatim API requests per OSM usage policy.

#### Scenario: Set User-Agent header
- **WHEN** the system makes a Nominatim API request
- **THEN** request includes User-Agent header identifying the application
