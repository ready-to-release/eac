// Package contracts provides generalized contract-based validation framework
//
// This package enables commands to:
// - Load contract specifications from YAML files
// - Apply anti-corruption filters to AI output
// - Validate AI output against contracts
// - Build AI prompts with contract context
package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ContractLoader handles loading validation contracts
// Type is inferred from the contract path automatically
type ContractLoader struct {
	workspaceRoot string
	contractPath  string       // e.g., "commit-message" (AI) or "src-core" (domain)
	version       string       // Only used for domain contracts
	contractType  ContractType // Inferred from path prefix
}

// NewContractLoader creates a loader with automatic type inference
// For AI configs: contractPath is command name (e.g., "commit-message", "specifications")
// For domain contracts: contractPath is module name (e.g., "src-core")
// Version is only used for domain contracts; AI configs are unversioned
func NewContractLoader(workspaceRoot string, contractPath string, version string) *ContractLoader {
	// Infer type from path prefix
	contractType := ContractTypeDomain
	aiPath := contractPath
	if strings.HasPrefix(contractPath, "ai/") {
		contractType = ContractTypeAI
		aiPath = strings.TrimPrefix(contractPath, "ai/")
	}

	return &ContractLoader{
		workspaceRoot: workspaceRoot,
		contractPath:  aiPath,
		version:       version,
		contractType:  contractType,
	}
}

// GetType returns the inferred contract type
func (cl *ContractLoader) GetType() ContractType {
	return cl.contractType
}

// IsAI returns true if this loader is for an AI contract
func (cl *ContractLoader) IsAI() bool {
	return cl.contractType == ContractTypeAI
}

// LoadContract loads contract.yml and sets the type
func (cl *ContractLoader) LoadContract() (*Contract, error) {
	contractPath := filepath.Join(cl.workspaceRoot, cl.getConfigDir(), "contract.yml")
	data, err := os.ReadFile(contractPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read contract file %s: %w", contractPath, err)
	}

	var contract Contract
	var rawData map[string]interface{}

	// Parse into typed struct
	if err := yaml.Unmarshal(data, &contract); err != nil {
		return nil, fmt.Errorf("failed to parse contract YAML: %w", err)
	}

	// Also parse into raw map for custom access
	if err := yaml.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse contract YAML into map: %w", err)
	}

	contract.RawData = rawData

	// Set type from loader (inferred from path)
	contract.Type = cl.contractType

	// Validate contract schema
	if err := validateContractSchema(&contract); err != nil {
		return nil, fmt.Errorf("contract validation failed: %w", err)
	}

	return &contract, nil
}

// LoadAntiCorruptionRules loads anti-corruption.yml (only for AI contracts)
func (cl *ContractLoader) LoadAntiCorruptionRules() (*AntiCorruptionRules, error) {
	if !cl.IsAI() {
		return nil, fmt.Errorf("anti-corruption rules are only for AI contracts (path: %s)", cl.contractPath)
	}

	rulesPath := filepath.Join(cl.workspaceRoot, cl.getConfigDir(), "anti-corruption.yml")
	data, err := os.ReadFile(rulesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read anti-corruption rules file %s: %w", rulesPath, err)
	}

	var rules AntiCorruptionRules
	var rawData map[string]interface{}

	// Parse into typed struct
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("failed to parse anti-corruption rules YAML: %w", err)
	}

	// Also parse into raw map for custom access
	if err := yaml.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse anti-corruption rules YAML into map: %w", err)
	}

	rules.RawData = rawData
	return &rules, nil
}

// LoadReferencedFile loads a file referenced by the contract (e.g., tags.yml, taxonomy.yml)
func (cl *ContractLoader) LoadReferencedFile(relativePath string) ([]byte, error) {
	fullPath := filepath.Join(cl.workspaceRoot, relativePath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read referenced file %s: %w", fullPath, err)
	}
	return data, nil
}

// LoadPrompt loads a prompt file with fallback chain (only for AI configs)
// Priority: AI config → embedded
func (cl *ContractLoader) LoadPrompt(promptName string, embeddedPrompt string) (string, string, error) {
	if !cl.IsAI() {
		return "", "", fmt.Errorf("prompts are only for AI configs (path: %s)", cl.contractPath)
	}

	// Try 1: AI config file in .r2r/eac/ai/<command>/<prompt>.md
	aiConfigPrompt := filepath.Join(cl.workspaceRoot, cl.getConfigDir(), promptName)
	if data, err := os.ReadFile(aiConfigPrompt); err == nil {
		return string(data), "AI config", nil
	}

	// Try 2: Embedded prompt (fallback)
	if embeddedPrompt != "" {
		return embeddedPrompt, "embedded default", nil
	}

	return "", "", fmt.Errorf("prompt not found: %s (tried %s and embedded)", promptName, aiConfigPrompt)
}

// GetContractPath returns the full path to the config directory
func (cl *ContractLoader) GetContractPath() string {
	return filepath.Join(cl.workspaceRoot, cl.getConfigDir())
}

// getConfigDir returns the full config directory path
// AI configs: .r2r/eac/ai/<command> (unversioned)
// Domain contracts: contracts/<module>/<version> (versioned)
func (cl *ContractLoader) getConfigDir() string {
	if cl.IsAI() {
		return filepath.Join(".r2r", "eac", "ai", cl.contractPath)
	}
	return filepath.Join("contracts", cl.contractPath, cl.version)
}

// ExtractStringList is a helper to extract string arrays from YAML data
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

// ExtractString is a helper to extract a string value from YAML data
func ExtractString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// ExtractMap is a helper to extract a nested map from YAML data
func ExtractMap(data map[string]interface{}, key string) map[string]interface{} {
	if val, ok := data[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

// validateContractSchema validates that a contract meets minimum requirements
func validateContractSchema(contract *Contract) error {
	// Required fields
	if contract.Version == "" {
		return fmt.Errorf("missing required field: version")
	}
	if contract.Name == "" {
		return fmt.Errorf("missing required field: name")
	}

	// Version format validation (semantic versioning)
	if !isValidVersion(contract.Version) {
		return fmt.Errorf("invalid version format: %s (expected semantic version like 1.0.0)", contract.Version)
	}

	// Check for valid data structure
	if contract.RawData == nil {
		return fmt.Errorf("contract has no data")
	}

	return nil
}

// isValidVersion checks if a version string follows semantic versioning
func isValidVersion(version string) bool {
	// Semantic version: X.Y.Z
	re := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	return re.MatchString(version)
}
