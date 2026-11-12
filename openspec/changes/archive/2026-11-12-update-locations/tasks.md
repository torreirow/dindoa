# Tasks

## 1. Gebruikersbestand lezen en schrijven

- [x] 1.1 Voeg `ReadUserFile(path)` toe in `internal/locations/userfile.go`: een ontbrekend bestand geeft een leeg `file` met `Version: 1`, een onleesbaar bestand geeft een fout met het pad erin. Klaar als tests in `userfile_test.go` beide gevallen dekken.
- [x] 1.2 Voeg `WriteUserFile(path, f)` toe: maakt de map aan, schrijft naar een tijdelijk bestand in dezelfde map, daarna `os.Rename`, met `MarshalIndent` en gesorteerde sleutels. Klaar als een test aantoont dat de inhoud na lezen-schrijven-lezen gelijk blijft, en dat een mislukte write (de map read-only) het origineel laat staan.

## 2. Bronnen

- [x] 2.1 Definieer de interface `Source` in `internal/locations/source.go` en deel de User-Agent-constante met `internal/scraper`. Klaar als `go build ./...` slaagt en de scraper dezelfde constante gebruikt.
- [x] 2.2 Implementeer `RepoSource` (raw-URL van `main`, time-out 15 s, `parse`, filter op de ontbrekende sleutels, lege regels overslaan). Klaar als `httptest`-tests een treffer, een lege regel, een sleutel buiten de lijst, HTTP 404 en ongeldige JSON dekken.
- [x] 2.3 Implementeer het splitsen van een locatiestring in naam en plaats (D3.1). Klaar als een tabeltest slaagt met in elk geval `Dindoa-hal ERMELO`, `Hoenderdaal DRIEBERGEN-RIJSENBURG`, `Topsportcentrum De Koog KOOG AAN DE ZAAN`, `'t Dok DRONTEN` en een string zonder hoofdletterplaats.
- [x] 2.4 Implementeer `NominatimSource` met de zoekopdracht, de tweede poging en de matchregels uit D3 (soortfilter, gelijke plaats, de 150 m-regel, voorkeur voor `sports_hall`) en de opbouw van de regel met `source: "osm-auto"`. Klaar als `httptest`-fixtures slagen voor een eenduidige treffer, een straat, een andere plaats, twee kandidaten ver uit elkaar, en hal plus complex binnen 150 m.
- [x] 2.5 Voeg de tempobegrenzing (instelbaar interval, standaard 1100 ms) en het stoppen bij 429, 403 of een netwerkfout toe, waarbij de resterende locaties de reden "niet opgezocht" krijgen. Klaar als tests aantonen dat er tussen verzoeken minstens het interval zit en dat na een 429 geen verzoek meer volgt.
- [x] 2.6 Werk de package-doc van `internal/locations` bij: lookups blijven offline, alleen `Source`-implementaties gebruiken het netwerk. Klaar als `go doc ./internal/locations` de nieuwe tekst toont.

## 3. Update-logica

- [x] 3.1 Implementeer `Update(ctx, missing, user, sources...)`: bronnen in volgorde, alleen nieuwe sleutels, en een `Result` met per sleutel de bron of de reden. Klaar als tests met nepbronnen dekken: repo vóór OSM, een bestaande gebruikersregel die nooit wordt overschreven, een falende repo-bron die doorvalt naar OSM, en een lege `missing` die geen bron aanroept.

## 4. Handmatige invoer

