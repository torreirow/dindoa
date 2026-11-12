package locations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DefaultNominatimURL is the public OpenStreetMap geocoder.
const DefaultNominatimURL = "https://nominatim.openstreetmap.org"

// clusterRadius is how far apart candidates may lie and still count as one
// venue, e.g. a sports hall and the sports centre around it.
const clusterRadius = 150.0 // metres

// NominatimSource finds venues in OpenStreetMap. It only takes over a hit that
// is a sports venue in the town named in the venue string; anything else is
// reported as not found, because a wrong address is worse than a missing one.
//
// It sends at most one request per Interval, as the Nominatim usage policy
// asks, and stops at the first failed request.
type NominatimSource struct {
	BaseURL  string
	Client   *http.Client
	Interval time.Duration
	// Progress, when set, receives one line per venue looked up.
	Progress io.Writer

	now     func() time.Time
	sleep   func(context.Context, time.Duration) error
	last    time.Time
	stopped error
}

// NewNominatimSource returns a source for the public Nominatim service.
func NewNominatimSource() *NominatimSource {
	return &NominatimSource{
		BaseURL:  DefaultNominatimURL,
		Client:   &http.Client{Timeout: 20 * time.Second},
		Interval: 1100 * time.Millisecond,
	}
}

// Name implements Source.
func (n *NominatimSource) Name() string { return "OpenStreetMap" }

// nominatimHit is the part of a Nominatim jsonv2 result this package uses.
type nominatimHit struct {
	OSMType   string            `json:"osm_type"`
	OSMID     int64             `json:"osm_id"`
	Lat       string            `json:"lat"`
	Lon       string            `json:"lon"`
	Category  string            `json:"category"`
	Type      string            `json:"type"`
	Name      string            `json:"name"`
	Display   string            `json:"display_name"`
	Address   map[string]string `json:"address"`
	ExtraTags map[string]string `json:"extratags"`
}

func (h nominatimHit) osmRef() string { return fmt.Sprintf("%s/%d", h.OSMType, h.OSMID) }

func (h nominatimHit) coords() (float64, float64) {
	lat, _ := strconv.ParseFloat(h.Lat, 64)
	lon, _ := strconv.ParseFloat(h.Lon, 64)
	return lat, lon
}

// place returns the town the hit lies in, as OpenStreetMap spells it.
func (h nominatimHit) place() string {
	for _, k := range []string{"city", "town", "village", "hamlet", "municipality"} {
		if v := h.Address[k]; v != "" {
			return v
		}
	}
	return ""
}

// inTown reports whether the hit lies in town, compared after normalisation.
func (h nominatimHit) inTown(town string) bool {
	want := NormalizeKey(town)
	for _, k := range []string{"city", "town", "village", "hamlet", "municipality"} {
		if v := h.Address[k]; v != "" && NormalizeKey(v) == want {
			return true
		}
	}
	return false
}

// kindRank says how well the kind of hit fits an indoor korfbal venue; lower
// is better and 0 means it is not a sports venue at all.
func (h nominatimHit) kindRank() int {
	switch {
	case h.Type == "sports_hall" && (h.Category == "leisure" || h.Category == "building"):
		return 1
	case h.Category == "leisure" && h.Type == "sports_centre":
		return 2
	case h.Category == "leisure" && h.Type == "pitch" && strings.Contains(h.ExtraTags["sport"], "korfball"):
		return 3
	}
	return 0
}

// Find implements Source.
func (n *NominatimSource) Find(ctx context.Context, venues []string) (map[string]Entry, map[string]string, error) {
	found := map[string]Entry{}
	reasons := map[string]string{}
	if n.Progress != nil && len(venues) > 0 {
		// Two requests per venue at most, one per Interval.
		est := time.Duration(2*len(venues)) * n.Interval
		fmt.Fprintf(n.Progress, "Looking up %d venue(s) in OpenStreetMap (up to %s)...\n",
			len(venues), est.Round(time.Second))
	}

	for i, v := range venues {
		k := NormalizeKey(v)
		if err := ctx.Err(); err != nil {
			markNotLookedUp(reasons, venues[i:], "interrupted")
			return found, reasons, err
		}
		if n.stopped != nil {
			markNotLookedUp(reasons, venues[i:], n.stopped.Error())
			return found, reasons, n.stopped
		}
		if n.Progress != nil {
			fmt.Fprintf(n.Progress, "  [%d/%d] %s\n", i+1, len(venues), v)
		}

		name, town, ok := SplitVenue(v)
		if !ok {
			reasons[k] = "no town recognised in the name"
			continue
		}
		e, reason, err := n.findOne(ctx, name, town)
		if err != nil {
			markNotLookedUp(reasons, venues[i:], err.Error())
			return found, reasons, err
		}
		if reason != "" {
			reasons[k] = reason
			continue
		}
		found[k] = e
	}
	return found, reasons, nil
}

func markNotLookedUp(reasons map[string]string, venues []string, why string) {
	for _, v := range venues {
		reasons[NormalizeKey(v)] = "not looked up: " + why
	}
}

