# Aanvangstijd valideren in plaats van stil normaliseren

## Why

De parser accepteert elke niet-lege waarde in de tijdkolom, en `parseMatchDateTime` doet `fmt.Sscanf(match.Time, "%d:%d", &hour, &minute)` zonder de uitkomst te controleren. `Sscanf` geeft geen fout terug die iemand leest, en `time.Date` normaliseert waarden buiten bereik zonder te klagen. Gemeten:

```
"13:45"  ->  20260905T114500Z    correct
"13.45"  ->  20260905T110000Z    45 minuten mis
"1345"   ->  20261031T000000Z    56 dagen later
""       ->  20260904T220000Z    middernacht
"25:99"  ->  20260906T003900Z    volgende dag
```

`1345` wordt gelezen als uur 1345, en dat normaliseert naar eind oktober. Er komt geen waarschuwing en geen foutmelding: de gebruiker krijgt een geldig ICS-bestand met wedstrijden op de verkeerde dag.

Dit is niet theoretisch. De oefenwedstrijden-pagina op dindoa.nl gebruikt nu al `19.00` met een punt. Zodra die notatie op de programma-pagina verschijnt, of de kolomvolgorde verandert zodat er iets anders in de eerste cel staat, gaat het precies zo mis.

Alle 210 rijen die nu gepubliceerd zijn gebruiken `HH:MM`, dus er is vandaag niets fout. Het is een latent gat, en het is dezelfde klasse fout waardoor deze tool eerder stil nul wedstrijden opleverde: een bron die verandert, code die dat niet merkt.

## What Changes

- De tijdkolom wordt gevalideerd op `H:MM` of `HH:MM` met uur 0–23 en minuut 0–59. Een afwijking levert een fout op die de rij en de aangetroffen waarde noemt, in dezelfde geest als de bestaande controle op de kolomkoppen.
- Het samenstellen van datum en tijd rekent niet langer met waarden die het niet heeft kunnen valideren. Er is geen pad meer waarlangs een onleesbare tijd tot een gedateerd event leidt.
- De foutmelding legt uit wat er werd verwacht, zodat duidelijk is dat de bron gewijzigd is en niet de gebruiker iets fout deed.

Wat dit **niet** doet: tolerant worden voor andere notaties. Een punt in plaats van een dubbele punt raden lijkt vriendelijk, maar `19.00` en `19:00` zijn niet noodzakelijk hetzelfde bedoeld, en stil gokken is precies het gedrag dat hier wordt weggehaald. Verandert de site van notatie, dan is een duidelijke fout het juiste antwoord.

## Capabilities

### Modified Capabilities

- `team-scraping`: De aanvangstijd wordt gevalideerd bij het parsen; een tijd die niet als `HH:MM` te lezen is levert een fout op in plaats van een stil verschoven datum of tijd.

## Impact

- `internal/scraper/parser.go` — validatie bij het lezen van de rij, naast de bestaande controle op de kolomkoppen.
- `internal/ics/generator.go` — `parseMatchDateTime` hoeft niet langer te gokken; de uren en minuten zijn dan al geldig.
- `internal/scraper/parser_test.go` en `internal/ics/generator_test.go` — dekking voor de afwijkende notaties die nu stil door glippen.
- `CHANGELOG.md` onder `## NEXT VERSION`.
- Geen wijziging aan de README: dit gaat over gedrag bij een gewijzigde bron, niet over het gebruik.
- Geen nieuwe dependencies.
