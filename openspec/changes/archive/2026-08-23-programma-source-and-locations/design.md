# Design: wedstrijdprogramma als bron + locatie-mappingfile

## Context

Zie `proposal.md` — Why voor de motivatie en de gemeten symptomen.

Wat de aanpak vormt:

**De programma-pagina is één blok HTML met datums buiten de tabellen.** De structuur is:

```html
<div class="page-content table-responsive">
  <h3>5 september</h3>                                      <- datum, geen jaartal
  <table class="table inverse table-hover">
    <thead><tr><th>Tijd</th><th>Thuis</th><th>Uit</th>
               <th>Kleur</th><th>Locatie</th><th>Scheidsrechter</th></tr></thead>
    <tbody>
      <tr><td>11:50</td><td>Dindoa J3</td>
          <td>Antilopen/Bloemendal Bouw J3</td>
          <td><span class="colorcode rood">Rood</span></td>
          <td>De Zanderij (Dindoa) ERMELO</td><td></td></tr>
  </table>
  <h3>9 september</h3> <table>…                              <- 12 blokken, 210 rijen
```

De datum staat dus niet in de rij maar in de voorgaande `<h3>`. Een parser die per `<table>` itereert verliest de datum; je moet over de siblings van de container lopen en de laatst gezien `<h3>` als state meenemen.

**Locatiestrings hebben een strikt patroon.** `<venue> <STAD IN CAPS>`, soms met `(club)` of `(plaats)` ertussen. Splitsen op precies het laatste whitespace-token levert 35 van 35 correct. Greedy op *alle* trailing caps-tokens gaat fout, want `Veld MIA AMERSFOORT`, `Veld ASVD DRONTEN` en `Sportveld DWS KOOTWIJKERBROEK` hebben ook een caps-clubafkorting vóór de stad. Steden met een koppelteken (`DRIEBERGEN-RIJSENBURG`) zijn één token en gaan goed.

**Elke locatie heeft precies één thuisclub.** Over alle 210 wedstrijden is de relatie locatie ↔ thuisclub 1-op-1; geen enkele locatie wordt door twee clubs gebruikt. Dat maakt de clubnaam een bruikbare tweede sleutel bij het opzoeken van adressen, en het is de reden dat 7 van de 10 best-onderbouwde mappingregels via de club gevonden zijn en niet via de venuenaam.

**De bestaande cache heeft al de juiste vorm.** `internal/geocode/cache.go` is `map[normalizeKey(query)]Result` in JSON met een `Version`-veld. Dat is precies wat de mappingfile nodig heeft. Wat ontbreekt is versiebeheer, uitlevering met de binary, en een bron die geen onzin produceert.

**Bestaande correcte gedrag dat behouden moet blijven.** De tijdzonebehandeling is goed: `golang-ical` v0.3.5 `SetStartAt` doet `t.UTC().Format(icalTimestampFormatUtc)`, dus een `time.Date(..., Europe/Amsterdam)` van 13:45 CEST wordt `DTSTART:20260905T114500Z`. Daar is geen bug en die code hoeft niet aangeraakt.

## Goals / Non-Goals

**Goals:**

- Eén HTTP-request naar één pagina levert alles: wedstrijden, teams, categorieën, kleuren en locatiestrings.
- Het genereren van een ICS werkt volledig offline, zonder rate limiting en zonder wachttijd.
- Een ontbrekende of foute locatie kan de uitvoer nooit stiller of slechter maken dan de website zelf al is.
- Locatiegegevens zijn in versiebeheer, na te controleren via een OSM-referentie, en door een gebruiker te overschrijven zonder release.
- Verkeerde jaartallen worden gedetecteerd in plaats van geraden.

**Non-Goals:**

