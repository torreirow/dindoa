package scraper

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// expectedHeaders are the column headers of the match programme table. The
// parser verifies these before reading any row, so an unexpected layout is
// reported as an error instead of silently yielding zero matches.
var expectedHeaders = []string{"tijd", "thuis", "uit", "kleur", "locatie", "scheidsrechter"}

// dutchMonths maps the month names used in the programme's date headings.
var dutchMonths = map[string]time.Month{
	"januari":   time.January,
	"februari":  time.February,
	"maart":     time.March,
	"april":     time.April,
	"mei":       time.May,
	"juni":      time.June,
	"juli":      time.July,
	"augustus":  time.August,
	"september": time.September,
	"oktober":   time.October,
	"november":  time.November,
	"december":  time.December,
}

// minWeekdayFraction is the share of match dates that must fall on a Saturday
// or Wednesday for a derived year to be accepted. Korfbal is played on those
// two days, which makes the weekday pattern a check on the year we inferred.
const minWeekdayFraction = 0.8

// Programma holds everything parsed from the match programme page.
type Programma struct {
	Matches []Match
}

// Parser handles HTML parsing logic
type Parser struct{}

// NewParser creates a new Parser
func NewParser() *Parser {
	return &Parser{}
}

// rawRow is a match row before its year has been resolved.
type rawRow struct {
	day      int
	month    time.Month
	time     string
	home     string
	away     string
	colour   string
	location string
	referee  string
}

// ParseProgramma extracts every match from the programme page. The date lives
// in the <h3> above each table rather than in the row, so the parser walks the
// container in document order and carries the last heading it saw.
func (p *Parser) ParseProgramma(doc *goquery.Document, now time.Time) (*Programma, error) {
	container := doc.Find("div.page-content").First()
	scope := container.Children()
	if container.Length() == 0 {
		// Fall back to scanning the whole document if the wrapper changed.
		scope = doc.Find("h3, table")
	}

	var (
		rows        []rawRow
		curDay      int
		curMonth    time.Month
		haveDate    bool
		sawTable    bool
		headerError error
	)

	scope.Each(func(_ int, sel *goquery.Selection) {
		if sel.Is("h3") {
			if d, m, ok := parseDateHeading(sel.Text()); ok {
				curDay, curMonth, haveDate = d, m, true
			}
			return
		}

		table := sel
		if !table.Is("table") {
			table = sel.Find("table").First()
			if table.Length() == 0 {
				return
			}
		}

		if err := verifyHeaders(table); err != nil {
			if headerError == nil {
				headerError = err
			}
			return
		}
		sawTable = true

		if !haveDate {
			if headerError == nil {
				headerError = fmt.Errorf("match table found without a preceding date heading")
			}
			return
		}

		table.Find("tr").Each(func(_ int, tr *goquery.Selection) {
			cells := tr.Find("td")
			if cells.Length() < 5 {
				return // header row or malformed row
			}
			r := rawRow{
				day:      curDay,
				month:    curMonth,
				time:     cellText(cells.Eq(0)),
				home:     cellText(cells.Eq(1)),
				away:     cellText(cells.Eq(2)),
				colour:   cellText(cells.Eq(3)),
				location: cellText(cells.Eq(4)),
			}
			if cells.Length() >= 6 {
				r.referee = cellText(cells.Eq(5))
			}
			if r.time == "" || r.home == "" || r.away == "" {
				return
			}
			// Validate here, where the source is interpreted, next to the
			// column check. A changed time format affects the whole page, so
			// failing on the page is right; skipping the row would hand the
			// user half a calendar, which is worse than none.
			if _, _, err := ParseKickOff(r.time); err != nil {
				if headerError == nil {
					headerError = fmt.Errorf("match on %d %s between %q and %q: %w",
						r.day, dutchMonthName(r.month), r.home, r.away, err)
				}
				return
			}
			rows = append(rows, r)
		})
	})

	if headerError != nil {
		return nil, headerError
	}
	if !sawTable {
		return nil, fmt.Errorf("no match table with the expected columns (%s) found on the programme page",
			strings.Join(expectedHeaders, ", "))
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("match table found but it contains no match rows")
	}

	year, err := resolveSeasonYear(rows, now)
	if err != nil {
		return nil, err
	}

	matches := make([]Match, 0, len(rows))
	for _, r := range rows {
		matches = append(matches, Match{
			Date:     time.Date(yearFor(r.month, year), r.month, r.day, 0, 0, 0, 0, time.UTC),
			Time:     r.time,
			Home:     r.home,
			Away:     r.away,
			Colour:   r.colour,
			Location: r.location,
			Referee:  r.referee,
		})
	}

	return &Programma{Matches: matches}, nil
}

