package locations

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// LineReader reads lines in the background so a read can be abandoned when the
// context is cancelled, e.g. on Ctrl+C.
type LineReader struct {
	lines chan string
	done  chan struct{}
	err   error
}

// NewLineReader starts reading lines from r.
func NewLineReader(r io.Reader) *LineReader {
	l := &LineReader{lines: make(chan string), done: make(chan struct{})}
	go func() {
		s := bufio.NewScanner(r)
		for s.Scan() {
			l.lines <- s.Text()
		}
		l.err = s.Err()
		close(l.done)
	}()
	return l
}

// ReadLine returns the next line. At the end of the input it returns io.EOF,
// and when ctx is cancelled it returns ctx.Err().
func (l *LineReader) ReadLine(ctx context.Context) (string, error) {
	select {
	case s := <-l.lines:
		return strings.TrimSpace(s), nil
	case <-l.done:
		if l.err != nil {
			return "", l.err
		}
		return "", io.EOF
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// AddressLookup looks up an address the user typed in.
type AddressLookup func(ctx context.Context, street, postcode, town string) (AddressHit, bool, error)

// Dialog asks the user for the address of a venue that no source found.
type Dialog struct {
	In     *LineReader
	Out    io.Writer
	Lookup AddressLookup
}

// Ask runs the questions for one venue. ok is false when the user skipped it.
// An error means the input ended or was interrupted; the caller should stop
// asking and keep what was confirmed so far.
func (d *Dialog) Ask(ctx context.Context, venue string, matches int) (Entry, bool, error) {
	fmt.Fprintf(d.Out, "\n%s (%d match(es)) was not found automatically.\n", venue, matches)

	street, err := d.ask(ctx, "  Street and number (empty to skip): ")
	if err != nil || street == "" {
		return Entry{}, false, err
	}
	postcode, err := d.ask(ctx, "  Postcode: ")
	if err != nil {
		return Entry{}, false, err
	}
	name, rawTown, split := SplitVenue(venue)
	if !split {
		name = venue
	}
	defTown := TownTitle(rawTown)
	prompt := "  Town: "
	if defTown != "" {
		prompt = fmt.Sprintf("  Town [%s]: ", defTown)
	}
	town, err := d.ask(ctx, prompt)
	if err != nil {
		return Entry{}, false, err
	}
	if town == "" {
		town = defTown
	}

	// The address stored is always what the user typed; OpenStreetMap only
	// supplies the coordinates.
	e := Entry{Name: name, Address: formatAddress(street, "", postcode, town), Source: "manual"}

	hit, found, lerr := d.Lookup(ctx, street, postcode, town)
	switch {
	case lerr != nil && ctx.Err() != nil:
		return Entry{}, false, ctx.Err()
	case lerr != nil:
		fmt.Fprintf(d.Out, "  Could not look up the address: %v\n", lerr)
	case found:
		fmt.Fprintf(d.Out, "  Found: %s\n", hit.Display)
		use, err := d.confirm(ctx, "  Use this? [Y/n] ", true)
		if err != nil {
			return Entry{}, false, err
		}
		if use {
			e.Lat, e.Lon, e.OSM = hit.Lat, hit.Lon, hit.OSM
			return e, true, nil
		}
	default:
		fmt.Fprintln(d.Out, "  Not found in OpenStreetMap.")
	}

	gps, err := d.confirm(ctx, "  Enter GPS coordinates? [y/N] ", false)
	if err != nil {
		return Entry{}, false, err
	}
	if !gps {
		return e, true, nil
	}
	for {
		s, err := d.ask(ctx, "  Latitude, longitude (empty to skip): ")
		if err != nil {
			return Entry{}, false, err
		}
		if s == "" {
			return e, true, nil
		}
		lat, lon, perr := ParseCoordinates(s)
		if perr != nil {
			fmt.Fprintf(d.Out, "  %v\n", perr)
			continue
		}
		if !nearNetherlands(lat, lon) {
			anyway, err := d.confirm(ctx, "  That point lies outside the Netherlands and its neighbours. Use it anyway? [y/N] ", false)
			if err != nil {
				return Entry{}, false, err
			}
			if !anyway {
				continue
			}
		}
		e.Lat, e.Lon = lat, lon
		return e, true, nil
	}
}

func (d *Dialog) ask(ctx context.Context, prompt string) (string, error) {
	fmt.Fprint(d.Out, prompt)
	s, err := d.In.ReadLine(ctx)
	if err != nil {
		fmt.Fprintln(d.Out)
	}
	return s, err
}

func (d *Dialog) confirm(ctx context.Context, prompt string, def bool) (bool, error) {
	for {
		s, err := d.ask(ctx, prompt)
		if err != nil {
			return false, err
		}
		switch strings.ToLower(s) {
		case "":
			return def, nil
		case "y", "yes", "j", "ja":
			return true, nil
		case "n", "no", "nee":
			return false, nil
		}
		fmt.Fprintln(d.Out, "  Please answer y or n.")
	}
}

// ParseCoordinates reads "lat, lon" or "lat lon" in decimal degrees.
func ParseCoordinates(s string) (lat, lon float64, err error) {
	fields := strings.Fields(strings.ReplaceAll(s, ",", " "))
	if len(fields) != 2 {
		return 0, 0, errors.New("enter two numbers, latitude then longitude, e.g. 52.2870, 5.5965")
	}
	lat, err1 := strconv.ParseFloat(fields[0], 64)
	lon, err2 := strconv.ParseFloat(fields[1], 64)
	if err1 != nil || err2 != nil {
		return 0, 0, errors.New("enter two numbers with a decimal point, e.g. 52.2870, 5.5965")
	}
	if lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("latitude %v is out of range: it must lie between -90 and 90", lat)
	}
	if lon < -180 || lon > 180 {
		return 0, 0, fmt.Errorf("longitude %v is out of range: it must lie between -180 and 180", lon)
	}
	return lat, lon, nil
}

// nearNetherlands is a rough box around the Netherlands and its neighbours,
// to catch swapped or mistyped coordinates.
func nearNetherlands(lat, lon float64) bool {
	return lat >= 50 && lat <= 54 && lon >= 3 && lon <= 8
}

// ErrInputEnded tells whether err means the user ended or interrupted input.
func ErrInputEnded(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, context.Canceled)
}
