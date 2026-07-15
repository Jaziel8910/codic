package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// registryPath returns the path for a named entry inside a sub-directory of
// ~/.codic (e.g. instruments/foo.json).
func registryPath(sub, name string) string {
	return filepath.Join(CodicDir(), sub, sanitize(name)+".json")
}

// registryList returns the names of all entries in a sub-directory.
func registryList(sub string) ([]string, error) {
	dir := filepath.Join(CodicDir(), sub)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		names = append(names, e.Name()[:len(e.Name())-len(".json")])
	}
	return names, nil
}

// registryLoad reads a named entry into v (JSON). Returns os.IsNotExist if
// missing.
func registryLoad(sub, name string, v interface{}) error {
	data, err := os.ReadFile(registryPath(sub, name))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// registrySave writes a named entry as JSON, creating the directory.
func registrySave(sub, name string, v interface{}) error {
	path := registryPath(sub, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// registryRemove deletes a named entry.
func registryRemove(sub, name string) error {
	err := os.Remove(registryPath(sub, name))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// loadJSONFile reads a JSON file into v; missing files are not an error.
func loadJSONFile(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// saveJSONFile writes v as JSON to path.
func saveJSONFile(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// errNotFound is a friendly helper.
func errNotFound(kind, name string) error {
	return fmt.Errorf("%s %q not found", kind, name)
}
