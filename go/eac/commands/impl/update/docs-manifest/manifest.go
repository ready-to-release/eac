package docsmanifest

import (
	"encoding/json"
	"os"
)

// LoadLegacyManifest loads an existing legacy manifest from the given path.
// Returns nil, nil if the file doesn't exist.
// This is used for migration from the old single-file format.
func LoadLegacyManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

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

// ExtractDescriptionsFromLegacy extracts descriptions from a legacy manifest
// for migration to the new format.
func ExtractDescriptionsFromLegacy(legacy *Manifest) map[string]Description {
	descriptions := make(map[string]Description)
	if legacy == nil || legacy.Assets == nil {
		return descriptions
	}

	for path, asset := range legacy.Assets {
		descriptions[path] = Description{
			Active:      true, // All existing assets are active
			Description: asset.Description,
		}
	}

	return descriptions
}