// dutchMonthName renders a month as it appears in the page headings, so an
// error message points at the same date the reader sees on the site.
func dutchMonthName(m time.Month) string {
	for name, month := range dutchMonths {
		if month == m {
			return name
		}
	}
	return m.String()
}

// verifyHeaders reports an error unless the table carries the expected columns.
func verifyHeaders(table *goquery.Selection) error {
	var got []string
	table.Find("th").Each(func(_ int, th *goquery.Selection) {
		got = append(got, strings.ToLower(strings.TrimSpace(th.Text())))
	})
	if len(got) == 0 {
		return nil // not a data table at all; skip quietly
	}
	if len(got) < len(expectedHeaders) {
		return fmt.Errorf("unexpected match table layout: expected columns [%s], found [%s]",
			strings.Join(expectedHeaders, ", "), strings.Join(got, ", "))
	}
	for i, want := range expectedHeaders {
		if got[i] != want {
			return fmt.Errorf("unexpected match table layout: expected columns [%s], found [%s]",
				strings.Join(expectedHeaders, ", "), strings.Join(got[:len(expectedHeaders)], ", "))
		}
	}
	return nil
}

// parseDateHeading reads a heading such as "5 september" into day and month.
// The heading carries no year.
func parseDateHeading(s string) (int, time.Month, bool) {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(s)))
	if len(fields) != 2 {
		return 0, 0, false
	}
	day, err := strconv.Atoi(fields[0])
	if err != nil || day < 1 || day > 31 {
		return 0, 0, false
	}
	month, ok := dutchMonths[fields[1]]
	if !ok {
		return 0, 0, false
	}
	return day, month, true
}

// seasonOpeningYear returns the calendar year in which the current season
// started. A season runs from August through July.
func seasonOpeningYear(now time.Time) int {
	if now.Month() >= time.August {
		return now.Year()
	}
	return now.Year() - 1
}

// yearFor maps a month to a calendar year within a season.
func yearFor(m time.Month, openingYear int) int {
	if m >= time.August {
		return openingYear
	}
	return openingYear + 1
}

// resolveSeasonYear derives the season's opening year and validates it against
// the weekday pattern of the parsed dates.
func resolveSeasonYear(rows []rawRow, now time.Time) (int, error) {
	opening := seasonOpeningYear(now)
	frac := weekdayFraction(rows, opening)
	if frac >= minWeekdayFraction {
		return opening, nil
	}
	return 0, fmt.Errorf(
		"could not establish the season year: assuming %d/%d puts only %.0f%% of the %d matches on a Saturday or Wednesday, "+
			"which is how korfbal is scheduled. The programme's date headings carry no year, so this is either an unexpected "+
			"page layout or a wrong system clock",
		opening, opening+1, frac*100, len(rows))
}

// weekdayFraction reports the share of dates falling on a Saturday or Wednesday.
func weekdayFraction(rows []rawRow, openingYear int) float64 {
	if len(rows) == 0 {
		return 0
	}
	ok := 0
	for _, r := range rows {
		d := time.Date(yearFor(r.month, openingYear), r.month, r.day, 0, 0, 0, 0, time.UTC)
		switch d.Weekday() {
		case time.Saturday, time.Wednesday:
			ok++
		}
	}
	return float64(ok) / float64(len(rows))
}

