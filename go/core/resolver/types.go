// Package resolver provides unified component-to-tool resolution for build, lint, and scan commands.
// It is the single source of truth for how components map to executable work units.
package resolver

import (
	"github.com/ready-to-release/eac/go/core/tool"
)

// Phase represents a command phase (build, lint, test, scan).
type Phase string

const (
	PhaseBuild Phase = "build"
	PhaseLint  Phase = "lint"
	PhaseTest  Phase = "test"
	PhaseScan  Phase = "scan"
)

// ScanCategory represents a scanner category.
type ScanCategory string

const (
	ScanCategorySBOM       ScanCategory = "sbom"
	ScanCategoryVuln       ScanCategory = "vuln"
	ScanCategorySecrets    ScanCategory = "secrets"
	ScanCategorySAST       ScanCategory = "sast"
	ScanCategoryIAC        ScanCategory = "iac"
	ScanCategoryCompliance ScanCategory = "compliance"
	ScanCategoryZAP        ScanCategory = "zap"
)

// AllScanCategories returns all valid scan categories.
func AllScanCategories() []ScanCategory {
	return []ScanCategory{
		ScanCategorySBOM,
		ScanCategoryVuln,
		ScanCategorySecrets,
		ScanCategorySAST,
		ScanCategoryIAC,
		ScanCategoryCompliance,
		ScanCategoryZAP,
	}
}

// DefaultScanCategories returns the default scan categories to run when none specified.
func DefaultScanCategories() []ScanCategory {
	return []ScanCategory{
		ScanCategorySBOM,
		ScanCategoryVuln,
		ScanCategorySecrets,
	}
}

// ToolAssignment represents a resolved tool for a component and phase.
type ToolAssignment struct {
	// ComponentName is the component name from the module (e.g., "go", "dockerfile")
	ComponentName string

	// ComponentType is the component type from component-types.yml
	ComponentType string

	// Phase is the operation phase (build, lint, test, scan)
	Phase Phase

	// ToolName is the resolved tool name
	ToolName string

	// ToolDef is the full tool definition (may be nil if handler is native)
	ToolDef *tool.ToolDefinition

	// Handler is the executable handler (for build operations)
	Handler tool.BuildHandler

	// IsContainer indicates if the tool runs in a container
	IsContainer bool

	// Weight is the scheduling weight for this assignment
	Weight int

	// DependsOn lists component names that must complete before this one
	DependsOn []string

	// ScanCategory is set only for scan phase to indicate the scanner category
	ScanCategory ScanCategory
}

// PhaseConfig holds configuration for a build/lint/test phase.
// This mirrors the YAML structure from component-types.yml.
type PhaseConfig struct {
	// Default is the default tool for this phase
	Default string `yaml:"default,omitempty" json:"default,omitempty"`

	// Options lists available tool alternatives
	Options []string `yaml:"options,omitempty" json:"options,omitempty"`

	// Command is tool-specific command override
	Command []string `yaml:"command,omitempty" json:"command,omitempty"`
}

// ScanPhaseConfig holds scan-specific configuration with category mappings.
// This mirrors the YAML structure from component-types.yml.
type ScanPhaseConfig struct {
	// Default lists default scanner categories to run
	Default []ScanCategory `yaml:"default,omitempty" json:"default,omitempty"`

	// CategoryTools maps scanner categories to specific tool names
	SBOM       string `yaml:"sbom,omitempty" json:"sbom,omitempty"`
	Vuln       string `yaml:"vuln,omitempty" json:"vuln,omitempty"`
	Secrets    string `yaml:"secrets,omitempty" json:"secrets,omitempty"`
	SAST       string `yaml:"sast,omitempty" json:"sast,omitempty"`
	IAC        string `yaml:"iac,omitempty" json:"iac,omitempty"`
	Compliance string `yaml:"compliance,omitempty" json:"compliance,omitempty"`
	ZAP        string `yaml:"zap,omitempty" json:"zap,omitempty"`
}

