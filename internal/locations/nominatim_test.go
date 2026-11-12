package locations

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// Trimmed real-shaped Nominatim answers.
const (
	hitKoploper  = `{"osm_type":"way","osm_id":47261671,"lat":"52.5085","lon":"5.4720","category":"leisure","type":"sports_centre","name":"Sportcentrum De Koploper","display_name":"Sportcentrum De Koploper, Badweg 21, Lelystad","address":{"road":"Badweg","house_number":"21","postcode":"8224EJ","city":"Lelystad"}}`
	hitStreet    = `{"osm_type":"way","osm_id":1,"lat":"52.29","lon":"5.60","category":"highway","type":"residential","name":"Dindoastraat","display_name":"Dindoastraat, Ermelo","address":{"road":"Dindoastraat","town":"Ermelo"}}`
	hitElsewhere = `{"osm_type":"node","osm_id":2,"lat":"52.10","lon":"5.10","category":"leisure","type":"sports_hall","name":"De Voorde","display_name":"De Voorde, Utrecht","address":{"road":"Laan","city":"Utrecht"}}`
	hitFarA      = `{"osm_type":"node","osm_id":3,"lat":"52.3000","lon":"5.6000","category":"leisure","type":"sports_hall","name":"Oost","display_name":"Oost","address":{"town":"Nijkerk"}}`
	hitFarB      = `{"osm_type":"node","osm_id":4,"lat":"52.2300","lon":"5.4800","category":"leisure","type":"sports_hall","name":"Oost 2","display_name":"Oost 2","address":{"town":"Nijkerk"}}`
	hitCentre    = `{"osm_type":"way","osm_id":5,"lat":"52.22690","lon":"5.47500","category":"leisure","type":"sports_centre","name":"Sportpark Watergoor","display_name":"Sportpark Watergoor","address":{"road":"Watergoorweg","postcode":"3861MA","town":"Nijkerk"}}`
	hitHall      = `{"osm_type":"node","osm_id":6,"lat":"52.22697","lon":"5.47504","category":"leisure","type":"sports_hall","name":"Sporthal Watergoor","display_name":"Sporthal Watergoor","address":{"road":"Watergoorweg","house_number":"44","postcode":"3861MA","town":"Nijkerk"}}`
)

// nominatimFake answers by the q parameter; unknown queries get [].
func nominatimFake(t *testing.T, answers map[string]string) (*NominatimSource, *int32) {
	t.Helper()
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		if !strings.HasPrefix(r.Header.Get("User-Agent"), "Dindoa-ICS-Generator/") {
			t.Errorf("User-Agent = %q", r.Header.Get("User-Agent"))
		}
		if strings.Contains(r.Header.Get("User-Agent"), "@") {
			t.Error("User-Agent contains an e-mail address")
		}
		q := r.URL.Query().Get("q")
		if q == "" {
			q = "street=" + r.URL.Query().Get("street")
		}
		body, ok := answers[q]
		if !ok {
			body = "[]"
		}
		if body == "429" {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return &NominatimSource{BaseURL: srv.URL, Client: srv.Client(), Interval: 0}, &calls
}

func TestNominatimMatchRules(t *testing.T) {
	src, _ := nominatimFake(t, map[string]string{
		"Sportcentrum De Koploper, Lelystad": "[" + hitKoploper + "]",
		"Dindoa-hal, Ermelo":                 "[" + hitStreet + "]",
		"De Voorde, Voorthuizen":             "[" + hitElsewhere + "]",
		"Oost, Nijkerk":                      "[" + hitFarA + "," + hitFarB + "]",
		"Watergoor, Nijkerk":                 "[" + hitCentre + "," + hitHall + "]",
	})
	venues := []string{
		"Sportcentrum De Koploper LELYSTAD",
		"Dindoa-hal ERMELO",
		"De Voorde VOORTHUIZEN",
		"Oost NIJKERK",
		"Watergoor NIJKERK",
	}
	found, reasons, err := src.Find(context.Background(), venues)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}

	k := found["sportcentrum de koploper lelystad"]
	if k.Address != "Badweg 21, 8224 EJ Lelystad" || k.OSM != "way/47261671" || k.Source != "osm-auto" || k.Lat == 0 {
		t.Errorf("unambiguous hit = %+v", k)
	}
	for key, want := range map[string]string{
		"dindoa hal ermelo":     "no sports venue found in Ermelo",
		"de voorde voorthuizen": "no sports venue found in Voorthuizen",
		"oost nijkerk":          "ambiguous: 2 sports venues in Nijkerk",
	} {
		if _, ok := found[key]; ok {
			t.Errorf("%s: taken over, want not found", key)
		}
		if reasons[key] != want {
			t.Errorf("%s: reason %q, want %q", key, reasons[key], want)
		}
	}
	w := found["watergoor nijkerk"]
	if w.OSM != "node/6" || w.Address != "Watergoorweg 44, 3861 MA Nijkerk" {
		t.Errorf("hall plus centre within 150 m: got %+v, want the sports hall", w)
	}
}

