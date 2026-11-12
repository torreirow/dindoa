# location-update Specification

## Purpose
Laat een gebruiker de eigen locatiemapping tussen releases door aanvullen. Ontbrekende locaties uit het actuele wedstrijdprogramma worden opgehaald uit de gecontroleerde mapping in de repository en daarna uit OpenStreetMap, zonder handmatige correcties te overschrijven.
## Requirements
### Requirement: Alleen ontbrekende locaties aanvullen

Het systeem SHALL bij het bijwerken alleen locaties behandelen die in het actuele wedstrijdprogramma voorkomen en geen bruikbare regel hebben in de samengevoegde mapping (meegeleverd plus gebruikersbestand).

#### Scenario: Gemapte locatie wordt overgeslagen

- **WHEN** een locatie uit het programma al een bruikbare regel heeft in de meegeleverde mapping of het gebruikersbestand
- **THEN** zoekt het systeem die locatie niet op en verandert het niets aan de regel

#### Scenario: Locatie buiten het programma wordt genegeerd

- **WHEN** de gecontroleerde mapping in de repository een regel bevat voor een locatie die niet in het actuele programma voorkomt
- **THEN** neemt het systeem die regel niet op in het gebruikersbestand

#### Scenario: Alles al gemapt

- **WHEN** alle locaties uit het programma al een bruikbare regel hebben
- **THEN** meldt het systeem dat er niets aan te vullen is, doet het geen verzoek aan een locatiebron en laat het het gebruikersbestand ongewijzigd

### Requirement: Gecontroleerde mapping gaat voor OpenStreetMap

Het systeem SHALL eerst de gecontroleerde mapping van de hoofdbranch van de repository raadplegen en SHALL alleen voor locaties die daarna nog ontbreken OpenStreetMap raadplegen.

#### Scenario: Locatie staat in de gecontroleerde mapping

- **WHEN** een ontbrekende locatie een bruikbare regel heeft in de gecontroleerde mapping in de repository
- **THEN** neemt het systeem die regel ongewijzigd over in het gebruikersbestand en zoekt het die locatie niet op in OpenStreetMap

#### Scenario: Repository is onbereikbaar

- **WHEN** de gecontroleerde mapping niet opgehaald of niet gelezen kan worden
- **THEN** meldt het systeem dat, en gaat het verder met OpenStreetMap voor alle ontbrekende locaties

### Requirement: Alleen eenduidige OpenStreetMap-treffers

Het systeem SHALL een locatie uit OpenStreetMap alleen overnemen als de treffer een sportlocatie is in de plaats uit de locatiestring. Bij geen of twijfelachtige treffer SHALL het systeem de locatie als ontbrekend laten.

#### Scenario: Eenduidige treffer

- **WHEN** OpenStreetMap voor een ontbrekende locatie een sporthal of sportcentrum geeft in de genoemde plaats
- **THEN** schrijft het systeem een regel weg met naam, adres voor zover bekend, coördinaten, de verwijzing naar het OpenStreetMap-object, en bron "osm-auto"

#### Scenario: Treffer in een andere plaats

- **WHEN** de enige treffer in een andere plaats ligt dan de plaats in de locatiestring
- **THEN** schrijft het systeem niets weg voor die locatie en meldt het die als niet gevonden

#### Scenario: Geen sportlocatie

- **WHEN** de treffer geen sportlocatie is, bijvoorbeeld een straat of een school
- **THEN** schrijft het systeem niets weg voor die locatie en meldt het die als niet gevonden

#### Scenario: Meerdere geschikte treffers

- **WHEN** OpenStreetMap meerdere verschillende sportlocaties in de genoemde plaats geeft en niet één daarvan duidelijk de gezochte is
- **THEN** schrijft het systeem niets weg voor die locatie en meldt het die als niet eenduidig

### Requirement: Handmatige invoer voor niet gevonden locaties

Wanneer het systeem interactief draait, SHALL het voor elke locatie die na alle bronnen nog ontbreekt de gebruiker om straat met huisnummer, postcode en woonplaats vragen. Dat adres SHALL het in OpenStreetMap opzoeken. Als dat mislukt, SHALL het vragen of de gebruiker GPS-coördinaten wil invoeren.

#### Scenario: Adres ingevoerd en gevonden

- **WHEN** een locatie niet gevonden is en de gebruiker straat, postcode en woonplaats invoert die OpenStreetMap kent
- **THEN** toont het systeem het gevonden adres. Na bevestiging door de gebruiker schrijft het een regel weg met het ingevoerde adres, de coördinaten en de verwijzing naar het OpenStreetMap-object, en bron "manual"

#### Scenario: Gevonden adres afgekeurd

- **WHEN** de gebruiker het door OpenStreetMap gevonden adres niet bevestigt
- **THEN** vraagt het systeem of de gebruiker GPS-coördinaten wil invoeren, zoals bij een adres dat niet gevonden is

#### Scenario: Adres niet gevonden, coördinaten ingevoerd

