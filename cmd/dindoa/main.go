package main

import (
	"flag"
	"fmt"
	"math"
	"os"
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

	if len(args) > 0 && args[0] == "start" {
		p := ui.NewInteractiveApp()
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

func handleListLocations() error {
	prog, err := loadProgramma()
	if err != nil {
		return err
	}
	mapping, err := loadMapping()
	if err != nil {
		return err
	}

	counts := prog.LocationCounts()
	venues := make([]string, 0, len(counts))
	total := 0
	for v, n := range counts {
		venues = append(venues, v)
		total += n
	}
	// Most matches first, so the venue worth fixing shows up at the top.
	sort.Slice(venues, func(i, j int) bool {
		if counts[venues[i]] != counts[venues[j]] {
			return counts[venues[i]] > counts[venues[j]]
		}
		return venues[i] < venues[j]
	})

	fmt.Printf("Venues in the match programme (%d venues, %d matches):\n\n", len(venues), total)

	var missing []string
	mappedVenues, mappedMatches := 0, 0
	for _, v := range venues {
		entry, ok := mapping.Lookup(v)
		if ok {
			mappedVenues++
			mappedMatches += counts[v]
			detail := entry.Address
			if detail == "" {
				detail = "(coordinates only)"
			}
			fmt.Printf("  ✓ %4dx  %-42s %s\n", counts[v], v, detail)
			continue
		}
		missing = append(missing, v)
		fmt.Printf("  ✗ %4dx  %-42s — not in the address list\n", counts[v], v)
	}

	// Round down, so the summary never claims 100% while a venue is missing.
	pct := 0.0
	if total > 0 {
		pct = math.Floor(100 * float64(mappedMatches) / float64(total))
	}
	fmt.Printf("\n%d/%d venues mapped (%d/%d matches = %.0f%%)\n",
		mappedVenues, len(venues), mappedMatches, total, pct)
	fmt.Printf("Your address list: %s\n", mapping.UserPath())
	if !mapping.UserFileLoaded {
		fmt.Println("  (does not exist yet; create it to add or correct venues)")
	}

	if len(missing) > 0 {
		fmt.Printf("\nPaste this into %s and fill in the blanks:\n\n%s\n",
			mapping.UserPath(), locations.Skeleton(missing))
	}

	return nil
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
	fmt.Printf("  Add them to %s — run 'dindoa --list-locations' for a fragment to paste.\n", userPath)
}

func printUsage() {
	fmt.Println("Dindoa ICS Generator - Generate calendar files for Dindoa korfbal team matches")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dindoa                                    Show this help message")
	fmt.Println("  dindoa start                              Start interactive mode")
	fmt.Println("  dindoa --list-categories                  List all categories")
	fmt.Println("  dindoa --category <name> --list-teams     List teams in category")
	fmt.Println("  dindoa --list-all-teams                   List all teams by category")
	fmt.Println("  dindoa --list-locations                   List venues and their address status")
	fmt.Println("  dindoa --team <name>                      Generate ICS for team")
	fmt.Println("  dindoa --team <name> --output <file>      Generate ICS with custom filename")
	fmt.Println("  dindoa --help                             Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  dindoa start                              Start interactive menu")
	fmt.Println("  dindoa --team j3                          Generate dindoa-j3.ics")
	fmt.Println("  dindoa --team j3 --output eigen.ics       Generate with a custom filename")
	fmt.Println("  dindoa --category rood --list-teams       List teams in Rood category")
	fmt.Println("  dindoa --list-locations                   See which venues need an address")
	fmt.Println()
	fmt.Printf("Match data comes from %s\n", scraper.ProgrammaURL())
	fmt.Printf("Venue addresses come from %s (optional; built-in list is used otherwise)\n",
		locations.UserFilePath())
	fmt.Println()
	fmt.Println("Options (a single leading dash also works, e.g. -team):")
	flag.PrintDefaults()
}
