## 1. Programma-pagina ophalen en parsen

- [x] 1.1 Voeg in `internal/scraper/fetcher.go` een fetch toe voor `https://dindoa.nl/ws/competitie-programma/`
- [x] 1.2 Breid `Match` in `internal/scraper/types.go` uit met de kleur en met gestructureerde locatiegegevens (naam, adres, coördinaten), en verwijder de aanname dat `Location` overschreven wordt
- [x] 1.3 Implementeer een kolomcontrole die vóór het parsen van rijen verifieert dat de koppen `Tijd | Thuis | Uit | Kleur | Locatie | Scheidsrechter` zijn, en een expliciete fout geeft bij een andere vorm — nooit een lege lijst
- [x] 1.4 Implementeer het parsen van de container: loop over de siblings van `div.page-content`, houd de laatst geziene `<h3>` als datum-state bij, en koppel elke tabelrij aan die datum
- [x] 1.5 Implementeer het parsen van Nederlandse maandnamen uit de `<h3>` (`5 september` → dag 5, maand 9)
- [x] 1.6 Implementeer de jaarafleiding: augustus t/m december → openingsjaar van het seizoen, januari t/m juli → volgend jaar
- [x] 1.7 Implementeer de validatie van het afgeleide jaar op dag-van-de-week (zaterdag/woensdag) en geef een duidelijke fout als het patroon niet klopt
- [x] 1.8 Parse de kleur uit `<span class="colorcode ...">` en behandel een lege cel als "geen kleur" in plaats van als ontbrekende data
- [x] 1.9 Herleid `ParseCategories` naar de kleurkolom, met een aparte categorie voor teams zonder kleur (senioren, MW, U-teams)
- [x] 1.10 Herleid `ParseTeams` naar de Dindoa-teams uit het programma, ontdubbeld, gegroepeerd per categorie
- [x] 1.11 Herschrijf `NormalizeTeamName` zodat gebruikersinvoer (`j4`, `J4`, `dindoa j4`, `Dindoa J4`) naar de displaynaam `Dindoa J4` resolveert, en houd de slugvorm apart voor de standaard bestandsnaam
- [x] 1.12 Implementeer exacte matching op volledige teamnaam bij het filteren van wedstrijden
- [x] 1.13 Bepaal thuis/uit door de volledige naam van het gekozen team te vergelijken, zodat een Dindoa-tegen-Dindoa wedstrijd eenduidig blijft
- [x] 1.14 Voeg tests toe met vastgelegde HTML uit de echte pagina voor: J4 levert 6 wedstrijden; J1 levert 6 en niet 66; J2 levert 6 en niet 36; `Dindoa 4` is niet `Dindoa J4`; een tegenstander `Unitas/Perspectief J4` komt niet in de J4-selectie
- [x] 1.15 Voeg een test toe die aantoont dat een gewijzigde kolomvorm een fout geeft en geen lege lijst

## 2. Locatie-mapping

- [x] 2.1 Maak een nieuw pakket voor de locatie-mapping met het type voor een mappingregel (naam, adres, coördinaten, OSM-referentie, bron)
- [x] 2.2 Implementeer sleutelnormalisatie: lowercase, witruimte samentrekken, interpunctie strippen — zonder benaderende matching
- [x] 2.3 Voeg het meegeleverde mappingbestand toe en laad het met `go:embed`
- [x] 2.4 Implementeer het inlezen van het gebruikersbestand uit `xdg.ConfigHome` (niet `CacheHome`) en het per-sleutel overschrijven en aanvullen van de meegeleverde mapping
- [x] 2.5 Zorg dat een ontbrekend gebruikersbestand geen waarschuwing geeft, en een onleesbaar gebruikersbestand een melding met het pad geeft waarna de meegeleverde mapping gewoon gebruikt wordt
- [x] 2.6 Implementeer het opzoeken van een locatie, met een resultaat dat expliciet "niet gevonden" kan zijn
- [x] 2.7 Voeg tests toe voor: overschrijven van één sleutel laat andere regels ongemoeid; toevoegen van een nieuwe sleutel; normalisatie vangt hoofdletter-, witruimte- en interpunctieverschillen; een bijna-treffer wordt als ontbrekend behandeld

## 3. ICS-uitvoer

