package scraper

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// refNow is a moment inside the 2026/2027 season, used so the year inference
// is deterministic regardless of when the tests run.
var refNow = time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)

func loadProgramma(t *testing.T) *Programma {
	t.Helper()
	f, err := os.Open("testdata/programma.html")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}

	prog, err := NewParser().ParseProgramma(doc, refNow)
	if err != nil {
		t.Fatalf("ParseProgramma: %v", err)
	}
	return prog
}

func TestParseProgrammaReadsEveryRow(t *testing.T) {
	prog := loadProgramma(t)
	if got, want := len(prog.Matches), 210; got != want {
		t.Errorf("matches = %d, want %d", got, want)
	}
}

func TestParseProgrammaDatesComeFromHeadings(t *testing.T) {
	prog := loadProgramma(t)

	first := prog.Matches[0]
	want := time.Date(2026, time.September, 5, 0, 0, 0, 0, time.UTC)
	if !first.Date.Equal(want) {
		t.Errorf("first match date = %s, want %s", first.Date.Format("2006-01-02"), want.Format("2006-01-02"))
	}
	if first.Date.Weekday() != time.Saturday {
		t.Errorf("first match falls on %s, want Saturday", first.Date.Weekday())
	}
}

// The programme's headings carry no year. Only 2026 puts these dates on the
// Saturday/Wednesday pattern korfbal is scheduled on: 2025 gives Friday/Tuesday
// and 2027 gives Sunday/Thursday.
func TestParseProgrammaRejectsWrongSeasonYear(t *testing.T) {
	f, err := os.Open("testdata/programma.html")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()
	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}

	// A clock a year off derives the wrong season, which the weekday check catches.
	wrong := time.Date(2025, time.August, 23, 12, 0, 0, 0, time.UTC)
	if _, err := NewParser().ParseProgramma(doc, wrong); err == nil {
		t.Fatal("expected an error for a season year that does not fit the weekday pattern")
	} else if !strings.Contains(err.Error(), "season year") {
		t.Errorf("error should name the season year, got: %v", err)
	}
}

