package locations

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/torreirow/dindoa/internal/useragent"
)

// Source looks up mapping entries for venues that are not mapped yet. Sources
// are the only part of this package that uses the network, and only the
// explicit update command calls them; Lookup never does.
type Source interface {
	// Name is shown in the update report, e.g. "repository".
	Name() string
	// Find looks up the given venue strings. found and reasons are keyed by
	// NormalizeKey(venue); reasons explains each venue that was not found.
	// A non-nil error means the source failed; found may still hold the
	// entries it established before failing.
	Find(ctx context.Context, venues []string) (found map[string]Entry, reasons map[string]string, err error)
}

// DefaultRepoURL is the reviewed mapping on the main branch of the repository.
const DefaultRepoURL = "https://raw.githubusercontent.com/torreirow/dindoa/main/internal/locations/locations.json"

// RepoSource reads the reviewed mapping from the repository, so venues added
// there reach users without a new release.
type RepoSource struct {
	URL    string
	Client *http.Client
}

// NewRepoSource returns a RepoSource for the main branch on GitHub.
func NewRepoSource() *RepoSource {
	return &RepoSource{URL: DefaultRepoURL, Client: &http.Client{Timeout: 15 * time.Second}}
}

// Name implements Source.
func (r *RepoSource) Name() string { return "repository" }

// Find implements Source.
func (r *RepoSource) Find(ctx context.Context, venues []string) (map[string]Entry, map[string]string, error) {
	data, err := httpGet(ctx, r.Client, r.URL)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch reviewed mapping: %w", err)
	}
	remote, err := parse(data)
	if err != nil {
		return nil, nil, fmt.Errorf("read reviewed mapping from %s: %w", r.URL, err)
	}
	byKey := make(map[string]Entry, len(remote.Locations))
	for k, v := range remote.Locations {
		byKey[NormalizeKey(k)] = v
	}

	found := map[string]Entry{}
	reasons := map[string]string{}
	for _, v := range venues {
		k := NormalizeKey(v)
		if e, ok := byKey[k]; ok && !e.IsEmpty() {
			found[k] = e
			continue
		}
		reasons[k] = "not in the reviewed list"
	}
	return found, reasons, nil
}

// statusError is a non-200 HTTP response.
type statusError struct {
	URL    string
	Status int
}

func (e *statusError) Error() string { return fmt.Sprintf("%s: HTTP %d", e.URL, e.Status) }

func httpGet(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", useragent.Value)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, &statusError{URL: url, Status: resp.StatusCode}
	}
	// A mapping file or a Nominatim answer is small; cap it anyway.
	return io.ReadAll(io.LimitReader(resp.Body, 4<<20))
}
