// Package paths provides centralized path constants and utilities for the EAC repository.
package paths

import (
	"path/filepath"
	"strings"
)

// PathConfig provides repository-specific path conventions.
// Used by config-aware path functions to respect customized directory layouts.
// All fields have sensible defaults matching the standard EAC layout.
type PathConfig struct {
	SpecsRoot  string
	Templates  string
	OutBuild   string
	OutTest    string
	OutLogs    string
	OutScan    string
	OutTools   string

	// Conventions
	BuildLog               string
	BuildTiming            string
	TestTiming             string
	Specification          string
	RiskCatalog            string
	RiskControlsDir        string
	TemplateSpecsDir       string
	TemplateReportsDir     string
	TemplateRiskCatalogDir string
	GodogTest              string
}

// DefaultPathConfig returns a PathConfig with all standard defaults.
// These values match the hardcoded constants used by the non-configurable
// path functions in this package.
func DefaultPathConfig() PathConfig {
	return PathConfig{
		SpecsRoot:              SpecsDir,
		Templates:              TemplatesDir,
		OutBuild:               OutBuildRelPath,
		OutTest:                OutTestRelPath,
		OutLogs:                OutLogsRelPath,
		OutScan:                OutSecurityRelPath,
		OutTools:               OutDir + "/" + ToolsDir,
		BuildLog:               "build.log",
		BuildTiming:            "build-timing.txt",
		TestTiming:             "test-timing.txt",
		Specification:          "specification.feature",
		RiskCatalog:            "controls.catalog.json",
		RiskControlsDir:        RiskControlsDir,
		TemplateSpecsDir:       "specs",
		TemplateReportsDir:     "reports",
		TemplateRiskCatalogDir: "risk-catalog",
		GodogTest:              "godog_test.go",
	}
}

// SanitizeMoniker converts a moniker to a filesystem-safe path component.
// Always replaces colons with underscores (colons are invalid on Windows
// and can cause issues with some tools on other platforms).
// This is the canonical sanitization function; use it when constructing paths
// from monikers that may contain colons (e.g., "docs:site" → "docs_site").
func SanitizeMoniker(moniker string) string {
	return strings.ReplaceAll(moniker, ":", "_")
}

// SpecsPathC returns the path to a module's specs directory using config values.
// If repoRoot is empty, returns a relative path.
func SpecsPathC(repoRoot, moniker string, pc PathConfig) string {
	if repoRoot == "" {
		return pc.SpecsRoot + "/" + moniker
	}
	return filepath.Join(repoRoot, pc.SpecsRoot, moniker)
}

// BuildOutputPathC returns the path to a module's build output using config values.
// If repoRoot is empty, returns a relative path.
func BuildOutputPathC(repoRoot, moniker string, pc PathConfig) string {
	if repoRoot == "" {
		return pc.OutBuild + "/" + moniker
	}
	return filepath.Join(repoRoot, pc.OutBuild, moniker)
}

// BuildLogPathC returns the path to a module's build log using config conventions.
// If repoRoot is empty, returns a relative path.
func BuildLogPathC(repoRoot, moniker string, pc PathConfig) string {
	if repoRoot == "" {
		return pc.OutBuild + "/" + moniker + "/" + pc.BuildLog
	}
	return filepath.Join(repoRoot, pc.OutBuild, moniker, pc.BuildLog)
}

// BuildTimingPathC returns the path to a module's build timing file.
// If repoRoot is empty, returns a relative path.
func BuildTimingPathC(repoRoot, moniker string, pc PathConfig) string {
	if repoRoot == "" {
		return pc.OutBuild + "/" + moniker + "/" + pc.BuildTiming
	}
	return filepath.Join(repoRoot, pc.OutBuild, moniker, pc.BuildTiming)
}

// TestModuleDirC returns the path to a module's test output directory.
// Sanitizes moniker for filesystem safety (replaces : with _).
// If repoRoot is empty, returns a relative path.
func TestModuleDirC(repoRoot, moniker string, pc PathConfig) string {
	safe := SanitizeMoniker(moniker)
	if repoRoot == "" {
		return pc.OutTest + "/" + safe
	}
	return filepath.Join(repoRoot, pc.OutTest, safe)
}

