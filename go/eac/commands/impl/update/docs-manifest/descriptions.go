package docsmanifest

import (
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

// LoadDescriptions loads the descriptions file from the given path.
// Returns an empty map if the file doesn't exist.
func LoadDescriptions(path string) (map[string]Description, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]Description), nil
		}
		return nil, err
	}

	var file DescriptionsFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, err
	}

	if file.Assets == nil {
		return make(map[string]Description), nil
	}

	return file.Assets, nil
}

// SaveDescriptions writes the descriptions to the given path in YAML format.
// Assets are sorted alphabetically for consistent output.
func SaveDescriptions(path string, descriptions map[string]Description) error {
	// Sort keys for deterministic output
	keys := make([]string, 0, len(descriptions))
	for k := range descriptions {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build ordered map for YAML output
	file := DescriptionsFile{
		Assets: make(map[string]Description),
	}
	for _, k := range keys {
		file.Assets[k] = descriptions[k]
	}

	// Marshal with custom encoder for nice formatting
	data, err := yaml.Marshal(&file)
	if err != nil {
		return err
	}

	// Add header comment
	header := []byte("# Human-maintained asset descriptions and status\n# This file is git-tracked. Edit descriptions and active status here.\n# Auto-generated data (usedIn, stats) is in .manifest-cache.json (git-ignored)\n")
	data = append(header, data...)

	return os.WriteFile(path, data, 0o644)
}
