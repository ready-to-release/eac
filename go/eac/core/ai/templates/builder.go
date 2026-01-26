// Package templates provides utilities for building AI prompts using Go templates.
package templates

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"gopkg.in/yaml.v3"
)

// PromptData holds data for prompt template execution.
type PromptData struct {
	Contract    string              // Contract structure as YAML
	ContractRaw *contracts.Contract // Raw contract object for advanced templates
	Custom      map[string]string   // Custom data provided by caller
}

// BuildPromptWithTemplate builds an AI prompt using Go text/template
//
// This function uses Go's standard text/template package to render prompts.
// The template uses {{.FieldName}} syntax for replacements.
//
// Available template fields:
// - {{.Contract}} - Contract structure as YAML string
// - {{.ContractRaw}} - Raw contract object (for accessing specific fields)
// - {{.Custom.TagsSpec}} - Custom data by key
//
// Example template:
//
//	## Contract Requirements
//	{{.Contract}}
//
//	## Tags
//	{{.Custom.TagsSpec}}
//
// Returns rendered prompt or error if template execution fails.
func BuildPromptWithTemplate(
	promptTemplate string,
	contract *contracts.Contract,
	customData map[string]string,
) (string, error) {
	// Parse template
	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	// Prepare template data
	data := PromptData{
		ContractRaw: contract,
		Custom:      customData,
	}

	// Marshal contract to YAML string
	if contract != nil {
		contractYAML, err := yaml.Marshal(contract.RawData)
		if err != nil {
			return "", fmt.Errorf("failed to marshal contract to YAML: %w", err)
		}
		data.Contract = string(contractYAML)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render prompt template: %w\n\nHint: Check that all template variables are defined in the contract.\nCommon issues:\n  - Missing custom data fields ({{.Custom.FieldName}})\n  - Accessing undefined contract fields\n  - Template syntax errors", err)
	}

	return buf.String(), nil
}
