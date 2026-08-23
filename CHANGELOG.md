# Changelog


## NEXT VERSION

### Fixed

- **Aanvangstijd werd stil verkeerd gelezen**: De tijdkolom werd met `fmt.Sscanf("%d:%d")` gelezen
  zonder te controleren of dat lukte, en `time.Date` normaliseert waarden buiten bereik zonder te
  klagen. Een gewijzigde notatie leverde daardoor een geldig ICS-bestand op het verkeerde moment op:
  `13.45` gaf 45 minuten verschil, en `1345` schoof het event 56 dagen op naar eind oktober. Zonder
  waarschuwing.
  - De tijd wordt nu gevalideerd waar de bron wordt gelezen, naast de bestaande controle op de
    kolomkoppen. Een afwijking geeft een fout die de wedstrijd en de aangetroffen waarde noemt.
  - Er wordt bewust niet gegokt. `19.00` naar `19:00` raden lijkt vriendelijk, maar dat is een
    aanname over de bedoeling van een gewijzigde bron.
  - Een fout laat de hele pagina falen in plaats van één rij over te slaan; een halve agenda is
    misleidender dan geen agenda.
  - Alle 210 op dit moment gepubliceerde rijen gebruiken `HH:MM`, dus er was niets actief fout. Dit
    sluit een latent gat.
- **Doodlopend foutscherm in de interactieve modus**: Ging er iets mis nadat je een team had gekozen,
  dan kon je alleen nog afsluiten en de tool opnieuw starten — inclusief het opnieuw ophalen van het
  wedstrijdprogramma. Nu keer je met enter terug naar de teamlijst en kies je een ander team; de
  programmagegevens zitten al in het geheugen, dus dat kost geen nieuw netwerkverzoek. Gaat het
  ophalen van het programma zelf mis, dan is er niets om naar terug te keren en biedt het scherm
  alleen afsluiten. Het foutscherm noemt nu de toetsen die er werkelijk werken.
- **`--category` zonder `--list-teams`**: Dit startte de help in plaats van de interactieve modus,
  terwijl de specificatie voorschrijft dat de categorie dan voorgeselecteerd wordt. Nu kom je met
  `dindoa --category rood` direct in de teamlijst van Rood. Bestaat de categorie niet, dan verschijnt
  het gewone keuzemenu met een melding erboven in plaats van een doodlopende foutmelding. Ook
  `dindoa start --category rood` werkt zo. Deze afwijking bestond al vóór de overgang naar het
  wedstrijdprogramma als bron.

### Added

- **`SEQUENCE` in ICS-events**: Elk event krijgt een revisienummer dat bij elke keer genereren
  oploopt, zodat een agenda-app die revisies bijhoudt een nieuw bestand als nieuwer herkent. Zonder
  dit veld gedraagt de uitvoer zich volgens RFC 5545 als revisie 0, waardoor een strikte client een
  echte wijziging kon negeren. Het wedstrijdprogramma wordt in blokken gepubliceerd, dus opnieuw
  genereren is de normale gang van zaken en niet de uitzondering.
  - De waarde is afgeleid van het genereermoment, niet van de wedstrijdgegevens. Deze tool houdt
    geen staat bij en weet dus niet welke revisie een wedstrijd al had; een teller op basis van de
    inhoud kan niet monotoon oplopen zoals de standaard vereist.
  - Alle events in één bestand hebben dezelfde waarde: de revisie hoort bij de uitgave van de
    kalender, net als `DTSTAMP`.
  - De waarde loopt ook op als er niets veranderd is. Dat is onschadelijk, en het is de reden dat er
    bewust niet per wedstrijd geteld wordt.

### Changed

- **Wedstrijdprogramma als bron**: Wedstrijden, teams, categorieën en kleuren komen nu uit
  `dindoa.nl/ws/competitie-programma/` in plaats van uit de teampagina's en de teamsoverzichtspagina.
  Die pagina's zijn voor het seizoen 2026/2027 niet meer gevuld met wedstrijddata, waardoor de tool
  geen wedstrijden meer vond en de lijstcommando's met een parse-fout afbraken. Eén pagina levert nu
  alles, inclusief de kleur per team.
  - Datum komt uit de kop boven elke tabel; die bevat geen jaartal, dus het seizoen wordt afgeleid
    en gecontroleerd op de dag van de week (korfbal speelt zaterdag en woensdag)
  - Teams worden op volledige naam gefilterd. `Dindoa J1` is een prefix van `J10` t/m `J19`, dus
    benaderend matchen leverde 66 wedstrijden op in plaats van 6
  - Onverwachte tabelopmaak geeft nu een duidelijke fout in plaats van stil nul wedstrijden