// GetToolForCategory returns the tool name for a given scanner category.
func (s *ScanPhaseConfig) GetToolForCategory(cat ScanCategory) string {
	if s == nil {
		return ""
	}
	switch cat {
	case ScanCategorySBOM:
		return s.SBOM
	case ScanCategoryVuln:
		return s.Vuln
	case ScanCategorySecrets:
		return s.Secrets
	case ScanCategorySAST:
		return s.SAST
	case ScanCategoryIAC:
		return s.IAC
	case ScanCategoryCompliance:
		return s.Compliance
	case ScanCategoryZAP:
		return s.ZAP
	default:
		return ""
	}
}

// GetDefaultCategories returns the default scan categories, or global defaults if not set.
func (s *ScanPhaseConfig) GetDefaultCategories() []ScanCategory {
	if s == nil || len(s.Default) == 0 {
		return DefaultScanCategories()
	}
	return s.Default
}

// ComponentTools holds phase-to-tool mappings for a component type.
// This represents the component-tools section in tool-config.yml.
type ComponentTools struct {
	// Build is the tool for build operations
	Build string `yaml:"build,omitempty" json:"build,omitempty"`

	// Lint is the tool for lint operations
	Lint string `yaml:"lint,omitempty" json:"lint,omitempty"`

	// Test is the tool for test operations
	Test string `yaml:"test,omitempty" json:"test,omitempty"`

	// Scan maps scanner categories to tools
	Scan *ScanToolMapping `yaml:"scan,omitempty" json:"scan,omitempty"`

	// Legacy fields (for backward compatibility)
	Builder  string   `yaml:"builder,omitempty" json:"builder,omitempty"`
	Linter   string   `yaml:"linter,omitempty" json:"linter,omitempty"`
	Tester   string   `yaml:"tester,omitempty" json:"tester,omitempty"`
	Scanners []string `yaml:"scanners,omitempty" json:"scanners,omitempty"`

	// Phase-specific argument overrides
	BuildArgs []string `yaml:"build_args,omitempty" json:"build_args,omitempty"`
	LintArgs  []string `yaml:"lint_args,omitempty" json:"lint_args,omitempty"`
}

// GetBuildTool returns the build tool, preferring new field over legacy.
func (c *ComponentTools) GetBuildTool() string {
	if c == nil {
		return ""
	}
	if c.Build != "" {
		return c.Build
	}
	return c.Builder
}

// GetLintTool returns the lint tool, preferring new field over legacy.
func (c *ComponentTools) GetLintTool() string {
	if c == nil {
		return ""
	}
	if c.Lint != "" {
		return c.Lint
	}
	return c.Linter
}

// GetTestTool returns the test tool, preferring new field over legacy.
func (c *ComponentTools) GetTestTool() string {
	if c == nil {
		return ""
	}
	if c.Test != "" {
		return c.Test
	}
	return c.Tester
}

// ScanToolMapping maps scanner categories to specific tool names.
type ScanToolMapping struct {
	SBOM       string `yaml:"sbom,omitempty" json:"sbom,omitempty"`
	Vuln       string `yaml:"vuln,omitempty" json:"vuln,omitempty"`
	Secrets    string `yaml:"secrets,omitempty" json:"secrets,omitempty"`
	SAST       string `yaml:"sast,omitempty" json:"sast,omitempty"`
	IAC        string `yaml:"iac,omitempty" json:"iac,omitempty"`
	Compliance string `yaml:"compliance,omitempty" json:"compliance,omitempty"`
	ZAP        string `yaml:"zap,omitempty" json:"zap,omitempty"`
}

// GetToolForCategory returns the tool name for a given scanner category.
func (m *ScanToolMapping) GetToolForCategory(cat ScanCategory) string {
	if m == nil {
		return ""
	}
	switch cat {
	case ScanCategorySBOM:
		return m.SBOM
	case ScanCategoryVuln:
		return m.Vuln
	case ScanCategorySecrets:
		return m.Secrets
	case ScanCategorySAST:
		return m.SAST
	case ScanCategoryIAC:
		return m.IAC
	case ScanCategoryCompliance:
		return m.Compliance
	case ScanCategoryZAP:
		return m.ZAP
	default:
		return ""
	}
}

