package generation

import (
	"fmt"
	"path/filepath"

	"github.com/ready-to-release/eac/go/core/paths"
)

// AI type names used for template and schema resolution.
const (
	TypeCommitMessage = "commit-message"
	TypeDesign        = "design"
	TypeSpecs         = "specs"
	TypeRiskProfile   = "risk-profile"
	TypeRiskAssess    = "risk-assess"
	TypeSquashMessage = "squash-message"
	TypeUpdateDesign  = "update-design"
)

// Path constants for template and contract resolution.
const (
	// AITemplatesPath is the base path for AI templates.
	AITemplatesPath = "templates/ai"
)

// ContractSchemaPath returns the relative path to contract schema files.
// This is a function (not a constant) to use the centralized paths.DefaultsVersion.
func ContractSchemaPath() string {
	return filepath.Join(paths.ContractsDir, paths.EACCoreModule, paths.DefaultsVersion)
}

// Retry prompt constants for consistent formatting.
const (
	// RetryPromptSeparator is the visual separator for retry sections.
	RetryPromptSeparator = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

	// RetryPromptWarning is the warning header for validation errors.
	RetryPromptWarning = "⚠️  VALIDATION ERRORS"

	// RetryPromptFocus is the focus header for key areas.
	RetryPromptFocus = "🎯 FOCUS AREAS"
)

// Phase 1 instructions by format type

// JSONPhaseInstruction is the standard instruction for Phase 1 JSON generation.
const JSONPhaseInstruction = `
IMPORTANT: Generate ONLY valid JSON matching the schema.
- No markdown fences (no ` + "```json" + `)
- No explanations or commentary before/after the JSON
- Just pure JSON starting with { and ending with }
- All string fields must use double quotes
- Use proper JSON escaping for special characters
`

// GherkinPhaseInstruction is the instruction for direct Gherkin generation.
const GherkinPhaseInstruction = `
IMPORTANT: Generate ONLY valid Gherkin matching the specification format.
- Use proper Gherkin syntax (Feature, Rule, Scenario)
- No markdown fences or code blocks
- No explanations or commentary before/after the Gherkin
- Start with Feature: declaration
- Follow proper indentation (2 spaces per level)
`

// OSCALPhaseInstruction is the instruction for OSCAL JSON generation.
const OSCALPhaseInstruction = `
IMPORTANT: Generate ONLY valid OSCAL JSON matching the OSCAL 1.1.2+ specification.
- No markdown fences (no ` + "```json" + `)
- Valid OSCAL profile or catalog structure
- Include required fields: uuid, metadata, imports
- Just pure JSON starting with { and ending with }
- Follow OSCAL schema requirements
`

// StructurizrPhaseInstruction is the instruction for Structurizr DSL generation.
const StructurizrPhaseInstruction = `
IMPORTANT: Generate ONLY valid Structurizr DSL syntax.
- No markdown fences or code blocks
- Valid workspace { } structure
- Include model { } and views { } sections
- No explanations or commentary
- Follow Structurizr DSL syntax requirements
`

// PlainTextPhaseInstruction is the instruction for plain text generation.
const PlainTextPhaseInstruction = `
IMPORTANT: Generate the final output directly.
- No markdown fences or formatting
- No explanations or meta-commentary
- Just the requested content
`

// GetPhaseInstruction returns the appropriate instruction for a format.
// Returns error if format is unsupported (NO DEFAULT FALLBACK).
func GetPhaseInstruction(format StructuredFormat) (string, error) {
	switch format {
	case FormatJSON:
		return JSONPhaseInstruction, nil
	case FormatGherkin:
		return GherkinPhaseInstruction, nil
	case FormatOSCALCatalog, FormatOSCALProfile:
		return OSCALPhaseInstruction, nil
	case FormatStructurizr:
		return StructurizrPhaseInstruction, nil
	case FormatPlainText:
		return PlainTextPhaseInstruction, nil
	default:
		return "", fmt.Errorf("unsupported format: %s (must be json, gherkin, oscal-catalog, oscal-profile, structurizr, or plaintext)", format)
	}
}
