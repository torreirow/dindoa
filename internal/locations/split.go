package locations

import (
	"strings"
	"unicode"
)

// SplitVenue splits a venue string from the match programme into the venue
// name and the town. The website writes the town in capitals at the end:
//
//	"Sportcentrum De Koploper LELYSTAD"        -> "Sportcentrum De Koploper", "LELYSTAD"
//	"Topsportcentrum De Koog KOOG AAN DE ZAAN" -> "Topsportcentrum De Koog", "KOOG AAN DE ZAAN"
//
// ok is false when no town in capitals is recognised.
func SplitVenue(venue string) (name, town string, ok bool) {
	words := strings.Fields(venue)
	i := len(words)
	for i > 0 && isUpperWord(words[i-1]) {
		i--
	}
	if i == len(words) {
		return "", "", false
	}
	if i == 0 {
		// Everything is in capitals; only the last word can be the town.
		if len(words) < 2 {
			return "", "", false
		}
		i = len(words) - 1
	}
	return strings.Join(words[:i], " "), strings.Join(words[i:], " "), true
}

// isUpperWord reports whether w has at least one letter and no lower-case
// letters, e.g. "ERMELO" or "DRIEBERGEN-RIJSENBURG".
func isUpperWord(w string) bool {
	hasLetter := false
	for _, r := range w {
		if unicode.IsLower(r) {
			return false
		}
		if unicode.IsLetter(r) {
			hasLetter = true
		}
	}
	return hasLetter
}

// TownTitle turns a town in capitals into its usual spelling as far as that
// can be guessed: "DRIEBERGEN-RIJSENBURG" -> "Driebergen-Rijsenburg".
func TownTitle(town string) string {
	b := []rune(strings.ToLower(town))
	start := true
	for i, r := range b {
		if start && unicode.IsLetter(r) {
			b[i] = unicode.ToUpper(r)
		}
		start = r == ' ' || r == '-'
	}
	return string(b)
}
