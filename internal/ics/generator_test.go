package ics

import (
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/torreirow/dindoa/internal/locations"
	"github.com/torreirow/dindoa/internal/scraper"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// j4Match mirrors the first J4 match of the 2026/2027 season.
func j4Match() scraper.Match {
	return scraper.Match{
		Date:     day(2026, time.September, 5),
		Time:     "13:45",
		Home:     "Unitas/Perspectief J4",
		Away:     "Dindoa J4",
		Colour:   "Rood",
		Location: "Het Slingerbos HARDERWIJK",
		IsHome:   false,
	}
}

// refGenTime is a fixed generation moment, so SEQUENCE is deterministic.
var refGenTime = time.Date(2026, time.August, 23, 18, 44, 0, 0, time.UTC)

// sixMatches mirrors the shape of a real team calendar: several matches, mixed
// home and away, more than one venue.
func sixMatches() []scraper.Match {
	out := make([]scraper.Match, 0, 6)
	for i := 0; i < 6; i++ {
		m := j4Match()
		m.Date = day(2026, time.September, 5+7*i)
		if i%2 == 1 {
			m.Home, m.Away, m.IsHome = "Dindoa J4", "Revival J1", true
			m.Location = "De Zanderij (Dindoa) ERMELO"
		}
		out = append(out, m)
	}
	return out
}

func generate(t *testing.T, matches []scraper.Match) string {
	t.Helper()
	m, err := locations.Load()
	if err != nil {
		t.Fatalf("load mapping: %v", err)
	}
	g, err := NewGenerator(m)
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}
	out := filepath.Join(t.TempDir(), "out.ics")
	if _, err := g.Generate("Dindoa J4", matches, out); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	return string(b)
}

func TestTimezoneStaysCorrect(t *testing.T) {
	// 13:45 CEST is 11:45 UTC. This behaviour was already right and must stay.
	got := generate(t, []scraper.Match{j4Match()})
	if !strings.Contains(got, "DTSTART:20260905T114500Z") {
		t.Errorf("DTSTART missing or wrong:\n%s", got)
	}
}

func TestDTENDIsOneHourAfterDTSTART(t *testing.T) {
	got := generate(t, []scraper.Match{j4Match()})
	if !strings.Contains(got, "DTEND:20260905T124500Z") {
		t.Errorf("DTEND missing or not one hour later:\n%s", got)
	}
}

func TestLocationKeepsNameAndAddsAddress(t *testing.T) {
	got := generate(t, []scraper.Match{j4Match()})
	// Het Slingerbos is in the shipped mapping; the readable name must survive.
	if !strings.Contains(got, "Het Slingerbos") {
		t.Errorf("readable venue name lost from LOCATION:\n%s", got)
	}
	if !strings.Contains(got, "Slingerbos 1") {
		t.Errorf("mapped address missing from LOCATION:\n%s", got)
	}
}

func TestGeoPresentForKnownVenue(t *testing.T) {
	got := generate(t, []scraper.Match{j4Match()})
	if !strings.Contains(got, "GEO:") {
		t.Errorf("GEO missing for a mapped venue:\n%s", got)
	}
}

func TestUnknownVenueKeepsWebsiteStringAndOmitsGeo(t *testing.T) {
	m := j4Match()
	m.Location = "Sporthal Nergens NERGENSHUIZEN"

	got := generate(t, []scraper.Match{m})
	if !strings.Contains(got, "LOCATION:Sporthal Nergens NERGENSHUIZEN") {
		t.Errorf("unknown venue should keep the website string verbatim:\n%s", got)
	}
	if strings.Contains(got, "GEO:") {
		t.Errorf("GEO must be omitted for an unmapped venue:\n%s", got)
	}
}

