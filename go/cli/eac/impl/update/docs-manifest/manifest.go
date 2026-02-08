package docsmanifest

import (
	"encoding/json"
	"os"
)

// LoadCache loads an existing cache manifest from the given path.
// Returns nil, nil if the file doesn't exist.
func LoadCache(path string) (*CacheManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var m CacheManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

// SaveCache writes the cache manifest to the given path.
func SaveCache(m *CacheManifest, path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