- Geen automatische OSM-resolver in de tool. Het opzoeken van adressen is eenmalig maintainerwerk dat buiten de binary gebeurt; het zou netwerk, rate limiting en ~50% foutmarge terugbrengen in een pad dat we juist deterministisch maken.
- Geen fuzzy matching op locatiesleutels tijdens runtime. Wel normaliseren, niet gokken — een bijna-treffer die het verkeerde adres oplevert is erger dan een gemelde misser.
- Geen ondersteuning voor de oefenwedstrijden-pagina. Die staat op een verouderde datum (`16.04.2026`), gebruikt een ander datum- en tijdformaat (`19.00`), heeft geen locatie- of kleurkolom en bevat alleen Dindoa 1 en 2.
- Geen trainingstijden in de ICS. Die staan op een derde pagina met een derde teamnotatie (`J4` zonder "Dindoa") en zijn een aparte capability.
- Geen wijziging aan de schermen of navigatie van de interactieve UI.

## Decisions

### Programma-pagina als enige bron, geen fallback op de teampagina

Overwogen alternatief: de programma-pagina alleen gebruiken als de teampagina 0 wedstrijden geeft. Dat houdt twee parsers in leven voor een bron die dit seizoen geen wedstrijddata meer bevat, en het levert een tool die stil van gedrag verandert afhankelijk van wat de vereniging in WordPress doet. Bovendien is de programma-pagina rijker: hij heeft kleur en scheidsrechter die de teampagina nooit had.

De teampagina blijft wel bestaan en is nog nuttig voor trainingstijden — maar dat is buiten scope.

Let op bij het opruimen: `/ws/dindoa-j4/` geeft **301 → `/ws/dindoa-j4-3/`** (WordPress-slugcollisie). Go's `http.Client` volgt dat standaard, dus het viel niet op. Elk ontwerp dat een teamnaam uit een URL-slug afleidt zou hier "Dindoa J4 3" produceren; dat is een extra reden om teamnamen uitsluitend uit het programma te halen.

### Jaartal afleiden en valideren op dag-van-de-week

De `<h3>` geeft `5 september`. Afleiden uit het seizoen (augustus–december → huidig jaar, januari–juli → volgend jaar) is de basisregel, maar die is stil fout rond de jaarwisseling en bij een verkeerde systeemklok.

De validatie is goedkoop en sluitend: korfbal speelt zaterdag en woensdag. Voor de 12 gepubliceerde datums geeft alleen 2026 het patroon `za wo za wo …`; 2025 geeft `vr di`, 2027 geeft `zo do`. Als het afgeleide jaar geen zaterdag/woensdag-patroon oplevert, is de aanname fout en moet dat gemeld worden in plaats van doorgerekend.

Overwogen alternatief: een `--season`-vlag. Dat schuift een vraag naar de gebruiker die de tool zelf betrouwbaar kan beantwoorden.

### Exacte matching op volledige teamnaam

`strings.Contains` op de teamcode is onbruikbaar: filteren op `J4` haalt `Revival J4`, `Regio '72 J4` en `Unitas/Perspectief J4` binnen. Ook binnen het Dindoa-domein is prefix-matching onveilig — `Dindoa J1` zit in `J10`–`J19` en `Dindoa J2` in `J20`–`J24`, wat 66 respectievelijk 36 rijen geeft in plaats van 6. Voor de overige 33 teams valt het samen, dus het probleem is makkelijk te missen in een test die J4 gebruikt.

De gebruikersinvoer wordt daarom genormaliseerd naar de displaynaam van het programma (`j4` → `Dindoa J4`), en daarna exact vergeleken met de celtekst. `NormalizeTeamName` produceert nu een URL-slug (`dindoa-j4`); die vorm is niet langer nodig voor het ophalen van data, maar blijft bruikbaar voor de standaard bestandsnaam.

Aandachtspunt voor `IsHome`: nu bepaald met "staat Dindoa in de thuiskolom". In het huidige programma zijn er 0 Dindoa-tegen-Dindoa wedstrijden, dus dat is nu eenduidig. Zodra die voorkomen is de vraag welk van de twee het geselecteerde team is — vergelijken met de volledige naam van het gekozen team lost dat op en kost niets extra.

### Gelaagde mappingfile: embed als basis, gebruikersbestand als bovenlaag

Nix- en binary-installaties hebben geen repo-checkout, dus het meegeleverde bestand moet in de binary (`go:embed`). Maar één foute regel mag een gebruiker niet dwingen op een release te wachten, dus er is een gebruikersbestand dat **per key** overschrijft en aanvult — geen alles-of-niets vervanging.

