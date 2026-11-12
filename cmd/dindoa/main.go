package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

	"github.com/torreirow/dindoa/internal/ics"
	"github.com/torreirow/dindoa/internal/locations"
	"github.com/torreirow/dindoa/internal/scraper"
	"github.com/torreirow/dindoa/internal/ui"
)

var (
	listCategories = flag.Bool("list-categories", false, "List all team categories")
	category       = flag.String("category", "", "Filter by category")
	listTeams      = flag.Bool("list-teams", false, "List teams in a category")
	listAllTeams   = flag.Bool("list-all-teams", false, "List all teams sorted by category")
	listLocations  = flag.Bool("list-locations", false, "List all venues with their address mapping status")
	updateLocs     = flag.Bool("update-locations", false, "Add missing venues to your address list (uses the network)")
	team           = flag.String("team", "", "Team to generate ICS for")
	output         = flag.String("output", "", "Custom output filename")
	help           = flag.Bool("help", false, "Show usage information")
	helpShort      = flag.Bool("h", false, "Show usage information")
)

func main() {
	flag.Parse()

	if *help || *helpShort {
		printUsage()
		os.Exit(0)
	}

	args := flag.Args()

	// --update-locations writes a file; refuse to guess what a combination
	// with another command was meant to do.
	if *updateLocs {
		if *team != "" || *output != "" || *category != "" || *listTeams ||
			*listCategories || *listAllTeams || *listLocations || len(args) > 0 {
			fmt.Fprintln(os.Stderr, "Error: --update-locations cannot be combined with other commands or flags")
			os.Exit(2)
		}
		run(handleUpdateLocations)
	}

	if len(args) > 0 && args[0] == "start" {
		p := ui.NewInteractiveApp(*category)
		if _, err := p.Run(); err != nil {
			fail(err)
		}
		os.Exit(0)
	}

	switch {
	case *listCategories:
		run(handleListCategories)
	case *category != "" && *listTeams:
		run(func() error { return handleListTeamsInCategory(*category) })
	case *listAllTeams:
		run(handleListAllTeams)
	case *listLocations:
		run(handleListLocations)
	case *team != "":
		run(func() error { return handleGenerateICS(*team, *output) })
	case *category != "":
		// A category on its own launches the interactive UI with that category
		// pre-selected.
		p := ui.NewInteractiveApp(*category)
		if _, err := p.Run(); err != nil {
			fail(err)
		}
		os.Exit(0)
	case *output != "":
		fmt.Fprintf(os.Stderr, "Error: --output requires --team flag\n")
		os.Exit(1)
	default:
		printUsage()
		os.Exit(0)
	}
}

func run(fn func() error) {
	if err := fn(); err != nil {
		fail(err)
	}
	os.Exit(0)
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

// loadProgramma fetches and parses the match programme, the single source for
// matches, teams, categories and colours.
func loadProgramma() (*scraper.Programma, error) {
	doc, err := scraper.NewFetcher().FetchProgrammaPage()
	if err != nil {
		return nil, fmt.Errorf("fetch match programme: %w", err)
	}
	prog, err := scraper.NewParser().ParseProgramma(doc, time.Now())
	if err != nil {
		return nil, fmt.Errorf("parse match programme (%s): %w", scraper.ProgrammaURL(), err)
	}
	return prog, nil
}

// loadMapping loads the location mapping, printing any non-fatal warnings.
func loadMapping() (*locations.Mapping, error) {
	m, err := locations.Load()
	if err != nil {
		return nil, err
	}
	for _, w := range m.Warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", w)
	}
	return m, nil
}

func handleListCategories() error {
	prog, err := loadProgramma()
	if err != nil {
		return err
	}
	for _, name := range prog.CategoryNames() {
		fmt.Println(name)
	}
	return nil
}

func handleListTeamsInCategory(categoryName string) error {
	prog, err := loadProgramma()
	if err != nil {
		return err
	}

	want := strings.ToLower(strings.TrimSpace(categoryName))
	for _, cat := range prog.Categories() {
		if strings.ToLower(cat.Name) == want {
			for _, t := range cat.Teams {
				fmt.Println(t.Name)
			}
			return nil
		}
	}

	return fmt.Errorf("category %q not found; available: %s",
		categoryName, strings.Join(prog.CategoryNames(), ", "))
}