- **WHEN** OpenStreetMap het ingevoerde adres niet vindt en de gebruiker kiest voor het invoeren van GPS-coördinaten
- **THEN** vraagt het systeem om breedte- en lengtegraad en schrijft het een regel weg met het ingevoerde adres, die coördinaten en bron "manual"

#### Scenario: Ongeldige coördinaten

- **WHEN** de gebruiker coördinaten invoert die geen getal zijn of buiten het bereik van breedtegraad (-90 tot 90) of lengtegraad (-180 tot 180) vallen
- **THEN** meldt het systeem wat er mis is en vraagt het opnieuw, zonder iets weg te schrijven

#### Scenario: Geen coördinaten

- **WHEN** OpenStreetMap het ingevoerde adres niet vindt en de gebruiker geen GPS-coördinaten wil invoeren
- **THEN** schrijft het systeem een regel weg met alleen het ingevoerde adres en bron "manual"

#### Scenario: Locatie overslaan

- **WHEN** de gebruiker bij de vraag om een adres niets invoert
- **THEN** slaat het systeem die locatie over, schrijft het er niets voor weg en vraagt het naar de volgende ontbrekende locatie

#### Scenario: Niet-interactief

- **WHEN** de standaardinvoer of standaarduitvoer geen terminal is
- **THEN** stelt het systeem geen vragen en meldt het de niet gevonden locaties zoals bij een gewone rapportage

#### Scenario: Afbreken tijdens de vragen

- **WHEN** de gebruiker de invoer afbreekt (Ctrl+C of einde van de invoer)
- **THEN** schrijft het systeem alle tot dan toe gevonden en bevestigde regels weg en stopt het zonder verdere vragen

### Requirement: Gebruikersregels worden nooit overschreven

Het systeem SHALL bestaande regels in het gebruikersbestand nooit wijzigen of verwijderen. Het SHALL alleen nieuwe sleutels toevoegen, en de rest van het bestand inhoudelijk ongemoeid laten.

#### Scenario: Handmatige correctie blijft staan

- **WHEN** het gebruikersbestand een regel bevat voor een sleutel waarvoor ook de repository of OpenStreetMap gegevens heeft
- **THEN** blijft de regel in het gebruikersbestand ongewijzigd

#### Scenario: Gebruikersbestand bestaat nog niet

- **WHEN** er nog geen gebruikersbestand is en er iets aan te vullen valt
- **THEN** maakt het systeem het bestand en de bovenliggende map aan, met alleen de toegevoegde regels

#### Scenario: Gebruikersbestand is onleesbaar

- **WHEN** het gebruikersbestand bestaat maar geen geldige inhoud heeft
- **THEN** schrijft het systeem niets, meldt het welk bestand niet gelezen kon worden en eindigt het met een niet-nul exitcode

#### Scenario: Schrijven is atomair

- **WHEN** het schrijven van het gebruikersbestand halverwege mislukt
- **THEN** blijft het vorige gebruikersbestand intact

### Requirement: OpenStreetMap netjes gebruiken

Het systeem SHALL OpenStreetMap hooguit één keer per seconde bevragen, SHALL zich identificeren met een generieke applicatienaam zonder persoonsgegevens, en SHALL het netwerk alleen gebruiken wanneer de gebruiker het bijwerkcommando expliciet uitvoert.

#### Scenario: Snelheidsbegrenzing

- **WHEN** het systeem meerdere locaties in OpenStreetMap opzoekt
- **THEN** zit er tussen twee opeenvolgende verzoeken minstens één seconde

#### Scenario: Identificatie zonder persoonsgegevens

- **WHEN** het systeem een verzoek aan OpenStreetMap doet
- **THEN** bevat het verzoek een applicatienaam en versie, en geen e-mailadres of andere gegevens van de gebruiker

#### Scenario: OpenStreetMap weigert of is onbereikbaar

- **WHEN** OpenStreetMap een fout geeft of niet bereikbaar is
- **THEN** stopt het systeem met verdere OpenStreetMap-verzoeken, schrijft het wel weg wat al gevonden was, en meldt het welke locaties niet zijn opgezocht

### Requirement: Rapportage van de bijwerking

Het systeem SHALL na het bijwerken per locatie melden uit welke bron die is toegevoegd of waarom die ontbreekt, en SHALL de dekking na afloop tonen.

#### Scenario: Overzicht na afloop

- **WHEN** het bijwerken klaar is
- **THEN** toont het systeem de toegevoegde locaties per bron, de locaties die nog ontbreken met reden, het pad van het gebruikersbestand, en het aantal gemapte locaties en het gedekte deel van de wedstrijden

#### Scenario: Automatische regels zijn te controleren

- **WHEN** er regels met bron "osm-auto" zijn toegevoegd
- **THEN** wijst het systeem de gebruiker erop dat die regels automatisch gevonden zijn en het best gecontroleerd kunnen worden

#### Scenario: Exitcode met ontbrekende locaties

- **WHEN** het bijwerken lukt maar er nog locaties ontbreken
- **THEN** eindigt het systeem met exitcode 0
