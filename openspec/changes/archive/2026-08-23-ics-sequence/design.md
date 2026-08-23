# Design: SEQUENCE in ICS-events

## Context

Zie `proposal.md` — Why.

De beperking die alles bepaalt: deze tool is staatloos. Elke run haalt het wedstrijdprogramma op en schrijft een nieuw bestand; er is geen opslag van eerdere uitvoer en dus geen kennis van welke revisie een wedstrijd al had. RFC 5545 eist dat `SEQUENCE` monotoon oploopt per `UID`. Een teller die uit de wedstrijdgegevens wordt afgeleid kan dat niet garanderen: een wedstrijd die van 13:45 naar 12:30 wordt verzet zou een lagere waarde opleveren dan de vorige uitgave.

`golang-ical` v0.3.5 kent `ComponentPropertySequence`; er is geen setter met eigen logica, dus de waarde wordt als property geschreven.

## Goals / Non-Goals

**Goals:**

- Een client die revisies bijhoudt herkent een nieuw gegenereerd bestand als nieuwer dan wat hij al heeft.
- De waarde blijft binnen het bereik dat RFC 5545 toestaat voor een geheel getal.
- Geen staat op schijf, geen configuratie, geen extra dependency.

**Non-Goals:**

- Per wedstrijd bijhouden of die daadwerkelijk gewijzigd is. Dat vraagt opslag van eerdere uitvoer en dat is een grotere verandering met een eigen afweging.
- Iets doen aan de `UID`. Die werkt en is getest; `SEQUENCE` is een aanvulling, geen vervanging.

## Decisions

### Waarde afleiden van het genereermoment, in minuten sinds de Unix-epoch

De enige grootheid die zonder staat monotoon oploopt is de tijd. Minuten sinds 1 januari 1970 geeft nu ongeveer 29,7 miljoen — ruim binnen de 2147483647 die een 32-bits geheel getal toestaat, en dat blijft zo tot ver na het jaar 6000.

Overwogen alternatieven:

- **Seconden sinds de epoch.** Werkt ook, maar het getal is 60× groter zonder dat het iets oplevert; twee runs binnen dezelfde minuut zijn in de praktijk hetzelfde bestand.
- **Dagen sinds de epoch.** Klein en leesbaar, maar twee regeneraties op dezelfde dag krijgen dezelfde waarde. Juist op de dag dat de vereniging het programma bijwerkt zou een tweede run dan door een strikte client genegeerd kunnen worden. Dat is precies het geval dat we willen dekken.
- **Vaste waarde 0.** Semantisch identiek aan het huidige gedrag, want RFC 5545 zegt dat een afwezige `SEQUENCE` als 0 geldt. Zou de indruk wekken dat er iets is opgelost.
- **Hash van de wedstrijdgegevens.** Verandert wel bij een wijziging, maar niet monotoon, en dat is precies wat de standaard verbiedt.

### Eén waarde voor alle events in een bestand

`SEQUENCE` hoort bij de uitgave, net als `DTSTAMP`. De waarde wordt één keer per `Generate` bepaald en aan alle events meegegeven. Per event de klok opnieuw lezen zou binnen één bestand verschillende waarden kunnen geven zonder dat daar betekenis aan hangt.

### Klok injecteerbaar maken voor de test

Een test die monotonie aantoont moet twee uitgaven met een verschillend moment kunnen maken zonder een minuut te wachten. De generator krijgt daarom een instelbare tijdsbron; standaard `time.Now`.

## Risks / Trade-offs

**`SEQUENCE` loopt op zonder dat er iets veranderd is.** → Onschadelijk: de client ziet dezelfde `UID` met een hogere revisie en identieke inhoud, en werkt het item bij met wat er al stond. De alternatieven die dit voorkomen vragen staat op schijf.

**Twee gegenereerde bestanden zijn niet langer byte-identiek.** Dat gold al voor `DTSTAMP`, nu ook voor `SEQUENCE`. → Een vergelijking van uitvoer moet beide velden negeren. Dit staat in het proposal en wordt in de tests zo gedaan.

**Een gebruiker die het bestand twee keer importeert zonder tussentijdse wijziging.** → Zelfde uitkomst als nu: één item per wedstrijd, bijgewerkt. Het gedrag dat we eerder verifieerden verandert hier niet.
