# Dindoa ICS Generator

[![Latest Release](https://img.shields.io/github/v/release/torreirow/dindoa)](https://github.com/torreirow/dindoa/releases/latest)
[![Downloads](https://img.shields.io/github/downloads/torreirow/dindoa/total)](https://github.com/torreirow/dindoa/releases)
[![License](https://img.shields.io/github/license/torreirow/dindoa)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/torreirow/dindoa)](go.mod)

Een CLI tool om wedstrijdschema's van Dindoa korfbal teams te exporteren naar ICS kalender bestanden.

## Features

- 🏐 Scrape wedstrijdschema's van dindoa.nl
- 📅 Genereer ICS bestanden voor import in elke kalender app
- 🗺️ Automatische geocoding van locaties via OpenStreetMap
- 💾 Cross-platform caching voor snelle herhaalde uitvoeringen
- 🎨 Interactieve TUI of CLI flags voor scripting
- 🌍 Cross-platform (Linux, macOS, Windows)

## Installatie

<details>
<summary><b>📦 Pre-built binaries (aanbevolen)</b></summary>

Download de nieuwste release voor jouw platform:

**Beschikbare platforms:**
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

**📥 [Download nieuwste release](https://github.com/torreirow/dindoa/releases/latest)**

```bash
# Linux amd64 voorbeeld (vervang VERSION met de nieuwste release):
wget https://github.com/torreirow/dindoa/releases/latest/download/dindoa-VERSION-linux-amd64.tar.gz
tar xzf dindoa-VERSION-linux-amd64.tar.gz
sudo mv dindoa /usr/local/bin/

# Of via GitHub CLI:
gh release download --repo torreirow/dindoa --pattern '*linux-amd64.tar.gz'
tar xzf dindoa-*-linux-amd64.tar.gz
sudo mv dindoa /usr/local/bin/

# Windows: download .zip van releases pagina en extract naar een directory in je PATH
```

</details>

<details>
<summary><b>❄️ NixOS / Nix</b></summary>

### Standalone gebruik (zonder installatie)

```bash
# Direct runnen vanaf GitHub (nieuwste versie)
nix run github:torreirow/dindoa -- start

# Met specifieke versie/tag
nix run github:torreirow/dindoa/v0.1.2 -- --team j3

# Alle commando's werken
nix run github:torreirow/dindoa -- --list-categories
nix run github:torreirow/dindoa -- --team j3 --output wedstrijden.ics
```

### NixOS configuratie

Voeg dindoa toe als input in je `flake.nix`:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    dindoa.url = "github:torreirow/dindoa";
  };

  outputs = { self, nixpkgs, dindoa, ... }: {
    nixosConfigurations.your-hostname = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        {
          environment.systemPackages = [
            dindoa.packages.x86_64-linux.dindoa
          ];
        }
      ];
    };
  };
}
```

### Home Manager

Voeg dindoa toe aan je Home Manager configuratie:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager.url = "github:nix-community/home-manager";
    dindoa.url = "github:torreirow/dindoa";
  };

  outputs = { self, nixpkgs, home-manager, dindoa, ... }: {
    homeConfigurations.your-username = home-manager.lib.homeManagerConfiguration {
      pkgs = nixpkgs.legacyPackages.x86_64-linux;
      modules = [
        {
          home.packages = [
            dindoa.packages.x86_64-linux.dindoa
          ];
        }
      ];
    };
  };
}
```

### Nix profile (zonder flakes)

```bash
# Installeer in je profiel (nieuwste versie)
nix profile install github:torreirow/dindoa

# Of met specifieke versie/tag
nix profile install github:torreirow/dindoa/v0.1.2
```

### Development shell

```bash
# Clone de repository
git clone https://github.com/torreirow/dindoa.git
cd dindoa

# Start development shell met Go tooling
nix develop

# Build in de shell
go build -o dindoa cmd/dindoa/main.go
```

</details>

<details>
<summary><b>🔧 Vanaf bron (Go)</b></summary>

### Vereisten
- Go 1.25.7 of hoger

### Build

```bash
git clone https://github.com/torreirow/dindoa.git
cd dindoa
go build -o dindoa cmd/dindoa/main.go
```

### Direct installeren met Go

```bash
go install github.com/torreirow/dindoa/cmd/dindoa@latest
```

</details>

## Gebruik

### Help tonen

Zonder argumenten toont de tool de help message:

```bash
dindoa
# of
dindoa --help
dindoa -h
```

<details>
<summary><b>🎨 Interactive Mode</b></summary>

Start de tool met het `start` commando voor een interactieve interface:

```bash
dindoa start
```

Dit opent een terminal UI waar je:
1. Een categorie selecteert (Senioren, Rood, Oranje, etc.)
2. Een team kiest
3. Automatisch een ICS bestand wordt gegenereerd

</details>

<details>
<summary><b>⌨️ CLI Mode (voor scripting)</b></summary>

### Lijst alle categorieën

```bash
dindoa --list-categories
```

Voorbeeld output:
```
Senioren
Wedstrijdsport
Rood
Oranje
Geel
Groen
Blauw
```

### Lijst teams in een categorie

```bash
dindoa --category rood --list-teams
```

Voorbeeld output:
```
Dindoa J1
Dindoa J2
Dindoa J3
Dindoa J4
```

Parameters zijn case-insensitive:
```bash
dindoa --category ROOD --list-teams
dindoa --category Rood --list-teams
```

### Lijst alle teams gesorteerd per categorie

```bash
dindoa --list-all-teams
```

Voorbeeld output:
```
Rood:
  Dindoa J1
  Dindoa J2
  Dindoa J3

Oranje:
  Dindoa J5
  Dindoa J6
  ...
```

### Genereer ICS voor een team

```bash
dindoa --team j3
```

Dit genereert `dindoa-j3.ics` met alle wedstrijden.

Team namen zijn flexibel:
- `dindoa --team j3` ✓
- `dindoa --team J3` ✓
- `dindoa --team "Dindoa J3"` ✓
- `dindoa --team "dindoa j3"` ✓

### Custom output bestand

```bash
dindoa --team j3 --output mijn-wedstrijden.ics
```

</details>

<details>
<summary><b>📅 ICS Bestand Details</b></summary>

Gegenereerde ICS bestanden bevatten:

- **Titel**: Correcte formatting (thuiswedstrijd: "Dindoa J3 - Tegenstander", uitwedstrijd: "Tegenstander - Dindoa J3")
- **Datum/Tijd**: In Europe/Amsterdam timezone (automatisch CET/CEST)
- **Locatie**: Volledig adres via OpenStreetMap geocoding (met fallback naar originele tekst)
- **UID**: Unieke identifier per wedstrijd

</details>

<details>
<summary><b>💾 Caching</b></summary>

Geocoding resultaten worden gecached op:

- **Linux**: `~/.cache/dindoa/geocode.json`
- **macOS**: `~/Library/Caches/dindoa/geocode.json`
- **Windows**: `%LOCALAPPDATA%\dindoa\cache\geocode.json`

Dit zorgt voor:
- Snellere herhaalde uitvoeringen
- Minder belasting op OpenStreetMap Nominatim API
- Handmatig editeerbare cache (JSON format)

De eerste run kan langzaam zijn door geocoding (1 request/seconde rate limit), maar daarna is het instant.

</details>

## Voorbeelden

<details>
<summary><b>Basis workflow</b></summary>

```bash
# Toon help
dindoa

# Start interactief menu
dindoa start

# Of direct via CLI
dindoa --team j3

# Importeer dindoa-j3.ics in je kalender app
```

</details>

<details>
<summary><b>Alle teams van een categorie</b></summary>

```bash
# Zie welke teams er zijn
dindoa --category rood --list-teams

# Genereer ICS voor elk team
dindoa --team j1
dindoa --team j2
dindoa --team j3
```

</details>

<details>
<summary><b>Scripting</b></summary>

```bash
#!/usr/bin/env bash
# Genereer ICS voor alle teams in Rood

for team in $(dindoa --category rood --list-teams | grep -o 'J[0-9]'); do
  echo "Generating ICS for $team..."
  dindoa --team "$team"
done
```

**Met Nix:**

```bash
#!/usr/bin/env bash
# Geen lokale installatie nodig

for team in j1 j2 j3 j4; do
  echo "Generating ICS for $team..."
  nix run github:torreirow/dindoa -- --team "$team"
done
```

</details>

## Troubleshooting

<details>
<summary><b>Geocoding faalt</b></summary>

Als een locatie niet gevonden wordt door OpenStreetMap:
- De originele locatie tekst wordt gebruikt als fallback
- Er verschijnt een waarschuwing in de output
- De ICS wordt nog steeds gegenereerd

Je kunt de cache handmatig bewerken:
```bash
# Linux
nano ~/.cache/dindoa/geocode.json

# Pas het adres aan voor een specifieke locatie
```

</details>

<details>
<summary><b>Team niet gevonden</b></summary>

```
Error: fetch team page: status 404
```

Mogelijke oorzaken:
- Typo in team naam (probeer `--list-teams`)
- Team bestaat niet (controleer dindoa.nl)
- Website structuur veranderd

</details>

<details>
<summary><b>Geen wedstrijden</b></summary>

```
No matches found for this team
```

Dit team heeft (nog) geen geplande wedstrijden dit seizoen.

</details>

## Technische Details

- **Taal**: Go 1.25.7
- **Dependencies**:
  - Bubbletea (TUI framework)
  - goquery (HTML parsing)
  - golang-ical (ICS generatie)
  - xdg (cross-platform paths)
- **Data source**: https://dindoa.nl/ws/
- **Geocoding**: OpenStreetMap Nominatim (rate limited: 1 req/sec)
- **Platforms**: Linux, macOS, Windows (amd64 & arm64)
- **Package managers**: Nix, Go modules

## License

MIT

## Bijdragen

Pull requests zijn welkom!

1. Fork het project
2. Maak een feature branch (`git checkout -b feature/amazing-feature`)
3. Commit je changes (`git commit -m 'Add amazing feature'`)
4. Push naar de branch (`git push origin feature/amazing-feature`)
5. Open een Pull Request
