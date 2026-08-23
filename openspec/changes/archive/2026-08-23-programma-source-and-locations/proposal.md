# Wedstrijdprogramma als bron + locatie-mappingfile

## Why

De tool is op dit moment volledig kapot voor het seizoen 2026/2027. Gemeten tegen de live site:

- `dindoa --list-categories` → `Error: parse categories: no categories found in HTML`. De pagina `/ws/teams/` heeft de categoriekoppen (Senioren, Wedstrijdsport, Rood, Oranje, Geel, Groen) maar **nul team-links**. De kleurenpagina's `/ws/teams-rood/` t/m `/ws/teams-blauw/` zijn leeg. Dat blokkeert ook `--list-teams`, `--list-all-teams` en `dindoa start`.
- `dindoa --team j4` → `No matches found for this team`, en er wordt **geen bestand geschreven** met exitcode 0. De teampagina bevat dit seizoen alleen een trainingstabel (`Dag | Tijd | Veld | Trainer(s)`); wedstrijddata staat er niet meer op.

Tegelijk is het volledige wedstrijdprogramma wél gepubliceerd op `/ws/competitie-programma/`: server-rendered HTML, 12 datumblokken, 210 wedstrijden, 35 Dindoa-teams, met kolommen `Tijd | Thuis | Uit | Kleur | Locatie | Scheidsrechter`. Eén GET levert wedstrijden, locaties **en** de kleur/categorie per team — waardoor de teams- en kleurenpagina's helemaal niet meer nodig zijn.

Daarnaast levert de runtime-geocoding aantoonbaar slechte data. Op de 35 unieke locaties van dit seizoen gemeten via Nominatim met de huidige parameters: circa 40% treffers, en de mislukkingen zijn niet het ergste. `Het Slingerbos HARDERWIJK` levert "Het Nachthok" op — **702 meter** naast de werkelijke locatie — en de code overschrijft daarmee een leesbare venuenaam met een verkeerd adres. De eigen accommodatie `De Zanderij (Dindoa) ERMELO`, 105 van de 210 wedstrijden, wordt nooit gevonden. En omdat `fallback()` een `Result{Address: query, Lat: 0, Lng: 0}` teruggeeft die in `main.go` onvoorwaardelijk gecached wordt zonder TTL of invalidatie, zet de eerste run die fouten permanent vast.

## What Changes

### Bron: wedstrijdprogramma in plaats van teampagina's

- **BREAKING** De primaire bron wordt `/ws/competitie-programma/`. De teampagina `/ws/<slug>/` en de teamsoverzichtspagina `/ws/teams/` worden niet meer gebruikt voor wedstrijden, teams of categorieën.
- Datum komt uit de `<h3>`-kop boven elke tabel (`5 september`) — **zonder jaartal**. Het jaar wordt afgeleid uit het seizoen en gevalideerd tegen de dag van de week: korfbal speelt zaterdag en woensdag, en alleen 2026 levert voor de 12 gepubliceerde datums het patroon `za wo za wo …` (2025 geeft `vr di`, 2027 geeft `zo do`). Die controle is het vangnet rond de jaarwisseling.
- De kolom `Kleur` (`<span class="colorcode rood">Rood</span>`) levert categorie per team. Senioren-, MW- en U-teams hebben een lege Kleur en vallen onder Senioren/Wedstrijdsport.
- Teamlijst en categorieën worden afgeleid uit het programma zelf. Daarmee werken `--list-categories`, `--list-teams`, `--list-all-teams` en `dindoa start` weer.
- Teams worden op **volledige naam** gefilterd (`Dindoa J4`), niet op teamcode. Filteren op `J4` haalt tegenstanders als `Revival J4`, `Regio '72 J4` en `Unitas/Perspectief J4` binnen. Exacte matching is ook nodig omdat `Dindoa J1` een prefix is van `J10`–`J19` en `Dindoa J2` van `J20`–`J24`: naïef matchen geeft daar 66 respectievelijk 36 rijen in plaats van 6.
- Gebruikersinvoer blijft tolerant: `j4`, `J4`, `dindoa j4` en `Dindoa J4` verwijzen alle naar hetzelfde team.