func TestNominatimSecondAttempt(t *testing.T) {
	src, calls := nominatimFake(t, map[string]string{
		"sporthal Sportcentrum De Koploper, Lelystad": "[" + hitKoploper + "]",
	})
	found, _, err := src.Find(context.Background(), []string{"Sportcentrum De Koploper LELYSTAD"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := found["sportcentrum de koploper lelystad"]; !ok {
		t.Error("second attempt with 'sporthal' was not used")
	}
	if *calls != 2 {
		t.Errorf("%d requests, want 2", *calls)
	}
}

func TestNominatimRateLimit(t *testing.T) {
	src, calls := nominatimFake(t, nil)
	src.Interval = 1100 * time.Millisecond
	clock := time.Date(2026, 11, 12, 20, 0, 0, 0, time.UTC)
	var sleeps []time.Duration
	src.now = func() time.Time { return clock }
	src.sleep = func(_ context.Context, d time.Duration) error {
		sleeps = append(sleeps, d)
		clock = clock.Add(d)
		return nil
	}

	src.Find(context.Background(), []string{"A ERMELO", "B PUTTEN"})
	// Two venues, two attempts each: four requests, three waits in between.
	if *calls != 4 {
		t.Fatalf("%d requests, want 4", *calls)
	}
	if len(sleeps) != 3 {
		t.Fatalf("%d waits, want 3", len(sleeps))
	}
	for _, d := range sleeps {
		if d < src.Interval {
			t.Errorf("waited %v between requests, want at least %v", d, src.Interval)
		}
	}
}

func TestNominatimStopsOnError(t *testing.T) {
	src, calls := nominatimFake(t, map[string]string{"A, Ermelo": "429"})
	found, reasons, err := src.Find(context.Background(), []string{"A ERMELO", "B PUTTEN", "C NIJKERK"})
	if err == nil {
		t.Fatal("want error after HTTP 429")
	}
	if *calls != 1 {
		t.Errorf("%d requests, want no more after the 429", *calls)
	}
	if len(found) != 0 {
		t.Errorf("found %v", found)
	}
	for _, k := range []string{"a ermelo", "b putten", "c nijkerk"} {
		if !strings.HasPrefix(reasons[k], "not looked up") {
			t.Errorf("%s: reason %q, want not looked up", k, reasons[k])
		}
	}
	// A later address lookup does not contact Nominatim again either.
	if _, ok, err := src.LookupAddress(context.Background(), "Weg 1", "1234AB", "Ermelo"); ok || err == nil {
		t.Error("LookupAddress after a stop: want error")
	}
	if *calls != 1 {
		t.Errorf("%d requests after the stop", *calls)
	}
}

func TestLookupAddress(t *testing.T) {
	src, _ := nominatimFake(t, map[string]string{
		"street=Watergoorweg 44": "[" + hitHall + "]",
	})
	ctx := context.Background()

	hit, ok, err := src.LookupAddress(ctx, "Watergoorweg 44", "3861 ma", "Nijkerk")
	if err != nil || !ok || hit.OSM != "node/6" {
		t.Errorf("matching postcode: %+v %v %v", hit, ok, err)
	}
	if _, ok, _ := src.LookupAddress(ctx, "Watergoorweg 44", "3861 XX", "Nijkerk"); ok {
		t.Error("other postcode: want not found")
	}
	if _, ok, _ := src.LookupAddress(ctx, "Nergensweg 1", "1234 AB", "Nijkerk"); ok {
		t.Error("no hit: want not found")
	}
}

func TestFormatPostcode(t *testing.T) {
	for in, want := range map[string]string{"3853pt": "3853 PT", "3853 PT": "3853 PT", "B-1000": "B-1000", "": ""} {
		if got := FormatPostcode(in); got != want {
			t.Errorf("FormatPostcode(%q) = %q, want %q", in, got, want)
		}
	}
}
