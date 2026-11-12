# Design

## Context

- `internal/locations` laadt de meegeleverde mapping (`//go:embed locations.json`) en legt
  daar per sleutel `$XDG_CONFIG_HOME/dindoa/locations.json` overheen. Het pakket leest
  alleen, en de package-doc belooft "Lookups never reach the network".
- `handleListLocations` in `cmd/dindoa/main.go` bepaalt al welke locaties uit het
  programma ontbreken (`prog.LocationCounts()` plus `mapping.Lookup`). Het bijwerkcommando
  gebruikt dezelfde bepaling.
- Een locatiestring heeft de vorm `<naam> <PLAATS IN HOOFDLETTERS>`, bijvoorbeeld
  `Sportcentrum De Koploper LELYSTAD` of `Hoenderdaal DRIEBERGEN-RIJSENBURG`.
- Een handmatige proef op de 61 hallen van het zaalseizoen liet zien dat een vrije zoekopdracht
  in Nominatim op `<naam>, <plaats>` voor eigennamen als `Dindoa-hal` vaak niets oplevert,
  en anders soms een straat of school. Een filter op de soort treffer is dus nodig.
- `internal/scraper` heeft al een HTTP-client met time-out en de User-Agent
  `Dindoa-ICS-Generator/1.0`.

## Goals / Non-Goals

**Goals:**
- Liever een gemiste dan een verkeerde locatie, dezelfde houding als de bestaande
  "geen benaderende treffer"-regel.
- Bronnen zijn te vervangen door nepversies, zodat de merge- en matchlogica zonder netwerk
  te testen is.
- Het commando is herhaalbaar: een tweede run zonder wijzigingen in het programma voegt
  niets toe en doet geen verzoeken.

**Non-Goals:**
- Geen automatische update bij gewone commando's, en geen achtergrond- of periodieke update.
- Regels in het gebruikersbestand die de repository inmiddels beter heeft, worden niet
  bijgewerkt of opgeschoond. De gebruiker verwijdert zo'n regel zelf en draait het commando
  opnieuw.
- Geen interactieve keuze tussen meerdere OSM-kandidaten.
- Geen cache van OSM-antwoorden: het gebruikersbestand is de cache.

## Decisions

### D1. Code-indeling

- `internal/locations/userfile.go`: `ReadUserFile(path)` en `WriteUserFile(path, file)`.
  Het schrijven gaat via een tijdelijk bestand in dezelfde map plus `os.Rename`, zodat een
  half geschreven bestand nooit het origineel vervangt. Mappermissie `0o755`,
  bestandspermissie `0o644`.
- `internal/locations/update.go`: `Update(ctx, missing []string, user *file, sources ...Source) Result`.
  Dit is pure logica: ontbrekende sleutels langs de bronnen in volgorde, alleen
  sleutels toevoegen die nog niet in `user` staan, en per locatie de herkomst of de reden
  van de misser vastleggen.
- `internal/locations/source.go`: de interface
  `Source { Name() string; Find(ctx, venues []string) (map[string]Entry, map[string]string /*reden*/, error) }`
  met twee implementaties: `RepoSource` (GitHub raw) en `NominatimSource`.
- `cmd/dindoa/main.go`: de handler `handleUpdateLocations` verbindt alles en drukt het
  rapport af.

*Alternatief:* een apart package `internal/geocode`. Dat hebben we niet gedaan, omdat de
bronnen `Entry`, `NormalizeKey` en `parse` nodig hebben. Een apart package zou die
moeten exporteren zonder dat iets anders ze gebruikt.

### D2. Repository-bron

- Deze bron haalt
  `https://raw.githubusercontent.com/torreirow/dindoa/main/internal/locations/locations.json`
  op met een time-out van 15 s en parseert die met dezelfde `parse` als de meegeleverde mapping.
- Hij neemt alleen regels over waarvan de genormaliseerde sleutel in `missing` staat en die
  niet `IsEmpty()` zijn. De regel gaat ongewijzigd over, inclusief `source`.
- Als ophalen of parsen mislukt, geeft de bron een fout terug. `Update` meldt die en gaat
  verder met de volgende bron voor alle ontbrekende sleutels.

*Alternatief:* de mapping uit de laatste GitHub Release halen. Een nieuwe hal komt
dan pas na een release beschikbaar, en juist dat gat wil dit commando dichten. Daarom
lezen we `main`. Het risico van een onaf commit op `main` is klein, omdat dit bestand
alleen data bevat en altijd via de parse loopt.

### D3. Nominatim-bron en matchregels

