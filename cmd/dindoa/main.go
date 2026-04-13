package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/torreirow/dindoa/internal/geocode"
	"github.com/torreirow/dindoa/internal/ics"
	"github.com/torreirow/dindoa/internal/scraper"
	"github.com/torreirow/dindoa/internal/ui"
)

var (
	listCategories = flag.Bool("list-categories", false, "List all team categories")
	category       = flag.String("category", "", "Filter by category")
	listTeams      = flag.Bool("list-teams", false, "List teams in a category")
	listAllTeams   = flag.Bool("list-all-teams", false, "List all teams sorted by category")
	team           = flag.String("team", "", "Team to generate ICS for")
	output         = flag.String("output", "", "Custom output filename")
	help           = flag.Bool("help", false, "Show usage information")
	helpShort      = flag.Bool("h", false, "Show usage information")
)

func main() {
	flag.Parse()

	// Show help
	if *help || *helpShort {
		printUsage()
		os.Exit(0)
	}

	// Check for positional arguments (like "start")
	args := flag.Args()

	// Handle "start" command for interactive mode
	if len(args) > 0 && args[0] == "start" {
		p := ui.NewInteractiveApp()
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Determine mode based on flags
	if *listCategories {
		if err := handleListCategories(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *category != "" && *listTeams {
		if err := handleListTeamsInCategory(*category); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *listAllTeams {
		if err := handleListAllTeams(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *team != "" {
		outputFile := *output
		if outputFile == "" {
			outputFile = ics.DefaultOutputFilename(*team)
		}
		if err := handleGenerateICS(*team, outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Check for --output without --team
	if *output != "" {
		fmt.Fprintf(os.Stderr, "Error: --output requires --team flag\n")
		os.Exit(1)
	}

	// No flags or commands provided - show help
	printUsage()
	os.Exit(0)
}

func handleListCategories() error {
	fetcher := scraper.NewFetcher()
	parser := scraper.NewParser()

	doc, err := fetcher.FetchTeamsPage()
	if err != nil {
		return fmt.Errorf("fetch teams page: %w", err)
	}

	categories, err := parser.ParseCategories(doc)
	if err != nil {
		return fmt.Errorf("parse categories: %w", err)
	}

	for _, cat := range categories {
		fmt.Println(cat)
	}

	return nil
}

func handleListTeamsInCategory(categoryName string) error {
	fetcher := scraper.NewFetcher()
	parser := scraper.NewParser()

	doc, err := fetcher.FetchTeamsPage()
	if err != nil {
		return fmt.Errorf("fetch teams page: %w", err)
	}

	categories, err := parser.ParseTeams(doc)
	if err != nil {
		return fmt.Errorf("parse teams: %w", err)
	}

	// Find matching category (case-insensitive)
	categoryLower := strings.ToLower(categoryName)
	for _, cat := range categories {
		if strings.ToLower(cat.Name) == categoryLower {
			for _, team := range cat.Teams {
				fmt.Println(team.Name)
			}
			return nil
		}
	}

	return fmt.Errorf("category '%s' not found", categoryName)
}

func handleListAllTeams() error {
	fetcher := scraper.NewFetcher()
	parser := scraper.NewParser()

	doc, err := fetcher.FetchTeamsPage()
	if err != nil {
		return fmt.Errorf("fetch teams page: %w", err)
	}

	categories, err := parser.ParseTeams(doc)
	if err != nil {
		return fmt.Errorf("parse teams: %w", err)
	}

	// Sort categories by name
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	for _, cat := range categories {
		fmt.Printf("\n%s:\n", cat.Name)
		for _, team := range cat.Teams {
			fmt.Printf("  %s\n", team.Name)
		}
	}

	return nil
}

func handleGenerateICS(teamName string, outputFile string) error {
	// Normalize team name
	teamSlug := scraper.NormalizeTeamName(teamName)

	// Fetch team page
	fetcher := scraper.NewFetcher()
	parser := scraper.NewParser()

	doc, err := fetcher.FetchTeamPage(teamSlug)
	if err != nil {
		return fmt.Errorf("fetch team page: %w", err)
	}

	// Parse matches
	matches, err := parser.ParseMatches(doc)
	if err != nil {
		return fmt.Errorf("parse matches: %w", err)
	}

	if len(matches) == 0 {
		fmt.Println("No matches found for this team")
		return nil
	}

	// Initialize geocoding
	cache, err := geocode.NewCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not initialize cache: %v\n", err)
		// Continue without cache
	}

	rateLimiter := geocode.NewRateLimiter()
	geocoder := geocode.NewClient(rateLimiter)

	// Geocode locations
	fmt.Printf("Geocoding %d locations...\n", len(matches))
	for i := range matches {
		location := matches[i].Location

		// Check cache first
		if cache != nil {
			if result, found := cache.Lookup(location); found {
				matches[i].Location = result.Address
				continue
			}
		}

		// Geocode
		result := geocoder.Geocode(location)
		matches[i].Location = result.Address

		// Store in cache
		if cache != nil {
			if err := cache.Store(result); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not cache result: %v\n", err)
			}
		}

		fmt.Printf("  ✓ %s\n", location)
	}

	// Generate ICS
	generator, err := ics.NewGenerator()
	if err != nil {
		return fmt.Errorf("create ICS generator: %w", err)
	}

	// Use original team name for display (not slug)
	displayName := teamName
	if !strings.Contains(strings.ToLower(teamName), "dindoa") {
		displayName = "Dindoa " + strings.ToUpper(teamName)
	}

	if err := generator.Generate(displayName, matches, outputFile); err != nil {
		return fmt.Errorf("generate ICS: %w", err)
	}

	fmt.Printf("\n✓ ICS file created: %s\n", outputFile)
	fmt.Printf("  Matches: %d\n", len(matches))

	return nil
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
	fmt.Println("  dindoa --team <name>                      Generate ICS for team")
	fmt.Println("  dindoa --team <name> --output <file>      Generate ICS with custom filename")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  dindoa start                              Start interactive menu")
	fmt.Println("  dindoa --team j3                          Generate dindoa-j3.ics")
	fmt.Println("  dindoa --category rood --list-teams       List teams in Rood category")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
}
