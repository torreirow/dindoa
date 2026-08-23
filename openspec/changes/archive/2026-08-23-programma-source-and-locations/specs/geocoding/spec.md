## REMOVED Requirements

### Requirement: Geocode venue names using OpenStreetMap Nominatim

**Reason**: Gemeten op de 35 unieke locaties van het seizoen levert deze route circa 40% treffers op, en de mislukkingen zijn niet het ergste. `Het Slingerbos HARDERWIJK` levert een resultaat 702 meter naast de werkelijke locatie op, en omdat de code de locatie van de website overschrijft met de `display_name` van de API, verdwijnt daarbij een correcte leesbare venuenaam. De eigen accommodatie `De Zanderij (Dindoa) ERMELO`, 105 van de 210 wedstrijden, wordt nooit gevonden. Een gestructureerde zoekopdracht is geen uitweg: Nominatim staat het combineren van een vrije zoekterm met een plaatsparameter niet toe en antwoordt met HTTP 400, en er is geen parameter voor een accommodatienaam.

**Migration**: Locatiegegevens komen uit de capability `location-mapping`. De mapping wordt met de binary meegeleverd, dus er is geen actie nodig van gebruikers. Wie een locatie wil aanpassen of aanvullen gebruikt het gebruikersbestand in de configmap; `dindoa --list-locations` toont het pad en de ontbrekende locaties.

### Requirement: Rate limit geocoding requests

**Reason**: Zonder verzoeken aan Nominatim is er niets te beperken. Het wegvallen hiervan is een direct voordeel voor de gebruiker: het genereren van een ICS kostte minimaal één seconde per onbekende locatie.

**Migration**: Niet van toepassing; er worden tijdens het genereren geen netwerkverzoeken voor locaties meer gedaan.

### Requirement: Cache geocoding results

**Reason**: De cache verergerde het probleem in plaats van het te verzachten. `fallback()` geeft bij een mislukking een resultaat terug met de oorspronkelijke tekst en coördinaten 0,0, en dat wordt onvoorwaardelijk opgeslagen. Er is geen geldigheidsduur en geen invalidatie — het `Version`-veld bestaat, maar wordt bij het inlezen overschreven en nooit gecontroleerd. De eerste run zette daardoor zowel de foute treffer als de mislukkingen permanent vast.

**Migration**: De mappingfile neemt deze rol over, maar als door mensen onderhouden invoer in de configmap in plaats van een herbouwbare afgeleide in de cachemap. Een bestaand `geocode.json` in de cachemap wordt niet meer gelezen en kan verwijderd worden.

### Requirement: Geocode all match locations

**Reason**: Er wordt niet meer gegeocodeerd. Het onderliggende doel — elke wedstrijd, thuis en uit, krijgt een bruikbare locatie — blijft gelden en is opnieuw vastgelegd in `location-mapping`.

**Migration**: Zie `location-mapping`, requirement "Locatie opzoeken via mappingfile" en "Ontbrekende locatie blokkeert de uitvoer niet".

### Requirement: Include User-Agent in API requests

**Reason**: Er worden geen verzoeken aan Nominatim meer gedaan waarin een User-Agent meegestuurd moet worden.

**Migration**: Niet van toepassing.
