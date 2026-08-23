package ics

import (
	"os"
	"path/filepath"
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
