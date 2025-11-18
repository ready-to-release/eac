// Package contract provides generalized contract-based AI validation framework
//
// This package enables commands to:
// - Load contract specifications from YAML files
// - Apply anti-corruption filters to AI output
// - Validate AI output against contracts
// - Build AI prompts with contract context
package contract

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// SpecContract represents a specification/validation contract (e.g., commit-message, gherkin)
// This is different from module Contract which represents module metadata
type SpecContract struct {
	Version     string                 `yaml:"version"`
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	RawData     map[string]interface{} // Full contract data for custom access
}

// AntiCorruptionRules represents noise filtering rules
type AntiCorruptionRules struct {
	Version             string                 `yaml:"version"`
	Name                string                 `yaml:"name"`
	Description         string                 `yaml:"description"`
	ForbiddenPrefixes   []string               `yaml:"forbidden_prefixes"`
	ForbiddenContains   []string               `yaml:"forbidden_contains"`
	ForbiddenEmojis     []string               `yaml:"forbidden_emojis"`
	NoiseHeaderKeywords []string               `yaml:"noise_header_keywords"`
	AgentSignatures     []string               `yaml:"agent_signatures"`
	RawData             map[string]interface{} // Full rules data for custom access
}

// ValidationError represents a contract violation
type ValidationError struct {
	Code     string
	Message  string
	Line     int
	Severity string // "error" or "warning"
}

