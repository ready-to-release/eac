// Package templates provides helper functions for the templates command
package templates

import (
	"encoding/json"
	"fmt"
	"os"
)

// TemplateValues represents the key-value pairs for template substitution
type TemplateValues map[string]interface{}

// LoadValuesFromJSON reads and parses a JSON file into template values
func LoadValuesFromJSON(filePath string) (TemplateValues, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read values file %s: %w", filePath, err)
	}

	// Parse JSON
	var values TemplateValues
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from %s: %w", filePath, err)
	}

	// Validate non-empty
	if len(values) == 0 {
		return nil, fmt.Errorf("values file %s is empty", filePath)
	}

	return values, nil
}

// ValidateValues checks that all required values are present
func ValidateValues(values TemplateValues, required []string) error {
	for _, key := range required {
		if _, exists := values[key]; !exists {
			return fmt.Errorf("required value %q not found in values file", key)
		}
	}
	return nil
}
