package generation

import (
	"fmt"
	"path/filepath"

	"github.com/ready-to-release/eac/go/core/ai/config"
	configpkg "github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/validation"
	"github.com/ready-to-release/eac/go/core/validation/formats/gherkin"
	"github.com/ready-to-release/eac/go/core/validation/formats/oscal"
	"github.com/ready-to-release/eac/go/core/validation/formats/structurizr"
)

// loadValidatorForFormat loads the appropriate validator for a given format.
// Each format has a dedicated loader function to keep the factory maintainable.
func loadValidatorForFormat(templateRoot, typeName string, format StructuredFormat, tagsConfig *configpkg.TestingTagsConfig) (validation.Validator, error) {
	switch format {
	case FormatJSON:
		return loadJSONValidator(templateRoot, typeName)
	case FormatGherkin:
		return loadGherkinValidator(templateRoot, typeName, tagsConfig)
	case FormatOSCALCatalog:
		return loadOSCALValidator(oscal.TypeCatalog)
	case FormatOSCALProfile:
		return loadOSCALValidator(oscal.TypeProfile)
	case FormatStructurizr:
		return loadStructurizrValidator(templateRoot, typeName)
	case FormatPlainText:
		return &domain.NoOpValidator{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s (must be json, gherkin, oscal-catalog, oscal-profile, structurizr, or plaintext)", format)
	}
}

// loadJSONValidator loads a JSON schema validator for the given type.
func loadJSONValidator(templateRoot, typeName string) (validation.Validator, error) {
	schemaPath := filepath.Join(templateRoot, ContractSchemaPath(), typeName+".schema.json")
	validator, err := domain.NewJSONSchemaValidator(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load JSON schema for type '%s': %w", typeName, err)
	}
	return validator, nil
}

// loadGherkinValidator loads a Gherkin validator with tags configuration.
func loadGherkinValidator(templateRoot, typeName string, tagsConfig *configpkg.TestingTagsConfig) (validation.Validator, error) {
	loader := config.NewContractLoader(templateRoot, typeName, paths.DefaultsVersion)
	contractData, err := loader.LoadContract()
	if err != nil {
		return nil, fmt.Errorf("failed to load contract for Gherkin validator: %w", err)
	}
	return gherkin.NewValidator(contractData, tagsConfig), nil
}

// loadOSCALValidator creates an OSCAL validator for the given document type.
func loadOSCALValidator(oscalType oscal.DocumentType) (validation.Validator, error) {
	validator, err := oscal.NewValidator(oscalType)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSCAL validator: %w", err)
	}
	return validator, nil
}

// loadStructurizrValidator creates a Structurizr quick validator,
// optionally enhanced with contract data if available.
func loadStructurizrValidator(templateRoot, typeName string) (validation.Validator, error) {
	loader := config.NewContractLoader(templateRoot, typeName, paths.DefaultsVersion)
	contract, err := loader.LoadContract()
	if err == nil && contract != nil {
		return structurizr.NewQuickValidatorWithContract(contract), nil
	}
	return structurizr.NewQuickValidator(), nil
}
