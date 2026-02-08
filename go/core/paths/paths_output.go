// Package paths provides centralized path constants and utilities for the EAC repository.
package paths

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Path builder functions

// BuildOutputPath returns the path to a module's build output directory.
func BuildOutputPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, OutDir, BuildDir, moniker)
}

// ComponentBuildOutputPath returns the path to a component's build output directory.
// Structure: out/build/<module>/<component>.
func ComponentBuildOutputPath(repoRoot, module, component string) string {
	return filepath.Join(repoRoot, OutDir, BuildDir, module, component)
}

// TestOutputPath returns the path to a module's test output directory.
func TestOutputPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, OutDir, TestDir, moniker)
}

// ComponentTestOutputPath returns the path to a component's test output directory.
// Structure: out/test/<module>/<component>.
func ComponentTestOutputPath(repoRoot, module, component string) string {
	return filepath.Join(repoRoot, OutDir, TestDir, module, component)
}

// LintOutputPath returns the path to a module's lint output directory.
// On Windows, colons in monikers are replaced with underscores since colons
// are not valid in Windows file paths (except for drive letters).
func LintOutputPath(repoRoot, moniker string) string {
	safeName := sanitizeMonikerForPath(moniker)
	return filepath.Join(repoRoot, OutDir, LintDir, safeName)
}

// sanitizeMonikerForPath replaces characters that are invalid in file paths.
// On Windows, colons are not allowed in file/directory names.
func sanitizeMonikerForPath(moniker string) string {
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(moniker, ":", "_")
	}
	return moniker
}

// ComponentLintOutputPath returns the path to a component's lint output directory.
// Structure: out/lint/<module>/<component>.
func ComponentLintOutputPath(repoRoot, module, component string) string {
	return filepath.Join(repoRoot, OutDir, LintDir, module, component)
}

// TestOutputDir returns the root test output directory.
func TestOutputDir(repoRoot string) string {
	return filepath.Join(repoRoot, OutDir, TestDir)
}

// EvidenceOutputPath returns the path to a module's evidence output directory.
func EvidenceOutputPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, OutDir, EvidenceDir, moniker)
}

// AISummaryOutputPath returns the path to a module's AI summary output directory.
func AISummaryOutputPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, OutDir, AISummaryDir, moniker)
}

// StagingPath returns the path to a module's staging directory.
func StagingPath(repoRoot, name string) string {
	return filepath.Join(repoRoot, OutDir, StagingDir, name)
}

// SecurityOutputPath returns the path to security scan output.
func SecurityOutputPath(repoRoot, scanner string) string {
	return filepath.Join(repoRoot, OutDir, SecurityDir, scanner)
}

// ComponentScanOutputPath returns the path to a component's scan output directory.
// Structure: out/scan/<module>/<component>.
func ComponentScanOutputPath(repoRoot, module, component string) string {
	return filepath.Join(repoRoot, OutDir, SecurityDir, module, component)
}

// RiskProfilePath returns the path to a module's OSCAL profile.
func RiskProfilePath(repoRoot, moduleName string) string {
	if moduleName == "" {
		moduleName = "common"
	}
	return filepath.Join(repoRoot, SpecsDir, RiskControlsDir, moduleName+".profile.json")
}

// RiskAssessmentResultsPath returns the path to a module's OSCAL assessment-results with timestamp.
func RiskAssessmentResultsPath(repoRoot, moduleName string) string {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("assessment-results-%s.json", timestamp)
	return filepath.Join(repoRoot, OutDir, RiskDir, moduleName, filename)
}

// RiskOutputPath returns the path to a module's risk output directory.
func RiskOutputPath(repoRoot, moduleName string) string {
	return filepath.Join(repoRoot, OutDir, RiskDir, moduleName)
}

// TemplatePath returns the path to a template file or directory.
func TemplatePath(repoRoot string, subpaths ...string) string {
	parts := append([]string{repoRoot, TemplatesDir}, subpaths...)
	return filepath.Join(parts...)
}

// CommandLogsPath returns the path to a command's log directory
// Supports flexible path construction with optional path segments:
//
//	CommandLogsPath(root, "design") → out/design/
//	CommandLogsPath(root, "build", "core") → out/build/core/
//	CommandLogsPath(root, "templates", "apply") → out/templates/apply/
func CommandLogsPath(repoRoot, command string, pathSegments ...string) string {
	parts := make([]string, 0, 3+len(pathSegments))
	parts = append(parts, repoRoot, OutDir, command)
	parts = append(parts, pathSegments...)
	return filepath.Join(parts...)
}

