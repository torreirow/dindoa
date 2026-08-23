# Design: aanvangstijd valideren

## Context

Zie `proposal.md` — Why, met de gemeten uitkomsten per notatie.

Twee plekken werken nu samen om de fout stil te houden:

```go
// parser.go — accepteert elke niet-lege waarde
if r.time == "" || r.home == "" || r.away == "" { return }

// generator.go — leest wat er staat, negeert of het gelukt is
fmt.Sscanf(match.Time, "%d:%d", &hour, &minute)
return time.Date(..., hour, minute, 0, 0, g.timezone)
```

`Sscanf` geeft wel een aantal en een fout terug, maar die worden niet gelezen. En `time.Date` documenteert expliciet dat het waarden buiten bereik normaliseert: uur 1345 wordt 56 dagen erbij. Beide gedragingen zijn afzonderlijk redelijk; samen leveren ze een geldig ICS-bestand met de verkeerde datum.

De bestaande controle op de kolomkoppen laat zien hoe dit in deze codebase hoort: eerst vaststellen dat de bron eruitziet zoals verwacht, en anders een fout die de aangetroffen vorm noemt.

## Goals / Non-Goals

**Goals:**

- Een tijd die niet als `HH:MM` te lezen is levert een fout op, niet een gok.
- De foutmelding noemt de wedstrijd en de aangetroffen waarde, zodat zichtbaar is dat de bron veranderd is.
- Na het parsen zijn uur en minuut geldig, zodat het samenstellen van de datum niets meer hoeft aan te nemen.

**Non-Goals:**

- Geen tolerantie voor andere notaties. `19.00` naar `19:00` raden lijkt vriendelijk, maar het is gokken over de bedoeling van een gewijzigde bron. Dat is precies het gedrag dat hier weggaat.
- Geen poging tot herstel of doorgaan met de overige rijen. Als de tijdkolom van vorm verandert, geldt dat voor de hele pagina; één rij overslaan zou een halve agenda opleveren, wat erger is dan geen agenda.
- Geen validatie van de andere kolommen in deze change. Teamnamen en locaties zijn vrije tekst; daar is geen vorm om tegen te toetsen.

## Decisions

### Valideren in de parser, niet in de generator

De generator zou de fout ook kunnen vangen, maar dan is de rij al door de parser gekomen en is de wedstrijd al onderdeel van een lijst. De parser is de plek waar de bron wordt geïnterpreteerd en waar de kolomcontrole al staat; daar hoort ook de vaststelling dat de tijdkolom een tijd bevat.

Gevolg: `parseMatchDateTime` hoeft niets te veranderen aan zijn werkwijze, maar mag er wel op vertrouwen dat de waarde geldig is. Om te voorkomen dat een toekomstige aanroeper die aanname stilzwijgend breekt, wordt het ontleden van de tijd één functie die zowel de parser als de generator gebruikt.

### Falen op de hele pagina, niet per rij

Overwogen alternatief: de rij overslaan en doorgaan. Dat is hier verkeerd. Een gewijzigde tijdnotatie treft niet één rij maar de hele pagina, en een agenda waarin een deel van de wedstrijden ontbreekt is misleidender dan een foutmelding. Dit sluit ook aan bij hoe de kolomcontrole werkt.

### Uur 0–23 en minuut 0–59 expliciet begrenzen

`time.Date` normaliseert buiten bereik, dus zonder deze grens is `25:99` een geldige invoer die naar de volgende dag schuift. De grens is wat een klok toestaat, niet wat de korfbalcompetitie plausibel maakt; een wedstrijd om 07:00 is ongebruikelijk maar niet fout, en de parser hoort daar geen mening over te hebben.

## Risks / Trade-offs

**De tool valt volledig stil als dindoa.nl de tijdnotatie wijzigt.** → Dat is de bedoeling. Het alternatief is een agenda met wedstrijden op de verkeerde dag, en dat merkt niemand tot iemand voor een dichte deur staat. De foutmelding noemt de aangetroffen waarde, dus de oorzaak is direct duidelijk en de reparatie is klein.

**Een enkele rommelige rij blokkeert het hele team.** → Aanvaardbaar, en in de huidige data komt het niet voor: alle 210 rijen zijn `HH:MM`. Blijkt het in de praktijk toch voor te komen bij één afwijkende rij, dan is per-rij overslaan mét melding een aparte afweging.
