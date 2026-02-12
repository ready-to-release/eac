package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/ready-to-release/eac/go/cli/clie/internal/github"
	"github.com/ready-to-release/eac/go/cli/clie/internal/logging"
	"gopkg.in/yaml.v3"
)

// findOrCreateConfigFile locates the existing config file or creates a new one.
// Returns the config file path and parsed config map.
func findOrCreateConfigFile(repoRoot string) (string, map[string]interface{}, error) {
	configPaths := []string{
		filepath.Join(repoRoot, ".clie", "clie.yml"),
		filepath.Join(repoRoot, ".clie", "clie.yaml"),
	}

	var configPath string
	var configMap map[string]interface{}

	for _, cp := range configPaths {
		if _, err := os.Stat(cp); err == nil {
			configPath = cp
			break
		}
	}

	if configPath == "" {
		configPath = filepath.Join(repoRoot, ".clie", "clie.yml")
		clieDir := filepath.Join(repoRoot, ".clie")
		if err := os.MkdirAll(clieDir, 0o755); err != nil { //nolint:gosec // G301: config dir needs to be accessible
			return "", nil, fmt.Errorf("failed to create .clie directory: %w", err)
		}
		logging.Infof("📝 Creating %s", configPath)
		configMap = map[string]interface{}{
			"version":    "1.0",
			"extensions": []interface{}{},
		}
	} else {
		configData, err := os.ReadFile(configPath)
		if err != nil {
			return "", nil, fmt.Errorf("failed to read config file: %w", err)
		}
		if err := yaml.Unmarshal(configData, &configMap); err != nil {
			return "", nil, fmt.Errorf("failed to parse config: %w", err)
		}
	}

	if configMap["version"] == nil {
		configMap["version"] = "1.0"
	}

	return configPath, configMap, nil
}

// getExtensionsList extracts the extensions list from a config map.
func getExtensionsList(configMap map[string]interface{}) []interface{} {
	if exts, ok := configMap["extensions"].([]interface{}); ok {
		return exts
	}
	return []interface{}{}
}

// extensionExistsInConfig checks if an extension already exists in the config.
func extensionExistsInConfig(extensions []interface{}, extensionName string) bool {
	for _, ext := range extensions {
		if extMap, ok := ext.(map[string]interface{}); ok {
			if name, ok := extMap["name"].(string); ok && name == extensionName {
				return true
			}
		}
	}
	return false
}

// verifyAndAddExtension verifies the extension exists in the registry and adds it.
func verifyAndAddExtension(extensions []interface{}, extensionName string) ([]interface{}, error) {
	client, err := github.NewRegistryClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create registry client: %w", err)
	}

	imagePath := fmt.Sprintf("ready-to-release/ext-%s", extensionName)
	tags, err := client.ListTags(imagePath)
	if err != nil || len(tags) == 0 {
		if os.Getenv("GITHUB_TOKEN") == "" {
			return nil, fmt.Errorf("extension not found in registry: %s\nNote: Set GITHUB_TOKEN and GITHUB_USERNAME environment variables to access private extensions", extensionName)
		}
		return nil, fmt.Errorf("extension not found in registry: %s", extensionName)
	}

	extensions = append(extensions, map[string]interface{}{
		"name":        extensionName,
		"description": fmt.Sprintf("%s development environment", cases.Title(language.English).String(extensionName)),
		"image":       fmt.Sprintf("ghcr.io/%s:latest", imagePath),
	})
	return extensions, nil
}

// GetLatestSHA queries the registry directly for the latest SHA tag of an extension.
// Returns the SHA tag string (e.g., "sha-abc123") or an error if unavailable.
func GetLatestSHA(extensionName string) (string, error) {
	client, err := github.NewRegistryClient()
	if err != nil {
		return "", fmt.Errorf("failed to create registry client: %w", err)
	}

	imagePath := fmt.Sprintf("ready-to-release/ext-%s", extensionName)
	tags, err := client.ListTags(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to list tags for %s: %w", extensionName, err)
	}

	// Find the latest sha- tag (they are returned newest-first from GHCR)
	for _, tag := range tags {
		if strings.HasPrefix(tag, "sha-") {
			return tag, nil
		}
	}

	return "", fmt.Errorf("no SHA tag found for extension %s", extensionName)
}

// resolveLatestSHA resolves the latest SHA tag for an extension using GetLatestSHA.
func resolveLatestSHA(_ []interface{}, extensionName string) (map[string]string, error) {
	logging.Infof("📌 Getting latest SHA tag for %s...", extensionName)

	sha, err := GetLatestSHA(extensionName)
	if err != nil {
		return nil, err
	}

	shaMap := map[string]string{
		extensionName: sha,
	}
	return shaMap, nil
}

// updateExtensionWithSHA updates a specific extension in the list with its SHA tag.
func updateExtensionWithSHA(extensions []interface{}, extensionName string, shaMap map[string]string) ([]interface{}, error) {
	for i, ext := range extensions {
		extMap, ok := ext.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := extMap["name"].(string)   //nolint:errcheck // type assertion with fallback to empty string
		image, _ := extMap["image"].(string) //nolint:errcheck // type assertion with fallback to empty string

		if name != extensionName {
			continue
		}

		if image == "" {
			continue
		}

		if strings.Contains(image, ":sha-") {
			logging.Infof("✅ %s already pinned", name)
			return extensions, nil
		}

		baseImage := image
		if idx := strings.LastIndex(image, ":"); idx > 0 {
			baseImage = image[:idx]
		}

		latestSHA, ok := shaMap[name]
		if !ok || latestSHA == "" || latestSHA == "sha-<unavailable>" {
			return nil, fmt.Errorf("failed to get latest SHA for %s", name)
		}

		extMap["image"] = baseImage + ":" + latestSHA
		extMap["image_pull_policy"] = "IfNotPresent"
		extensions[i] = extMap
		logging.Infof("📌 %s configured with %s", name, latestSHA)
		return extensions, nil
	}

	return nil, fmt.Errorf("failed to update extension %s", extensionName)
}

// writeConfigFile writes the updated config back to the YAML file.
func writeConfigFile(configPath string, configMap map[string]interface{}, extensions []interface{}) error {
	type OrderedConfig struct {
		Version    string        `yaml:"version"`
		Extensions []interface{} `yaml:"extensions"`
	}

	orderedConfig := OrderedConfig{
		Version:    configMap["version"].(string), //nolint:errcheck // version field always present in valid config
		Extensions: extensions,
	}

	updatedConfig, err := yaml.Marshal(&orderedConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, updatedConfig, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	logging.Infof("✅ Configuration updated in %s", configPath)
	return nil
}
