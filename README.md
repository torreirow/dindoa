# Dindoa ICS Generator

Een CLI tool om wedstrijdschema's van Dindoa korfbal teams te exporteren naar ICS kalender bestanden.

## Features

- 🏐 Scrape wedstrijdschema's van dindoa.nl
- 📅 Genereer ICS bestanden voor import in elke kalender app
- 🗺️ Automatische geocoding van locaties via OpenStreetMap
- 💾 Cross-platform caching voor snelle herhaalde uitvoeringen
- 🎨 Interactieve TUI of CLI flags voor scripting
- 🌍 Cross-platform (Linux, macOS, Windows)

## Installatie

### Vanaf bron

```bash
git clone https://github.com/torreirow/dindoa.git
cd dindoa
go build -o dindoa cmd/dindoa/main.go
```

### Direct installeren met Go

```bash
go install github.com/torreirow/dindoa/cmd/dindoa@latest
```

## Gebruik

### Help tonen

Zonder argumenten toont de tool de help message:

```bash
dindoa
# of
dindoa --help
dindoa -h
```

### Interactive Mode

Start de tool met het `start` commando voor een interactieve interface:

```bash
dindoa start
```

Dit opent een terminal UI waar je:
1. Een categorie selecteert (Senioren, Rood, Oranje, etc.)
2. Een team kiest
3. Automatisch een ICS bestand wordt gegenereerd

### CLI Mode (voor scripting)

#### Lijst alle categorieën

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

#### Lijst teams in een categorie

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

#### Lijst alle teams gesorteerd per categorie

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

#### Genereer ICS voor een team

```bash
dindoa --team j3
```

Dit genereert `dindoa-j3.ics` met alle wedstrijden.

Team namen zijn flexibel:
- `dindoa --team j3` ✓
- `dindoa --team J3` ✓
- `dindoa --team "Dindoa J3"` ✓
- `dindoa --team "dindoa j3"` ✓

#### Custom output bestand

```bash
dindoa --team j3 --output mijn-wedstrijden.ics
```


## ICS Bestand Details

Gegenereerde ICS bestanden bevatten:

- **Titel**: Correcte formatting (thuiswedstrijd: "Dindoa J3 - Tegenstander", uitwedstrijd: "Tegenstander - Dindoa J3")
- **Datum/Tijd**: In Europe/Amsterdam timezone (automatisch CET/CEST)
- **Locatie**: Volledig adres via OpenStreetMap geocoding (met fallback naar originele tekst)
- **UID**: Unieke identifier per wedstrijd

## Caching

Geocoding resultaten worden gecached op:

- **Linux**: `~/.cache/dindoa/geocode.json`
- **macOS**: `~/Library/Caches/dindoa/geocode.json`
- **Windows**: `%LOCALAPPDATA%\dindoa\cache\geocode.json`

Dit zorgt voor:
- Snellere herhaalde uitvoeringen
- Minder belasting op OpenStreetMap Nominatim API
- Handmatig editeerbare cache (JSON format)

De eerste run kan langzaam zijn door geocoding (1 request/seconde rate limit), maar daarna is het instant.

## Voorbeelden

### Basis workflow

```bash
# Toon help
dindoa

# Start interactief menu
dindoa start

# Of direct via CLI
dindoa --team j3

# Importeer dindoa-j3.ics in je kalender app
```

### Alle teams van een categorie

```bash
# Zie welke teams er zijn
dindoa --category rood --list-teams

# Genereer ICS voor elk team
dindoa --team j1
dindoa --team j2
dindoa --team j3
```

### Scripting

```bash
#!/usr/bin/env bash
# Genereer ICS voor alle teams in Rood

for team in $(dindoa --category rood --list-teams | grep -o 'J[0-9]'); do
  echo "Generating ICS for $team..."
  dindoa --team "$team"
done
```

## Troubleshooting

### Geocoding faalt

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

### Team niet gevonden

```
Error: fetch team page: status 404
```

Mogelijke oorzaken:
- Typo in team naam (probeer `--list-teams`)
- Team bestaat niet (controleer dindoa.nl)
- Website structuur veranderd

### Geen wedstrijden

```
No matches found for this team
```

Dit team heeft (nog) geen geplande wedstrijden dit seizoen.

## Technische Details

- **Taal**: Go
- **Dependencies**:
  - Bubbletea (TUI framework)
  - goquery (HTML parsing)
  - golang-ical (ICS generatie)
  - xdg (cross-platform paths)
- **Data source**: https://dindoa.nl/ws/
- **Geocoding**: OpenStreetMap Nominatim (rate limited: 1 req/sec)

## License

MIT

## Bijdragen

Pull requests zijn welkom!

1. Fork het project
2. Maak een feature branch (`git checkout -b feature/amazing-feature`)
3. Commit je changes (`git commit -m 'Add amazing feature'`)
4. Push naar de branch (`git push origin feature/amazing-feature`)
5. Open een Pull Request