func TestVerifyHeadersRejectsUnexpectedLayout(t *testing.T) {
	// The old team page layout: Datum | Tijd | Thuis | Uit | Locatie. Pointing
	// the parser at it must fail loudly rather than yield zero matches.
	html := `<html><body><div class="page-content"><h3>5 september</h3>
	<table><thead><tr><th>Datum</th><th>Tijd</th><th>Thuis</th><th>Uit</th><th>Locatie</th></tr></thead>
	<tbody><tr><td>05-09-2026</td><td>11:50</td><td>Dindoa J3</td><td>Anders J3</td><td>Ergens ERMELO</td></tr></tbody>
	</table></div></body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	_, err = NewParser().ParseProgramma(doc, refNow)
	if err == nil {
		t.Fatal("expected an error for an unexpected table layout, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected match table layout") {
		t.Errorf("error should describe the layout mismatch, got: %v", err)
	}
}

func TestMatchesForJ4(t *testing.T) {
	prog := loadProgramma(t)
	matches := prog.MatchesFor("Dindoa J4")

	if got, want := len(matches), 6; got != want {
		t.Fatalf("Dindoa J4 matches = %d, want %d", got, want)
	}

	var home, away int
	for _, m := range matches {
		if m.IsHome {
			home++
		} else {
			away++
		}
		if m.Colour != "Rood" {
			t.Errorf("colour = %q, want Rood", m.Colour)
		}
	}
	if home != 3 || away != 3 {
		t.Errorf("home/away = %d/%d, want 3/3", home, away)
	}

	first := matches[0]
	if first.Time != "13:45" || first.Home != "Unitas/Perspectief J4" || first.Away != "Dindoa J4" {
		t.Errorf("first match = %s %s - %s, want 13:45 Unitas/Perspectief J4 - Dindoa J4",
			first.Time, first.Home, first.Away)
	}
	if first.Location != "Het Slingerbos HARDERWIJK" {
		t.Errorf("first location = %q", first.Location)
	}
	if first.IsHome {
		t.Error("first match should be an away match")
	}
}

// Dindoa J1 is a prefix of J10 through J19; substring matching returns 66 rows
// instead of 6. Dindoa J2 is a prefix of J20 through J24, giving 36.
func TestMatchesForPrefixTeamsDoNotLeak(t *testing.T) {
	prog := loadProgramma(t)
	for _, name := range []string{"Dindoa J1", "Dindoa J2", "Dindoa J4"} {
		if got, want := len(prog.MatchesFor(name)), 6; got != want {
			t.Errorf("%s matches = %d, want %d", name, got, want)
		}
	}
}

// "Dindoa 4" is the senior team, a different team from "Dindoa J4".
func TestSeniorTeamIsDistinctFromJuniorTeam(t *testing.T) {
	prog := loadProgramma(t)
	senior := prog.MatchesFor("Dindoa 4")
	junior := prog.MatchesFor("Dindoa J4")

	if len(senior) == 0 {
		t.Fatal("Dindoa 4 should have matches")
	}
	for _, s := range senior {
		for _, j := range junior {
			if s.Date.Equal(j.Date) && s.Time == j.Time && s.Home == j.Home && s.Away == j.Away {
				t.Errorf("senior and junior selections overlap on %s %s", s.Date.Format("2006-01-02"), s.Time)
			}
		}
	}
}

// An opponent sharing the team code must never end up in the selection.
func TestOpponentWithSameTeamCodeExcluded(t *testing.T) {
	prog := loadProgramma(t)
	for _, m := range prog.MatchesFor("Dindoa J4") {
		if m.Home != "Dindoa J4" && m.Away != "Dindoa J4" {
			t.Errorf("match %s - %s does not involve Dindoa J4", m.Home, m.Away)
		}
	}

	// These opponents exist in the programme and share the "J4" code.
	for _, opponent := range []string{"Revival J4", "Regio '72 J4", "Unitas/Perspectief J4"} {
		found := false
		for _, m := range prog.Matches {
			if m.Home == opponent || m.Away == opponent {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("fixture no longer contains opponent %q; the guard is meaningless", opponent)
		}
	}
}

func TestTeamsAndCategories(t *testing.T) {
	prog := loadProgramma(t)

	if got, want := len(prog.Teams()), 35; got != want {
		t.Errorf("teams = %d, want %d", got, want)
	}

	byName := map[string]string{}
	for _, tm := range prog.Teams() {
		byName[tm.Name] = tm.Category
	}
	for name, want := range map[string]string{
		"Dindoa J4":    "Rood",
		"Dindoa J14":   "Geel",
		"Dindoa J24":   "Blauw",
		"Dindoa J15":   "Groen",
		"Dindoa J5":    "Oranje",
		"Dindoa 4":     CategoryOther,
		"Dindoa U15-1": CategoryOther,
		"Dindoa MW1":   CategoryOther,
	} {
		if got := byName[name]; got != want {
			t.Errorf("%s category = %q, want %q", name, got, want)
		}
	}

	// Only Dindoa teams are listed.
	for _, tm := range prog.Teams() {
		if !IsDindoaTeam(tm.Name) {
			t.Errorf("non-Dindoa team in list: %q", tm.Name)
		}
	}
}

func TestLocationCounts(t *testing.T) {
	prog := loadProgramma(t)
	counts := prog.LocationCounts()

	if got, want := len(counts), 35; got != want {
		t.Errorf("distinct locations = %d, want %d", got, want)
	}
	if got, want := counts["De Zanderij (Dindoa) ERMELO"], 105; got != want {
		t.Errorf("home venue matches = %d, want %d", got, want)
	}
}

func TestNormalizeTeamInput(t *testing.T) {
	for in, want := range map[string]string{
		"j3":           "Dindoa J3",
		"J3":           "Dindoa J3",
		"  j3  ":       "Dindoa J3",
		"dindoa j3":    "Dindoa J3",
		"Dindoa J3":    "Dindoa J3",
		"DINDOA J3":    "Dindoa J3",
		"dindoa-j3":    "Dindoa J3",
		"4":            "Dindoa 4",
		"u15-1":        "Dindoa U15-1",
		"mw1":          "Dindoa MW1",
		"Dindoa U19-1": "Dindoa U19-1",
	} {
		if got := NormalizeTeamInput(in); got != want {
			t.Errorf("NormalizeTeamInput(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizedInputResolvesToRealTeams(t *testing.T) {
	prog := loadProgramma(t)
	for _, in := range []string{"j4", "J4", "dindoa j4", "Dindoa J4", "DINDOA J4"} {
		name := NormalizeTeamInput(in)
		if !prog.HasTeam(name) {
			t.Errorf("input %q resolved to %q, which is not in the programme", in, name)
		}
		if got := len(prog.MatchesFor(name)); got != 6 {
			t.Errorf("input %q gave %d matches, want 6", in, got)
		}
	}
}

func TestTeamSlug(t *testing.T) {
	for in, want := range map[string]string{
		"Dindoa J3":    "dindoa-j3",
		"Dindoa 4":     "dindoa-4",
		"Dindoa U15-1": "dindoa-u15-1",
		"Dindoa MW1":   "dindoa-mw1",
	} {
		if got := TeamSlug(in); got != want {
			t.Errorf("TeamSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNoDindoaVersusDindoaInFixture(t *testing.T) {
	// Home/away is currently unambiguous because no such match exists. If one
	// ever appears, MatchesFor still resolves it by exact name, but this test
	// records the assumption.
	prog := loadProgramma(t)
	for _, m := range prog.Matches {
		if IsDindoaTeam(m.Home) && IsDindoaTeam(m.Away) {
			t.Logf("Dindoa vs Dindoa: %s - %s on %s", m.Home, m.Away, m.Date.Format("2006-01-02"))
		}
	}
}

func TestParseKickOffAcceptsPublishedForms(t *testing.T) {
	for in, want := range map[string][2]int{
		"13:45": {13, 45},
		"9:30":  {9, 30},
		"09:30": {9, 30},
		"00:00": {0, 0},
		"23:59": {23, 59},
		" 9:30": {9, 30},
	} {
		h, m, err := ParseKickOff(in)
		if err != nil {
			t.Errorf("ParseKickOff(%q) failed: %v", in, err)
			continue
		}
		if h != want[0] || m != want[1] {
			t.Errorf("ParseKickOff(%q) = %d:%02d, want %d:%02d", in, h, m, want[0], want[1])
		}
	}
}

// Each of these used to pass through and produce a calendar entry at the wrong
// moment; "1345" landed 56 days away because time.Date normalises hour 1345.
func TestParseKickOffRejectsEverythingElse(t *testing.T) {
	for _, in := range []string{"13.45", "1345", "", "abc", "25:99", "24:00", "13:60", "13:4", "13:456", "-1:00"} {
		if _, _, err := ParseKickOff(in); err == nil {
			t.Errorf("ParseKickOff(%q) should have failed", in)
		} else if !strings.Contains(err.Error(), in) && in != "" {
			t.Errorf("error for %q should quote the value, got: %v", in, err)
		}
	}
}

func TestUnreadableTimeFailsTheWholePage(t *testing.T) {
	// A changed time format affects every row, so half a calendar would be
	// more misleading than none.
	html := `<html><body><div class="page-content"><h3>5 september</h3>
	<table><thead><tr><th>Tijd</th><th>Thuis</th><th>Uit</th><th>Kleur</th><th>Locatie</th><th>Scheidsrechter</th></tr></thead>
	<tbody>
	<tr><td>11:50</td><td>Dindoa J3</td><td>Anders J3</td><td>Rood</td><td>Ergens ERMELO</td><td></td></tr>
	<tr><td>13.45</td><td>Dindoa J4</td><td>Anders J4</td><td>Rood</td><td>Ergens ERMELO</td><td></td></tr>
	</tbody></table></div></body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	prog, err := NewParser().ParseProgramma(doc, refNow)
	if err == nil {
		t.Fatalf("expected an error; got %d matches", len(prog.Matches))
	}
	for _, want := range []string{"13.45", "5 september", "Dindoa J4"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should mention %q, got: %v", want, err)
		}
	}
}

