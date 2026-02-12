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

	// ComponentType is the component type from blueprints.yml component-kinds
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
// This mirrors the YAML structure from blueprints.yml component-kinds.
type PhaseConfig struct {
	// Default is the default tool for this phase
	Default string `yaml:"default,omitempty" json:"default,omitempty"`

	// Options lists available tool alternatives
	Options []string `yaml:"options,omitempty" json:"options,omitempty"`

	// Command is tool-specific command override
	Command []string `yaml:"command,omitempty" json:"command,omitempty"`
}

// ScanPhaseConfig holds scan-specific configuration with category mappings.
// This mirrors the YAML structure from blueprints.yml component-kinds.
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

	// Phase-specific argument overrides
	BuildArgs []string `yaml:"build_args,omitempty" json:"build_args,omitempty"`
	LintArgs  []string `yaml:"lint_args,omitempty" json:"lint_args,omitempty"`
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
type ExtendedComponentType struct {
	// Extensions are the file extensions for this component type
	Extensions []string `yaml:"extensions" json:"extensions"`

	// Builders are the build tool IDs
	Builders []string `yaml:"builders,omitempty" json:"builders,omitempty"`

	// Linters are the lint tool IDs
	Linters []string `yaml:"linters,omitempty" json:"linters,omitempty"`

	// Testers are the test tool IDs
	Testers []string `yaml:"testers,omitempty" json:"testers,omitempty"`

	// Scanners are scanner categories
	Scanners []string `yaml:"scanners,omitempty" json:"scanners,omitempty"`

	// BuildAfter lists component types that must complete first
	BuildAfter []string `yaml:"build_after,omitempty" json:"build_after,omitempty"`

	// DefaultRoot is the default root path pattern
	DefaultRoot string `yaml:"default_root,omitempty" json:"default_root,omitempty"`

	// Tools is the phase-specific tool configuration (alternative to arrays above)
	Tools *ToolsConfig `yaml:"tools,omitempty" json:"tools,omitempty"`
}

// ToolsConfig holds phase-specific tool configurations.
type ToolsConfig struct {
	Build *PhaseConfig     `yaml:"build,omitempty" json:"build,omitempty"`
	Lint  *PhaseConfig     `yaml:"lint,omitempty" json:"lint,omitempty"`
	Test  *PhaseConfig     `yaml:"test,omitempty" json:"test,omitempty"`
	Scan  *ScanPhaseConfig `yaml:"scan,omitempty" json:"scan,omitempty"`
}

// GetBuilders returns the builder tool IDs.
func (e *ExtendedComponentType) GetBuilders() []string {
	if e == nil {
		return nil
	}
	if len(e.Builders) > 0 {
		return e.Builders
	}
	if e.Tools != nil && e.Tools.Build != nil && e.Tools.Build.Default != "" {
		return []string{e.Tools.Build.Default}
	}
	return nil
}

// GetLinters returns the linter tool IDs.
func (e *ExtendedComponentType) GetLinters() []string {
	if e == nil {
		return nil
	}
	if len(e.Linters) > 0 {
		return e.Linters
	}
	if e.Tools != nil && e.Tools.Lint != nil && e.Tools.Lint.Default != "" {
		return []string{e.Tools.Lint.Default}
	}
	return nil
}

// GetTesters returns the tester tool IDs.
func (e *ExtendedComponentType) GetTesters() []string {
	if e == nil {
		return nil
	}
	if len(e.Testers) > 0 {
		return e.Testers
	}
	if e.Tools != nil && e.Tools.Test != nil && e.Tools.Test.Default != "" {
		return []string{e.Tools.Test.Default}
	}
	return nil
}

// GetScanCategories returns the scan categories for this component type.
func (e *ExtendedComponentType) GetScanCategories() []ScanCategory {
	if e == nil {
		return nil
	}
	if e.Tools != nil && e.Tools.Scan != nil {
		return e.Tools.Scan.GetDefaultCategories()
	}
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

// IsBuildable returns true if this component type has builders configured.
func (e *ExtendedComponentType) IsBuildable() bool {
	return len(e.GetBuilders()) > 0
}

// IsLintable returns true if this component type has linters configured.
func (e *ExtendedComponentType) IsLintable() bool {
	return len(e.GetLinters()) > 0
}

// IsTestable returns true if this component type has testers configured.
func (e *ExtendedComponentType) IsTestable() bool {
	return len(e.GetTesters()) > 0
}

// IsScannable returns true if this component type has scanners configured.
func (e *ExtendedComponentType) IsScannable() bool {
	return len(e.GetScanCategories()) > 0
}
