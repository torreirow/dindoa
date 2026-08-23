// Package locations resolves the venue strings published on dindoa.nl to a
// readable name, an address and coordinates, using a mapping file shipped with
// the binary that a user can override.
//
// Lookups never reach the network. A venue that is not in the mapping is
// reported as missing so the caller can say so, rather than silently
// substituting something that might be wrong.
package locations

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/adrg/xdg"
)

//go:embed locations.json
var embedded embed.FS

const (
	embeddedFile = "locations.json"
	// userFileName is looked up in the config directory, not the cache
	// directory: this is human-maintained input, not a rebuildable derivative.
	userFileName = "locations.json"
	appDir       = "dindoa"
)

// Entry is one venue in the mapping.
type Entry struct {
	// Name is the readable venue name, e.g. "Dindoa, korfbalveld De Zanderij".
	Name string `json:"name,omitempty"`
	// Address is the postal address, e.g. "Watervalweg 170, 3853 PT Ermelo".
	// It may be street-level only, in which case Lat/Lon carry the precision.
	Address string `json:"address,omitempty"`
	// Lat and Lon are the coordinates. Both zero means "no coordinates".
	Lat float64 `json:"lat,omitempty"`
	Lon float64 `json:"lon,omitempty"`
	// OSM references the OpenStreetMap object this entry came from, e.g.
	// "way/563472182", so the entry can be verified later.
	OSM string `json:"osm,omitempty"`
	// Source records how the entry was established: "manual", "osm-tags",
	// "osm-reverse". An empty value is allowed.
	Source string `json:"source,omitempty"`
}

// HasCoordinates reports whether the entry carries usable coordinates.
func (e Entry) HasCoordinates() bool {
	return e.Lat != 0 || e.Lon != 0
}

// IsEmpty reports whether the entry carries nothing usable. Entries that only
// exist as a placeholder for work still to be done look like this.
func (e Entry) IsEmpty() bool {
	return e.Name == "" && e.Address == "" && !e.HasCoordinates()
}

// file is the on-disk shape of a mapping file.
type file struct {
	Version   int              `json:"version"`
	Locations map[string]Entry `json:"locations"`
}

// Mapping holds the merged mapping: the embedded entries with the user's
// entries laid over them, per key.
type Mapping struct {
	entries  map[string]Entry
	userPath string
	// UserFileLoaded reports whether a user file was found and read.
	UserFileLoaded bool
	// Warnings holds non-fatal problems, e.g. an unreadable user file.
	Warnings []string
}

var (
	punctuation = regexp.MustCompile(`[^\p{L}\p{N}\s]+`)
	whitespace  = regexp.MustCompile(`\s+`)
)

// NormalizeKey turns a venue string from the website into a lookup key.
// Case, surrounding whitespace, repeated whitespace and punctuation are
// levelled out, so an editorial tweak on the website does not cause a miss.
//
//	"De Zanderij (Dindoa)  ERMELO" -> "de zanderij dindoa ermelo"
//
// Normalization is deliberately the only tolerance: there is no fuzzy
// matching, because a near miss that yields the wrong address is worse than a
// reported miss.
func NormalizeKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = punctuation.ReplaceAllString(s, " ")
	s = whitespace.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// UserFilePath returns the path where the user's mapping file is looked up.
func UserFilePath() string {
	return filepath.Join(xdg.ConfigHome, appDir, userFileName)
}

// Load reads the embedded mapping and lays the user's file over it. A missing
// user file is normal and produces no warning. An unreadable user file
// produces a warning and is skipped; the embedded mapping is still returned.
func Load() (*Mapping, error) {
	m := &Mapping{
		entries:  map[string]Entry{},
		userPath: UserFilePath(),
	}

	data, err := embedded.ReadFile(embeddedFile)
	if err != nil {
		return nil, fmt.Errorf("read embedded location mapping: %w", err)
	}
	base, err := parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse embedded location mapping: %w", err)
	}
	for k, v := range base.Locations {
		m.entries[NormalizeKey(k)] = v
	}

	userData, err := os.ReadFile(m.userPath)
	switch {
	case err == nil:
		user, perr := parse(userData)
		if perr != nil {
			m.Warnings = append(m.Warnings,
				fmt.Sprintf("could not read %s (%v); using the built-in location mapping", m.userPath, perr))
			break
		}
		// Per-key override: entries not mentioned by the user stay as they are.
		for k, v := range user.Locations {
			m.entries[NormalizeKey(k)] = v
		}
		m.UserFileLoaded = true
	case os.IsNotExist(err):
		// Normal: no user file.
	default:
		m.Warnings = append(m.Warnings,
			fmt.Sprintf("could not read %s (%v); using the built-in location mapping", m.userPath, err))
	}

	return m, nil
}

func parse(data []byte) (*file, error) {
	var f file
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	if f.Locations == nil {
		f.Locations = map[string]Entry{}
	}
	return &f, nil
}

// Lookup resolves a venue string. found is false when the venue is not in the
// mapping, or when its entry carries nothing usable.
func (m *Mapping) Lookup(venue string) (Entry, bool) {
	e, ok := m.entries[NormalizeKey(venue)]
	if !ok || e.IsEmpty() {
		return Entry{}, false
	}
	return e, true
}

// UserPath returns the path of the user's mapping file.
func (m *Mapping) UserPath() string { return m.userPath }

// Len returns the number of usable entries.
func (m *Mapping) Len() int {
	n := 0
	for _, e := range m.entries {
		if !e.IsEmpty() {
			n++
		}
	}
	return n
}

// Skeleton renders a JSON fragment a user can paste into their mapping file to
// fill in the given venues.
func Skeleton(venues []string) string {
	sorted := append([]string(nil), venues...)
	sort.Strings(sorted)

	f := file{Version: 1, Locations: map[string]Entry{}}
	for _, v := range sorted {
		f.Locations[NormalizeKey(v)] = Entry{Source: "manual"}
	}

	// Marshal by hand so the placeholder fields stay visible; omitempty would
	// strip the very fields the user needs to fill in.
	var b strings.Builder
	b.WriteString("{\n  \"version\": 1,\n  \"locations\": {\n")
	for i, v := range sorted {
		key, _ := json.Marshal(NormalizeKey(v))
		comment, _ := json.Marshal(v)
		b.WriteString(fmt.Sprintf("    %s: {\n", key))
		b.WriteString(fmt.Sprintf("      \"name\": %s,\n", comment))
		b.WriteString("      \"address\": \"\",\n")
		b.WriteString("      \"lat\": 0,\n")
		b.WriteString("      \"lon\": 0,\n")
		b.WriteString("      \"osm\": \"\",\n")
		b.WriteString("      \"source\": \"manual\"\n")
		if i == len(sorted)-1 {
			b.WriteString("    }\n")
		} else {
			b.WriteString("    },\n")
		}
	}
	b.WriteString("  }\n}")
	return b.String()
}