### Locaties: mappingfile in plaats van runtime-geocoding

- **BREAKING** Runtime-geocoding via Nominatim verdwijnt uit het normale pad. Geen netwerkverkeer, geen rate limiting, geen wachttijd bij het genereren van een ICS.
- Er komt een gelaagde mappingfile: een meegeleverd bestand in de binary (`go:embed`), met daarbovenop een optioneel gebruikersbestand in de configmap dat per key overschrijft en aanvult. Zo werkt de tool correct uit de doos en kan een gebruiker een foute regel zelf repareren zonder op een release te wachten.
- Sleutel is de genormaliseerde locatiestring van de website. De waarde bevat een leesbare naam, een adres, coördinaten en een OSM-referentie zodat een regel later na te controleren is.
- **Een ontbrekende locatie blokkeert nooit.** De ICS wordt gegenereerd met de originele websitestring als locatie, en de tool meldt expliciet welke locaties ontbreken.
- Nieuwe vlag `--list-locations` toont alle locaties uit het wedstrijdschema met hun mappingstatus, en geeft voor ontbrekende locaties een plak-klaar JSON-fragment.

### ICS-uitvoer

- `LOCATION` bevat de leesbare naam **en** het adres (`De Zanderij (Dindoa), Watervalweg 170, 3853 PT Ermelo`) in plaats van een overschreven Nominatim-`display_name`. Informatie van de website kan zo nooit meer verloren gaan.
- `GEO` krijgt de coördinaten, zodat agenda-apps kaartweergave en routeplanning bieden. Dit is de reden dat een adres op straatniveau zonder huisnummer bruikbaar blijft.
- `DTEND` wordt toegevoegd, standaard 60 minuten. Nu heeft elk event alleen `DTSTART` en is het volgens RFC 5545 dus nul seconden lang.
- De UID bevat niet langer de aanvangstijd. Nu is dat `dindoa-j4-2026-09-05-1345@dindoa.nl`; een verzette wedstrijd levert daardoor een tweede event op in plaats van een wijziging van het bestaande.
- De kleur wordt vastgelegd in `CATEGORIES`, zodat die informatie niet verloren gaat.

### Documentatie

- Help: nieuwe vlag in het handgeschreven `Usage:`-blok van `printUsage()` én in de voorbeelden. Let op dat `flag.PrintDefaults()` de `Options:`-lijst automatisch vult, waardoor beide helften uiteen kunnen lopen — een nieuwe vlag verschijnt automatisch onderaan maar niet in de usage-regels.
- README: de feature-bullet over automatische geocoding, de sectie over de locatie van het cachebestand, de troubleshooting-sectie die nu adviseert `~/.cache/dindoa/geocode.json` met de hand te bewerken, en de technische noot over de Nominatim-rate-limit.
- README krijgt een tabel met de platformspecifieke paden van het gebruikersbestand in de **config**map, naast de bestaande tabel met cachepaden. Die twee verwarren is de meest waarschijnlijke gebruikersfout.
- CHANGELOG onder `## NEXT VERSION`.

## Capabilities

### New Capabilities

- `location-mapping`: Gelaagde locatie-mappingfile — meegeleverd bestand plus gebruikersbestand dat per key overschrijft, sleutelnormalisatie, opzoeken van een locatie, expliciete melding en veilige terugval bij ontbrekende locaties, en `--list-locations` om de status te tonen.

### Modified Capabilities

