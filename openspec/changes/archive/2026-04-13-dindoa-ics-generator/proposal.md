## Why

Dindoa korfbal team members need an easy way to import their match schedules into their personal calendars. Currently, they must manually check the website and add matches one by one. This tool automates the process by scraping match data from dindoa.nl and generating standard ICS calendar files that can be imported into any calendar application.

## What Changes

- New CLI tool `dindoa` written in Go that scrapes match schedules from dindoa.nl
- Interactive terminal UI for selecting team categories and specific teams
- CLI flag-based mode for scripting and automation
- ICS file generation with properly formatted match events (correct titles, geocoded locations, timezone handling)
- OpenStreetMap-based geocoding to convert venue names to full addresses
- Cross-platform JSON cache for geocoded locations to minimize API calls
- Support for all Dindoa team categories (Senioren, Wedstrijdsport, youth divisions)

## Capabilities

### New Capabilities

- `team-scraping`: Scrape team categories and match schedules from dindoa.nl HTML pages
- `geocoding`: Convert venue names to full addresses using OpenStreetMap Nominatim API with rate limiting and caching
- `ics-generation`: Generate ICS calendar files with matches, including proper event formatting and timezone handling
- `interactive-ui`: Terminal-based user interface for team selection using Bubbletea
- `cli-interface`: Command-line flags for listing teams/categories and generating ICS files

### Modified Capabilities

<!-- No existing capabilities are being modified -->

## Impact

- New Go project with dependencies: bubbletea, bubbles, goquery, golang-ical, adrg/xdg
- Creates cache directory at platform-specific location (~/.cache/dindoa/ on Linux)
- Makes HTTP requests to dindoa.nl for scraping team data
- Makes HTTP requests to OpenStreetMap Nominatim API for geocoding (rate-limited to 1 req/sec)
- Produces ICS files as output (default: dindoa-{team}.ics)