// ExtendedComponentType extends ComponentType with tools mapping.
// This represents the component-types.yml structure with the new tools field.
type ExtendedComponentType struct {
	// Extensions are the file extensions for this component type
	Extensions []string `yaml:"extensions" json:"extensions"`

	// Builder is the legacy build handler field
	Builder string `yaml:"builder,omitempty" json:"builder,omitempty"`

	// Scanners is the legacy scanners field
	Scanners []string `yaml:"scanners,omitempty" json:"scanners,omitempty"`

	// BuildAfter lists component types that must complete first
	BuildAfter []string `yaml:"build_after,omitempty" json:"build_after,omitempty"`

	// Requirements are system dependencies needed
	Requirements []string `yaml:"requirements,omitempty" json:"requirements,omitempty"`

	// DefaultRoot is the default root path pattern
	DefaultRoot string `yaml:"default_root,omitempty" json:"default_root,omitempty"`

	// Tools is the new phase-specific tool configuration
	Tools *ToolsConfig `yaml:"tools,omitempty" json:"tools,omitempty"`
}

// ToolsConfig holds phase-specific tool configurations.
type ToolsConfig struct {
	Build *PhaseConfig     `yaml:"build,omitempty" json:"build,omitempty"`
	Lint  *PhaseConfig     `yaml:"lint,omitempty" json:"lint,omitempty"`
	Test  *PhaseConfig     `yaml:"test,omitempty" json:"test,omitempty"`
	Scan  *ScanPhaseConfig `yaml:"scan,omitempty" json:"scan,omitempty"`
}

// GetBuildTool returns the build tool from new or legacy fields.
func (e *ExtendedComponentType) GetBuildTool() string {
	if e == nil {
		return ""
	}
	// New field takes precedence
	if e.Tools != nil && e.Tools.Build != nil && e.Tools.Build.Default != "" {
		return e.Tools.Build.Default
	}
	// Fall back to legacy
	return e.Builder
}

// GetLintTool returns the lint tool from new fields.
func (e *ExtendedComponentType) GetLintTool() string {
	if e == nil || e.Tools == nil || e.Tools.Lint == nil {
		return ""
	}
	return e.Tools.Lint.Default
}

// GetTestTool returns the test tool from new fields.
func (e *ExtendedComponentType) GetTestTool() string {
	if e == nil || e.Tools == nil || e.Tools.Test == nil {
		return ""
	}
	return e.Tools.Test.Default
}

// GetScanCategories returns the default scan categories for this component type.
func (e *ExtendedComponentType) GetScanCategories() []ScanCategory {
	if e == nil {
		return nil
	}
	// New field takes precedence
	if e.Tools != nil && e.Tools.Scan != nil {
		return e.Tools.Scan.GetDefaultCategories()
	}
	// Fall back to legacy scanners field
	if len(e.Scanners) > 0 {
		cats := make([]ScanCategory, 0, len(e.Scanners))
		for _, s := range e.Scanners {
			cats = append(cats, ScanCategory(s))
		}
		return cats
	}
	return nil
}

// GetScannerTool returns the tool for a specific scanner category.
func (e *ExtendedComponentType) GetScannerTool(cat ScanCategory) string {
	if e == nil || e.Tools == nil || e.Tools.Scan == nil {
		return ""
	}
	return e.Tools.Scan.GetToolForCategory(cat)
}

// HasBuilder returns true if this component type has a builder configured.
func (e *ExtendedComponentType) HasBuilder() bool {
	return e.GetBuildTool() != ""
}

// IsScannable returns true if this component type has scanners configured.
func (e *ExtendedComponentType) IsScannable() bool {
	return len(e.GetScanCategories()) > 0
}