// CommandsLogPath returns path to the unified commands log file
// Example: CommandsLogPath(root) -> <root>/out/commands.log.
func CommandsLogPath(repoRoot string) string {
	return filepath.Join(repoRoot, OutDir, "commands.log")
}

// ToolsPath returns the path to the tools output directory.
func ToolsPath(repoRoot string) string {
	return filepath.Join(repoRoot, OutDir, ToolsDir)
}

// BuildStatePath returns the path to the build state file.
func BuildStatePath(repoRoot, stateFileName string) string {
	return filepath.Join(repoRoot, OutDir, BuildDir, stateFileName)
}

// TestStatePath returns the path to the test state file.
func TestStatePath(repoRoot, stateFileName string) string {
	return filepath.Join(repoRoot, OutDir, TestDir, stateFileName)
}

// ============================================================================
// Additional Path Helpers for Common Patterns
// ============================================================================

// BuildOutputDir returns the root build output directory (out/build).
func BuildOutputDir(repoRoot string) string {
	return filepath.Join(repoRoot, OutDir, BuildDir)
}

// BuildLogPath returns the path to a module's build.log file.
func BuildLogPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, OutDir, BuildDir, moniker, "build.log")
}

// ComponentBuildLogPath returns the path to a component's build.log file.
// Structure: out/build/<module>/<component>/build.log.
func ComponentBuildLogPath(repoRoot, module, component string) string {
	return filepath.Join(repoRoot, OutDir, BuildDir, module, component, "build.log")
}

// BuildTimingPath returns the path to a module's build-timing.txt file.
func BuildTimingPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, OutDir, BuildDir, moniker, "build-timing.txt")
}

// ComponentBuildTimingPath returns the path to a component's build-timing.txt file.
// Structure: out/build/<module>/<component>/build-timing.txt.
func ComponentBuildTimingPath(repoRoot, module, component string) string {
	return filepath.Join(repoRoot, OutDir, BuildDir, module, component, "build-timing.txt")
}

// TestModuleDir returns the path to a module's test output directory
// Path: out/test/<module>.
func TestModuleDir(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, OutDir, TestDir, moniker)
}

// TestModuleTimingPath returns the path to a module's test-timing.txt file
// Path: out/test/<module>/test-timing.txt.
func TestModuleTimingPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, OutDir, TestDir, moniker, "test-timing.txt")
}

// RiskControlsPath returns the path to the risk controls directory.
func RiskControlsPath(repoRoot string) string {
	return filepath.Join(repoRoot, SpecsDir, RiskControlsDir)
}

// RiskCatalogPath returns the path to the risk catalog file.
func RiskCatalogPath(repoRoot string) string {
	return filepath.Join(repoRoot, TemplatesDir, "specs", "risk-catalog", "controls.catalog.json")
}

// TemplateSpecsPath returns the path to specs templates subdirectory.
func TemplateSpecsPath(repoRoot string, subpaths ...string) string {
	parts := append([]string{repoRoot, TemplatesDir, "specs"}, subpaths...)
	return filepath.Join(parts...)
}

// TemplateReportsPath returns the path to reports templates subdirectory.
func TemplateReportsPath(repoRoot string, subpaths ...string) string {
	parts := append([]string{repoRoot, TemplatesDir, "reports"}, subpaths...)
	return filepath.Join(parts...)
}

// SpecsFeaturePath returns the path to a feature specification file
// For module-scoped features: SpecsFeaturePath(root, moduleName, featureName)
// For top-level features: SpecsFeaturePath(root, "", featureName) or just pass featureName as first param.
func SpecsFeaturePath(repoRoot, moduleName, featureName string) string {
	if moduleName == "" {
		return filepath.Join(repoRoot, SpecsDir, featureName, "specification.feature")
	}
	return filepath.Join(repoRoot, SpecsDir, moduleName, featureName, "specification.feature")
}

// BuildLockPath returns the path to a build lock file.
func BuildLockPath(buildDir, moniker string) string {
	return filepath.Join(buildDir, fmt.Sprintf(".lock-%s", moniker))
}