- [x] 3.1 Zet `LOCATION` samen uit de leesbare naam en het adres, en gebruik bij een onbekende locatie de originele websitestring ongewijzigd
- [x] 3.2 Voeg `GEO` toe met `event.SetGeo(lat, lng)` wanneer coördinaten bekend zijn, en laat de property weg wanneer niet
- [x] 3.3 Voeg `DTEND` toe op één uur na `DTSTART`
- [x] 3.4 Voeg `CATEGORIES` toe met de kleur, en laat de property weg bij teams zonder kleur
- [x] 3.5 Herschrijf `generateUID` zodat de UID uit team, datum en tegenstander bestaat en de aanvangstijd er niet in zit
- [x] 3.6 Laat de tijdzonebehandeling ongemoeid — `SetStartAt` doet al `t.UTC().Format(...)` en levert correct `20260905T114500Z` voor 13:45 CEST
- [x] 3.7 Voeg tests toe voor: een gewijzigde aanvangstijd levert dezelfde UID; twee wedstrijden op dezelfde datum tegen verschillende tegenstanders leveren verschillende UID's; `GEO` ontbreekt netjes bij een onbekende locatie; `DTEND` staat een uur na `DTSTART`

## 4. CLI en meldingen

- [x] 4.1 Voeg de vlag `--list-locations` toe die alle locaties uit het programma toont met mappingstatus en adres
- [x] 4.2 Sorteer die uitvoer op aantal wedstrijden, aflopend, zodat de locatie met de meeste impact bovenaan staat
- [x] 4.3 Toon per locatie het aantal wedstrijden, en onderaan een samenvatting van hoeveel locaties gemapt zijn en welk deel van de wedstrijden dat dekt
- [x] 4.4 Geef voor ontbrekende locaties een plak-klaar fragment plus het pad van het gebruikersbestand
- [x] 4.5 Laat `--list-locations` bij een onbereikbare programma-pagina een fout geven met exitcode ongelijk 0
- [x] 4.6 Meld na het genereren van een ICS elke ontbrekende locatie met het aantal betrokken wedstrijden, en eindig met exitcode 0 omdat het bestand geldig is
- [x] 4.7 Geef bij een onbekend team een fout die de beschikbare teams opsomt, met exitcode ongelijk 0
- [x] 4.8 Zorg dat een team zonder wedstrijden in het gepubliceerde deel een duidelijke melding geeft in plaats van stil niets te doen — geïmplementeerd, maar niet live te reproduceren: de teamlijst wordt uit de wedstrijdrijen afgeleid, dus een team dat in de lijst staat heeft per definitie minstens één wedstrijd. De tak is defensief
- [x] 4.9 Voeg `--list-locations` toe aan het handgeschreven `Usage:`-blok van `printUsage()` én aan `Examples:` — `flag.PrintDefaults()` vult alleen de `Options:`-lijst
- [x] 4.10 Loop `printUsage()` na op consistentie tussen beide helften; `--output` staat er nu met twee streepjes in de usage-regels en met één in de gegenereerde lijst

## 5. Interactieve UI aansluiten

- [x] 5.1 Laat het laden van categorieën in `internal/ui` de programma-bron gebruiken in plaats van de teamsoverzichtspagina
- [x] 5.2 Laat de teamselectie de teams uit het programma gebruiken, inclusief de categorie voor teams zonder kleur
- [x] 5.3 Toon ontbrekende locaties in het afrondingsscherm, in dezelfde bewoording als de CLI
- [x] 5.4 Controleer dat `dindoa start` end-to-end werkt: categorie kiezen, team kiezen, bestand geschreven

## 6. Geocoding uitfaseren

- [x] 6.1 Verwijder de aanroep van de geocoder uit het genereerpad in `cmd/dindoa/main.go`
- [x] 6.2 Verwijder `internal/geocode` of reduceer het tot wat de mapping nog nodig heeft; de HTTP-client, de rate limiter en de resultatencache vervallen
- [x] 6.3 Controleer dat er tijdens het genereren van een ICS geen enkel netwerkverzoek naar Nominatim meer uitgaat
- [x] 6.4 Verifieer dat er na het ophalen van het programma geen netwerk meer aan te pas komt — gemeten via een lokale proxy: precies 1 verbinding (`CONNECT dindoa.nl:443`) voor een volledige `--team j4`. De oorspronkelijke formulering ("genereert zonder internetverbinding") is niet haalbaar zonder een cache van de programma-pagina; dat is buiten scope

