package locations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrg/xdg"
)

func TestNormalizeKey(t *testing.T) {
	// All of these are the same venue written differently.
	same := []string{
		"De Zanderij (Dindoa) ERMELO",
		"de zanderij (dindoa) ermelo",
		"  De Zanderij (Dindoa)  ERMELO  ",
		"De Zanderij Dindoa ERMELO",
		"De Zanderij, Dindoa - ERMELO",
	}
	want := NormalizeKey(same[0])
	for _, s := range same[1:] {
		if got := NormalizeKey(s); got != want {
			t.Errorf("NormalizeKey(%q) = %q, want %q", s, got, want)
		}
	}

	if got, want := NormalizeKey("Burg. Bode Sportpark ELBURG"), "burg bode sportpark elburg"; got != want {
		t.Errorf("NormalizeKey = %q, want %q", got, want)
	}
}

func TestEmbeddedMappingLoads(t *testing.T) {
	m, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if m.Len() == 0 {
		t.Fatal("embedded mapping is empty")
	}

	e, ok := m.Lookup("De Zanderij (Dindoa) ERMELO")
	if !ok {
		t.Fatal("home venue not found in the embedded mapping")
	}
	if e.Address == "" {
		t.Error("home venue has no address")
	}
	if !e.HasCoordinates() {
		t.Error("home venue has no coordinates")
	}
	if e.OSM == "" {
		t.Error("home venue has no OSM reference, so the entry cannot be verified later")
	}
}

func TestEmbeddedEntriesAreUsable(t *testing.T) {
	m, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// A shipped entry with no address or coordinates would be reported as
	// found while carrying nothing, which is the trap this guards against.
	for key, e := range m.entries {
		if e.IsEmpty() {
			t.Errorf("shipped entry %q is empty; leave it out instead", key)
		}
		if e.Address == "" && !e.HasCoordinates() {
			t.Errorf("shipped entry %q has neither address nor coordinates", key)
		}
	}
}

func TestLookupIsCaseAndPunctuationInsensitive(t *testing.T) {
	m, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	for _, v := range []string{
		"HET SLINGERBOS HARDERWIJK",
		"het slingerbos harderwijk",
		"Het  Slingerbos   HARDERWIJK",
	} {
		if _, ok := m.Lookup(v); !ok {
			t.Errorf("Lookup(%q) missed", v)
		}
	}
}

func TestLookupDoesNotMatchApproximately(t *testing.T) {
	m, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Close to a real key, but not equal after normalization. A near miss must
	// be reported as missing rather than resolved to the wrong address.
	for _, v := range []string{
		"Het Slingerbosch HARDERWIJK",
		"Slingerbos HARDERWIJK",
		"Het Slingerbos AMERSFOORT",
	} {
		if e, ok := m.Lookup(v); ok {
			t.Errorf("Lookup(%q) matched %q; approximate matching must not happen", v, e.Address)
		}
	}
}

func TestUnknownVenueIsMissing(t *testing.T) {
	m, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, ok := m.Lookup("Sporthal Nergens NERGENSHUIZEN"); ok {
		t.Error("unknown venue reported as found")
	}
}

// withUserFile points xdg's config home at a temp dir holding the given JSON.
func withUserFile(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if content == "" {
		return
	}
	if err := os.MkdirAll(filepath.Join(dir, appDir), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, appDir, userFileName), []byte(content), 0o644); err != nil {
		t.Fatalf("write user file: %v", err)
	}
}

func TestUserFileOverridesOneKeyOnly(t *testing.T) {
	base, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	other, ok := base.Lookup("Sportpark Rielerenk DEVENTER")
	if !ok {
		t.Fatal("fixture entry missing; test cannot distinguish override from replacement")
	}

	withUserFile(t, `{"version":1,"locations":{
	  "het slingerbos harderwijk":{"name":"Eigen naam","address":"Eigen adres 1, 1234 AB Harderwijk","lat":1.5,"lon":2.5,"source":"manual"}
	}}`)
	xdg.Reload()

	m, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !m.UserFileLoaded {
		t.Fatal("user file was not loaded")
	}
	if len(m.Warnings) != 0 {
		t.Errorf("unexpected warnings: %v", m.Warnings)
	}

	got, ok := m.Lookup("Het Slingerbos HARDERWIJK")
	if !ok {
		t.Fatal("overridden entry not found")
	}
	if got.Address != "Eigen adres 1, 1234 AB Harderwijk" || got.Lat != 1.5 {
		t.Errorf("override not applied: %+v", got)
	}

	// Every other entry must be untouched: this is a per-key override, not a
	// wholesale replacement of the mapping.
	still, ok := m.Lookup("Sportpark Rielerenk DEVENTER")
	if !ok {
		t.Fatal("unrelated entry disappeared after loading a user file")
	}
	if still.Address != other.Address {
		t.Errorf("unrelated entry changed: %q -> %q", other.Address, still.Address)
	}
	if m.Len() < base.Len() {
		t.Errorf("mapping shrank from %d to %d entries", base.Len(), m.Len())
	}
}

func TestUserFileAddsNewKey(t *testing.T) {
	// A key that is not in the shipped list, so this really tests adding.
	withUserFile(t, `{"version":1,"locations":{
	  "sporthal onbekend nergenshuizen":{"name":"Sporthal Onbekend","address":"Ergens 1, 1234 XX Nergenshuizen","lat":52.1,"lon":5.5,"source":"manual"}
	}}`)
	xdg.Reload()

	m, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	e, ok := m.Lookup("Sporthal Onbekend NERGENSHUIZEN")
	if !ok {
		t.Fatal("venue added by the user file was not found")
	}
	if e.Address != "Ergens 1, 1234 XX Nergenshuizen" {
		t.Errorf("address = %q", e.Address)
	}
}

func TestMissingUserFileIsSilent(t *testing.T) {
	withUserFile(t, "")
	xdg.Reload()

	m, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if m.UserFileLoaded {
		t.Error("UserFileLoaded should be false when there is no user file")
	}
	if len(m.Warnings) != 0 {
		t.Errorf("a missing user file must not warn, got: %v", m.Warnings)
	}
	if m.Len() == 0 {
		t.Error("embedded entries should still be available")
	}
}

func TestUnreadableUserFileWarnsAndFallsBack(t *testing.T) {
	withUserFile(t, `{this is not json`)
	xdg.Reload()

	m, err := Load()
	if err != nil {
		t.Fatalf("Load should not fail on a broken user file: %v", err)
	}
	if m.UserFileLoaded {
		t.Error("UserFileLoaded should be false for an unreadable file")
	}
	if len(m.Warnings) == 0 {
		t.Fatal("expected a warning for an unreadable user file")
	}
	if m.Len() == 0 {
		t.Error("embedded mapping should still be usable")
	}
	// The warning must name the path so the user knows what to fix.
	if !strings.Contains(m.Warnings[0], UserFilePath()) {
		t.Errorf("warning should contain the path, got: %q", m.Warnings[0])
	}
}

func TestSkeletonIsPasteable(t *testing.T) {
	s := Skeleton([]string{"Sportpark Overbeek TERSCHUUR", "Veld ASVD DRONTEN"})
	for _, want := range []string{
		`"sportpark overbeek terschuur"`,
		`"veld asvd dronten"`,
		`"address": ""`,
		`"source": "manual"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("skeleton missing %q:\n%s", want, s)
		}
	}
	// It must parse as JSON, otherwise pasting it breaks the user's file.
	if _, err := parse([]byte(s)); err != nil {
		t.Errorf("skeleton is not valid JSON: %v\n%s", err, s)
	}
}