func TestMissingVenuesAreReported(t *testing.T) {
	known := j4Match()
	// A venue that is deliberately not in the mapping, so this test does not
	// depend on how complete the shipped address list happens to be.
	unknown1 := j4Match()
	unknown1.Location = "Sporthal Onbekend NERGENSHUIZEN"
	unknown2 := j4Match()
	unknown2.Location = "Sporthal Onbekend NERGENSHUIZEN"
	unknown2.Date = day(2026, time.October, 10)

	mp, err := locations.Load()
	if err != nil {
		t.Fatalf("load mapping: %v", err)
	}
	g, err := NewGenerator(mp)
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}

	out := filepath.Join(t.TempDir(), "out.ics")
	missing, err := g.Generate("Dindoa J4", []scraper.Match{known, unknown1, unknown2}, out)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if got, want := missing["Sporthal Onbekend NERGENSHUIZEN"], 2; got != want {
		t.Errorf("missing count = %d, want %d", got, want)
	}
	if _, reported := missing["Het Slingerbos HARDERWIJK"]; reported {
		t.Error("a mapped venue must not be reported as missing")
	}

	// The file is written regardless: a missing venue never blocks output.
	if _, err := os.Stat(out); err != nil {
		t.Errorf("file should exist even with missing venues: %v", err)
	}
}

func TestCategoriesCarriesColour(t *testing.T) {
	got := generate(t, []scraper.Match{j4Match()})
	if !strings.Contains(got, "CATEGORIES:Rood") {
		t.Errorf("CATEGORIES missing:\n%s", got)
	}
}

func TestCategoriesOmittedWithoutColour(t *testing.T) {
	m := j4Match()
	m.Colour = ""
	got := generate(t, []scraper.Match{m})
	if strings.Contains(got, "CATEGORIES") {
		t.Errorf("CATEGORIES must be omitted for a team without a colour:\n%s", got)
	}
}

func TestUIDIsStableAcrossKickOffChange(t *testing.T) {
	g, err := NewGenerator(nil)
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}

	early := j4Match()
	moved := j4Match()
	moved.Time = "15:30"

	a := g.generateUID("Dindoa J4", early)
	b := g.generateUID("Dindoa J4", moved)
	if a != b {
		t.Errorf("UID changed when only the kick-off time moved: %q vs %q", a, b)
	}
	if strings.Contains(a, "1345") || strings.Contains(a, "1530") {
		t.Errorf("UID still contains the kick-off time: %q", a)
	}
}

func TestUIDDiffersForTwoMatchesOnOneDay(t *testing.T) {
	g, err := NewGenerator(nil)
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}

	first := j4Match()
	second := j4Match()
	second.Home = "Revival J1"

	if a, b := g.generateUID("Dindoa J4", first), g.generateUID("Dindoa J4", second); a == b {
		t.Errorf("two matches on the same day share a UID: %q", a)
	}
}

func TestUIDIsDeterministic(t *testing.T) {
	g, err := NewGenerator(nil)
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}
	m := j4Match()
	if a, b := g.generateUID("Dindoa J4", m), g.generateUID("Dindoa J4", m); a != b {
		t.Errorf("UID not deterministic: %q vs %q", a, b)
	}
}

func TestTitleOrdersTeamsByHomeAway(t *testing.T) {
	g, err := NewGenerator(nil)
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}

	away := j4Match()
	if got, want := g.formatTitle("Dindoa J4", away), "Unitas/Perspectief J4 - Dindoa J4"; got != want {
		t.Errorf("away title = %q, want %q", got, want)
	}

	home := j4Match()
	home.Home, home.Away, home.IsHome = "Dindoa J4", "Revival J1", true
	if got, want := g.formatTitle("Dindoa J4", home), "Dindoa J4 - Revival J1"; got != want {
		t.Errorf("home title = %q, want %q", got, want)
	}
}

func TestGeneratorWorksWithoutMapping(t *testing.T) {
	// A nil mapping must degrade to the website string, not panic.
	g, err := NewGenerator(nil)
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}
	out := filepath.Join(t.TempDir(), "out.ics")
	missing, err := g.Generate("Dindoa J4", []scraper.Match{j4Match()}, out)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got := missing["Het Slingerbos HARDERWIJK"]; got != 1 {
		t.Errorf("without a mapping the venue should be reported missing, got %d", got)
	}
	b, _ := os.ReadFile(out)
	if !strings.Contains(string(b), "LOCATION:Het Slingerbos HARDERWIJK") {
		t.Errorf("website string should be used verbatim:\n%s", b)
	}
}

func TestDefaultOutputFilename(t *testing.T) {
	for in, want := range map[string]string{
		"Dindoa J4":    "dindoa-j4.ics",
		"Dindoa U15-1": "dindoa-u15-1.ics",
		"Dindoa 4":     "dindoa-4.ics",
	} {
		if got := DefaultOutputFilename(in); got != want {
			t.Errorf("DefaultOutputFilename(%q) = %q, want %q", in, got, want)
		}
	}
}