func handleListAllTeams() error {
	prog, err := loadProgramma()
	if err != nil {
		return err
	}
	for _, cat := range prog.Categories() {
		fmt.Printf("\n%s:\n", cat.Name)
		for _, t := range cat.Teams {
			fmt.Printf("  %s\n", t.Name)
		}
	}
	return nil
}

// coverage is how far the mapping covers the venues in the programme.
type coverage struct {
	venues        []string // most matches first
	counts        map[string]int
	missing       []string // in the same order as venues
	mappedVenues  int
	mappedMatches int
	totalMatches  int
}

func computeCoverage(prog *scraper.Programma, mapping *locations.Mapping) coverage {
	c := coverage{counts: prog.LocationCounts()}
	for v, n := range c.counts {
		c.venues = append(c.venues, v)
		c.totalMatches += n
	}
	// Most matches first, so the venue worth fixing shows up at the top.
	sort.Slice(c.venues, func(i, j int) bool {
		if c.counts[c.venues[i]] != c.counts[c.venues[j]] {
			return c.counts[c.venues[i]] > c.counts[c.venues[j]]
		}
		return c.venues[i] < c.venues[j]
	})
	for _, v := range c.venues {
		if _, ok := mapping.Lookup(v); ok {
			c.mappedVenues++
			c.mappedMatches += c.counts[v]
			continue
		}
		c.missing = append(c.missing, v)
	}
	return c
}

// summary is the coverage line shared by --list-locations and --update-locations.
func (c coverage) summary() string {
	// Round down, so the summary never claims 100% while a venue is missing.
	pct := 0.0
	if c.totalMatches > 0 {
		pct = math.Floor(100 * float64(c.mappedMatches) / float64(c.totalMatches))
	}
	return fmt.Sprintf("%d/%d venues mapped (%d/%d matches = %.0f%%)",
		c.mappedVenues, len(c.venues), c.mappedMatches, c.totalMatches, pct)
}

func handleListLocations() error {
	prog, err := loadProgramma()
	if err != nil {
		return err
	}
	mapping, err := loadMapping()
	if err != nil {
		return err
	}
	c := computeCoverage(prog, mapping)

	fmt.Printf("Venues in the match programme (%d venues, %d matches):\n\n", len(c.venues), c.totalMatches)
	for _, v := range c.venues {
		entry, ok := mapping.Lookup(v)
		if !ok {
			fmt.Printf("  ✗ %4dx  %-42s — not in the address list\n", c.counts[v], v)
			continue
		}
		detail := entry.Address
		if detail == "" {
			detail = "(coordinates only)"
		}
		if entry.Source == "osm-auto" {
			detail += "  [osm-auto]"
		}
		fmt.Printf("  ✓ %4dx  %-42s %s\n", c.counts[v], v, detail)
	}

	fmt.Printf("\n%s\n", c.summary())
	fmt.Printf("Your address list: %s\n", mapping.UserPath())
	if !mapping.UserFileLoaded {
		fmt.Println("  (does not exist yet; create it to add or correct venues)")
	}

	if len(c.missing) > 0 {
		fmt.Printf("\nRun 'dindoa --update-locations' to look them up, or paste this into %s and fill in the blanks:\n\n%s\n",
			mapping.UserPath(), locations.Skeleton(c.missing))
	}
	if hasOSMAuto(mapping, c.venues) {
		fmt.Println("\n[osm-auto] entries were found in OpenStreetMap by name only; worth checking.")
	}

	return nil
}

func hasOSMAuto(mapping *locations.Mapping, venues []string) bool {
	for _, v := range venues {
		if e, ok := mapping.Lookup(v); ok && e.Source == "osm-auto" {
			return true
		}
	}
	return false
}