## 7. Mappingdata vullen

- [x] 7.1 Neem de 23 reeds opgezochte locaties op in het meegeleverde bestand, met OSM-referentie en bron per regel
- [x] 7.2 Bevestigd door de gebruiker (23-08-2026): Watervalweg 170, 3853 PT Ermelo. Kandidaten waren: OSM heeft drie kandidaten binnen 350 meter — het naamloze `sport=korfball`-veld `way/563472182`, `Sporthal Dindoa` `way/285497494` (Watervalweg 170) en het gemeentelijke `Sporthallen Zanderij` `node/12449716882` (Oude Telgterweg 203). Dit raakt 105 van de 210 wedstrijden
- [x] 7.3 Alle 12 aangevuld met adressen van de vereniging (23-08-2026); coordinaten afgeleid en op postcode gecontroleerd. Lijst C is leeg: 35/35 locaties, 210/210 wedstrijden
- [x] 7.4 10 van de 11 aangevuld met huisnummer door de vereniging (23-08-2026); bij 7 daarvan ook straat of postcode gecorrigeerd, coordinaten herbepaald en op postcode gecontroleerd. Sportpark De Woerd blijft op straatniveau: huisnummer niet vastgesteld
- [x] 7.5 Verwijder de annotatievelden die per seizoen verlopen uit het uitgeleverde bestand en houd de bronverwijzing

## 8. Documentatie

- [x] 8.1 Werk de feature-bullet op `README.md:14` bij; automatische geocoding via OpenStreetMap dekt de lading niet meer
- [x] 8.2 Werk `README.md:284` bij: de locatie in het event is nu naam plus adres, met coördinaten in `GEO`
- [x] 8.3 Herschrijf de cachesectie `README.md:292-303`; cache en mapping zijn twee verschillende dingen
- [x] 8.4 Voeg een tabel toe met de platformspecifieke paden van het gebruikersbestand in de configmap, naast de bestaande tabel met cachepaden, en benoem expliciet dat dit twee verschillende mappen zijn
- [x] 8.5 Herschrijf de troubleshooting-sectie `README.md:372-384`, die nu adviseert `~/.cache/dindoa/geocode.json` met de hand te bewerken
- [x] 8.6 Verwijder de technische noot over de Nominatim-rate-limit op `README.md:423`
- [x] 8.7 Documenteer `--list-locations` in de gebruikssectie van de README
- [x] 8.8 Documenteer dat het gepubliceerde programma een momentopname is, dat opnieuw genereren bestaande events bijwerkt, en dat een eerder geïmporteerde agenda uit een oudere versie eenmalig verwijderd moet worden vanwege de gewijzigde UID-vorm
- [x] 8.9 Voeg een noot toe dat een bestaand `geocode.json` in de cachemap niet meer gebruikt wordt en verwijderd kan worden
- [x] 8.10 Werk `CHANGELOG.md` bij onder `## NEXT VERSION` met Added, Changed en Fixed

## 9. Verificatie tegen de echte site

- [x] 9.1 `dindoa --list-categories` levert de categorieën zonder fout
- [x] 9.2 `dindoa --list-all-teams` levert de 35 Dindoa-teams gegroepeerd per categorie
- [x] 9.3 `dindoa --team j4` schrijft een bestand met 6 events, 3 thuis en 3 uit, kleur Rood
- [x] 9.4 Controleer in dat bestand: `DTSTART` van de eerste wedstrijd is `20260905T114500Z`, `DTEND` staat een uur later, `LOCATION` bevat naam plus adres, en `GEO` staat er waar bekend
- [x] 9.5 `dindoa --team j4` meldt de ontbrekende locatie in Terschuur en eindigt met exitcode 0
- [x] 9.6 `dindoa --list-locations` toont 35 locaties met de dekkingssamenvatting
- [ ] 9.7 Importeer het gegenereerde bestand in een agenda-app en controleer dat de wedstrijd tijd inneemt en dat de locatie aantikbaar is
- [x] 9.8 Genereer twee keer achter elkaar en controleer dat de tweede import bijwerkt in plaats van verdubbelt