- [x] 4.1 Implementeer de vragen per locatie (D6) met een `io.Reader` en `io.Writer`: adres invoeren of overslaan, met de plaats uit de string als standaard. Klaar als tests met een gescripte invoer dekken: overslaan, alle drie de velden ingevuld, en EOF midden in een vraag.
- [x] 4.2 Voeg gestructureerd adres zoeken toe aan `NominatimSource` (met dezelfde tempobegrenzing, en de postcode moet overeenkomen). Klaar als `httptest`-tests een treffer, geen treffer en een treffer met een andere postcode dekken.
- [x] 4.3 Implementeer bevestigen of afkeuren van de treffer, de GPS-vraag, het parsen van coördinaten (komma of spatie) en de bereikcontrole, plus de extra vraag buiten NL en buurlanden. Klaar als een tabeltest voor het parsen slaagt (geldig, ongeldig, buiten bereik) en de dialogtests de paden adres-gevonden, afgekeurd-naar-GPS, GPS-geweigerd (alleen het adres) en ongeldige-invoer-opnieuw dekken.
- [x] 4.4 Schakel de vragen alleen in als stdin en stdout een terminal zijn, en schrijf bij SIGINT of EOF wel weg wat al bevestigd is. Klaar als `dindoa --update-locations < /dev/null` geen vragen stelt, en als een handmatige run met Ctrl+C na één bevestigde locatie die locatie wel in het bestand heeft staan.

## 5. CLI

- [x] 5.1 Voeg de vlag `--update-locations` toe met de controle op combinaties (exitcode 2, vóór er iets wordt opgehaald). Klaar als `dindoa --update-locations --team j3` een fout geeft en exitcode 2.
- [x] 5.2 Implementeer `handleUpdateLocations`: programma ophalen, ontbrekende locaties bepalen (dezelfde logica als `--list-locations`, in een gedeelde helper), gebruikersbestand lezen, `Update`, alleen schrijven als er iets is toegevoegd, voortgang op stderr, en het rapport in drie blokken plus dekking en pad. Klaar als een handmatige run met een tijdelijke `XDG_CONFIG_HOME` het bestand aanmaakt en een tweede run "niets aan te vullen" meldt zonder netwerkverzoek naar een locatiebron.
- [x] 5.3 Laat `--list-locations` `[osm-auto]` tonen bij regels met die bron. Klaar als een test of handmatige run met een gebruikersbestand de markering laat zien.
- [x] 5.4 Werk de helptekst bij (usage-overzicht en voorbeeld, met de vermelding dat het commando het netwerk gebruikt). Klaar als `dindoa --help` de vlag in beide helften toont.
- [x] 5.5 Werk de README bij met een sectie over `--update-locations`: de bronnen, `osm-auto` controleren, hoe je een regel terugzet. Klaar als de sectie in de README staat en de voorbeeldcommando's werken.

## 6. Integratie

- [x] 6.1 `go test ./...` en `go vet ./...` slagen. Klaar als beide schoon zijn.
- [x] 6.2 Doe een echte run tegen het huidige zaalprogramma met een tijdelijke `XDG_CONFIG_HOME` en controleer een steekproef van 5 `osm-auto`-regels op de kaart. Klaar als het resultaat (aantallen per bron, aantal missers en de steekproef) in de PR-beschrijving staat.
  - Resultaat 2026-11-12 (ingebouwde lijst van `main`, lege `XDG_CONFIG_HOME`, 61 zaalhallen): repository 0 (de hallen stonden nog niet op `main`), OpenStreetMap 33 in 1m45s, 28 ontbrekend met reden. Vergeleken met de gecontroleerde lijst: 32 van 33 binnen 25 m; 1 fout: De Sypel Harderwijk → Gelreweg (1,4 km naast Slingerbos 1). Interactief pad via een pty getest: adres invoeren → OSM-treffer → bevestigen → Ctrl+D en Ctrl+C bewaren de invoer.
- [x] 6.3 `nix build .#dindoa` slaagt zonder `vendorHash`-wijziging, omdat er geen nieuwe dependencies zijn. Klaar als de build slaagt.

## Workflow follow-up

- Vul de gecontroleerde `internal/locations/locations.json` op `main` aan met de hallen van het zaalseizoen (een aparte change of commit). Daarmee wordt de repository-bron de primaire bron voor alle gebruikers.
- Archiveer de change na review met `/opsx:archive`.
