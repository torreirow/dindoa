# Dindoa ICS Generator

[![Latest Release](https://img.shields.io/github/v/release/torreirow/dindoa)](https://github.com/torreirow/dindoa/releases/latest)
[![Downloads](https://img.shields.io/github/downloads/torreirow/dindoa/total)](https://github.com/torreirow/dindoa/releases)
[![License](https://img.shields.io/github/license/torreirow/dindoa)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/torreirow/dindoa)](go.mod)

Een CLI tool om wedstrijdschema's van Dindoa korfbal teams te exporteren naar ICS kalender bestanden.

## Features

- 🏐 Scrape wedstrijdschema's van dindoa.nl
- 📅 Genereer ICS bestanden voor import in elke kalender app
- 🗺️ Adressen en coördinaten van speellocaties uit een meegeleverde adressenlijst
- 💾 Werkt zonder wachttijd: één pagina ophalen, verder geen netwerk
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
1. Een categorie selecteert (Rood, Oranje, Geel, Groen, Blauw, Senioren/Wedstrijdsport)
2. Een team kiest
3. Automatisch een ICS bestand wordt gegenereerd

Locaties zonder adres worden in het afrondingsscherm gemeld, net als bij de CLI.

</details>

<details>
<summary><b>⌨️ CLI Mode (voor scripting)</b></summary>

### Lijst alle categorieën

```bash
dindoa --list-categories
```

Voorbeeld output:
```
Blauw
Geel
Groen
Oranje
Rood
Senioren/Wedstrijdsport
```

De kleuren komen uit de kolom `Kleur` van het wedstrijdprogramma. Senioren-, midweek- en U-teams hebben daar geen kleur en vallen samen onder `Senioren/Wedstrijdsport`.

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

Laat je `--list-teams` weg, dan opent de interactieve modus met die categorie al gekozen:

```bash
dindoa --category rood
```

Je komt direct in de teamlijst van Rood terecht. Bestaat de categorie niet, dan verschijnt het gewone keuzemenu met een melding erboven.

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

Na het genereren meldt de tool welke locaties nog geen adres hebben. Het bestand wordt altijd geschreven en de exitcode blijft `0`:

```
✓ ICS file created: dindoa-j4.ics
  Team:    Dindoa J4
  Matches: 6

⚠ 1 venue(s) not in the address list; the name from the website was used:
    Sportpark Overbeek TERSCHUUR               (1 match(es))
  Add them to ~/.config/dindoa/locations.json — run 'dindoa --list-locations' for a fragment to paste.
```

### Bekijk de speellocaties en hun adressen

```bash
dindoa --list-locations
```

Voorbeeld output:
```
Venues in the match programme (35 venues, 210 matches):

  ✓  105x  De Zanderij (Dindoa) ERMELO                Watervalweg 170, 3853 PT Ermelo
  ✓   15x  Het Slingerbos HARDERWIJK                  Slingerbos 1, 3844AC Harderwijk
  ✗    5x  Veld ASVD DRONTEN                          — not in the address list
  ...

23/35 venues mapped (185/210 matches = 88%)
Your address list: /home/je-naam/.config/dindoa/locations.json
```

Gesorteerd op aantal wedstrijden, dus de locatie die het meest oplevert staat bovenaan. Voor de ontbrekende locaties volgt een JSON-fragment om te plakken. Zie **Adressenlijst van speellocaties** hieronder.

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
- **Locatie**: De naam van de locatie zoals dindoa.nl die publiceert, plus het adres uit de adressenlijst (`De Zanderij (Dindoa), Watervalweg 170, 3853 PT Ermelo`). Staat een locatie niet in de lijst, dan wordt de naam van de website ongewijzigd gebruikt
- **Duur**: `DTSTART` en `DTEND`, standaard één uur, zodat de wedstrijd ruimte inneemt in je agenda
- **Coördinaten**: `GEO` wanneer bekend, zodat je agenda-app kaartweergave en routeplanning kan bieden
- **Categorie**: `CATEGORIES` met de kleur van het team (Rood, Oranje, Geel, Groen, Blauw)
- **UID**: Stabiele identifier per wedstrijd, opgebouwd uit team, datum en tegenstander. Wordt een wedstrijd naar een andere aanvangstijd verzet, dan **werkt** een nieuwe import het bestaande item bij in plaats van er een tweede naast te zetten
- **SEQUENCE**: Revisienummer dat bij elke keer genereren oploopt, zodat een agenda-app die revisies bijhoudt een nieuw bestand als nieuwer herkent. De waarde hoort bij de *uitgave*, niet bij de individuele wedstrijd: alle events in één bestand hebben dezelfde waarde, en die loopt ook op als er niets veranderd is

</details>

<details>
<summary><b>📍 Adressenlijst van speellocaties</b></summary>

Het wedstrijdprogramma op dindoa.nl noemt locaties met een naam en een plaats, bijvoorbeeld `De Zanderij (Dindoa) ERMELO`. De tool zet die om naar een adres met coördinaten via een adressenlijst die **in de binary is meegeleverd**. Er is geen netwerk voor nodig en er is geen wachttijd.

### Zelf een locatie toevoegen of corrigeren

Je kunt de meegeleverde lijst per locatie overschrijven of aanvullen met een eigen bestand:

| Platform | Pad van je eigen adressenlijst |
|---|---|
| **Linux** | `~/.config/dindoa/locations.json` |
| **macOS** | `~/Library/Application Support/dindoa/locations.json` |
| **Windows** | `%APPDATA%\dindoa\locations.json` |

> **Let op:** dit is de **config**map, niet de cachemap. Een oudere versie van deze tool schreef een `geocode.json` in de cachemap (`~/.cache/dindoa/` op Linux). Dat bestand wordt niet meer gebruikt en kan verwijderd worden.

Jouw bestand overschrijft de meegeleverde lijst **per locatie**. Locaties die je niet noemt blijven gewoon werken, dus je hoeft nooit de hele lijst te kopiëren.

Begin met:

```bash
dindoa --list-locations
```

Dat toont alle locaties uit het wedstrijdprogramma, welke al een adres hebben, en hoeveel wedstrijden er aan elke locatie hangen. Voor de locaties die nog ontbreken krijg je een fragment dat je rechtstreeks in je eigen bestand kunt plakken:

```json
{
  "version": 1,
  "locations": {
    "sportpark overbeek terschuur": {
      "name": "Sportpark Overbeek",
      "address": "Overbeeksestraat 1, 3784 XX Terschuur",
      "lat": 52.1654,
      "lon": 5.5195,
      "osm": "way/123456789",
      "source": "manual"
    }
  }
}
```

| Veld | Betekenis |
|---|---|
| sleutel | De locatienaam van de website, genormaliseerd: kleine letters, zonder interpunctie, enkele spaties |
| `name` | De leesbare naam die in je agenda komt |
| `address` | Het adres. Mag ook alleen straatniveau zijn; `lat`/`lon` leveren dan de precisie |
| `lat` / `lon` | Coördinaten. Laat op `0` staan als je ze niet weet |
| `osm` | Optionele verwijzing naar het OpenStreetMap-object, zodat de regel later na te controleren is |
| `source` | Waar de gegevens vandaan komen: `manual`, `osm-tags`, `osm-reverse` |

Een locatie die nergens in de lijst staat blokkeert nooit: de ICS wordt gewoon gegenereerd met de naam van de website, en de tool meldt welke locaties ontbreken.

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
<summary><b>Een locatie heeft geen adres</b></summary>

```
⚠ 1 venue(s) not in the address list; the name from the website was used:
    Sportpark Overbeek TERSCHUUR               (1 match(es))
```

Dit is geen fout. De ICS is aangemaakt en geldig; de betreffende wedstrijd heeft alleen de locatienaam van de website als locatie, zonder adres en zonder coördinaten. De exitcode blijft `0`.

Om het aan te vullen:

```bash
dindoa --list-locations
```

Dat geeft een JSON-fragment dat je in je eigen adressenlijst kunt plakken. Zie **Adressenlijst van speellocaties** hierboven.

</details>

<details>
<summary><b>Team niet gevonden</b></summary>

```
Error: team "Dindoa J99" does not appear in the match programme.
Available teams:
  Dindoa 1
  ...
```

De tool somt de beschikbare teams op. Mogelijke oorzaken:
- Typo in de teamnaam
- Het team komt (nog) niet voor in het gepubliceerde deel van het wedstrijdprogramma

</details>

<details>
<summary><b>Nog geen wedstrijden</b></summary>

```
Dindoa J3 has no matches in the published part of the programme.
```

Het wedstrijdprogramma op dindoa.nl wordt in blokken gepubliceerd, niet in één keer voor het hele seizoen. Loop je hier tegenaan, kijk dan later in het seizoen opnieuw.

Om dezelfde reden is elk ICS-bestand een momentopname. Opnieuw genereren en importeren is veilig: de UID van een wedstrijd is stabiel, dus je agenda-app werkt bestaande items bij in plaats van ze te verdubbelen.

> **Eenmalig bij het bijwerken vanaf v0.1.3 of ouder:** de opbouw van de UID is gewijzigd. Verwijder een eerder geïmporteerde Dindoa-agenda voordat je een nieuw bestand importeert, anders staan de oude en nieuwe items naast elkaar.

</details>

<details>
<summary><b>Foutmelding over de opmaak van de pagina</b></summary>

```
Error: parse match programme (...): unexpected match table layout: expected columns [...]
```

De opmaak van het wedstrijdprogramma op dindoa.nl is gewijzigd. De tool geeft hier bewust een fout in plaats van stil nul wedstrijden op te leveren. Meld dit als issue.

</details>

<details>
<summary><b>Foutmelding over het seizoensjaar</b></summary>

```
Error: could not establish the season year: ...
```

De datumkoppen op de pagina bevatten geen jaartal, dus de tool leidt het seizoen af en controleert dat op de dag van de week: korfbal speelt zaterdag en woensdag. Klopt dat patroon niet, dan is er iets mis met de systeemklok of met de pagina.

</details>

## Technische Details

- **Taal**: Go 1.25.7
- **Dependencies**:
  - Bubbletea (TUI framework)
  - goquery (HTML parsing)
  - golang-ical (ICS generatie)
  - xdg (cross-platform paths)
- **Data source**: https://dindoa.nl/ws/
- **Locaties**: meegeleverde adressenlijst (`locations.json`, via `go:embed`), te overschrijven met een eigen bestand in de configmap
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
