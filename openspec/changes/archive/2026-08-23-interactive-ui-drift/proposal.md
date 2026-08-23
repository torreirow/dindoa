# interactive-ui in lijn brengen met de werkelijkheid

## Why

Bij `cli-interface` bleek een requirement gedrag te beschrijven dat de code nooit had: `dindoa --category rood` zou de interactieve modus starten met die categorie voorgeselecteerd, maar de tool toonde de help. Dat is inmiddels opgelost.

Dezelfde controle op `interactive-ui` levert drie afwijkingen op:

1. **"Show geocoding progress"** beschrijft dat de UI voortgang per gegeocodeerde locatie toont. Geocoding is verwijderd; het verwerkingsscherm zegt nu alleen dat de wedstrijden verwerkt worden. Dit is een gemiste delta: de vorige change noemde "geen wijziging aan de interactieve UI" als non-goal, terwijl dat scherm wél veranderde.

2. **"Run without flags"** zegt dat `dindoa` zonder argumenten de interactieve UI start. De tool toont de help. Code, README en de helptekst zelf zeggen alle drie hetzelfde, en er is een apart `start`-commando voor de UI. De spec is hier achterhaald, niet de code.

3. **"Return to previous screen on error"** zegt dat de gebruiker bij een fout terug kan naar de categoriekeuze. De code doet `tea.Quit` en toont alleen `[enter: afsluiten]`. Dit is een doodlopend scherm en het enige punt van de drie waar de code moet veranderen.

Zulke afwijkingen zijn niet onschuldig: een spec die gedrag belooft dat er niet is, maakt de rest van de spec minder betrouwbaar als naslag.

## What Changes

- **Spec volgt de code** voor het gedrag zonder argumenten: `dindoa` toont de help, `dindoa start` opent de interactieve modus. Dat is de bestaande, gedocumenteerde werking; er verandert niets voor de gebruiker.
- **Het scenario over geocoding-voortgang vervalt.** Het verwerkingsscherm meldt dat de wedstrijden van het gekozen team verwerkt worden; er is geen stap per locatie meer om voortgang van te tonen.
- **Terug naar de categoriekeuze bij een fout wordt geïmplementeerd.** Bij een fout tijdens het genereren kan de gebruiker terug naar de teamlijst of de categoriekeuze in plaats van alleen afsluiten. Bij een fout tijdens het ophalen van het wedstrijdprogramma is er nog geen lijst om naar terug te keren, dus daar blijft afsluiten het enige zinvolle.
- De UI meldt al welke locaties geen adres hebben; dat blijft ongewijzigd en krijgt een scenario zodat het vastligt.

## Capabilities

### Modified Capabilities

- `interactive-ui`: Het scenario over geocoding-voortgang vervalt, het gedrag zonder argumenten wordt vastgelegd zoals het is, en terugkeren naar een eerder scherm na een fout wordt een werkende eis in plaats van een belofte.

## Impact

- `internal/ui/interactive.go` — toetsafhandeling in de foutstaat, en onthouden naar welk scherm teruggekeerd kan worden.
- `internal/ui/models.go` — `viewError` moet de beschikbare toetsen tonen die er werkelijk zijn.
- `CHANGELOG.md` onder `## NEXT VERSION`.
- Geen wijziging aan `README.md`: het gedrag zonder argumenten staat daar al goed beschreven.
- Geen nieuwe dependencies.
