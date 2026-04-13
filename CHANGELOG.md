# Changelog

## NEXT VERSION

### Fixed
- **Nix build compatibility**: Updated flake.nix to use nixpkgs-unstable for Go 1.25.7 support
- **GitHub Actions**: Updated workflow to use Go 1.25.7 to match go.mod requirements

## 0.1.1 - 13 Apr 2026

### Added
- **Nix packaging**: Added flake.nix for Nix package manager support
  - Automatic vendorHash calculation in release.sh when Go dependencies change
  - Development shell with Go tooling
  - Multi-platform builds via buildGoModule

## 0.1.0 - 13 Apr 2026

### Added
- **Initial release**: CLI tool to generate ICS calendar files for Dindoa korfbal team matches
- **Team scraping**: Scrape team categories and match schedules from dindoa.nl
- **Geocoding**: Convert venue names to full addresses using OpenStreetMap Nominatim
- **Caching**: Cross-platform JSON cache for geocoded locations
- **Interactive UI**: Terminal-based interface using Bubbletea for team selection
- **CLI interface**: Command-line flags for listing teams/categories and generating ICS files
  - `dindoa start` - Start interactive mode
  - `dindoa --list-categories` - List all categories
  - `dindoa --category <name> --list-teams` - List teams in category
  - `dindoa --list-all-teams` - List all teams by category
  - `dindoa --team <name>` - Generate ICS file for team
- **Cross-platform support**: Builds for Linux, macOS, and Windows (amd64/arm64)
