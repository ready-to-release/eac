package contracts

import (
	"bytes"
	"fmt"
	"text/template"

	"gopkg.in/yaml.v3"
)

// PromptData holds data for prompt template execution
type PromptData struct {
	Contract          string               // Contract structure as YAML
	AntiCorruption    string               // Anti-corruption rules as YAML
	ContractRaw       *Contract            // Raw contract object for advanced templates
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
	contract *Contract,
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
