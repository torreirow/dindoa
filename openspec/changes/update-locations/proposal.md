# Proposal

## Why

De ingebouwde locatielijst is een momentopname per release. Begin november begon het
zaalseizoen en verscheen er een nieuwe reeks sporthallen in het programma: 61 hallen
met samen 343 wedstrijden, waarvan er nu geen enkele een adres heeft. Een gebruiker kan
dit nu alleen oplossen door zelf adressen op te zoeken en in
`~/.config/dindoa/locations.json` te plakken, of door op een nieuwe release te wachten.
Locaties wisselen per seizoen, releases niet, dus de gebruiker moet de eigen lijst
tussendoor kunnen aanvullen.

## What Changes

- Nieuw commando `dindoa --update-locations` dat het gebruikersbestand
  (`~/.config/dindoa/locations.json`, platformspecifiek via de configmap) aanvult met
  locaties die in het actuele wedstrijdprogramma voorkomen maar nog geen bruikbare regel
  hebben.
- Bronnen, in deze volgorde:
  1. de nieuwste gecontroleerde mapping op de `main`-branch van de repository op GitHub.
     Een regel die daar staat maar niet in de ingebouwde mapping of het gebruikersbestand,
     komt in het gebruikersbestand;
  2. OpenStreetMap (Nominatim) voor wat daarna nog ontbreekt. Het commando doet hooguit één
     verzoek per seconde en stuurt een generieke User-Agent zonder persoonsgegevens. Zulke
     regels krijgen `source: "osm-auto"`, zodat ze herkenbaar en later na te controleren zijn.
- Het commando overschrijft nooit een bestaande regel in het gebruikersbestand: een
  handmatige correctie gaat altijd voor.
- Een locatie waarvoor geen eenduidige treffer is, gokt het commando niet. In een
  terminal vraagt het de gebruiker om straat, postcode en woonplaats en zoekt het dat adres
  op in OpenStreetMap. Lukt dat niet, dan vraagt het of de gebruiker GPS-coördinaten wil
  invoeren. Ingevoerde regels krijgen `source: "manual"`. Buiten een terminal stelt het
  commando geen vragen en meldt het de locatie als ontbrekend.
- Na afloop toont het commando wat er uit welke bron is toegevoegd, wat ontbreekt en de
  nieuwe dekking.
- Gewone lookups bij het genereren van een ICS blijven volledig offline. Alleen dit
  expliciete commando gebruikt het netwerk.
- Het onderhoud van de ingebouwde lijst verandert niet: elke release levert de ingebouwde
  mapping mee, en het commando vult alleen aan.

## Capabilities

### New Capabilities
- `location-update`: de gebruikersmapping aanvullen uit externe bronnen (de gecontroleerde
  mapping in de repository, en daarna OpenStreetMap). Hieronder vallen de bronvolgorde,
  dat bestaande regels nooit worden overschreven, de eisen voor een treffer, het netjes
  gebruiken van OpenStreetMap en de rapportage.

### Modified Capabilities
- `location-mapping`: de eis "geen netwerk" geldt expliciet voor het opzoeken bij genereren
  en de locatiestatus. Het bijwerken van de gebruikersmapping is daarvan uitgezonderd.
  `osm-auto` wordt een erkende bronwaarde.
- `cli-interface`: nieuwe vlag `--update-locations`, die ook in de help staat.

## Impact

- `cmd/dindoa/main.go`: vlag, dispatch en helptekst.
- `internal/locations/`: functie om het gebruikersbestand terug te schrijven (nu wordt het
  alleen gelezen), merge-logica, en een nieuwe subpackage of nieuw bestand voor de twee
  bronnen (GitHub raw en Nominatim) met een injecteerbare HTTP-client voor tests.
- Netwerk: `raw.githubusercontent.com` en `nominatim.openstreetmap.org`, alleen bij dit
  commando. Er komen geen nieuwe Go-dependencies bij (alleen `net/http` en
  `encoding/json`), dus de `vendorHash` in `flake.nix` blijft gelijk.
- README: uitleg over het commando en over `osm-auto`-regels.