Het gebruikersbestand hoort in `xdg.ConfigHome`, niet in `xdg.CacheHome`. Het is door de gebruiker onderhouden invoer, geen herbouwbare afgeleide; een `cache clear` mag het niet weggooien. Dat geeft wel twee dindoa-mappen op één systeem, dus beide paden moeten in de README én in de uitvoer van `--list-locations` staan.

### Sleutelnormalisatie: agressiever dan de huidige cache

De huidige `normalizeKey` doet `ToLower(TrimSpace(...))`. Dat breekt op elke opmaakwijziging in WordPress. `Burgermeester Buiningpark` is nu al verkeerd gespeld; wordt die typo ooit gecorrigeerd, dan mist de key.

Normalisatie wordt daarom: lowercase, whitespace samentrekken, interpunctie strippen. Fuzzy matching bewust niet — zie Non-Goals. De stad is een betrouwbaar tweede anker als er ooit een tweede matchronde nodig blijkt, maar dat is nu niet nodig.

### Ontbrekende locatie: doorgaan, en het zeggen

De ICS wordt altijd gegenereerd. Een onbekende locatie levert `LOCATION` met de originele websitestring — die is leesbaar en meestal genoeg om de plek te vinden — en geen `GEO`. Daarna een expliciete melding met het aantal wedstrijden per ontbrekende locatie, zodat de prioriteit meteen duidelijk is.

Overwogen alternatief: falen met een non-zero exit. Dat maakt de tool onbruikbaar voor het ene team dat toevallig naar een niet-gemapte locatie moet, terwijl de rest van de agenda prima is. De huidige code doet het omgekeerde en het slechtste: hij valt stil terug en cachet de mislukking permanent.

### `LOCATION` behoudt de naam, `GEO` krijgt de coördinaten

De huidige code doet `matches[i].Location = result.Address`, en dat is waar informatie verdwijnt. `LOCATION` wordt de leesbare naam plus het adres, `GEO` de coördinaten. Dat lost ook het probleem van adressen op straatniveau op: 13 van de 23 opgezochte regels hebben geen huisnummer in OSM, maar de coördinaten zijn exact, dus navigatie werkt via `GEO` terwijl `LOCATION` leesbaar blijft.

`event.SetGeo(lat, lng)` bestaat in golang-ical v0.3.5.

### UID zonder aanvangstijd

Nu is de UID `dindoa-j4-2026-09-05-1345@dindoa.nl`. Omdat het programma slechts tot 14 oktober gepubliceerd is, moet iedereen later opnieuw genereren — en dan levert elke verzette aanvangstijd een tweede event op in plaats van een wijziging van het bestaande. De UID wordt daarom opgebouwd uit team, datum en tegenstander. Dan is opnieuw importeren een bijwerking.

### Wedstrijdduur vast op 60 minuten

Een `VEVENT` met alleen `DTSTART` als DATE-TIME is volgens RFC 5545 nul seconden lang. Een vlag voor de duur zou de gebruiker laten beslissen over iets waar hij geen betere informatie over heeft dan wij; 60 minuten is voor korfbal een redelijke standaard en is later per leeftijdscategorie te verfijnen zonder de interface te veranderen.

## Risks / Trade-offs

**Alles hangt aan één pagina.** Verandert de opmaak van `/ws/competitie-programma/`, dan werkt niets meer in plaats van een deel. → De parser moet een duidelijke fout geven bij een onverwachte tabelvorm in plaats van 0 wedstrijden terug te geven. Dat is precies de val waar de huidige code in loopt: `time.Parse("02-01-2006", "11:50")` faalt per rij, de rij wordt overgeslagen, en het resultaat is een lege lijst zonder enige melding. Een expliciete controle op de kolomkoppen (`Tijd | Thuis | Uit | Kleur | Locatie | Scheidsrechter`) vóór het parsen van rijen vangt dit.