// TestModuleTimingPathC returns the path to a module's test timing file.
// Sanitizes moniker for filesystem safety (replaces : with _).
// If repoRoot is empty, returns a relative path.
func TestModuleTimingPathC(repoRoot, moniker string, pc PathConfig) string {
	safe := SanitizeMoniker(moniker)
	if repoRoot == "" {
		return pc.OutTest + "/" + safe + "/" + pc.TestTiming
	}
	return filepath.Join(repoRoot, pc.OutTest, safe, pc.TestTiming)
}

// ScanModuleOutputPathC returns the path to a module's scan output directory.
// If repoRoot is empty, returns a relative path.
func ScanModuleOutputPathC(repoRoot, moduleName string, pc PathConfig) string {
	if repoRoot == "" {
		return pc.OutScan + "/" + moduleName
	}
	return filepath.Join(repoRoot, pc.OutScan, moduleName)
}

// TemplatePathC returns the path to a template file using config values.
// If repoRoot is empty, returns a relative path.
func TemplatePathC(repoRoot string, pc PathConfig, segments ...string) string {
	parts := make([]string, 0, 2+len(segments))
	if repoRoot != "" {
		parts = append(parts, repoRoot)
	}
	parts = append(parts, pc.Templates)
	parts = append(parts, segments...)
	return filepath.Join(parts...)
}

// TemplateSpecsPathC returns the path to specs templates subdirectory.
func TemplateSpecsPathC(repoRoot string, pc PathConfig, subpaths ...string) string {
	parts := make([]string, 0, 1+len(subpaths))
	parts = append(parts, pc.TemplateSpecsDir)
	parts = append(parts, subpaths...)
	return TemplatePathC(repoRoot, pc, parts...)
}

// TemplateReportsPathC returns the path to reports templates subdirectory.
func TemplateReportsPathC(repoRoot string, pc PathConfig, subpaths ...string) string {
	parts := make([]string, 0, 1+len(subpaths))
	parts = append(parts, pc.TemplateReportsDir)
	parts = append(parts, subpaths...)
	return TemplatePathC(repoRoot, pc, parts...)
}

// RiskControlsPathC returns the path to the risk controls directory.
// If repoRoot is empty, returns a relative path.
func RiskControlsPathC(repoRoot string, pc PathConfig) string {
	if repoRoot == "" {
		return pc.SpecsRoot + "/" + pc.RiskControlsDir
	}
	return filepath.Join(repoRoot, pc.SpecsRoot, pc.RiskControlsDir)
}

// RiskCatalogPathC returns the path to the risk catalog file.
func RiskCatalogPathC(repoRoot string, pc PathConfig) string {
	return TemplatePathC(repoRoot, pc, pc.TemplateSpecsDir, pc.TemplateRiskCatalogDir, pc.RiskCatalog)
}

// SpecsFeaturePathC returns the path to a feature specification file.
// For module-scoped features, pass moduleName; for top-level features, pass empty string.
// If repoRoot is empty, returns a relative path.
func SpecsFeaturePathC(repoRoot, moduleName, featureName string, pc PathConfig) string {
	if moduleName == "" {
		if repoRoot == "" {
			return pc.SpecsRoot + "/" + featureName + "/" + pc.Specification
		}
		return filepath.Join(repoRoot, pc.SpecsRoot, featureName, pc.Specification)
	}
	if repoRoot == "" {
		return pc.SpecsRoot + "/" + moduleName + "/" + featureName + "/" + pc.Specification
	}
	return filepath.Join(repoRoot, pc.SpecsRoot, moduleName, featureName, pc.Specification)
}

// IsGodogTestFileC checks if a filename matches the godog test convention.
func IsGodogTestFileC(filename string, pc PathConfig) bool {
	return filename == pc.GodogTest
}

// GetPathVariablesC returns a map of path variables for template substitution.
func GetPathVariablesC(pc PathConfig) map[string]string {
	return map[string]string{
		"specs_root": pc.SpecsRoot,
		"templates":  pc.Templates,
		"out_build":  pc.OutBuild,
		"out_test":   pc.OutTest,
		"out_logs":   pc.OutLogs,
		"out_scan":   pc.OutScan,
		"out_tools":  pc.OutTools,
	}
}
