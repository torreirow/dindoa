# Design: interactive-ui in lijn brengen

## Context

Zie `proposal.md` — Why.

Twee van de drie punten zijn tekstueel: de spec beschrijft gedrag dat er niet is of niet meer bestaat. Alleen het terugkeren na een fout vraagt code.

Hoe de UI nu in elkaar zit:

```
stateLoadingCategories ──► stateCategorySelection ──► stateTeamSelection ──► stateProcessing
                                                                                    │
                                                              stateDone ◄───────────┴──► stateError
```

`handleEnter` doet voor zowel `stateDone` als `stateError` een `tea.Quit`. Een fout is daarmee altijd een doodlopend scherm, ook als er een volkomen bruikbare teamlijst in het model staat.

Er zijn twee soorten fouten, en dat verschil bepaalt de oplossing:

- **Tijdens het ophalen van het wedstrijdprogramma.** Er is dan niets: geen categorieën, geen teams. Afsluiten is het enige zinvolle.
- **Tijdens het genereren voor een gekozen team.** De programmagegevens zitten al in het model. Terugkeren naar de teamlijst kost geen nieuw netwerkverzoek.

## Goals / Non-Goals

**Goals:**

- Een fout na teamkeuze is geen doodlopend scherm meer.
- Het foutscherm noemt alleen toetsen die daar werkelijk werken.
- De spec beschrijft na deze change het gedrag dat de tool heeft.

**Non-Goals:**

- Geen algemene terug-navigatie door de hele UI. Van de teamlijst terug naar de categoriekeuze is een aparte wens die niets met foutafhandeling te maken heeft; die staat niet in de spec en zit hier niet in.
- Geen wijziging aan het gedrag zonder argumenten. Dat blijft de help, en de spec wordt daarop aangepast.
- Geen herstelpoging na een netwerkfout. Opnieuw proberen is `dindoa start` opnieuw starten.

## Decisions

### Onthouden waar de fout optrad in plaats van waarnaar teruggekeerd moet worden

De verleiding is een veld met "het scherm waarnaar terug te gaan". Dat gaat schuiven zodra er een scherm bijkomt. In plaats daarvan onthoudt het model of het programma al geladen is; daaruit volgt of terugkeren mogelijk is. Eén afleidbare vraag in plaats van een tweede stuk toestand dat kan wegzakken.

Praktisch: als `m.programma` gevuld is en er teams in `m.teams` staan, is de teamlijst een geldige bestemming.

### De toetsenregel wordt uit de staat opgemaakt

`viewError` schreef `[enter: afsluiten]` hard op. Nu er twee mogelijkheden zijn, moet de regel volgen uit wat er kan. Anders belooft het scherm een toets die niets doet — precies het soort afwijking dat deze change opruimt.

### Escape en q blijven altijd afsluiten

Die werken nu al in elke staat via de globale toetsafhandeling. Dat blijft zo, zodat een gebruiker nooit vastzit.

## Risks / Trade-offs

**Terugkeren na een generatiefout leidt tot dezelfde fout als de gebruiker hetzelfde team opnieuw kiest.** → Dat is aan de gebruiker; hij kan een ander team kiezen of afsluiten. Een fout onthouden en dat team blokkeren is meer machinerie dan het waard is.

**De spec aanpassen aan de code kan een echte bug maskeren.** Bij "run without flags" is dat afgewogen: er is een apart `start`-commando, de README en de helptekst documenteren het huidige gedrag al, en een TUI die spontaan opent is hinderlijk in een script of pipe. Hier is de spec achterhaald, niet de code.