**De locatiesleutel is redactionele invoer.** Iemand kan de string in WordPress wijzigen en dan mist de mapping. → Normalisatie dempt opmaakverschillen; de melding maakt een misser zichtbaar in plaats van stil.

**Het seizoen is nog niet volledig gepubliceerd** (5 september t/m 14 oktober, 12 datums). Elke ICS is een momentopname. → Stabiele UID maakt opnieuw genereren een bijwerking. Dit is geen op te lossen probleem, alleen een te documenteren verwachting.

**De mappingfile is handwerk bij nieuwe locaties.** Bekerwedstrijden en een nieuwe poule-indeling brengen nieuwe venues. → Naar schatting een handjevol per seizoen, en `--list-locations` plus de melding bij het genereren maken ze zichtbaar. De scheve verdeling helpt: tien regels dekken 80% van de wedstrijden.

**Twee dindoa-mappen op één systeem** (cache en config) is verwarrend. → Beide paden expliciet in de README en in de uitvoer van `--list-locations`.

**Verlies van de geocoding-capability is onomkeerbaar in de specs.** Mocht de mappingfile onhoudbaar blijken, dan is terugvallen op Nominatim een nieuwe change. → Aanvaardbaar: de gemeten kwaliteit van die route is ~40% treffers met stille fouten, dus er gaat niets bruikbaars verloren.

**De helptekst kan uiteenlopen.** `printUsage()` heeft een handgeschreven `Usage:`-blok naast een automatisch gegenereerde `Options:`-lijst van `flag.PrintDefaults()`. Een nieuwe vlag verschijnt automatisch onderaan maar niet in de usage-regels, en de twee helften gebruiken al verschillende conventies (`--output` versus `-output`). → De taak voor de nieuwe vlag moet beide helften expliciet noemen.

## Migration Plan

De tool schrijft alleen een ICS-bestand en heeft geen server-state, dus er is geen migratie in de gebruikelijke zin. Wat wel aandacht vraagt:

1. **Bestaande cache.** `~/.cache/dindoa/geocode.json` wordt niet meer gelezen. Op dit systeem bestaat het bestand niet (`--team j4` stopte vóór de geocoding), dus er is geen vervuilde cache om op te ruimen. Bij gebruikers die de tool eerder succesvol gebruikten staat er mogelijk wel een bestand met foute entries; dat wordt simpelweg genegeerd. Het opruimen ervan is een noot in de README, geen code.

2. **Bestaande ICS-bestanden bij gebruikers.** De UID-vorm verandert, dus een eerder geïmporteerde agenda krijgt bij herimport nieuwe events naast de oude in plaats van bijwerkingen. Eenmalig, en niet te vermijden zonder de oude UID-vorm te blijven produceren. Documenteren als "verwijder de eerder geïmporteerde agenda voor je opnieuw importeert".

3. **Uitrolvolgorde.** De parser en de mappingfile kunnen onafhankelijk gebouwd worden: zolang de mapping leeg is valt alles terug op de originele websitestring, wat nog steeds beter is dan de huidige situatie. Er is dus geen big-bang nodig.

Terugrolstrategie: een eerdere release blijft beschikbaar, maar die is voor dit seizoen niet functioneel — de huidige versie levert nul wedstrijden. Terugrollen is dus geen zinvolle optie, wat de druk op punt 3 verhoogt: lever de parser niet uit zonder dat hij op de echte pagina getest is.

## Open Questions

- Het officiële adres van de eigen accommodatie. OSM heeft drie kandidaten binnen 350 meter: het naamloze `sport=korfball`-veld (`way/563472182`), `Sporthal Dindoa` (`way/285497494`, Watervalweg 170) en het gemeentelijke `Sporthallen Zanderij` (`node/12449716882`, Oude Telgterweg 203). De voorgestelde regel combineert de coördinaten van het korfbalveld met het adres van Sporthal Dindoa. Dit raakt 105 van de 210 wedstrijden en moet door iemand van de vereniging bevestigd worden — maar het is één regel in een datafile en verandert de specs, de aanpak noch de taakverdeling.
- Of de wedstrijdduur later per leeftijdscategorie moet verschillen. Te beantwoorden zonder de interface te wijzigen.
