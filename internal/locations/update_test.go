package locations

import (
	"context"
	"errors"
	"testing"
)

type fakeSource struct {
	name    string
	entries map[string]Entry // by normalized key
	err     error
	asked   [][]string
}

func (f *fakeSource) Name() string { return f.name }

func (f *fakeSource) Find(_ context.Context, venues []string) (map[string]Entry, map[string]string, error) {
	f.asked = append(f.asked, venues)
	if f.err != nil {
		return nil, nil, f.err
	}
	found, reasons := map[string]Entry{}, map[string]string{}
	for _, v := range venues {
		k := NormalizeKey(v)
		if e, ok := f.entries[k]; ok {
			found[k] = e
		} else {
			reasons[k] = f.name + " does not know it"
		}
	}
	return found, reasons, nil
}

func TestUpdateRepoBeforeOSM(t *testing.T) {
	repo := &fakeSource{name: "repository", entries: map[string]Entry{
		"de reeve kampen": {Name: "De Reeve", Address: "Horstsingel 1", Source: "manual"},
	}}
	osm := &fakeSource{name: "OpenStreetMap", entries: map[string]Entry{
		"de reeve kampen":   {Name: "Wrong", Address: "Elders"},
		"watergoor nijkerk": {Name: "Watergoor", Address: "Watergoorweg 44", Source: "osm-auto"},
	}}
	user := &File{Version: 1}

	res := Update(context.Background(), []string{"De Reeve KAMPEN", "Watergoor NIJKERK", "Onbekend DORP"}, user, repo, osm)

	if user.Locations["de reeve kampen"].Address != "Horstsingel 1" {
		t.Errorf("repository entry not used: %+v", user.Locations["de reeve kampen"])
	}
	if len(osm.asked) != 1 || len(osm.asked[0]) != 2 {
		t.Errorf("OpenStreetMap was asked %v, want only the two venues the repository lacked", osm.asked)
	}
	if user.Locations["watergoor nijkerk"].Source != "osm-auto" {
		t.Error("OpenStreetMap entry not added")
	}
	if len(res.Added) != 2 || res.Added[0].Source != "repository" || res.Added[1].Source != "OpenStreetMap" {
		t.Errorf("Added = %+v", res.Added)
	}
	if len(res.Missing) != 1 || res.Missing[0].Reason != "OpenStreetMap does not know it" {
		t.Errorf("Missing = %+v", res.Missing)
	}
}

func TestUpdateNeverOverwritesUserEntries(t *testing.T) {
	mine := Entry{Name: "Mijn correctie", Address: "Eigen adres 1"}
	user := &File{Version: 1, Locations: map[string]Entry{
		"de reeve kampen": mine,
		"lege hal dorp":   {},
	}}
	src := &fakeSource{name: "repository", entries: map[string]Entry{
		"de reeve kampen": {Name: "Other", Address: "Other 1"},
		"lege hal dorp":   {Name: "Filled", Address: "Weg 2"},
	}}

	res := Update(context.Background(), []string{"De Reeve KAMPEN", "Lege hal DORP"}, user, src)

	if user.Locations["de reeve kampen"] != mine {
		t.Errorf("user entry overwritten: %+v", user.Locations["de reeve kampen"])
	}
	if !user.Locations["lege hal dorp"].IsEmpty() {
		t.Error("empty user placeholder was filled")
	}
	if len(res.Added) != 0 {
		t.Errorf("Added = %+v, want nothing", res.Added)
	}
	if len(src.asked) != 0 {
		t.Errorf("source asked %v, want no lookup for keys the user already has", src.asked)
	}
}

func TestUpdateFailingRepoFallsThrough(t *testing.T) {
	repo := &fakeSource{name: "repository", err: errors.New("HTTP 503")}
	osm := &fakeSource{name: "OpenStreetMap", entries: map[string]Entry{
		"watergoor nijkerk": {Name: "Watergoor", Address: "Watergoorweg 44"},
	}}
	user := &File{Version: 1}

	res := Update(context.Background(), []string{"Watergoor NIJKERK"}, user, repo, osm)

	if len(osm.asked) != 1 {
		t.Fatal("OpenStreetMap not asked after the repository failed")
	}
	if _, ok := user.Locations["watergoor nijkerk"]; !ok {
		t.Error("entry from OpenStreetMap not added")
	}
	if len(res.Warnings) != 1 {
		t.Errorf("Warnings = %v, want the repository failure", res.Warnings)
	}
}

func TestUpdateNothingMissing(t *testing.T) {
	src := &fakeSource{name: "repository"}
	res := Update(context.Background(), nil, &File{Version: 1}, src)
	if len(src.asked) != 0 {
		t.Error("source called with nothing to look up")
	}
	if len(res.Added)+len(res.Missing) != 0 {
		t.Errorf("result = %+v", res)
	}
}