func cellText(sel *goquery.Selection) string {
	return multiSpace.ReplaceAllString(strings.TrimSpace(sel.Text()), " ")
}

// Teams returns every Dindoa team in the programme, with its category,
// deduplicated and sorted.
func (p *Programma) Teams() []Team {
	seen := map[string]string{} // display name -> colour
	for _, m := range p.Matches {
		for _, name := range []string{m.Home, m.Away} {
			if !IsDindoaTeam(name) {
				continue
			}
			// A non-empty colour wins over an empty one seen earlier.
			if existing, ok := seen[name]; !ok || (existing == "" && m.Colour != "") {
				seen[name] = m.Colour
			}
		}
	}

	teams := make([]Team, 0, len(seen))
	for name, colour := range seen {
		teams = append(teams, Team{
			Name:     name,
			Slug:     TeamSlug(name),
			Category: CategoryOf(colour),
		})
	}
	sort.Slice(teams, func(i, j int) bool { return lessTeamName(teams[i].Name, teams[j].Name) })
	return teams
}

// Categories groups the Dindoa teams by category.
func (p *Programma) Categories() []Category {
	byName := map[string][]Team{}
	for _, t := range p.Teams() {
		byName[t.Category] = append(byName[t.Category], t)
	}

	cats := make([]Category, 0, len(byName))
	for name, teams := range byName {
		cats = append(cats, Category{Name: name, Teams: teams})
	}
	sort.Slice(cats, func(i, j int) bool { return cats[i].Name < cats[j].Name })
	return cats
}

// CategoryNames returns the category names present in the programme.
func (p *Programma) CategoryNames() []string {
	cats := p.Categories()
	names := make([]string, 0, len(cats))
	for _, c := range cats {
		names = append(names, c.Name)
	}
	return names
}

// MatchesFor returns the matches of one team, selected by exact display name.
// Exact matching matters: "Dindoa J1" is a prefix of "Dindoa J10" through
// "Dindoa J19", and opponents such as "Revival J4" share a team code with
// "Dindoa J4".
func (p *Programma) MatchesFor(teamName string) []Match {
	var out []Match
	for _, m := range p.Matches {
		home := m.Home == teamName
		away := m.Away == teamName
		if !home && !away {
			continue
		}
		m.IsHome = home
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Date.Equal(out[j].Date) {
			return out[i].Time < out[j].Time
		}
		return out[i].Date.Before(out[j].Date)
	})
	return out
}

// HasTeam reports whether the programme contains a team with this exact name.
func (p *Programma) HasTeam(teamName string) bool {
	for _, t := range p.Teams() {
		if t.Name == teamName {
			return true
		}
	}
	return false
}

// TeamNames returns every Dindoa team name in the programme, sorted.
func (p *Programma) TeamNames() []string {
	teams := p.Teams()
	names := make([]string, 0, len(teams))
	for _, t := range teams {
		names = append(names, t.Name)
	}
	return names
}

// LocationCounts returns how many matches are played at each venue.
func (p *Programma) LocationCounts() map[string]int {
	counts := map[string]int{}
	for _, m := range p.Matches {
		if m.Location != "" {
			counts[m.Location]++
		}
	}
	return counts
}

// lessTeamName sorts team names naturally, so J2 comes before J10.
func lessTeamName(a, b string) bool {
	pa, na := splitTrailingNumber(a)
	pb, nb := splitTrailingNumber(b)
	if pa != pb {
		return pa < pb
	}
	return na < nb
}

func splitTrailingNumber(s string) (string, int) {
	i := len(s)
	for i > 0 && s[i-1] >= '0' && s[i-1] <= '9' {
		i--
	}
	if i == len(s) {
		return s, -1
	}
	n, err := strconv.Atoi(s[i:])
	if err != nil {
		return s, -1
	}
	return s[:i], n
}
