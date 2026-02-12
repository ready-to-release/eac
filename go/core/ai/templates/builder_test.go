package templates

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPromptWithTemplate_BasicTemplate(t *testing.T) {
	tmpl := "Hello, this is a prompt."
	result, err := BuildPromptWithTemplate(tmpl, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "Hello, this is a prompt.", result)
}

func TestBuildPromptWithTemplate_ContractRendering(t *testing.T) {
	contract := &domain.Contract{
		Version:     "1.0",
		Name:        "test-contract",
		Description: "A test contract",
		RawData: map[string]interface{}{
			"version": "1.0",
			"name":    "test-contract",
		},
	}

	tmpl := "## Contract\n{{.Contract}}"
	result, err := BuildPromptWithTemplate(tmpl, contract, nil)
	require.NoError(t, err)
	assert.Contains(t, result, "## Contract")
	assert.Contains(t, result, "name: test-contract")
	assert.Contains(t, result, "version: \"1.0\"")
}

func TestBuildPromptWithTemplate_NilContract(t *testing.T) {
	tmpl := "No contract: [{{.Contract}}]"
	result, err := BuildPromptWithTemplate(tmpl, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "No contract: []", result)
}

func TestBuildPromptWithTemplate_CustomData(t *testing.T) {
	customData := map[string]string{
		"TagsSpec": "@smoke @regression",
		"Extra":    "some-value",
	}

	tmpl := "Tags: {{.Custom.TagsSpec}}\nExtra: {{.Custom.Extra}}"
	result, err := BuildPromptWithTemplate(tmpl, nil, customData)
	require.NoError(t, err)
	assert.Contains(t, result, "Tags: @smoke @regression")
	assert.Contains(t, result, "Extra: some-value")
}

func TestBuildPromptWithTemplate_ContractAndCustomData(t *testing.T) {
	contract := &domain.Contract{
		Version: "2.0",
		Name:    "full-contract",
		RawData: map[string]interface{}{
			"version": "2.0",
			"name":    "full-contract",
		},
	}
	customData := map[string]string{
		"TagsSpec": "@critical",
	}

	tmpl := "## Contract\n{{.Contract}}\n## Tags\n{{.Custom.TagsSpec}}"
	result, err := BuildPromptWithTemplate(tmpl, contract, customData)
	require.NoError(t, err)
	assert.Contains(t, result, "## Contract")
	assert.Contains(t, result, "name: full-contract")
	assert.Contains(t, result, "## Tags")
	assert.Contains(t, result, "@critical")
}

func TestBuildPromptWithTemplate_ContractRawAccess(t *testing.T) {
	contract := &domain.Contract{
		Version:     "1.0",
		Name:        "raw-access",
		Description: "Testing raw access",
		RawData:     map[string]interface{}{},
	}

	tmpl := "Name: {{.ContractRaw.Name}}, Version: {{.ContractRaw.Version}}"
	result, err := BuildPromptWithTemplate(tmpl, contract, nil)
	require.NoError(t, err)
	assert.Equal(t, "Name: raw-access, Version: 1.0", result)
}

func TestBuildPromptWithTemplate_InvalidTemplateSyntax(t *testing.T) {
	tmpl := "{{ .Invalid..Syntax }}"
	_, err := BuildPromptWithTemplate(tmpl, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse prompt template")
}

func TestBuildPromptWithTemplate_UndefinedCustomField(t *testing.T) {
	// Accessing a key that does not exist in the Custom map returns empty string
	customData := map[string]string{
		"Exists": "yes",
	}

	tmpl := "Exists: {{.Custom.Exists}}, Missing: {{.Custom.Missing}}"
	result, err := BuildPromptWithTemplate(tmpl, nil, customData)
	require.NoError(t, err)
	assert.Contains(t, result, "Exists: yes")
	assert.Contains(t, result, "Missing: ")
}

func TestBuildPromptWithTemplate_EmptyTemplate(t *testing.T) {
	result, err := BuildPromptWithTemplate("", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestBuildPromptWithTemplate_NilCustomData(t *testing.T) {
	tmpl := "Static content only"
	result, err := BuildPromptWithTemplate(tmpl, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "Static content only", result)
}

func TestBuildPromptWithTemplate_ComplexContractRawData(t *testing.T) {
	contract := &domain.Contract{
		Version: "1.0",
		Name:    "complex",
		RawData: map[string]interface{}{
			"version": "1.0",
			"fields": []interface{}{
				map[string]interface{}{"name": "field1", "type": "string"},
				map[string]interface{}{"name": "field2", "type": "int"},
			},
		},
	}

	tmpl := "Contract:\n{{.Contract}}"
	result, err := BuildPromptWithTemplate(tmpl, contract, nil)
	require.NoError(t, err)
	assert.Contains(t, result, "Contract:")
	assert.Contains(t, result, "field1")
	assert.Contains(t, result, "field2")
}

func TestPromptData_StructFields(t *testing.T) {
	// Verify PromptData fields are accessible
	pd := PromptData{
		Contract:    "yaml-content",
		ContractRaw: &domain.Contract{Name: "test"},
		Custom:      map[string]string{"key": "value"},
	}
	assert.Equal(t, "yaml-content", pd.Contract)
	assert.Equal(t, "test", pd.ContractRaw.Name)
	assert.Equal(t, "value", pd.Custom["key"])
}
