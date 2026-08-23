# SEQUENCE in ICS-events

## Why

Het wedstrijdprogramma op dindoa.nl wordt in blokken gepubliceerd — bij het schrijven hiervan staat er 5 september tot en met 14 oktober. Opnieuw genereren en importeren is dus geen uitzondering maar de normale gang van zaken, en aanvangstijden verschuiven gedurende het seizoen.

De UID is daarop ingericht: die bestaat uit team, datum en tegenstander en bevat niet de aanvangstijd, zodat een verzette wedstrijd hetzelfde item bijwerkt. Getest en werkend in agenda-apps.

Wat ontbreekt is `SEQUENCE`. Dat veld nummert revisies van hetzelfde item, en een client die het respecteert gebruikt het om te bepalen of een binnenkomende versie nieuwer is dan wat hij al heeft. Zonder `SEQUENCE` gedraagt onze uitvoer zich volgens RFC 5545 als revisie 0 — bij elke regeneratie opnieuw. Een strikte client kan een echte wijziging daardoor als "niet nieuwer" afdoen en negeren.

## What Changes

- Elk `VEVENT` krijgt een `SEQUENCE` die bij elke regeneratie hoger is dan de vorige, zodat een client de nieuwe versie altijd als nieuwer herkent.
- De waarde wordt afgeleid van het genereermoment, niet van de wedstrijdgegevens. Deze tool houdt geen staat bij en weet dus niet wat de vorige revisie van een wedstrijd was; een teller die op de inhoud is gebaseerd zou niet monotoon kunnen oplopen zoals RFC 5545 vereist.
- Alle events in één bestand krijgen dezelfde waarde. `SEQUENCE` hoort bij "deze uitgave van de kalender", net als `DTSTAMP`.

Wat dit **niet** doet: onderscheiden of een individuele wedstrijd daadwerkelijk gewijzigd is. Bij een regeneratie waarin niets veranderde loopt `SEQUENCE` ook op. Dat is onschadelijk — de client ziet dezelfde UID met een hogere revisie en identieke inhoud — maar het is wel de reden dat we bewust geen poging doen om per wedstrijd te tellen.

Gevolg voor het vergelijken van uitvoer: twee bestanden van dezelfde ploeg verschillen nu in `DTSTAMP` **en** `SEQUENCE`. Een test die controleert dat regenereren hetzelfde resultaat geeft moet beide velden buiten beschouwing laten.

## Capabilities

### Modified Capabilities

- `ics-generation`: Elk event krijgt een `SEQUENCE` die per gegenereerd bestand oploopt, zodat clients die revisies bijhouden een nieuwe uitgave als nieuwer herkennen.

## Impact

- `internal/ics/generator.go` — `createEvent` en `Generate`; de waarde wordt één keer per bestand bepaald en niet per event.
- `internal/ics/generator_test.go` — de bestaande test op determinisme van de UID blijft gelden; er komt dekking bij voor de monotonie van `SEQUENCE` en voor het feit dat alle events in één bestand dezelfde waarde krijgen.
- `README.md` — de opsomming van velden in een gegenereerd event.
- `CHANGELOG.md` onder `## NEXT VERSION`.
- Geen nieuwe dependencies. `golang-ical` v0.3.5 kent `ComponentPropertySequence`.