- **Adressen uit een meegeleverde lijst**: De geocoding tijdens het genereren is vervangen door een
  adressenlijst die met de binary wordt meegeleverd. Genereren doet nu precies één netwerkverzoek
  en heeft geen wachttijd meer.
- **Locatie in het ICS-bestand**: `LOCATION` bevat nu de naam van de website **plus** het adres. De
  naam werd voorheen overschreven met het zoekresultaat, waardoor een leesbare locatienaam kon
  verdwijnen ten gunste van een verkeerd adres.
- **Stabiele UID**: De UID bevat niet langer de aanvangstijd, maar team, datum en tegenstander. Een
  verzette wedstrijd werkt daardoor het bestaande agenda-item bij in plaats van er een tweede naast
  te zetten. **Verwijder een eerder geïmporteerde Dindoa-agenda eenmalig voordat je opnieuw
  importeert.**

### Added

- **`--list-locations`**: Toont alle speellocaties uit het wedstrijdprogramma met hun adresstatus en
  het aantal wedstrijden per locatie, gesorteerd op impact. Voor ontbrekende locaties volgt een
  JSON-fragment om in je eigen adressenlijst te plakken.
- **Eigen adressenlijst**: Een optioneel bestand in de configmap (`~/.config/dindoa/locations.json`
  op Linux) overschrijft en vult de meegeleverde lijst aan, **per locatie**. Locaties die je niet
  noemt blijven gewoon werken.
- **`DTEND` in ICS-events**: Wedstrijden duren standaard een uur. Voorheen had een event alleen
  `DTSTART` en was het volgens RFC 5545 nul seconden lang.
- **`GEO` in ICS-events**: Coördinaten worden meegegeven wanneer bekend, zodat agenda-apps
  kaartweergave en routeplanning kunnen bieden.
- **`CATEGORIES` in ICS-events**: De kleur van het team (Rood, Oranje, Geel, Groen, Blauw) wordt
  vastgelegd.
- **Scheidsrechter in de omschrijving**: Wordt meegenomen wanneer die in het programma staat.

### Fixed

- **Stil falen bij het genereren**: `--team` schreef geen bestand en meldde niets bruikbaars wanneer
  er geen wedstrijden werden gevonden. Een onbekend team geeft nu een fout met de beschikbare teams
  en exitcode 1.
- **Ontbrekende locatie blokkeert niet meer**: Het bestand wordt altijd geschreven met de naam van de
  website, en de ontbrekende locaties worden expliciet gemeld met het aantal betrokken wedstrijden.
  De exitcode blijft 0, want het bestand is geldig.
- **Zelfvergiftigende cache verwijderd**: Mislukte zoekopdrachten werden onvoorwaardelijk gecached
  met coördinaten 0,0 en zonder geldigheidsduur, waardoor de eerste run een fout resultaat permanent
  vastzette. De cache in `~/.cache/dindoa/geocode.json` wordt niet meer gebruikt en kan verwijderd
  worden.

## 0.1.3 - 13 Apr 2026

### Changed
- **Documentation**: Comprehensive README update with collapsible sections
  - Added pre-built binaries installation instructions for all platforms
  - Extensive NixOS/Nix documentation (standalone, NixOS config, Home Manager)
  - Better organization with markdown collapsible blocks
  - Added shields.io badges (version, downloads, license, Go version)
  - Use GitHub's /releases/latest URL for automatic version linking
  - Added GitHub CLI download example

### Added
- **Release automation**: Script now auto-generates NEXT VERSION placeholder after releases
- **Nix vendorHash detection**: Release script now detects placeholder hashes and calculates them automatically

### Fixed
- **Nix build compatibility**: Updated flake.nix to use nixpkgs-unstable for Go 1.25.7 support
- **GitHub Actions**: Updated workflow to use Go 1.25.7 to match go.mod requirements
- **Nix vendorHash**: Set correct hash for Go module dependencies

## 0.1.2 - 13 Apr 2026

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