func (e ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("[%s] Line %d: %s", e.Code, e.Line, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// SpecValidator interface must be implemented by specific contract validators
type SpecValidator interface {
	// Validate validates output against the contract
	Validate(output string, context map[string]interface{}) []ValidationError

	// VerifyImplementation verifies that the validator implements all contract rules
	VerifyImplementation() []ValidationError
}

// SpecContractLoader handles loading specification/validation contracts and related files
type SpecContractLoader struct {
	workspaceRoot string
	contractDir   string // e.g., "contracts/commit-message" or "contracts/specifications"
	version       string // e.g., "0.1.0"
}

// NewSpecContractLoader creates a new specification contract loader
func NewSpecContractLoader(workspaceRoot string, contractDir string, version string) *SpecContractLoader {
	return &SpecContractLoader{
		workspaceRoot: workspaceRoot,
		contractDir:   contractDir,
		version:       version,
	}
}

// LoadContract loads the structure.yml contract
func (cl *SpecContractLoader) LoadContract() (*SpecContract, error) {
	contractPath := filepath.Join(cl.workspaceRoot, cl.contractDir, cl.version, "structure.yml")
	data, err := os.ReadFile(contractPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read contract file %s: %w", contractPath, err)
	}

	var contract SpecContract
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

	// Validate contract schema
	if err := validateContractSchema(&contract); err != nil {
		return nil, fmt.Errorf("contract validation failed: %w", err)
	}

	return &contract, nil
}

// LoadAntiCorruptionRules loads the anti-corruption.yml rules
func (cl *SpecContractLoader) LoadAntiCorruptionRules() (*AntiCorruptionRules, error) {
	rulesPath := filepath.Join(cl.workspaceRoot, cl.contractDir, cl.version, "anti-corruption.yml")
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
func (cl *SpecContractLoader) LoadReferencedFile(relativePath string) ([]byte, error) {
	fullPath := filepath.Join(cl.workspaceRoot, relativePath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read referenced file %s: %w", fullPath, err)
	}
	return data, nil
}

// GetContractPath returns the full path to the contract directory
func (cl *SpecContractLoader) GetContractPath() string {
	return filepath.Join(cl.workspaceRoot, cl.contractDir, cl.version)
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

// PromptData holds data for prompt template execution
type PromptData struct {
	Contract          string               // Contract structure as YAML
	AntiCorruption    string               // Anti-corruption rules as YAML
	ContractRaw       *SpecContract        // Raw contract object for advanced templates
	AntiCorruptionRaw *AntiCorruptionRules // Raw anti-corruption object
	Custom            map[string]string    // Custom data provided by caller
}

// BuildPromptWithTemplate builds an AI prompt using Go text/template
//
// This function uses Go's standard text/template package to render prompts.
// The template uses {{.FieldName}} syntax for replacements.
//
// Available template fields:
// - {{.Contract}} - Contract structure as YAML string
// - {{.AntiCorruption}} - Anti-corruption rules as YAML string
// - {{.ContractRaw}} - Raw contract object (for accessing specific fields)
// - {{.AntiCorruptionRaw}} - Raw anti-corruption object
// - {{.Custom.TagsSpec}} - Custom data by key
//
// Example template:
//   ## Contract Requirements
//   {{.Contract}}
//
//   ## Tags
//   {{.Custom.TagsSpec}}
//
// Returns rendered prompt or error if template execution fails.
func BuildPromptWithTemplate(
	promptTemplate string,
	contract *SpecContract,
	antiCorruption *AntiCorruptionRules,
	customData map[string]string,
) (string, error) {
	// Parse template
	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	// Prepare template data
	data := PromptData{
		ContractRaw:       contract,
		AntiCorruptionRaw: antiCorruption,
		Custom:            customData,
	}

	// Marshal contract to YAML string
	if contract != nil {
		contractYAML, err := yaml.Marshal(contract.RawData)
		if err != nil {
			return "", fmt.Errorf("failed to marshal contract to YAML: %w", err)
		}
		data.Contract = string(contractYAML)
	}

	// Marshal anti-corruption rules to YAML string
	if antiCorruption != nil {
		rulesYAML, err := yaml.Marshal(antiCorruption.RawData)
		if err != nil {
			return "", fmt.Errorf("failed to marshal anti-corruption rules to YAML: %w", err)
		}
		data.AntiCorruption = string(rulesYAML)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render prompt template: %w\n\nHint: Check that all template variables are defined in the contract.\nCommon issues:\n  - Missing custom data fields ({{.Custom.FieldName}})\n  - Accessing undefined contract fields\n  - Template syntax errors", err)
	}

	return buf.String(), nil
}

// BuildPromptWithContract is deprecated. Use BuildPromptWithTemplate instead.
// Kept for backward compatibility.
func BuildPromptWithContract(
	promptTemplate string,
	contract *SpecContract,
	antiCorruption *AntiCorruptionRules,
	replacements map[string]string,
) string {
	result, err := BuildPromptWithTemplate(promptTemplate, contract, antiCorruption, replacements)
	if err != nil {
		// Fallback to simple replacement for backward compatibility
		result = promptTemplate
		if contract != nil {
			contractYAML, _ := yaml.Marshal(contract.RawData)
			result = strings.ReplaceAll(result, "{CONTRACT_STRUCTURE}", string(contractYAML))
		}
		if antiCorruption != nil {
			rulesYAML, _ := yaml.Marshal(antiCorruption.RawData)
			result = strings.ReplaceAll(result, "{ANTI_CORRUPTION_RULES}", string(rulesYAML))
		}
		for k, v := range replacements {
			result = strings.ReplaceAll(result, "{"+k+"}", v)
		}
	}
	return result
}

// validateContractSchema validates that a contract meets minimum requirements
//
// This function ensures:
// - Required fields are present (version, name)
// - Version follows semantic versioning (X.Y.Z)
// - Contract has valid data structure
func validateContractSchema(contract *SpecContract) error {
	// Required fields
	if contract.Version == "" {
		return fmt.Errorf("missing required field: version")
	}
	if contract.Name == "" {
		return fmt.Errorf("missing required field: name")
	}

	// Version format validation
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
// Expected format: X.Y.Z where X, Y, Z are integers
func isValidVersion(version string) bool {
	// Semantic version: X.Y.Z
	re := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	return re.MatchString(version)
}
