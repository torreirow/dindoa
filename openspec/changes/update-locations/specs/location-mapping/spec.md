## MODIFIED Requirements

### Requirement: Locatie opzoeken via mappingfile

Het systeem SHALL locatiegegevens uitsluitend uit een mappingfile halen en SHALL tijdens het genereren van een ICS geen netwerkverzoek doen om een locatie op te lossen. Alleen het expliciete bijwerkcommando voor de gebruikersmapping gebruikt het netwerk voor locatiegegevens.

#### Scenario: Bekende locatie opgezocht

- **WHEN** het wedstrijdschema de locatie "De Zanderij (Dindoa) ERMELO" bevat en die staat in de mapping
- **THEN** levert het systeem de bijbehorende leesbare naam, adres en coördinaten

#### Scenario: Geen netwerk nodig

- **WHEN** de gebruiker een ICS genereert zonder internetverbinding nadat het wedstrijdschema beschikbaar is
- **THEN** worden alle locaties uit de mapping opgezocht zonder netwerkfout en zonder vertraging door rate limiting

#### Scenario: Coördinaten optioneel

- **WHEN** een mappingregel een adres bevat maar geen coördinaten
- **THEN** levert het systeem het adres en meldt het geen fout

#### Scenario: Locatiestatus gebruikt geen locatiebron

- **WHEN** de gebruiker de locatiestatus opvraagt
- **THEN** raadpleegt het systeem geen externe locatiebron, alleen de mappingfiles

### Requirement: Mappingregel is na te controleren

Het systeem SHALL per mappingregel een verwijzing naar de bron van de gegevens kunnen bevatten, zodat een regel later te verifiëren is.

#### Scenario: Regel met bronverwijzing

- **WHEN** een mappingregel een verwijzing naar een OpenStreetMap-object bevat
- **THEN** blijft die verwijzing bewaard en is die op te vragen bij het tonen van de locatiestatus

#### Scenario: Regel zonder bronverwijzing

- **WHEN** een mappingregel met de hand is ingevuld zonder bronverwijzing
- **THEN** is de regel geldig en bruikbaar

#### Scenario: Automatisch gevonden regel is herkenbaar

- **WHEN** een mappingregel door het bijwerkcommando uit OpenStreetMap is gehaald
- **THEN** draagt de regel de bron "osm-auto" en toont de locatiestatus die bron, zodat de gebruiker ziet welke regels nog niet handmatig gecontroleerd zijn