func handleUpdateLocations() error {
	// Read the user's file first: if it is unreadable, stop before anything
	// goes over the network.
	userPath := locations.UserFilePath()
	userFile, err := locations.ReadUserFile(userPath)
	if err != nil {
		return fmt.Errorf("%w\nFix or move that file, then run --update-locations again", err)
	}

	prog, err := loadProgramma()
	if err != nil {
		return err
	}
	mapping, err := loadMapping()
	if err != nil {
		return err
	}
	before := computeCoverage(prog, mapping)
	if len(before.missing) == 0 {
		fmt.Printf("Nothing to add: all %d venues in the programme have an address.\n", len(before.venues))
		fmt.Println(before.summary())
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fmt.Fprintf(os.Stderr, "%d venue(s) without an address. Checking the reviewed list in the repository...\n", len(before.missing))
	osm := locations.NewNominatimSource()
	osm.Progress = os.Stderr
	repo := locations.NewRepoSource()
	res := locations.Update(ctx, before.missing, userFile, repo, osm)

	// Ask about whatever is still missing, but only when someone can answer.
	var manual []locations.Outcome
	if len(res.Missing) > 0 && ctx.Err() == nil && isTerminal(os.Stdin) && isTerminal(os.Stdout) {
		fmt.Printf("\n%d venue(s) still have no address. Enter them by hand, or press Enter to skip one.\n", len(res.Missing))
		d := &locations.Dialog{In: locations.NewLineReader(os.Stdin), Out: os.Stdout, Lookup: osm.LookupAddress}
		var still []locations.Outcome
		for i, o := range res.Missing {
			if strings.HasPrefix(o.Reason, "your file has an empty entry") {
				still = append(still, o)
				continue
			}
			e, ok, err := d.Ask(ctx, o.Venue, before.counts[o.Venue])
			if err != nil {
				// Input ended or Ctrl+C: keep what was confirmed, ask no more.
				fmt.Println("\nStopped; saving what was entered so far.")
				still = append(still, res.Missing[i:]...)
				break
			}
			if !ok {
				o.Reason = "skipped"
				still = append(still, o)
				continue
			}
			userFile.Locations[o.Key] = e
			o.Source, o.Entry, o.Reason = "entered by hand", e, ""
			manual = append(manual, o)
		}
		res.Missing = still
	}

	added := len(res.Added) + len(manual)
	if added > 0 {
		if err := locations.WriteUserFile(userPath, userFile); err != nil {
			return err
		}
	}

	printUpdateReport(res, manual)

	after := before
	if added > 0 {
		if m, err := locations.Load(); err == nil {
			after = computeCoverage(prog, m)
		}
	}
	fmt.Printf("\n%s\n", after.summary())
	if added > 0 {
		fmt.Printf("Saved %d venue(s) to %s\n", added, userPath)
	} else {
		fmt.Printf("Nothing saved; %s is unchanged.\n", userPath)
	}
	return nil
}

func printUpdateReport(res locations.Result, manual []locations.Outcome) {
	for _, w := range res.Warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", w)
	}
	block := func(title string, list []locations.Outcome) {
		if len(list) == 0 {
			return
		}
		fmt.Printf("\n%s (%d):\n", title, len(list))
		for _, o := range list {
			if o.Reason != "" {
				fmt.Printf("  ✗ %-42s %s\n", o.Venue, o.Reason)
				continue
			}
			fmt.Printf("  ✓ %-42s %s\n", o.Venue, o.Entry.Address)
		}
	}
	var fromRepo, fromOSM []locations.Outcome
	for _, o := range res.Added {
		if o.Source == "repository" {
			fromRepo = append(fromRepo, o)
		} else {
			fromOSM = append(fromOSM, o)
		}
	}
	block("Added from the reviewed list in the repository", fromRepo)
	block("Added from OpenStreetMap (source osm-auto; please check these)", fromOSM)
	block("Entered by hand", manual)
	block("Still missing", res.Missing)
}

