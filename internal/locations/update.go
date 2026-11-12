package locations

import (
	"context"
	"fmt"
)

// Outcome is what happened to one venue during an update.
type Outcome struct {
	Venue string
	Key   string
	// Source names where an added entry came from; empty for a missing one.
	Source string
	// Reason explains why a venue is still missing; empty for an added one.
	Reason string
	Entry  Entry
}

// Result is the outcome of an update.
type Result struct {
	Added   []Outcome
	Missing []Outcome
	// Warnings holds source failures; the update carried on past them.
	Warnings []string
}

// Update looks up the missing venues in the sources, in order, and adds what
// it finds to user. A venue a source finds is not passed to later sources.
// Keys already present in user are never touched, not even empty ones: a
// placeholder the user pasted is the user's to fill or remove.
func Update(ctx context.Context, missing []string, user *File, sources ...Source) Result {
	var res Result
	if user.Locations == nil {
		user.Locations = map[string]Entry{}
	}

	reason := map[string]string{}
	var pending []string
	seen := map[string]bool{}
	for _, v := range missing {
		k := NormalizeKey(v)
		if seen[k] {
			continue
		}
		seen[k] = true
		if _, exists := user.Locations[k]; exists {
			res.Missing = append(res.Missing, Outcome{Venue: v, Key: k,
				Reason: "your file has an empty entry for it; fill it in or remove it"})
			continue
		}
		pending = append(pending, v)
	}

	for _, src := range sources {
		if len(pending) == 0 {
			break
		}
		found, reasons, err := src.Find(ctx, pending)
		if err != nil {
			res.Warnings = append(res.Warnings, fmt.Sprintf("%s: %v", src.Name(), err))
		}
		var still []string
		for _, v := range pending {
			k := NormalizeKey(v)
			if e, ok := found[k]; ok && !e.IsEmpty() {
				user.Locations[k] = e
				res.Added = append(res.Added, Outcome{Venue: v, Key: k, Source: src.Name(), Entry: e})
				continue
			}
			switch {
			case reasons[k] != "":
				reason[k] = reasons[k]
			case err != nil:
				reason[k] = "not looked up: " + err.Error()
			}
			still = append(still, v)
		}
		pending = still
	}

	for _, v := range pending {
		k := NormalizeKey(v)
		r := reason[k]
		if r == "" {
			r = "not found"
		}
		res.Missing = append(res.Missing, Outcome{Venue: v, Key: k, Reason: r})
	}
	return res
}