// findOne looks up one venue. It returns either an entry, or a reason why
// there is none, or an error when Nominatim itself failed.
func (n *NominatimSource) findOne(ctx context.Context, name, town string) (Entry, string, error) {
	townTitle := TownTitle(town)
	var cands []nominatimHit
	for _, q := range []string{name + ", " + townTitle, "sporthal " + name + ", " + townTitle} {
		hits, err := n.search(ctx, url.Values{
			"q":              {q},
			"format":         {"jsonv2"},
			"addressdetails": {"1"},
			"extratags":      {"1"},
			"countrycodes":   {"nl"},
			"limit":          {"10"},
		})
		if err != nil {
			return Entry{}, "", err
		}
		cands = suitable(hits, town)
		if len(cands) > 0 {
			break
		}
	}

	if len(cands) == 0 {
		return Entry{}, "no sports venue found in " + townTitle, nil
	}
	if !withinRadius(cands, clusterRadius) {
		return Entry{}, fmt.Sprintf("ambiguous: %d sports venues in %s", len(cands), townTitle), nil
	}

	best := cands[0]
	for _, c := range cands[1:] {
		if c.kindRank() < best.kindRank() {
			best = c
		}
	}
	lat, lon := best.coords()
	e := Entry{
		Name:    best.Name,
		Address: formatAddress(best.Address["road"], best.Address["house_number"], best.Address["postcode"], best.place()),
		Lat:     lat,
		Lon:     lon,
		OSM:     best.osmRef(),
		Source:  "osm-auto",
	}
	if e.Name == "" {
		e.Name = name
	}
	return e, "", nil
}

// suitable keeps the hits that are a sports venue in town, once per object.
func suitable(hits []nominatimHit, town string) []nominatimHit {
	var out []nominatimHit
	seen := map[string]bool{}
	for _, h := range hits {
		if h.kindRank() == 0 || !h.inTown(town) || seen[h.osmRef()] {
			continue
		}
		seen[h.osmRef()] = true
		out = append(out, h)
	}
	return out
}

// withinRadius reports whether every pair of hits lies within r metres.
func withinRadius(hits []nominatimHit, r float64) bool {
	for i := range hits {
		for j := i + 1; j < len(hits); j++ {
			if distance(hits[i], hits[j]) > r {
				return false
			}
		}
	}
	return true
}

// distance is an equirectangular approximation in metres, plenty for points
// a few hundred metres apart.
func distance(a, b nominatimHit) float64 {
	alat, alon := a.coords()
	blat, blon := b.coords()
	const earth = 6371000.0
	x := (blon - alon) * math.Pi / 180 * math.Cos((alat+blat)/2*math.Pi/180)
	y := (blat - alat) * math.Pi / 180
	return math.Hypot(x, y) * earth
}

// AddressHit is the result of looking up an address the user typed in.
type AddressHit struct {
	Lat, Lon float64
	OSM      string
	Display  string
}

// LookupAddress looks up a postal address. A hit only counts when its postcode
// equals the one given (ignoring spaces and case), so a near miss in another
// street is not taken for the address the user meant.
func (n *NominatimSource) LookupAddress(ctx context.Context, street, postcode, town string) (AddressHit, bool, error) {
	if n.stopped != nil {
		return AddressHit{}, false, n.stopped
	}
	params := url.Values{
		"street":         {street},
		"city":           {town},
		"format":         {"jsonv2"},
		"addressdetails": {"1"},
		"countrycodes":   {"nl"},
		"limit":          {"5"},
	}
	if postcode != "" {
		params.Set("postalcode", postcode)
	}
	hits, err := n.search(ctx, params)
	if err != nil {
		return AddressHit{}, false, err
	}
	want := compactPostcode(postcode)
	for _, h := range hits {
		if want != "" && compactPostcode(h.Address["postcode"]) != want {
			continue
		}
		lat, lon := h.coords()
		return AddressHit{Lat: lat, Lon: lon, OSM: h.osmRef(), Display: h.Display}, true, nil
	}
	return AddressHit{}, false, nil
}

// search sends one rate-limited request. Any failure stops the source: later
// calls return the same error without contacting Nominatim again.
func (n *NominatimSource) search(ctx context.Context, params url.Values) ([]nominatimHit, error) {
	if err := n.throttle(ctx); err != nil {
		return nil, err
	}
	data, err := httpGet(ctx, n.Client, strings.TrimRight(n.BaseURL, "/")+"/search?"+params.Encode())
	if err != nil {
		n.stopped = fmt.Errorf("OpenStreetMap: %w", err)
		return nil, n.stopped
	}
	var hits []nominatimHit
	if err := json.Unmarshal(data, &hits); err != nil {
		n.stopped = fmt.Errorf("OpenStreetMap: unreadable answer: %w", err)
		return nil, n.stopped
	}
	return hits, nil
}

// throttle waits until at least Interval has passed since the previous request.
func (n *NominatimSource) throttle(ctx context.Context) error {
	now, sleep := n.now, n.sleep
	if now == nil {
		now = time.Now
	}
	if sleep == nil {
		sleep = sleepCtx
	}
	if !n.last.IsZero() {
		if wait := n.Interval - now().Sub(n.last); wait > 0 {
			if err := sleep(ctx, wait); err != nil {
				return err
			}
		}
	}
	n.last = now()
	return nil
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// compactPostcode strips spaces and upper-cases: "3853 pt" -> "3853PT".
func compactPostcode(pc string) string {
	return strings.ToUpper(strings.ReplaceAll(pc, " ", ""))
}

// FormatPostcode writes a Dutch postcode as "1234 AB" and leaves anything
// else as it was typed.
func FormatPostcode(pc string) string {
	c := compactPostcode(pc)
	if len(c) == 6 && isDigits(c[:4]) && isLetters(c[4:]) {
		return c[:4] + " " + c[4:]
	}
	return strings.TrimSpace(pc)
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isLetters(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

// formatAddress builds "Street 1, 1234 AB Town", leaving out what is missing.
func formatAddress(road, number, postcode, town string) string {
	street := strings.TrimSpace(road + " " + number)
	place := strings.TrimSpace(FormatPostcode(postcode) + " " + town)
	switch {
	case street == "":
		return place
	case place == "":
		return street
	}
	return street + ", " + place
}
