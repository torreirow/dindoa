package locations

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestReadUserFileMissing(t *testing.T) {
	f, err := ReadUserFile(filepath.Join(t.TempDir(), "nope", "locations.json"))
	if err != nil {
		t.Fatalf("missing file: %v", err)
	}
	if f.Version != 1 || f.Locations == nil || len(f.Locations) != 0 {
		t.Fatalf("missing file = %+v, want empty version 1", f)
	}
}

func TestReadUserFileInvalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "locations.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ReadUserFile(path)
	if err == nil {
		t.Fatal("invalid file: want error")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error %q does not name the file", err)
	}
}

func TestWriteUserFileRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "locations.json")
	in := &File{Version: 1, Locations: map[string]Entry{
		"b hal plaats": {Name: "B", Address: "Weg 1, 1234 AB Plaats", Lat: 52.1, Lon: 5.2, OSM: "node/1", Source: "manual"},
		"a hal plaats": {Name: "A", Address: "Laan 2"},
	}}
	if err := WriteUserFile(path, in); err != nil {
		t.Fatalf("write: %v", err)
	}
	out, err := ReadUserFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round trip changed content:\n in  %+v\n out %+v", in, out)
	}

	data, _ := os.ReadFile(path)
	if strings.Index(string(data), `"a hal plaats"`) > strings.Index(string(data), `"b hal plaats"`) {
		t.Error("keys are not sorted")
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Errorf("directory holds %d files, want only the mapping file", len(entries))
	}
}

func TestWriteUserFileFailureKeepsOriginal(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "locations.json")
	original := []byte(`{"version":1,"locations":{"x":{"name":"X"}}}`)
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(dir, 0o755)

	err := WriteUserFile(path, &File{Version: 1, Locations: map[string]Entry{"y": {Name: "Y"}}})
	if err == nil {
		t.Fatal("write into read-only directory: want error")
	}
	got, _ := os.ReadFile(path)
	if string(got) != string(original) {
		t.Fatalf("original changed after failed write: %s", got)
	}
}