// isTerminal reports whether f is an interactive terminal.
func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func handleGenerateICS(teamInput, outputFile string) error {
	teamName := scraper.NormalizeTeamInput(teamInput)
	if teamName == "" {
		return fmt.Errorf("no team name given")
	}

	prog, err := loadProgramma()
	if err != nil {
		return err
	}

	if !prog.HasTeam(teamName) {
		return fmt.Errorf("team %q does not appear in the match programme.\nAvailable teams:\n  %s",
			teamName, strings.Join(prog.TeamNames(), "\n  "))
	}

	matches := prog.MatchesFor(teamName)
	if len(matches) == 0 {
		fmt.Printf("%s has no matches in the published part of the programme.\n", teamName)
		fmt.Printf("The programme at %s is published in blocks; check back later in the season.\n",
			scraper.ProgrammaURL())
		return nil
	}

	mapping, err := loadMapping()
	if err != nil {
		return err
	}

	generator, err := ics.NewGenerator(mapping)
	if err != nil {
		return fmt.Errorf("create ICS generator: %w", err)
	}

	if outputFile == "" {
		outputFile = ics.DefaultOutputFilename(teamName)
	}

	missing, err := generator.Generate(teamName, matches, outputFile)
	if err != nil {
		return fmt.Errorf("generate ICS: %w", err)
	}

	fmt.Printf("\n✓ ICS file created: %s\n", outputFile)
	fmt.Printf("  Team:    %s\n", teamName)
	fmt.Printf("  Matches: %d\n", len(matches))
	printMissingLocations(missing, mapping.UserPath())

	return nil
}

// printMissingLocations reports venues that were not in the mapping. This is
// informational: the file was written and is valid, so the exit code stays 0.
func printMissingLocations(missing map[string]int, userPath string) {
	if len(missing) == 0 {
		return
	}

	venues := make([]string, 0, len(missing))
	for v := range missing {
		venues = append(venues, v)
	}
	sort.Slice(venues, func(i, j int) bool {
		if missing[venues[i]] != missing[venues[j]] {
			return missing[venues[i]] > missing[venues[j]]
		}
		return venues[i] < venues[j]
	})

	fmt.Printf("\n⚠ %d venue(s) not in the address list; the name from the website was used:\n", len(venues))
	for _, v := range venues {
		fmt.Printf("    %-42s (%d match(es))\n", v, missing[v])
	}
	fmt.Printf("  Run 'dindoa --update-locations' to look them up, or add them to %s.\n", userPath)
}

func printUsage() {
	fmt.Println("Dindoa ICS Generator - Generate calendar files for Dindoa korfbal team matches")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dindoa                                    Show this help message")
	fmt.Println("  dindoa start                              Start interactive mode")
	fmt.Println("  dindoa --list-categories                  List all categories")
	fmt.Println("  dindoa --category <name> --list-teams     List teams in category")
	fmt.Println("  dindoa --category <name>                  Interactive mode, category pre-selected")
	fmt.Println("  dindoa --list-all-teams                   List all teams by category")
	fmt.Println("  dindoa --list-locations                   List venues and their address status")
	fmt.Println("  dindoa --update-locations                 Add missing venue addresses (uses the network)")
	fmt.Println("  dindoa --team <name>                      Generate ICS for team")
	fmt.Println("  dindoa --team <name> --output <file>      Generate ICS with custom filename")
	fmt.Println("  dindoa --help                             Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  dindoa start                              Start interactive menu")
	fmt.Println("  dindoa --team j3                          Generate dindoa-j3.ics")
	fmt.Println("  dindoa --team j3 --output eigen.ics       Generate with a custom filename")
	fmt.Println("  dindoa --category rood --list-teams       List teams in Rood category")
	fmt.Println("  dindoa --category rood                    Pick a Rood team interactively")
	fmt.Println("  dindoa --list-locations                   See which venues need an address")
	fmt.Println("  dindoa --update-locations                 Fill your address list from the repository and OpenStreetMap")
	fmt.Println()
	fmt.Printf("Match data comes from %s\n", scraper.ProgrammaURL())
	fmt.Printf("Venue addresses come from %s (optional; built-in list is used otherwise)\n",
		locations.UserFilePath())
	fmt.Println()
	fmt.Println("Options (a single leading dash also works, e.g. -team):")
	flag.PrintDefaults()
}
