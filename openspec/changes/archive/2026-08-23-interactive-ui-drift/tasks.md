## 1. Terugkeren na een fout

- [x] 1.1 Laat `handleEnter` in `stateError` terugkeren naar de teamlijst wanneer het programma al geladen is en er teams in het model staan, en anders afsluiten
- [x] 1.2 Wis de fout bij het terugkeren, zodat een volgend foutscherm niet de vorige melding toont
- [x] 1.3 Laat `viewError` de toetsen tonen die op dat scherm werkelijk werken, opgemaakt uit de staat in plaats van hard geschreven
- [x] 1.4 Controleer dat `q` en `esc` in beide gevallen blijven afsluiten

## 2. Verificatie

- [x] 2.1 Fout tijdens het ophalen van het programma (onbereikbare pagina): het foutscherm biedt alleen afsluiten
- [x] 2.2 Fout na teamkeuze: het foutscherm biedt terug naar de teamlijst, en na terugkeren is een ander team te kiezen
- [x] 2.3 Na terugkeren en opnieuw kiezen wordt een geldig ICS-bestand geschreven
- [x] 2.4 `q` sluit af vanuit het foutscherm in beide gevallen

## 3. Documentatie

- [x] 3.1 Werk `CHANGELOG.md` bij onder `## NEXT VERSION`
- [x] 3.2 Controleer dat `README.md` geen aanpassing nodig heeft; het gedrag zonder argumenten staat daar al goed beschreven
