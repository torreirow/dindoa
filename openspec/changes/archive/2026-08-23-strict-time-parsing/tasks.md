## 1. Implementatie

- [x] 1.1 Maak in `internal/scraper` één functie die een tijdstring ontleedt naar uur en minuut, met een fout bij een andere vorm dan `H:MM` of `HH:MM` en bij uur boven 23 of minuut boven 59
- [x] 1.2 Laat `ParseProgramma` die functie gebruiken bij het lezen van elke rij, en een fout teruggeven die de datum, de teams en de aangetroffen waarde noemt
- [x] 1.3 Laat `parseMatchDateTime` in `internal/ics` dezelfde functie gebruiken in plaats van `fmt.Sscanf`, zodat er geen tweede interpretatie van hetzelfde veld bestaat
- [x] 1.4 Controleer dat de bestaande controle op de kolomkoppen vóór de tijdvalidatie komt, zodat een gewijzigde kolomvolgorde als layoutfout wordt gemeld en niet als tijdfout

## 2. Tests

- [x] 2.1 `13.45`, `1345`, een lege cel en `abc` leveren een fout op met de aangetroffen waarde erin
- [x] 2.2 `25:99` levert een fout op
- [x] 2.3 `9:30` wordt gelezen als half tien
- [x] 2.4 De bestaande fixture met 210 rijen blijft zonder fout parsen
- [x] 2.5 Een fout in één rij laat de hele pagina falen en levert geen halve wedstrijdlijst op
- [x] 2.6 Regressietest op de gemeten gevallen: er is geen invoer meer die een event op een andere datum zet dan de kop boven de tabel

## 3. Verificatie tegen de echte site

- [x] 3.1 `dindoa --team j4` levert nog 6 wedstrijden met de juiste tijden
- [x] 3.2 `dindoa --list-locations` werkt nog, dus alle 210 rijen komen door de validatie

## 4. Documentatie

- [x] 4.1 Werk `CHANGELOG.md` bij onder `## NEXT VERSION`
- [x] 4.2 Voeg de nieuwe foutmelding toe aan de troubleshooting-sectie van `README.md`, naast die over de gewijzigde tabelopmaak
