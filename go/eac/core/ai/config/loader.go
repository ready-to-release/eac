// Package config handles AI configuration loading and management
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"gopkg.in/yaml.v3"
)

const (
	// Local AI configuration constants.
	promptExtension = ".md"
	promptsSubdir   = "prompts"
)

// AIConfigLoader loads the unified AI configuration.
type AIConfigLoader struct {
	workspaceRoot string
	config        *AIConfig
}

// NewAIConfigLoader creates a new AI config loader.
func NewAIConfigLoader(workspaceRoot string) *AIConfigLoader {
	return &AIConfigLoader{
		workspaceRoot: workspaceRoot,
	}
}

// Load loads the unified ai-config.yml file with fallback to system defaults and contracts defaults.
func (l *AIConfigLoader) Load() (*AIConfig, error) {
	if l.config != nil {
		return l.config, nil
	}

	// Try user override first (/var/task/.r2r/eac/ai-config.yml)
	configPath := filepath.Join(l.workspaceRoot, paths.R2RDir, paths.EACDir, paths.AIConfigFilename)
	data, err := os.ReadFile(configPath)

	// If not found, fall back to system default (/app/.r2r/eac/ai-config.yml)
	if os.IsNotExist(err) {
		systemRoot := os.Getenv(paths.ContainerRootEnv)
		if systemRoot == "" {
			// Local dev mode - try workspace root
			systemRoot = l.workspaceRoot
		}
		systemPath := filepath.Join(systemRoot, paths.R2RDir, paths.EACDir, paths.AIConfigFilename)
		data, err = os.ReadFile(systemPath)

		// If still not found, try contracts default (ultimate fallback)
		if os.IsNotExist(err) {
			// contracts/eac-core/0.1.0/defaults/ai-config.yml
			// Use workspaceRoot (in container, this is /app; in dev, this is repo root)
			contractsRoot := l.workspaceRoot
			contractsPath := filepath.Join(contractsRoot, "contracts", "eac-core", "0.1.0", "defaults", paths.AIConfigFilename)
			data, err = os.ReadFile(contractsPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read AI config from user repo (%s), system defaults (%s), or contracts defaults (%s): %w",
					configPath, systemPath, contractsPath, err)
			}
		} else if err != nil {
			return nil, fmt.Errorf("failed to read AI config from system defaults %s: %w", systemPath, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to read AI config %s: %w", configPath, err)
	}

	var config AIConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse AI config: %w", err)
	}

	l.config = &config
	return &config, nil
}

// GetType returns the configuration for a specific AI type.
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

// LoadPrompt loads a prompt file from ai/prompts/<name>.md.
func (l *AIConfigLoader) LoadPrompt(promptName string) (string, error) {
	promptPath := filepath.Join(l.workspaceRoot, paths.R2RDir, paths.EACDir, paths.AIDir, promptsSubdir, promptName)
	data, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt %s: %w", promptPath, err)
	}
	return string(data), nil
}

// LoadData loads a data file referenced by a type.
func (l *AIConfigLoader) LoadData(typeName, dataKey string) ([]byte, error) {
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

// LoadReferencedFile loads a file by path relative to workspace root.
// Falls back to system defaults if not found in user repository.
// This allows commands to work without requiring all files in user repos.
func (l *AIConfigLoader) LoadReferencedFile(relativePath string) ([]byte, error) {
	// Try user override first
	fullPath := filepath.Join(l.workspaceRoot, relativePath)
	data, err := os.ReadFile(fullPath)
	if err == nil {
		return data, nil
	}

	// If not found in user repo, fall back to system defaults
	if os.IsNotExist(err) {
		systemRoot := os.Getenv(paths.ContainerRootEnv)
		if systemRoot == "" {
			// Local dev mode - system defaults are same as workspace root
			// This means no fallback is available, return original error
			systemRoot = l.workspaceRoot
		}
		systemPath := filepath.Join(systemRoot, relativePath)
		data, err = os.ReadFile(systemPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read file from user repo (%s) or system defaults (%s): %w", fullPath, systemPath, err)
			}
			return nil, fmt.Errorf("failed to read file from system defaults (%s): %w", systemPath, err)
		}
		return data, nil
	}

	// Other error (not IsNotExist)
	return nil, fmt.Errorf("failed to read file %s: %w", fullPath, err)
}

// GetWorkspaceRoot returns the workspace root.
func (l *AIConfigLoader) GetWorkspaceRoot() string {
	return l.workspaceRoot
}

// ContractLoader provides backward-compatible API for loading AI configs
// Deprecated: Use AIConfigLoader directly for new code.
type ContractLoader struct {
	loader   *AIConfigLoader
	typeName string
}

// NewContractLoader creates a backward-compatible loader
// contractPath format: "ai/<type>" (e.g., "ai/specs", "ai/commit-message")
// version is ignored (unified config is unversioned).
func NewContractLoader(workspaceRoot, contractPath, version string) *ContractLoader {
	// Extract type name from path (e.g., "ai/specs" -> "specs")
	typeName := contractPath
	aiPrefix := paths.AIDir + "/"
	if strings.HasPrefix(contractPath, aiPrefix) {
		typeName = contractPath[len(aiPrefix):]
	}

	return &ContractLoader{
		loader:   NewAIConfigLoader(workspaceRoot),
		typeName: typeName,
	}
}

