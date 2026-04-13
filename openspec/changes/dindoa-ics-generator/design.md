## Context

This is a new Go CLI tool for the Dindoa korfbal club that scrapes match schedules from their WordPress website and generates ICS calendar files. The website doesn't provide an API, so we must parse HTML. The tool needs to work cross-platform (Linux, macOS, Windows) and support both interactive and scripted usage.

Key constraints:
- HTML structure may change (WordPress updates)
- OpenStreetMap Nominatim has strict rate limits (1 req/sec)
- Users expect simple invocation without complex setup
- Binary should be self-contained (no runtime dependencies)

## Goals / Non-Goals

**Goals:**
- Robust HTML scraping that handles common website changes gracefully
- Fast operation through aggressive caching of geocoding results
- Excellent UX in both interactive (Bubbletea) and CLI flag modes
- Correct ICS formatting with proper timezone handling
- Cross-platform cache storage following OS conventions

**Non-Goals:**
- Monitoring website changes or auto-updating ICS files (users re-run at season start)
- Web scraping for other Dutch korfbal clubs
- Calendar sync/update features (one-time import model)
- GUI or web interface (CLI-only tool)

## Decisions

### 1. HTML Scraping with goquery

**Decision**: Use github.com/PuerkitoBio/goquery for HTML parsing

**Rationale**:
- jQuery-like selectors are readable and maintainable
- Handles malformed HTML gracefully
- Standard library in Go ecosystem
- Better than regex parsing for structure extraction

**Alternative considered**:
- golang.org/x/net/html (lower level, more complex)
- regex (too brittle for HTML)

### 2. Bubbletea for Interactive UI

**Decision**: Use github.com/charmbracelet/bubbletea + bubbles for TUI

**Rationale**:
- Modern, well-maintained framework
- Elm architecture makes state management clear
- Built-in components (list, spinner) match our needs
- Excellent cross-platform terminal support

**Alternative considered**:
- survey/promptui (less flexible, no progress spinners)
- Basic fmt.Scan (poor UX)

### 3. Cross-Platform Cache with xdg

**Decision**: Use github.com/adrg/xdg for cache directory resolution

**Rationale**:
- Follows OS conventions automatically (XDG on Linux, ~/Library on Mac, AppData on Windows)
- Single dependency, well-tested
- No need to manually handle platform differences

**Cache format**: JSON file at `{XDG_CACHE_HOME}/dindoa/geocode.json`

**Structure**:
```json
{
  "version": "1.0",
  "locations": {
    "normalized-key": {
      "query": "Original Query",
      "address": "Full Address",
      "lat": 52.1234,
      "lng": 5.6789,
      "cached_at": "2026-04-13T12:00:00Z"
    }
  }
}
```

**Alternative considered**:
- SQLite (overkill for simple key-value storage)
- Hardcoded paths (breaks cross-platform support)

### 4. golang-ical for ICS Generation

**Decision**: Use github.com/arran4/golang-ical

**Rationale**:
- Handles ICS format spec details (escaping, folding, timestamps)
- Active maintenance
- Simple API for creating events

**Alternative considered**:
- Manual string formatting (error-prone, hard to validate)
- Other ICS libraries (less active or more complex APIs)

### 5. Rate Limiting Strategy

**Decision**: Simple time.Sleep-based rate limiter (1 second between requests)

**Rationale**:
- OSM Nominatim policy: max 1 req/sec
- Simple implementation (no need for token bucket complexity)
- Cache hits bypass rate limiting entirely

**Implementation**:
```go
type RateLimiter struct {
    lastRequest time.Time
    mu          sync.Mutex
}

func (r *RateLimiter) Wait() {
    r.mu.Lock()
    defer r.mu.Unlock()

    elapsed := time.Since(r.lastRequest)
    if elapsed < time.Second {
        time.Sleep(time.Second - elapsed)
    }
    r.lastRequest = time.Now()
}
```

### 6. Team Name Normalization

**Decision**: Auto-prefix "dindoa-" to team names, accept short forms

**Rationale**:
- All Dindoa teams follow "dindoa-{name}" URL pattern
- Users type less ("j3" vs "dindoa j3")
- Case-insensitive matching improves UX

**Normalization logic**:
```
Input: "j3" / "J3" / "Dindoa J3" / "dindoa j3"
Output: "dindoa-j3"
```

### 7. Error Handling Philosophy

**Decision**: Graceful degradation for geocoding, fail-fast for scraping

**Rationale**:
- Geocoding failure shouldn't stop ICS generation (use fallback text)
- Scraping failure means no data → fail with clear error
- Show warnings for non-fatal issues (verbose mode)

**Examples**:
- ✗ Team page 404 → Exit with error "Team 'j3' not found"
- ⚠ Geocode fails → Use original text, show warning
- ⚠ Cache write fails → Continue, show warning

### 8. Architecture: Internal Package Structure

**Decision**: Organize by capability in internal/ directory

```
dindoa/
├── cmd/dindoa/
│   └── main.go              # CLI parsing, mode selection
├── internal/
│   ├── scraper/
│   │   ├── fetcher.go       # HTTP client wrapper
│   │   ├── parser.go        # HTML parsing (goquery)
│   │   └── types.go         # Match, Team, Category structs
│   ├── geocode/
│   │   ├── client.go        # OSM Nominatim client
│   │   ├── cache.go         # JSON cache read/write
│   │   └── ratelimit.go     # Rate limiter
│   ├── ics/
│   │   └── generator.go     # ICS file creation
│   └── ui/
│       ├── interactive.go   # Bubbletea app
│       └── models.go        # Bubbletea models/update/view
└── go.mod
```

**Rationale**:
- internal/ prevents external import (tool-specific code)
- Clear separation of concerns
- Each package maps to a spec capability

## Risks / Trade-offs

### [Risk] Website HTML structure changes → Scraping breaks

**Mitigation**:
- Use flexible selectors where possible (find table, don't hardcode nth-child)
- Return clear errors when expected elements missing
- Document HTML structure assumptions in code comments
- Consider adding basic structure validation

### [Risk] OSM Nominatim returns wrong address

**Mitigation**:
- Cache includes original query for manual inspection
- Users can edit cache JSON if needed
- Future: Add --verify-geocoding flag to show results before committing

### [Risk] Rate limiting too aggressive (slow for many matches)

**Impact**: Acceptable trade-off
- Typical team has 4-10 matches per season
- With cache, only first run is slow (10 matches × 2 locations = 20 seconds)
- Subsequent runs use cache (instant)

### [Risk] Timezone handling edge cases

**Mitigation**:
- Use Go's time.LoadLocation("Europe/Amsterdam") for automatic DST handling
- Assume website times are already in local CET/CEST (confirmed by user)
- ICS includes TZID for clarity

### [Risk] Large teams pages cause memory issues

**Impact**: Unlikely
- Teams page has ~30 teams total
- Match tables have <20 rows
- DOM size is manageable for goquery

### [Risk] Windows users lack proper terminal for Bubbletea

**Mitigation**:
- Test on Windows Terminal (modern default)
- CLI flag mode works in any terminal
- Document terminal requirements in README

## Migration Plan

Not applicable - this is a new tool with no existing users or data to migrate.

## Open Questions

None - all major decisions are resolved. Implementation can proceed.
