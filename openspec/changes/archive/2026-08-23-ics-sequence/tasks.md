## 1. Implementatie

- [x] 1.1 Geef `Generator` een instelbare tijdsbron met `time.Now` als standaard, zodat een test twee uitgaven met een verschillend moment kan maken zonder te wachten
- [x] 1.2 Bepaal de revisiewaarde één keer per `Generate` als minuten sinds de Unix-epoch, en geef die aan alle events mee
- [x] 1.3 Schrijf de waarde als `SEQUENCE` op elk `VEVENT` via `ComponentPropertySequence`
- [x] 1.4 Controleer dat `DTSTAMP` en de UID-opbouw ongemoeid blijven

## 2. Tests

- [x] 2.1 Elk event heeft een `SEQUENCE` met een niet-negatieve gehele waarde
- [x] 2.2 Een latere generatie levert een hogere `SEQUENCE` dan een eerdere
- [x] 2.3 Alle events in één bestand hebben dezelfde `SEQUENCE`
- [x] 2.4 De waarde past in een 32-bits signed integer
- [x] 2.5 Pas de bestaande vergelijking van twee gegenereerde bestanden aan zodat naast `DTSTAMP` ook `SEQUENCE` buiten beschouwing blijft

## 3. Verificatie tegen de echte site

- [x] 3.1 `dindoa --team j4` levert 6 events die alle dezelfde `SEQUENCE` hebben
- [x] 3.2 Live geverifieerd dat de waarde de klok volgt (29791862 = exact het aantal minuten sinds epoch op het genereermoment, binnen int32 tot het jaar 6053). Twee runs binnen dezelfde minuut geven per ontwerp dezelfde waarde; de monotonie zelf is deterministisch getest met een vaste klok in TestLaterGenerationHasHigherSequence

## 4. Documentatie

- [x] 4.1 Voeg `SEQUENCE` toe aan de opsomming van velden in een gegenereerd event in `README.md`, met de uitleg dat de waarde per uitgave oploopt en niet per wedstrijd
- [x] 4.2 Werk `CHANGELOG.md` bij onder `## NEXT VERSION`