- `team-scraping`: Bron wordt het wedstrijdprogramma in plaats van de teampagina's en de teamsoverzichtspagina. Datum uit de `<h3>` met afleiding en validatie van het jaartal, zes-koloms tabelvorm, kleur per team, teamlijst en categorieën uit het programma, en exacte matching op volledige teamnaam.
- `geocoding`: Runtime-geocoding via Nominatim en de bijbehorende rate limiting en resultatencache verdwijnen uit het normale pad; locatiegegevens komen uit `location-mapping`.
- `ics-generation`: `LOCATION` behoudt de leesbare naam naast het adres, `GEO` met coördinaten, `DTEND` met standaardduur, UID zonder aanvangstijd zodat verzette wedstrijden bijwerken in plaats van verdubbelen, en `CATEGORIES` met de kleur.
- `cli-interface`: Nieuwe vlag `--list-locations`, en de helptekst die de nieuwe vlag beschrijft.

## Impact

**Code**

- `internal/scraper/fetcher.go` — nieuwe URL; `FetchTeamsPage` en `FetchTeamPage` vervallen als bron voor wedstrijden en teams.
- `internal/scraper/parser.go` — `ParseMatches` moet de zes-koloms vorm met datum in de `<h3>` aan; de huidige implementatie leest `cells.Eq(0)` als datum en doet `time.Parse("02-01-2006", …)`. Op de programma-tabel is cel 0 de tijd (`11:50`), de parse faalt en de rij wordt stil overgeslagen. De bestaande parser op deze URL richten levert dus **0 wedstrijden zonder foutmelding**. `ParseCategories` en `ParseTeams` worden herleid naar het programma.
- `internal/scraper/types.go` — `Match` krijgt kleur en gestructureerde locatiegegevens.
- `internal/geocode/` — vervalt als runtime-afhankelijkheid; de bestaande `CacheData`-structuur (`map[genormaliseerde key]Result` in JSON) is precies de vorm die de mappingfile nodig heeft en kan als basis dienen.
- `internal/ics/generator.go` — `createEvent`, `generateUID`, `parseMatchDateTime`. De tijdzonebehandeling is correct en blijft: `SetStartAt` in golang-ical v0.3.5 doet `t.UTC().Format(...)`, dus 13:45 CEST wordt `114500Z`.
- `cmd/dindoa/main.go` — nieuwe vlag, `printUsage()`, en de meldingen bij ontbrekende locaties.
- `internal/ui/` — haalt categorieën en teams via de nieuwe bron; de schermen zelf blijven gelijk.
- Nieuw: het meegeleverde mappingbestand plus de loader met `go:embed`.

**Afhankelijkheden**

- `github.com/PuerkitoBio/goquery` blijft. `github.com/adrg/xdg` blijft, nu ook voor `ConfigHome`. `github.com/arran4/golang-ical` blijft; `SetGeo` bestaat in v0.3.5.
- Geen nieuwe dependencies. De HTTP-client naar Nominatim verdwijnt uit het normale pad, waardoor de tool offline werkt.

**Data**

- 35 unieke locaties in het huidige seizoen. 23 zijn al opgezocht en dekken 185 van de 210 wedstrijden (88%); de resterende 12 clubs staan niet in OpenStreetMap en worden met de hand aangevuld. De verdeling is scheef: één locatie dekt 50% van alle wedstrijden, tien locaties dekken 80%.

**Risico's**

- De sleutel is een string die iemand in WordPress typt. `Burgermeester Buiningpark` is nu al verkeerd gespeld (moet Burgemeester zijn); wordt die typo gecorrigeerd, dan mist de key. Vandaar normalisatie plus een expliciete melding in plaats van stille terugval.
- Het gepubliceerde programma loopt van 5 september tot 14 oktober — het seizoen is nog niet volledig gepubliceerd. Elke ICS is dus per definitie een momentopname en moet later opnieuw gegenereerd worden; de stabiele UID zorgt dat dat een bijwerking is en geen duplicaat.
- Alles hangt nu aan één pagina. Verandert de opmaak van `/ws/competitie-programma/`, dan ligt de hele tool plat in plaats van een deel ervan.
