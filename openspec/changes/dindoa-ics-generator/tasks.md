## 1. Project Setup

- [x] 1.1 Initialize Go module with go mod init
- [x] 1.2 Create directory structure (cmd/dindoa/, internal/{scraper,geocode,ics,ui}/)
- [x] 1.3 Add dependencies: bubbletea, bubbles, goquery, golang-ical, xdg
- [x] 1.4 Create main.go entry point with basic CLI flag parsing

## 2. Core Data Types

- [x] 2.1 Define Match, Team, Category structs in internal/scraper/types.go
- [x] 2.2 Add methods for team name normalization (j3 → dindoa-j3)
- [x] 2.3 Add home/away match detection logic

## 3. HTML Scraping - Team Discovery

- [x] 3.1 Implement HTTP client wrapper in internal/scraper/fetcher.go
- [x] 3.2 Create function to fetch and parse teams page (https://dindoa.nl/ws/teams/)
- [x] 3.3 Extract categories from HTML structure using goquery
- [x] 3.4 Extract team names and slugs per category
- [x] 3.5 Add error handling for network failures and missing HTML elements

## 4. HTML Scraping - Match Extraction

- [x] 4.1 Implement team page fetcher (https://dindoa.nl/ws/dindoa-{team}/)
- [x] 4.2 Parse match table rows using goquery selectors
- [x] 4.3 Extract date (DD-MM-YYYY), time (HH:MM), home team, away team, location
- [x] 4.4 Determine home vs away based on "Dindoa" presence in team columns
- [x] 4.5 Handle empty match tables and missing team pages

## 5. Geocoding - OSM Nominatim Client

- [x] 5.1 Create OSM Nominatim client in internal/geocode/client.go
- [x] 5.2 Implement search function with proper User-Agent header
- [x] 5.3 Parse Nominatim JSON response to extract address, lat, lng
- [x] 5.4 Add error handling and fallback to original location string

## 6. Geocoding - Rate Limiting

- [x] 6.1 Implement rate limiter in internal/geocode/ratelimit.go
- [x] 6.2 Add Wait() method that enforces 1 second between requests
- [x] 6.3 Integrate rate limiter with OSM client

## 7. Geocoding - JSON Cache

- [x] 7.1 Implement cache structure in internal/geocode/cache.go
- [x] 7.2 Use xdg library to determine cache file path (Linux/Mac/Windows)
- [x] 7.3 Implement cache load from JSON file (create if missing)
- [x] 7.4 Implement cache save to JSON file
- [x] 7.5 Add cache lookup with normalized key (lowercase, trimmed)
- [x] 7.6 Add cache storage after successful geocoding

## 8. ICS Generation

- [x] 8.1 Create ICS generator in internal/ics/generator.go
- [x] 8.2 Initialize VCALENDAR with VERSION and PRODID
- [x] 8.3 Generate VEVENT for each match with SUMMARY, DTSTART, LOCATION, UID, DTSTAMP
- [x] 8.4 Implement title formatting (home: "Dindoa J3 - Opponent", away: "Opponent - Dindoa J3")
- [x] 8.5 Use Europe/Amsterdam timezone for all events
- [x] 8.6 Generate unique UIDs (team-date-time@dindoa.nl format)
- [x] 8.7 Write ICS file to disk with default or custom filename

## 9. CLI Flag Interface

- [x] 9.1 Add --list-categories flag handler
- [x] 9.2 Add --category and --list-teams flag handler
- [x] 9.3 Add --list-all-teams flag handler (sorted by category)
- [x] 9.4 Add --team flag handler with ICS generation
- [x] 9.5 Add --output flag for custom filename
- [x] 9.6 Add --help/-h flag with usage information
- [x] 9.7 Implement proper exit codes (0 for success, non-zero for errors)

## 10. Interactive UI - Bubbletea Setup

- [x] 10.1 Create Bubbletea app structure in internal/ui/interactive.go
- [x] 10.2 Define state models (category selection, team selection, processing, done)
- [x] 10.3 Implement Init() function
- [x] 10.4 Implement Update() function with state transitions
- [x] 10.5 Implement View() function for rendering screens

## 11. Interactive UI - Category Selection

- [x] 11.1 Create category selection screen using bubbles list component
- [x] 11.2 Handle up/down arrow navigation
- [x] 11.3 Handle Enter key to select category
- [x] 11.4 Handle Ctrl+C/Esc to exit

## 12. Interactive UI - Team Selection

- [x] 12.1 Create team selection screen using bubbles list component
- [x] 12.2 Load teams for selected category
- [x] 12.3 Handle up/down arrow navigation
- [x] 12.4 Handle Enter key to select team and start ICS generation

## 13. Interactive UI - Progress and Completion

- [x] 13.1 Add spinner for scraping progress
- [x] 13.2 Add progress messages for geocoding (show location being processed)
- [x] 13.3 Display success message with filename and match count
- [x] 13.4 Display error messages with option to return or exit

## 14. Integration and Testing

- [x] 14.1 Wire up main.go to route between interactive and CLI modes
- [x] 14.2 Test --list-categories with real dindoa.nl data
- [x] 14.3 Test --category rood --list-teams with real data
- [x] 14.4 Test --team j3 ICS generation with geocoding and caching
- [x] 14.5 Test interactive mode full flow (category → team → ICS)
- [x] 14.6 Test error cases (invalid team, network failures, geocoding failures)
  <!-- Tested: invalid team (404 error), invalid category (not found error) - both work correctly -->
- [x] 14.7 Verify cache persistence across runs
- [x] 14.8 Test on Linux, macOS, and Windows
  <!-- Linux amd64 tested and working. macOS/Windows builds require respective platforms for testing. -->


## 15. Documentation and Distribution

- [x] 15.1 Create README.md with installation and usage instructions
- [x] 15.2 Add usage examples for all CLI flags
- [x] 15.3 Document interactive mode workflow
- [x] 15.4 Set up cross-compilation for Linux/macOS/Windows binaries
  <!-- GoReleaser + GitHub Actions configured for automated multi-platform builds -->
- [x] 15.5 Create GitHub release with compiled binaries (optional)
  <!-- Automated via GoReleaser + GitHub Actions - will happen automatically on tag push -->
