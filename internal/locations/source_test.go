package locations

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSplitVenue(t *testing.T) {
	cases := []struct {
		in, name, town string
		ok             bool
	}{
		{"Dindoa-hal ERMELO", "Dindoa-hal", "ERMELO", true},
		{"Hoenderdaal DRIEBERGEN-RIJSENBURG", "Hoenderdaal", "DRIEBERGEN-RIJSENBURG", true},
		{"Topsportcentrum De Koog KOOG AAN DE ZAAN", "Topsportcentrum De Koog", "KOOG AAN DE ZAAN", true},
		{"'t Dok DRONTEN", "'t Dok", "DRONTEN", true},
		{"FBK hal ALMERE", "FBK hal", "ALMERE", true},
		{"DVO BENNEKOM", "DVO", "BENNEKOM", true},
		{"Sporthal Zonder Plaats", "", "", false},
		{"ERMELO", "", "", false},
	}
	for _, c := range cases {
		name, town, ok := SplitVenue(c.in)
		if name != c.name || town != c.town || ok != c.ok {
			t.Errorf("SplitVenue(%q) = %q, %q, %v; want %q, %q, %v", c.in, name, town, ok, c.name, c.town, c.ok)
		}
	}
}

func TestTownTitle(t *testing.T) {
	for in, want := range map[string]string{
		"ERMELO":                "Ermelo",
		"DRIEBERGEN-RIJSENBURG": "Driebergen-Rijsenburg",
		"KOOG AAN DE ZAAN":      "Koog Aan De Zaan",
	} {
		if got := TownTitle(in); got != want {
			t.Errorf("TownTitle(%q) = %q, want %q", in, got, want)
		}
	}
}

func repoServer(t *testing.T, status int, body string) *RepoSource {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ua := r.Header.Get("User-Agent"); ua == "" {
			t.Error("request without User-Agent")
		}
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return &RepoSource{URL: srv.URL, Client: srv.Client()}
}

func TestRepoSource(t *testing.T) {
	src := repoServer(t, 200, `{"version":1,"locations":{
		"De Reeve KAMPEN": {"name":"De Reeve","address":"Horstsingel 1, 8261 DB Kampen","source":"manual"},
		"lege hal plaats": {"name":"","address":""},
		"andere hal elders": {"name":"Andere hal","address":"Weg 1"}
	}}`)

	found, reasons, err := src.Find(context.Background(), []string{"De Reeve KAMPEN", "Lege hal PLAATS", "Onbekend DORP"})
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if e, ok := found["de reeve kampen"]; !ok || e.Address != "Horstsingel 1, 8261 DB Kampen" || e.Source != "manual" {
		t.Errorf("hit not taken over unchanged: %+v", found)
	}
	if _, ok := found["lege hal plaats"]; ok {
		t.Error("empty entry was taken over")
	}
	if _, ok := found["andere hal elders"]; ok {
		t.Error("entry for a venue outside the programme was taken over")
	}
	if len(found) != 1 {
		t.Errorf("found %d entries, want 1", len(found))
	}
	if reasons["onbekend dorp"] == "" || reasons["lege hal plaats"] == "" {
		t.Errorf("missing reasons: %v", reasons)
	}
}

func TestRepoSourceFailures(t *testing.T) {
	for name, src := range map[string]*RepoSource{
		"404":          repoServer(t, 404, "not found"),
		"invalid json": repoServer(t, 200, "{nope"),
	} {
		found, _, err := src.Find(context.Background(), []string{"De Reeve KAMPEN"})
		if err == nil {
			t.Errorf("%s: want error", name)
		}
		if len(found) != 0 {
			t.Errorf("%s: found %v, want nothing", name, found)
		}
	}
}