// LoadContract loads validation config as a Contract for backward compatibility
// Note: Validation is now done via JSON schema. This returns minimal contract info.
func (cl *ContractLoader) LoadContract() (*contracts.Contract, error) {
	typeConfig, err := cl.loader.GetType(cl.typeName)
	if err != nil {
		return nil, err
	}

	// Create minimal Contract from type config
	// Actual validation is done via JSON schema
	contract := &contracts.Contract{
		Version:     paths.DefaultsVersion,
		Name:        typeConfig.Name,
		Description: typeConfig.Description,
		Type:        contracts.ContractTypeAI,
		RawData:     make(map[string]interface{}), // Empty - using JSON schema now
	}

	return contract, nil
}

// LoadPrompt loads a prompt with three-tier priority:
// 1. Custom path (absolute path means --prompt flag was used)
// 2. Team override (.r2r/eac/templates/ai/<type>/<name>)
// 3. System default (templates/ai/<type>/<name>)
// 4. Fallback (if provided, for backward compatibility)
//
// Convention: If promptName is empty, uses type name as file name.
// Example: type "ai/specs" with promptName "" → "specs.md"
// Extension: Automatically adds ".md" if not present.
func (cl *ContractLoader) LoadPrompt(promptName, fallback string) (string, string, error) {
	// Convention: Use type name if no prompt name provided
	if promptName == "" {
		promptName = cl.getDefaultPromptName()
	}

	// Normalize: Add .md extension if missing (unless absolute path)
	if !filepath.IsAbs(promptName) && !strings.HasSuffix(promptName, promptExtension) {
		promptName = promptName + promptExtension
	}

	// Priority 1: Custom path from --prompt flag (absolute path)
	if filepath.IsAbs(promptName) {
		content, err := os.ReadFile(promptName)
		if err != nil {
			return "", "", fmt.Errorf("custom prompt not found: %w", err)
		}
		return string(content), "command flag", nil
	}

	// Priority 2: Team override (.r2r/eac/templates/ai/<type>/<name>)
	// Note: Team override uses "ai/<type>" path structure
	teamOverridePath := filepath.Join(cl.loader.workspaceRoot, paths.R2RDir, paths.EACDir, paths.TemplatesDir, paths.AIDir, cl.typeName, promptName)
	if content, err := os.ReadFile(teamOverridePath); err == nil {
		return string(content), "team override", nil
	}

	// Priority 3: System default from distribution root
	// In container: uses R2R_CONTAINER_ROOT (/app where Dockerfile copies templates)
	// In local dev: uses workspaceRoot (repo root where templates/ exists)
	// This ensures templates work in both container and development scenarios
	// Note: System default uses "ai/<type>" path structure
	distRoot := cl.loader.workspaceRoot
	if containerRoot := os.Getenv(paths.ContainerRootEnv); containerRoot != "" {
		distRoot = containerRoot
	}
	systemDefaultPath := filepath.Join(distRoot, paths.TemplatesDir, paths.AIDir, cl.typeName, promptName)
	if content, err := os.ReadFile(systemDefaultPath); err == nil {
		return string(content), "system default", nil
	}

	// Priority 4: Fallback (for backward compatibility)
	if fallback != "" {
		return fallback, "embedded fallback", nil
	}

	return "", "", fmt.Errorf("no prompt found: %s (checked team override and system default)", promptName)
}

// getDefaultPromptName derives the default prompt file name from the AI type name.
// Convention: "ai/specs" → "specs.md", "ai/commit-message" → "commit-message.md".
func (cl *ContractLoader) getDefaultPromptName() string {
	// Extract base name from type path
	// "ai/specs" → "specs"
	// "ai/commit-message" → "commit-message"
	// "ai/risk-profile" → "risk-profile"
	baseName := filepath.Base(cl.typeName)
	return baseName + promptExtension
}

// LoadPromptWithPriority loads a prompt with explicit priority handling
// This method provides explicit control over custom path vs prompt name.
func (cl *ContractLoader) LoadPromptWithPriority(promptName, customPath string) (string, string, error) {
	// Priority 1: Custom path from --prompt flag
	if customPath != "" {
		if !filepath.IsAbs(customPath) {
			customPath = filepath.Join(cl.loader.workspaceRoot, customPath)
		}
		content, err := os.ReadFile(customPath)
		if err != nil {
			return "", "", fmt.Errorf("custom prompt not found: %w", err)
		}
		return string(content), "command flag", nil
	}

	// Priority 2 & 3: Delegate to LoadPrompt
	return cl.LoadPrompt(promptName, "")
}

// LoadReferencedFile loads a file by path.
func (cl *ContractLoader) LoadReferencedFile(relativePath string) ([]byte, error) {
	return cl.loader.LoadReferencedFile(relativePath)
}

// GetContractPath returns the path to the AI config directory.
func (cl *ContractLoader) GetContractPath() string {
	return filepath.Join(cl.loader.workspaceRoot, paths.R2RDir, paths.EACDir, paths.AIDir, cl.typeName)
}

// IsAI returns true (all ContractLoader instances are for AI configs).
func (cl *ContractLoader) IsAI() bool {
	return true
}

// Helper functions for extracting typed values from validation maps

// ExtractStringList extracts a string array from a map.
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

// ExtractString extracts a string value from a map.
func ExtractString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// ExtractInt extracts an int value from a map.
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

// ExtractBool extracts a bool value from a map.
func ExtractBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// ExtractMap extracts a nested map from a map.
func ExtractMap(data map[string]interface{}, key string) map[string]interface{} {
	if val, ok := data[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

// LoadAIConfig is a convenience function to load AI configuration.
func LoadAIConfig(workspaceRoot string) (*AIConfig, error) {
	loader := NewAIConfigLoader(workspaceRoot)
	return loader.Load()
}
