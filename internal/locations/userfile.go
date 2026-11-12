package locations

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ReadUserFile reads the user's mapping file for updating. A missing file is
// not an error: it yields an empty File, ready to be filled. A file that exists
// but cannot be read or parsed is an error, so an update never replaces a file
// it did not understand.
func ReadUserFile(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &File{Version: 1, Locations: map[string]Entry{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	f, err := parse(data)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if f.Version == 0 {
		f.Version = 1
	}
	return f, nil
}

// WriteUserFile writes f to path. It writes to a temporary file in the same
// directory and renames it into place, so a failed write leaves the previous
// file intact. Keys come out sorted, because encoding/json sorts map keys.
func WriteUserFile(path string, f *File) error {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("encode location mapping: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".locations-*.json")
	if err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	tmpName := tmp.Name()
	// Remove the temporary file on any failure; after a successful rename it
	// no longer exists and this is a no-op.
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
