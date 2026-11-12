package locations

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

type lookupCall struct{ street, postcode, town string }

func dialog(input string, hit *AddressHit, lerr error) (*Dialog, *[]lookupCall, *bytes.Buffer) {
	var calls []lookupCall
	out := &bytes.Buffer{}
	d := &Dialog{
		In:  NewLineReader(strings.NewReader(input)),
		Out: out,
		Lookup: func(_ context.Context, street, postcode, town string) (AddressHit, bool, error) {
			calls = append(calls, lookupCall{street, postcode, town})
			if lerr != nil {
				return AddressHit{}, false, lerr
			}
			if hit == nil {
				return AddressHit{}, false, nil
			}
			return *hit, true, nil
		},
	}
	return d, &calls, out
}

var watergoor = &AddressHit{Lat: 52.2269706, Lon: 5.4750352, OSM: "node/2718958165", Display: "Optisport Sporthal Watergoor, 44, Watergoorweg, Nijkerk"}

func TestDialogSkip(t *testing.T) {
	d, calls, _ := dialog("\n", nil, nil)
	_, ok, err := d.Ask(context.Background(), "Watergoor NIJKERK", 5)
	if ok || err != nil {
		t.Fatalf("skip: ok=%v err=%v", ok, err)
	}
	if len(*calls) != 0 {
		t.Error("looked up an address after a skip")
	}
}

func TestDialogAddressFound(t *testing.T) {
	// Town left empty: the town from the venue string is used.
	d, calls, _ := dialog("Watergoorweg 44\n3861ma\n\n\n", watergoor, nil)
	e, ok, err := d.Ask(context.Background(), "Watergoor NIJKERK", 5)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if (*calls)[0] != (lookupCall{"Watergoorweg 44", "3861ma", "Nijkerk"}) {
		t.Errorf("lookup = %+v", (*calls)[0])
	}
	want := Entry{Name: "Watergoor", Address: "Watergoorweg 44, 3861 MA Nijkerk", Lat: watergoor.Lat, Lon: watergoor.Lon, OSM: watergoor.OSM, Source: "manual"}
	if e != want {
		t.Errorf("entry = %+v\nwant    %+v", e, want)
	}
}

func TestDialogRejectedThenGPS(t *testing.T) {
	d, _, _ := dialog("Watergoorweg 44\n3861 MA\nNijkerk\nn\ny\n52.2270, 5.4750\n", watergoor, nil)
	e, ok, err := d.Ask(context.Background(), "Watergoor NIJKERK", 5)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if e.Lat != 52.2270 || e.Lon != 5.4750 || e.OSM != "" {
		t.Errorf("entry = %+v, want typed coordinates and no OSM reference", e)
	}
}

func TestDialogNotFoundNoGPS(t *testing.T) {
	d, _, out := dialog("Nergensweg 1\n1234 AB\nErmelo\n\n", nil, nil)
	e, ok, err := d.Ask(context.Background(), "Dindoa-hal ERMELO", 173)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if e.Address != "Nergensweg 1, 1234 AB Ermelo" || e.HasCoordinates() || e.Source != "manual" {
		t.Errorf("entry = %+v, want the typed address only", e)
	}
	if !strings.Contains(out.String(), "Not found in OpenStreetMap") {
		t.Errorf("output does not say the address was not found:\n%s", out)
	}
}

func TestDialogInvalidCoordinatesAskedAgain(t *testing.T) {
	input := "Weg 1\n1234 AB\nErmelo\ny\nabc\n95, 5\n40.7, -74.0\nn\n52.29 5.60\n"
	d, _, out := dialog(input, nil, nil)
	e, ok, err := d.Ask(context.Background(), "Hal ERMELO", 1)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if e.Lat != 52.29 || e.Lon != 5.60 {
		t.Errorf("coordinates = %v, %v", e.Lat, e.Lon)
	}
	for _, s := range []string{"enter two numbers", "latitude 95 is out of range", "outside the Netherlands"} {
		if !strings.Contains(out.String(), s) {
			t.Errorf("output lacks %q:\n%s", s, out)
		}
	}
}

func TestDialogLookupFailureOffersGPS(t *testing.T) {
	d, _, out := dialog("Weg 1\n1234 AB\nErmelo\ny\n52.29, 5.60\n", nil, errors.New("HTTP 429"))
	e, ok, err := d.Ask(context.Background(), "Hal ERMELO", 1)
	if err != nil || !ok || e.Lat != 52.29 {
		t.Fatalf("e=%+v ok=%v err=%v", e, ok, err)
	}
	if !strings.Contains(out.String(), "Could not look up") {
		t.Errorf("failure not reported:\n%s", out)
	}
}

func TestDialogEOFMidway(t *testing.T) {
	d, _, _ := dialog("Weg 1\n", nil, nil)
	_, ok, err := d.Ask(context.Background(), "Hal ERMELO", 1)
	if ok || !errors.Is(err, io.EOF) || !ErrInputEnded(err) {
		t.Fatalf("ok=%v err=%v, want io.EOF", ok, err)
	}
}

func TestDialogCancelled(t *testing.T) {
	r, _ := io.Pipe() // never delivers a line
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	d := &Dialog{In: NewLineReader(r), Out: io.Discard}
	_, ok, err := d.Ask(ctx, "Hal ERMELO", 1)
	if ok || !ErrInputEnded(err) {
		t.Fatalf("ok=%v err=%v, want cancellation", ok, err)
	}
}

func TestParseCoordinates(t *testing.T) {
	good := map[string][2]float64{
		"52.29, 5.60": {52.29, 5.60},
		"52.29 5.60":  {52.29, 5.60},
		" -33.9,18.4": {-33.9, 18.4},
	}
	for in, want := range good {
		lat, lon, err := ParseCoordinates(in)
		if err != nil || lat != want[0] || lon != want[1] {
			t.Errorf("ParseCoordinates(%q) = %v, %v, %v", in, lat, lon, err)
		}
	}
	for _, in := range []string{"", "52.29", "a, b", "91, 5", "52, 181", "1 2 3"} {
		if _, _, err := ParseCoordinates(in); err == nil {
			t.Errorf("ParseCoordinates(%q): want error", in)
		}
	}
}
