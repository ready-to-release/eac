// Package generation provides AI-related utilities including format configuration
package generation

import (
	"github.com/ready-to-release/eac/go/eac/core/validation"
)

// StructuredFormat defines supported Phase 1 output formats
type StructuredFormat string

const (
	// FormatJSON indicates JSON with schema validation
	FormatJSON StructuredFormat = "json"

	// FormatGherkin indicates Gherkin .feature syntax
	FormatGherkin StructuredFormat = "gherkin"

	// FormatOSCALCatalog indicates OSCAL catalog JSON format
	FormatOSCALCatalog StructuredFormat = "oscal-catalog"

	// FormatOSCALProfile indicates OSCAL profile JSON format
	FormatOSCALProfile StructuredFormat = "oscal-profile"

	// FormatStructurizr indicates Structurizr DSL format
	FormatStructurizr StructuredFormat = "structurizr"

	// FormatPlainText indicates plain text (for simple outputs)
	FormatPlainText StructuredFormat = "plaintext"
)

// PhaseConfig defines generation behavior for a specific phase
type PhaseConfig struct {
	// Format is the expected output format
	Format StructuredFormat

	// Instruction is the format-specific AI instruction
	Instruction string

	// Validator is the format-specific validator
	Validator validation.Validator
}
