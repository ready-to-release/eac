// Package contracts provides generalized contract-based validation framework
package contracts

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/paths"
	"gopkg.in/yaml.v3"
)

// AIConfigLoader loads the unified AI configuration
type AIConfigLoader struct {
	workspaceRoot string
	config        *AIConfig
}

// NewAIConfigLoader creates a new AI config loader
func NewAIConfigLoader(workspaceRoot string) *AIConfigLoader {
	return &AIConfigLoader{
		workspaceRoot: workspaceRoot,
	}
}

// Load loads the unified ai-config.yml file
func (l *AIConfigLoader) Load() (*AIConfig, error) {
	if l.config != nil {
		return l.config, nil
	}

	configPath := filepath.Join(l.workspaceRoot, paths.R2RDir, paths.EACDir, "ai-config.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read AI config %s: %w", configPath, err)
	}

	var config AIConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse AI config: %w", err)
	}

	l.config = &config
	return &config, nil
}

// GetType returns the configuration for a specific AI type with merged anti-corruption rules
func (l *AIConfigLoader) GetType(typeName string) (*AITypeConfig, error) {
	config, err := l.Load()
	if err != nil {
		return nil, err
	}

	typeConfig, ok := config.Types[typeName]
	if !ok {
		return nil, fmt.Errorf("unknown AI type: %s", typeName)
	}

	return &typeConfig, nil
}

// GetAntiCorruptionRules returns merged anti-corruption rules for a type (defaults + type-specific)
func (l *AIConfigLoader) GetAntiCorruptionRules(typeName string) (*AntiCorruptionRules, error) {
	config, err := l.Load()
	if err != nil {
		return nil, err
	}

	typeConfig, ok := config.Types[typeName]
	if !ok {
		return nil, fmt.Errorf("unknown AI type: %s", typeName)
	}

	// Merge defaults with type-specific rules
	merged := config.Defaults.AntiCorruption.Merge(typeConfig.AntiCorruption)
	return &merged, nil
}

// GetStartMarker returns the content start marker for a type (nil means immediate start)
func (l *AIConfigLoader) GetStartMarker(typeName string) (*string, error) {
	typeConfig, err := l.GetType(typeName)
	if err != nil {
		return nil, err
	}
	return typeConfig.Output.StartMarker, nil
}

// GetValidation returns the validation rules for a type
func (l *AIConfigLoader) GetValidation(typeName string) (map[string]interface{}, error) {
	typeConfig, err := l.GetType(typeName)
	if err != nil {
		return nil, err
	}
	return typeConfig.Validation, nil
}

// LoadPrompt loads a prompt file from ai/prompts/<name>.md
func (l *AIConfigLoader) LoadPrompt(promptName string) (string, error) {
	promptPath := filepath.Join(l.workspaceRoot, paths.R2RDir, paths.EACDir, "ai", "prompts", promptName)
	data, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt %s: %w", promptPath, err)
	}
	return string(data), nil
}

// LoadData loads a data file referenced by a type
func (l *AIConfigLoader) LoadData(typeName string, dataKey string) ([]byte, error) {
	typeConfig, err := l.GetType(typeName)
	if err != nil {
		return nil, err
	}

	dataPath, ok := typeConfig.Data[dataKey]
	if !ok {
		return nil, fmt.Errorf("unknown data key %s for type %s", dataKey, typeName)
	}

	fullPath := filepath.Join(l.workspaceRoot, paths.R2RDir, paths.EACDir, dataPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read data file %s: %w", fullPath, err)
	}
	return data, nil
}

// LoadReferencedFile loads a file by path relative to workspace root
func (l *AIConfigLoader) LoadReferencedFile(relativePath string) ([]byte, error) {
	fullPath := filepath.Join(l.workspaceRoot, relativePath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", fullPath, err)
	}
	return data, nil
}

// GetWorkspaceRoot returns the workspace root
func (l *AIConfigLoader) GetWorkspaceRoot() string {
	return l.workspaceRoot
}

// ContractLoader provides backward-compatible API for loading AI configs
// Deprecated: Use AIConfigLoader directly for new code
type ContractLoader struct {
	loader   *AIConfigLoader
	typeName string
}

// NewContractLoader creates a backward-compatible loader
// contractPath format: "ai/<type>" (e.g., "ai/specs", "ai/commit-message")
// version is ignored (unified config is unversioned)
func NewContractLoader(workspaceRoot string, contractPath string, version string) *ContractLoader {
	// Extract type name from path (e.g., "ai/specs" -> "specs")
	typeName := contractPath
	if len(contractPath) > 3 && contractPath[:3] == "ai/" {
		typeName = contractPath[3:]
	}

	return &ContractLoader{
		loader:   NewAIConfigLoader(workspaceRoot),
		typeName: typeName,
	}
}

// LoadContract loads validation config as a Contract for backward compatibility
func (cl *ContractLoader) LoadContract() (*Contract, error) {
	typeConfig, err := cl.loader.GetType(cl.typeName)
	if err != nil {
		return nil, err
	}

	// Create Contract from type config
	contract := &Contract{
		Version:     "0.1.0",
		Name:        typeConfig.Name,
		Description: typeConfig.Description,
		Type:        ContractTypeAI,
		RawData:     typeConfig.Validation,
	}

	return contract, nil
}

// LoadAntiCorruptionRules loads merged anti-corruption rules
func (cl *ContractLoader) LoadAntiCorruptionRules() (*AntiCorruptionRules, error) {
	return cl.loader.GetAntiCorruptionRules(cl.typeName)
}

// LoadPrompt loads a prompt file from the new prompts directory
func (cl *ContractLoader) LoadPrompt(promptName string, fallback string) (string, string, error) {
	prompt, err := cl.loader.LoadPrompt(promptName)
	if err != nil {
		if fallback != "" {
			return fallback, "fallback", nil
		}
		return "", "", err
	}
	return prompt, "config", nil
}

// LoadReferencedFile loads a file by path
func (cl *ContractLoader) LoadReferencedFile(relativePath string) ([]byte, error) {
	return cl.loader.LoadReferencedFile(relativePath)
}

// GetContractPath returns the path to the AI config directory
func (cl *ContractLoader) GetContractPath() string {
	return filepath.Join(cl.loader.workspaceRoot, paths.R2RDir, paths.EACDir, "ai", cl.typeName)
}

// IsAI returns true (all ContractLoader instances are for AI configs)
func (cl *ContractLoader) IsAI() bool {
	return true
}

// Helper functions for extracting typed values from validation maps

// ExtractStringList extracts a string array from a map
func ExtractStringList(data map[string]interface{}, key string) []string {
	if val, ok := data[key]; ok {
		if list, ok := val.([]interface{}); ok {
			result := make([]string, 0, len(list))
			for _, item := range list {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return []string{}
}

// ExtractString extracts a string value from a map
func ExtractString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// ExtractInt extracts an int value from a map
func ExtractInt(data map[string]interface{}, key string) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

// ExtractBool extracts a bool value from a map
func ExtractBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// ExtractMap extracts a nested map from a map
func ExtractMap(data map[string]interface{}, key string) map[string]interface{} {
	if val, ok := data[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}