func TestColumnCheckTakesPrecedenceOverTimeCheck(t *testing.T) {
	// A reordered table should be reported as a layout problem, not as a
	// strange time, so the message points at the real cause.
	html := `<html><body><div class="page-content"><h3>5 september</h3>
	<table><thead><tr><th>Thuis</th><th>Tijd</th><th>Uit</th><th>Kleur</th><th>Locatie</th><th>Scheidsrechter</th></tr></thead>
	<tbody><tr><td>Dindoa J3</td><td>11:50</td><td>Anders J3</td><td>Rood</td><td>Ergens ERMELO</td><td></td></tr></tbody>
	</table></div></body></html>`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	_, err := NewParser().ParseProgramma(doc, refNow)
	if err == nil {
		t.Fatal("expected an error for a reordered table")
	}
	if !strings.Contains(err.Error(), "unexpected match table layout") {
		t.Errorf("should report the layout, got: %v", err)
	}
}

// The published programme must keep parsing: 210 rows, all HH:MM.
func TestRealFixtureStillParses(t *testing.T) {
	prog := loadProgramma(t)
	if got := len(prog.Matches); got != 210 {
		t.Errorf("matches = %d, want 210", got)
	}
	for _, m := range prog.Matches {
		if _, _, err := ParseKickOff(m.Time); err != nil {
			t.Errorf("fixture row has an unreadable time %q: %v", m.Time, err)
		}
	}
}