Per locatie:

1. Splits de string in naam en plaats. De plaats is het laatste blok hoofdletterwoorden
   (inclusief `-` en spaties); `KOOG AAN DE ZAAN` blijft één plaats. Lukt het splitsen
   niet, dan is de reden "geen plaats herkend".
2. Vraag op: `GET /search?q=<naam>, <plaats>&format=jsonv2&addressdetails=1&extratags=1&countrycodes=nl&limit=10`.
   Levert dat geen geschikte kandidaat op, dan volgt één tweede poging met `q=sporthal <naam>, <plaats>`.
3. Een kandidaat is geschikt als:
   - de soort `leisure=sports_hall`, `leisure=sports_centre`, `building=sports_hall`
     of `leisure=pitch` met `sport=korfball` is; en
   - de plaats in het adres (`city`, `town`, `village` of `municipality`) na normalisatie
     gelijk is aan de plaats uit de string.
4. Blijft er precies één geschikte kandidaat over, of liggen alle geschikte kandidaten
   binnen 150 m van elkaar (een hal en het complex eromheen), dan nemen we de kandidaat met
   de meest specifieke soort (eerst `sports_hall`, dan `sports_centre`). Bij meer dan één
   die verder uit elkaar liggen, is de reden "niet eenduidig (n kandidaten)".
5. De regel wordt: `name` = de naam uit OSM als die er is, anders de naam uit de string;
   `address` = `<road> <house_number>, <postcode> <plaats>`, waarbij ontbrekende delen
   wegvallen; `lat`/`lon` en `osm` = `<type>/<id>` uit het antwoord; `source` = `osm-auto`.

- Het tempo regelt een `time.Ticker` van 1100 ms rond elk verzoek, dus ook rond de
  tweede poging. Bij 61 hallen duurt een volledige run ruim een minuut. Daarom toont het
  commando per locatie voortgang op stderr.
- De User-Agent is `Dindoa-ICS-Generator/1.0 (+https://github.com/torreirow/dindoa)`:
  een constante in de code die het scraper-package deelt. Er gaat geen e-mailadres mee.
- Bij HTTP 429, 403 of een netwerkfout stopt de bron meteen. De resterende locaties krijgen
  de reden "niet opgezocht: <fout>", en wat al gevonden was, wordt toch weggeschreven.

*Alternatief:* Overpass, zoeken op `leisure=sports_hall` in de gemeentegrens. Dat is
nauwkeuriger voor halnamen, maar traag (de proef duurde minuten), zwaar voor de publieke
dienst, en het levert geen adres. Daarvoor zou nog een tweede, reverse-verzoek nodig
zijn. We houden het nu bij Nominatim. Overpass kan later als derde `Source` achter dezelfde
interface.

### D4. Gebruikersbestand lezen en schrijven

- `ReadUserFile` maakt onderscheid tussen een ontbrekend bestand (geeft een leeg `file` terug,
  `Version: 1`) en een onleesbaar bestand (geeft een fout, en het commando stopt vóór
  er een netwerkverzoek is gedaan).
- Bij het terugschrijven gaat het bestand als geheel opnieuw door `json.MarshalIndent`
  (twee spaties, sleutels gesorteerd). De inhoud van bestaande regels blijft gelijk, alleen
  de opmaak kan veranderen. Wie zijn bestand met de hand opmaakt, merkt dat één keer.
  Dat accepteren we liever dan zelf JSON te gaan patchen.
- Het commando schrijft alleen als er minstens één regel is toegevoegd.

### D5. CLI

- Nieuwe vlag `--update-locations`. Die is te combineren met niets behalve zichzelf. Bij
  `--team`, `--list-*`, `--category` of `--output` geeft het commando een fout met exitcode
  2, en het controleert dat vóórdat het programma wordt opgehaald.
- `--list-locations` toont achter het adres `[osm-auto]` als de regel die bron heeft, zodat
  automatische regels herkenbaar zijn (spec location-mapping).
- Het rapport bestaat uit drie blokken: `Added from repository`, `Added from OpenStreetMap
  (please verify)` en `Still missing`, elk met reden. Daarna volgen de dekkingsregel, die
  hetzelfde is als bij `--list-locations` (in een gedeelde helper), en het pad van het
  gebruikersbestand.

### D6. Handmatige invoer

- Dit draait alleen als zowel stdin als stdout een terminal is (`f.Stat()` met
  `os.ModeCharDevice`, alleen stdlib: `golang.org/x/term` staat niet in `go.mod`, en een
  nieuwe directe dependency zou de `vendorHash` wijzigen). Anders wordt het overgeslagen en blijft het rapport
  zoals in D5.
