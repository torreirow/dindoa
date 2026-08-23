# location-mapping Specification

## Purpose
Zet de locatiestrings van dindoa.nl om naar een leesbare naam, een adres en coördinaten via een meegeleverde mappingfile die een gebruiker kan overschrijven, zodat het genereren van een agenda geen netwerk nodig heeft en een ontbrekende locatie zichtbaar wordt in plaats van stil verkeerde gegevens op te leveren.
## Requirements
### Requirement: Locatie opzoeken via mappingfile

Het systeem SHALL locatiegegevens uitsluitend uit een mappingfile halen en SHALL tijdens het genereren van een ICS geen netwerkverzoek doen om een locatie op te lossen.

#### Scenario: Bekende locatie opgezocht

- **WHEN** het wedstrijdschema de locatie "De Zanderij (Dindoa) ERMELO" bevat en die staat in de mapping
- **THEN** levert het systeem de bijbehorende leesbare naam, adres en coördinaten

#### Scenario: Geen netwerk nodig

- **WHEN** de gebruiker een ICS genereert zonder internetverbinding nadat het wedstrijdschema beschikbaar is
- **THEN** worden alle locaties uit de mapping opgezocht zonder netwerkfout en zonder vertraging door rate limiting

#### Scenario: Coördinaten optioneel

- **WHEN** een mappingregel een adres bevat maar geen coördinaten
- **THEN** levert het systeem het adres en meldt het geen fout

### Requirement: Gelaagde mapping met gebruikersoverschrijving

Het systeem SHALL een meegeleverde mapping bevatten die zonder installatiestappen beschikbaar is, en SHALL een optioneel gebruikersbestand inlezen dat de meegeleverde mapping per sleutel overschrijft en aanvult.

#### Scenario: Meegeleverde mapping werkt uit de doos

- **WHEN** de gebruiker de tool voor het eerst gebruikt zonder eigen mappingbestand
- **THEN** zijn de meegeleverde locaties beschikbaar

#### Scenario: Gebruikersbestand overschrijft één sleutel

- **WHEN** het gebruikersbestand een regel bevat voor een sleutel die ook in de meegeleverde mapping staat
- **THEN** gebruikt het systeem de regel uit het gebruikersbestand en blijven alle andere meegeleverde regels ongewijzigd van kracht

#### Scenario: Gebruikersbestand voegt een sleutel toe

- **WHEN** het gebruikersbestand een sleutel bevat die niet in de meegeleverde mapping staat
- **THEN** is die locatie beschikbaar naast de meegeleverde locaties

#### Scenario: Gebruikersbestand ontbreekt

- **WHEN** er geen gebruikersbestand aanwezig is
- **THEN** werkt het systeem met alleen de meegeleverde mapping zonder waarschuwing of fout

#### Scenario: Gebruikersbestand is onleesbaar

- **WHEN** het gebruikersbestand bestaat maar geen geldige inhoud heeft
- **THEN** meldt het systeem welk bestand niet gelezen kon worden en gaat het verder met de meegeleverde mapping

#### Scenario: Gebruikersbestand staat in de configmap

- **WHEN** het systeem het gebruikersbestand zoekt
- **THEN** gebruikt het de platformspecifieke configmap, niet de cachemap, zodat het opruimen van caches het bestand niet verwijdert

### Requirement: Sleutelnormalisatie

Het systeem SHALL locatiesleutels normaliseren voordat het opzoekt, zodat verschillen in schrijfwijze op de website niet tot een misser leiden. Het systeem SHALL NOT benaderend matchen.

#### Scenario: Verschil in hoofdletters en witruimte

- **WHEN** de website "De Zanderij (Dindoa)  ERMELO" schrijft met dubbele witruimte en de mapping "de zanderij (dindoa) ermelo" als sleutel heeft
- **THEN** wordt de regel gevonden

#### Scenario: Verschil in interpunctie

- **WHEN** de website de interpunctie in een locatienaam wijzigt zonder de woorden te wijzigen
- **THEN** wordt de regel gevonden

#### Scenario: Geen benaderende treffer

- **WHEN** een locatiestring lijkt op een sleutel in de mapping maar er na normalisatie niet aan gelijk is
- **THEN** behandelt het systeem de locatie als ontbrekend in plaats van de bijna-treffer te gebruiken

### Requirement: Ontbrekende locatie blokkeert de uitvoer niet

Het systeem SHALL een ICS blijven genereren wanneer een locatie niet in de mapping staat, SHALL de originele locatiestring van de website gebruiken, en SHALL de ontbrekende locaties expliciet melden.

#### Scenario: Ontbrekende locatie in de uitvoer

- **WHEN** een wedstrijd op een locatie plaatsvindt die niet in de mapping staat
- **THEN** bevat het gegenereerde bestand de wedstrijd met de originele locatiestring van de website als locatie

#### Scenario: Ontbrekende locatie gemeld

- **WHEN** het genereren klaar is en één of meer locaties ontbraken
- **THEN** meldt het systeem elke ontbrekende locatie met het aantal wedstrijden dat eraan hangt

#### Scenario: Exitcode bij ontbrekende locaties

- **WHEN** het genereren slaagt maar er locaties ontbraken
- **THEN** eindigt het systeem met exitcode 0, omdat het bestand geldig en bruikbaar is

#### Scenario: Alle locaties bekend

- **WHEN** alle locaties in het wedstrijdschema in de mapping staan
- **THEN** meldt het systeem geen ontbrekende locaties

### Requirement: Locatiestatus opvragen

Het systeem SHALL een manier bieden om alle locaties uit het wedstrijdschema te tonen met hun mappingstatus, zodat een gebruiker weet wat er aangevuld moet worden.

#### Scenario: Overzicht van locaties

- **WHEN** de gebruiker de locatiestatus opvraagt
- **THEN** toont het systeem elke locatie uit het wedstrijdschema, of die in de mapping staat, en het bijbehorende adres wanneer bekend

#### Scenario: Aantal wedstrijden per locatie

- **WHEN** het systeem de locatiestatus toont
- **THEN** vermeldt het per locatie het aantal wedstrijden, zodat de gebruiker kan zien welke ontbrekende locatie de meeste impact heeft

#### Scenario: Plak-klaar fragment voor een ontbrekende locatie

- **WHEN** het systeem de locatiestatus toont en er locaties ontbreken
- **THEN** geeft het voor die locaties een fragment dat de gebruiker rechtstreeks in het gebruikersbestand kan plakken en aanvullen

#### Scenario: Paden worden getoond

- **WHEN** het systeem de locatiestatus toont
- **THEN** vermeldt het het pad van het gebruikersbestand, zodat de gebruiker weet waar hij moet bewerken

#### Scenario: Samenvatting van de dekking

- **WHEN** het systeem de locatiestatus toont
- **THEN** vermeldt het hoeveel locaties in de mapping staan en welk deel van de wedstrijden daarmee gedekt is

### Requirement: Mappingregel is na te controleren

Het systeem SHALL per mappingregel een verwijzing naar de bron van de gegevens kunnen bevatten, zodat een regel later te verifiëren is.

#### Scenario: Regel met bronverwijzing

- **WHEN** een mappingregel een verwijzing naar een OpenStreetMap-object bevat
- **THEN** blijft die verwijzing bewaard en is die op te vragen bij het tonen van de locatiestatus

#### Scenario: Regel zonder bronverwijzing

- **WHEN** een mappingregel met de hand is ingevuld zonder bronverwijzing
- **THEN** is de regel geldig en bruikbaar

