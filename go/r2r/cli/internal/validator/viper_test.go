//go:build L0

package validator

import (
	"encoding/json"
	"strings"
	"testing"
	"github.com/spf13/viper"
)

func TestViperUnmarshalStructure(t *testing.T) {
	yamlContent := `
extensions:
  - name: "test"
    image: "test:latest"
    image_pull_policy: "IfNotPresent"
`
	v := viper.New()
	v.SetConfigType("yaml")
	v.ReadConfig(strings.NewReader(yamlContent))
	
	// Get the raw unmarshaled map
	rawConfig := v.AllSettings()
	
	// Convert to JSON to see the actual structure
	jsonBytes, err := json.MarshalIndent(rawConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	
	t.Logf("Viper produces this JSON structure:\n%s", string(jsonBytes))
}