// generateAt renders a calendar with the clock pinned, so a test can produce
// two editions without waiting a minute.
func generateAt(t *testing.T, at time.Time, matches []scraper.Match) string {
	t.Helper()
	m, err := locations.Load()
	if err != nil {
		t.Fatalf("load mapping: %v", err)
	}
	g, err := NewGenerator(m)
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}
	g.now = func() time.Time { return at }

	out := filepath.Join(t.TempDir(), "out.ics")
	if _, err := g.Generate("Dindoa J4", matches, out); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	return string(b)
}

func sequences(t *testing.T, ics string) []int64 {
	t.Helper()
	var out []int64
	for _, line := range strings.Split(ics, "\n") {
		line = strings.TrimRight(line, "\r")
		if v, ok := strings.CutPrefix(line, "SEQUENCE:"); ok {
			n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			if err != nil {
				t.Fatalf("SEQUENCE is not an integer: %q", v)
			}
			out = append(out, n)
		}
	}
	return out
}

func TestEveryEventCarriesASequence(t *testing.T) {
	got := generateAt(t, refGenTime, sixMatches())

	seq := sequences(t, got)
	if len(seq) != 6 {
		t.Fatalf("found %d SEQUENCE properties, want one per event (6)", len(seq))
	}
	for _, n := range seq {
		if n < 0 {
			t.Errorf("SEQUENCE is negative: %d", n)
		}
	}
}

func TestOneSequencePerGeneratedFile(t *testing.T) {
	// SEQUENCE identifies the edition of the calendar, not the match, so every
	// event in one file must carry the same value.
	seq := sequences(t, generateAt(t, refGenTime, sixMatches()))
	for _, n := range seq[1:] {
		if n != seq[0] {
			t.Fatalf("events in one file carry different SEQUENCE values: %v", seq)
		}
	}
}

func TestLaterGenerationHasHigherSequence(t *testing.T) {
	early := sequences(t, generateAt(t, refGenTime, []scraper.Match{j4Match()}))
	later := sequences(t, generateAt(t, refGenTime.Add(2*time.Minute), []scraper.Match{j4Match()}))

	if len(early) != 1 || len(later) != 1 {
		t.Fatalf("expected one SEQUENCE each, got %d and %d", len(early), len(later))
	}
	if later[0] <= early[0] {
		t.Errorf("SEQUENCE did not increase: %d then %d", early[0], later[0])
	}
}

func TestSequenceFitsIn32BitSignedInteger(t *testing.T) {
	// The property is defined as an integer; keep it inside the 32-bit range
	// that calendar applications rely on.
	for _, at := range []time.Time{
		refGenTime,
		time.Date(2100, time.January, 1, 0, 0, 0, 0, time.UTC),
	} {
		for _, n := range sequences(t, generateAt(t, at, []scraper.Match{j4Match()})) {
			if n < 0 || n > math.MaxInt32 {
				t.Errorf("SEQUENCE %d at %s is outside the 32-bit signed range", n, at.Format("2006"))
			}
		}
	}
}

// Two editions differ only in the fields that identify the edition. Anything
// else changing would mean regenerating alters the match data.
func TestRegeneratingChangesOnlyEditionFields(t *testing.T) {
	a := generateAt(t, refGenTime, sixMatches())
	b := generateAt(t, refGenTime.Add(90*time.Minute), sixMatches())

	if stripEdition(a) != stripEdition(b) {
		t.Error("regenerating changed more than DTSTAMP and SEQUENCE")
	}
	if a == b {
		t.Error("expected the edition fields themselves to differ")
	}
}

// stripEdition removes the fields that identify the edition rather than the
// match: DTSTAMP and SEQUENCE.
func stripEdition(s string) string {
	var keep []string
	for _, line := range strings.Split(s, "\n") {
		t := strings.TrimRight(line, "\r")
		if strings.HasPrefix(t, "DTSTAMP:") || strings.HasPrefix(t, "SEQUENCE:") {
			continue
		}
		keep = append(keep, t)
	}
	return strings.Join(keep, "\n")
}
