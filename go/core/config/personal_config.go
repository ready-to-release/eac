package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// PersonalConfigSuffix is the suffix used for personal override files.
const PersonalConfigSuffix = ".personal"

// SupportedPersonalConfigs lists config files that support personal overrides.
var SupportedPersonalConfigs = []string{
	"repository.yml",
	"tools.yml",
	"environments.yml",
	"timeouts.yml",
}

// LoadWithPersonal loads a config file and merges any personal overrides.
// Personal overrides are stored in files with .personal suffix before extension.
// Example: repository.yml -> repository.personal.yml
//
// The merge strategy:
// - Maps: personal values override base values (deep merge)
// - Arrays: personal values replace base values
// - Scalars: personal values replace base values
func LoadWithPersonal[T any](configRoot, filename string) (*T, error) {
	basePath := filepath.Join(configRoot, filename)
	personalPath := ToPersonalPath(basePath)

	// Load base config (required)
	baseData, err := os.ReadFile(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", basePath)
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Load personal overrides (optional)
	personalData, _ := os.ReadFile(personalPath)

	// Merge if personal config exists
	var merged []byte
	if len(personalData) > 0 {
		merged, err = MergeYAML(baseData, personalData)
		if err != nil {
			return nil, fmt.Errorf("merging personal config: %w", err)
		}
	} else {
		merged = baseData
	}

	var result T
	if err := yaml.Unmarshal(merged, &result); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &result, nil
}

// ToPersonalPath converts a config path to its personal override path.
// Example: /path/to/repository.yml -> /path/to/repository.personal.yml
func ToPersonalPath(path string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	return base + PersonalConfigSuffix + ext
}

// HasPersonalConfig checks if a personal override file exists.
func HasPersonalConfig(configRoot, filename string) bool {
	basePath := filepath.Join(configRoot, filename)
	personalPath := ToPersonalPath(basePath)
	_, err := os.Stat(personalPath)
	return err == nil
}

// ListPersonalConfigs returns the list of personal config files that exist.
func ListPersonalConfigs(configRoot string) []string {
	var found []string
	for _, filename := range SupportedPersonalConfigs {
		if HasPersonalConfig(configRoot, filename) {
			found = append(found, ToPersonalPath(filepath.Join(configRoot, filename)))
		}
	}
	return found
}

// MergeYAML performs a deep merge of two YAML documents.
// Values from override take precedence over base.
// Maps are merged recursively; arrays and scalars are replaced.
func MergeYAML(base, override []byte) ([]byte, error) {
	if len(base) == 0 {
		return override, nil
	}
	if len(override) == 0 {
		return base, nil
	}

	var baseMap map[string]interface{}
	if err := yaml.Unmarshal(base, &baseMap); err != nil {
		return nil, fmt.Errorf("parsing base YAML: %w", err)
	}

	var overrideMap map[string]interface{}
	if err := yaml.Unmarshal(override, &overrideMap); err != nil {
		return nil, fmt.Errorf("parsing override YAML: %w", err)
	}

	merged := deepMerge(baseMap, overrideMap)

	result, err := yaml.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("serializing merged YAML: %w", err)
	}

	return result, nil
}

// deepMerge recursively merges two maps.
func deepMerge(base, override map[string]interface{}) map[string]interface{} {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	result := make(map[string]interface{})

	// Copy base values
	for k, v := range base {
		result[k] = v
	}

	// Merge/override with override values
	for k, v := range override {
		if baseVal, exists := result[k]; exists {
			// If both are maps, merge recursively
			baseMap, baseIsMap := baseVal.(map[string]interface{})
			overrideMap, overrideIsMap := v.(map[string]interface{})
			if baseIsMap && overrideIsMap {
				result[k] = deepMerge(baseMap, overrideMap)
				continue
			}
		}
		// Otherwise, override wins
		result[k] = v
	}

	return result
}

// PersonalConfigInfo provides information about a personal config file.
type PersonalConfigInfo struct {
	BasePath     string // Path to base config file
	PersonalPath string // Path to personal override file
	HasPersonal  bool   // Whether personal override exists
	BaseExists   bool   // Whether base file exists
}

// GetPersonalConfigInfo returns information about a config file and its personal override.
func GetPersonalConfigInfo(configRoot, filename string) PersonalConfigInfo {
	basePath := filepath.Join(configRoot, filename)
	personalPath := ToPersonalPath(basePath)

	_, baseErr := os.Stat(basePath)
	_, personalErr := os.Stat(personalPath)

	return PersonalConfigInfo{
		BasePath:     basePath,
		PersonalPath: personalPath,
		HasPersonal:  personalErr == nil,
		BaseExists:   baseErr == nil,
	}
}

// AllPersonalConfigInfo returns information about all supported personal config files.
func AllPersonalConfigInfo(configRoot string) []PersonalConfigInfo {
	var result []PersonalConfigInfo
	for _, filename := range SupportedPersonalConfigs {
		result = append(result, GetPersonalConfigInfo(configRoot, filename))
	}
	return result
}