- De handmatige invoer komt na de automatische bronnen. Eerst lopen alle locaties door de
  repository en Nominatim, daarna volgen de vragen per ontbrekende locatie. Zo hoeft de
  gebruiker niet te wachten tussen vragen door.
- De vragen gaan per locatie met gewone regel-invoer via `bufio.Scanner`, zonder bubbletea.
  Dat is een korte reeks vragen, geen scherm. Ook is het zo te testen met een `io.Reader` en
  `io.Writer`.
  1. `Street and number:` (leeg = overslaan), `Postcode:`, `Town:` (standaard de plaats uit de
     locatiestring).
  2. Gestructureerd zoeken in Nominatim:
     `/search?street=<straat>&postalcode=<pc>&city=<plaats>&countrycodes=nl&format=jsonv2&addressdetails=1&limit=1`,
     met dezelfde tempobegrenzing. Er is geen soortfilter, want de gebruiker heeft het adres
     gekozen. De postcode moet overeenkomen (na het weghalen van spaties en hoofdletters), anders telt het als niet gevonden.
  3. Na een treffer toont het commando het adres uit OSM en vraagt `Use this? [Y/n]`.
  4. Bij geen treffer of `n` vraagt het `Enter GPS coordinates? [y/N]`, daarna
     `Latitude, longitude:`. De invoer is één regel; `52.29, 5.60` en `52.29 5.60` werken allebei.
     De waarden worden gecontroleerd op bereik. Valt een punt buiten Nederland en de
     buurlanden (breedtegraad 50–54, lengtegraad 3–8), dan volgt nog een extra vraag.
  5. Het adres dat in de regel komt, is altijd wat de gebruiker heeft ingevoerd, genormaliseerd
     tot `Straat 1, 1234 AB Plaats`. Uit OSM komen alleen de coördinaten en `osm`.
- Bij Ctrl+C of EOF (Ctrl+D) stopt de invoer. Wat al bevestigd is, wordt weggeschreven. Daarom
  is de schrijfstap één keer aan het eind, ook na een onderbreking: SIGINT wordt afgevangen
  met `signal.NotifyContext` en de context wordt doorgegeven.

*Alternatief:* de vragen in de bestaande bubbletea-UI. Die UI is gebouwd rond het kiezen van een
team, en een formulier daarin kost veel meer code voor een paar regels invoer.

### D7. Testen

- Voor de bronnen gebruiken de tests `httptest.Server`. De basis-URL en de `*http.Client` zijn
  velden op de struct, en de standaardwaarden zet de constructor.
- Fixtures voor Nominatim: echte, ingekorte antwoorden voor een eenduidige treffer, een straat,
  een andere plaats en twee kandidaten ver uit elkaar.
- De begrenzing van het tempo is te testen met een in te stellen interval (in tests 0)
  plus een test die met een nep-klok nagaat dat de wachttijd minstens het interval is.

## Risks / Trade-offs

- [OSM vindt de verkeerde hal met de juiste soort en plaats, bijvoorbeeld de gymzaal van
  een school die ook `sports_hall` is] → `source: osm-auto`, het rapport vraagt om controle
  en `--list-locations` toont de bron. Een fout adres is in de agenda zichtbaar en met één
  regel in het gebruikersbestand te corrigeren.
- [Veel hallen worden niet gevonden, omdat korfbalverenigingen hun hal anders noemen dan OSM]
  → het rapport noemt per misser de reden, en de echte oplossing blijft de gecontroleerde
  lijst in de repository, die de eerste bron is. De onderhouder vult `locations.json` op
  `main` aan, en gebruikers krijgen dat met dit commando zonder release.
- [Nominatim-beleid verandert of blokkeert] → de bron faalt netjes, eerdere resultaten
  blijven staan en de repository-bron werkt onafhankelijk daarvan.
- [Een regel in het gebruikersbestand blijft achter terwijl de repository een betere heeft]
  → bewust buiten scope (zie Non-Goals). README legt uit: verwijder de regel en draai
  opnieuw.
- [Een opmaakwijziging in het gebruikersbestand bij het eerste terugschrijven] → zie D4,
  geaccepteerd.

## Migration Plan

Er is geen migratie nodig. Een bestaand gebruikersbestand blijft geldig en wordt alleen
aangevuld. Om terug te gaan verwijdert de gebruiker regels met `"source": "osm-auto"`,
of het hele bestand.
