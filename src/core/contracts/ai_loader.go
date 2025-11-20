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
	contractPath  string       // e.g., "ai/specifications" or "modules"
	version       string
	contractType  ContractType // Inferred from path
}

// NewAIContractLoader creates a loader for AI contracts
// contractName: "commit-message", "specifications"
func NewAIContractLoader(workspaceRoot string, contractName string, version string) *ContractLoader {
	return &ContractLoader{
		workspaceRoot: workspaceRoot,
		contractPath:  fmt.Sprintf("ai/%s", contractName),
		version:       version,
		contractType:  ContractTypeAI, // Set automatically
	}
}

// NewDomainContractLoader creates a loader for domain contracts
// contractName: "modules", "environments", "testing", "vscode-commit", etc.
func NewDomainContractLoader(workspaceRoot string, contractName string, version string) *ContractLoader {
	return &ContractLoader{
		workspaceRoot: workspaceRoot,
		contractPath:  contractName,
		version:       version,
		contractType:  ContractTypeDomain, // Set automatically
	}
}

// NewContractLoader creates a loader with automatic type inference
// contractPath: "ai/specifications", "modules", "environments", etc.
func NewContractLoader(workspaceRoot string, contractPath string, version string) *ContractLoader {
	// Infer type from path
	contractType := ContractTypeDomain
	if strings.HasPrefix(contractPath, "ai/") {
		contractType = ContractTypeAI
	}

	return &ContractLoader{
		workspaceRoot: workspaceRoot,
		contractPath:  contractPath,
		version:       version,
		contractType:  contractType,
	}
}

// NewSpecContractLoader creates a loader for backward compatibility
// DEPRECATED: Use NewContractLoader, NewAIContractLoader, or NewDomainContractLoader
func NewSpecContractLoader(workspaceRoot string, contractPath string, version string) *ContractLoader {
	return NewContractLoader(workspaceRoot, contractPath, version)
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
	contractPath := filepath.Join(cl.workspaceRoot, "contracts", cl.contractPath, cl.version, "contract.yml")
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

	rulesPath := filepath.Join(cl.workspaceRoot, "contracts", cl.contractPath, cl.version, "anti-corruption.yml")
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

// LoadPrompt loads a prompt file with fallback chain (only for AI contracts)
// Priority: .r2r/contracts → contracts → embedded
func (cl *ContractLoader) LoadPrompt(promptName string, embeddedPrompt string) (string, string, error) {
	if !cl.IsAI() {
		return "", "", fmt.Errorf("prompts are only for AI contracts (path: %s)", cl.contractPath)
	}

	// Try 1: .r2r/contracts/ai/<command>/<version>/<prompt>.md
	localContractPrompt := filepath.Join(cl.workspaceRoot, ".r2r", "contracts", cl.contractPath, cl.version, promptName)
	if data, err := os.ReadFile(localContractPrompt); err == nil {
		return string(data), "local contract override", nil
	}

	// Try 2: contracts/ai/<command>/<version>/<prompt>.md (built-in)
	repoContractPrompt := filepath.Join(cl.workspaceRoot, "contracts", cl.contractPath, cl.version, promptName)
	if data, err := os.ReadFile(repoContractPrompt); err == nil {
		return string(data), "repository contract", nil
	}

	// Try 3: Embedded prompt (fallback)
	if embeddedPrompt != "" {
		return embeddedPrompt, "embedded default", nil
	}

	return "", "", fmt.Errorf("prompt not found: %s (tried .r2r/contracts, contracts, and embedded)", promptName)
}

// GetContractPath returns the full path to the contract directory
func (cl *ContractLoader) GetContractPath() string {
	return filepath.Join(cl.workspaceRoot, "contracts", cl.contractPath, cl.version)
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
