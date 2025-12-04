package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Repository directory conventions
// These are the standard directory names used throughout the repository.
const (
	// OutDir is the root output directory for all generated artifacts
	OutDir = "out"

	// BuildDir is the subdirectory under OutDir for build outputs
	BuildDir = "build"

	// TestDir is the subdirectory under OutDir for test outputs
	TestDir = "test"

	// LogsDir is the subdirectory under OutDir for log files
	LogsDir = "logs"

	// SecurityDir is the subdirectory under OutDir for security scan outputs
	SecurityDir = "security"

	// ToolsDir is the subdirectory under OutDir for CI tools (not build outputs)
	ToolsDir = "tools"
	// RiskDir is the subdirectory under OutDir for risk assessment outputs
	RiskDir = "risk"

	// RiskControlsDir is the subdirectory under SpecsDir for OSCAL profiles
	RiskControlsDir = ".risk-controls"

	// SpecsDir is the root directory for specifications (Gherkin, Structurizr)
	SpecsDir = "specs"

	// ContractsDir is the root directory for contract definitions
	ContractsDir = "contracts"

	// SrcDir is the root directory for source code
	SrcDir = "src"

	// GoDir is the root directory for Go source code
	GoDir = "go"

	// TemplatesDir is the root directory for templates
	TemplatesDir = "templates"

	// R2RDir is the configuration directory (.r2r)
	R2RDir = ".r2r"

	// EACDir is the EAC configuration subdirectory under R2RDir
	EACDir = "eac"

	// DesignDir is the design workspace subdirectory under a module's specs
	DesignDir = ".design"

	// WorkspaceDSL is the standard name for Structurizr workspace files
	WorkspaceDSL = "workspace.dsl"
)

// Conventional filenames
const (
	// GodogTestFile is the conventional name for godog test files
	GodogTestFile = "godog_test.go"

	// PackageJSONFile is the conventional name for npm package files
	PackageJSONFile = "package.json"
)

// Relative path constants (combinations of directory names)
const (
	// EACConfigRelPath is the relative path from repo root to EAC configuration
	EACConfigRelPath = R2RDir + "/" + EACDir

	// OutBuildRelPath is the relative path from repo root to build output
	OutBuildRelPath = OutDir + "/" + BuildDir

	// OutTestRelPath is the relative path from repo root to test output
	OutTestRelPath = OutDir + "/" + TestDir

	// OutLogsRelPath is the relative path from repo root to logs output
	OutLogsRelPath = OutDir + "/" + LogsDir

	// OutSecurityRelPath is the relative path from repo root to security output
	OutSecurityRelPath = OutDir + "/" + SecurityDir

	// EACCommandsModule is the module name for the EAC commands binary
	EACCommandsModule = "eac-commands"
)

// Path builder functions for common paths

// BuildOutputPath returns the path to a module's build output directory
// Example: out/build/r2r-cli
func BuildOutputPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, OutDir, BuildDir, moniker)
}

// CommandsBinaryPath returns the full path to the pre-built eac-commands binary.
// This is THE canonical way to locate the commands binary for execution.
//
// When running in a container (R2R_CONTAINER_ROOT is set), uses the container's
// internal path (e.g., /app/out/tools/commands) where the binary
// was pre-built during container image creation.
//
// When running locally, uses the repo root's tools output directory
// (e.g., out/tools/commands.exe on Windows).
//
// The commands binary is stored in out/tools/ (not out/build/) because it's a
// CI tool used to create builds, not a build output itself. This separation
// ensures the tool binary isn't confused with or overwritten by module build outputs.
//
// Path Configuration:
// The tools directory path is defined in .r2r/eac/repository.yml under paths.out.tools.
// The default value "out/tools" is also defined as the ToolsDir constant in this package.
// GitHub Actions and Dockerfile must use this same path - they cannot read the config
// dynamically, so if the path changes, all locations must be updated together:
//   - .r2r/eac/repository.yml (paths.out.tools)
//   - go/eac/core/repository/paths.go (ToolsDir constant)
//   - .github/actions/setup-commands/action.yaml
//   - containers/ext-eac/Dockerfile
//
// Usage:
//
//	binaryPath := repository.CommandsBinaryPath(repoRoot)
//	cmd := exec.Command(binaryPath, "show", "modules")
func CommandsBinaryPath(repoRoot string) string {
	return CommandsBinaryPathWithToolsDir(repoRoot, "")
}

// CommandsBinaryPathWithToolsDir returns the full path to the commands binary,
// allowing the tools directory to be specified explicitly.
// If toolsDir is empty, uses the default ToolsDir constant.
// This variant is useful when the caller has access to configuration.
//
// If a .new version of the binary exists (staged by a build), this function
// performs an atomic replacement before returning the path:
// 1. Rename current binary to .old
// 2. Rename .new to the target name
// 3. Remove .old
// This lazy update ensures the binary is always fresh after a build.
func CommandsBinaryPathWithToolsDir(repoRoot string, toolsDir string) string {
	binaryName := "commands"
	if runtime.GOOS == "windows" {
		binaryName = "commands.exe"
	}

	if toolsDir == "" {
		toolsDir = filepath.Join(OutDir, ToolsDir)
	}

	// Use container root if running in container, otherwise use repo root
	effectiveRoot := GetEffectiveRoot(repoRoot)
	binaryPath := filepath.Join(effectiveRoot, toolsDir, binaryName)

	// Check for staged .new binary and perform atomic replacement
	newPath := binaryPath + ".new"
	if _, err := os.Stat(newPath); err == nil {
		oldPath := binaryPath + ".old"

		// Atomic replacement: current -> .old, .new -> current, remove .old
		os.Remove(oldPath)            // Clean up any stale .old
		os.Rename(binaryPath, oldPath) // Move current to .old (may fail if doesn't exist)
		if err := os.Rename(newPath, binaryPath); err == nil {
			os.Remove(oldPath) // Cleanup .old on success
		} else {
			// Restore on failure
			os.Rename(oldPath, binaryPath)
		}
	}

	return binaryPath
}

// CommandsBinaryExists checks if the commands binary exists at the expected path.
// Returns the path and true if it exists, empty string and false otherwise.
func CommandsBinaryExists(repoRoot string) (string, bool) {
	binaryPath := CommandsBinaryPath(repoRoot)
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, true
	}
	return "", false
}

// SpecsPath returns the path to a module's specs directory
// Example: specs/r2r-cli
func SpecsPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, SpecsDir, moniker)
}

// DesignPath returns the path to a module's design workspace directory
// Example: specs/r2r-cli/.design
func DesignPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, SpecsDir, moniker, DesignDir)
}

// WorkspaceDSLPath returns the path to a module's workspace.dsl file
// Example: specs/r2r-cli/.design/workspace.dsl
func WorkspaceDSLPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, SpecsDir, moniker, DesignDir, WorkspaceDSL)
}

// WorkspaceDSLFiles returns all validatable DSL files in a module's .design folder.
// Files starting with "_" are considered fragments (for !include) and are excluded.
// Returns paths sorted with workspace.dsl first if present.
// Example: specs/r2r-cli/.design/*.dsl (excluding _*.dsl)
func WorkspaceDSLFiles(repoRoot, moniker string) ([]string, error) {
	designDir := filepath.Join(repoRoot, SpecsDir, moniker, DesignDir)

	// Check if design directory exists
	if _, err := os.Stat(designDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(designDir)
	if err != nil {
		return nil, err
	}

	var files []string
	var hasMainWorkspace bool

	for _, entry := range entries {
		name := entry.Name()

		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Skip files not ending in .dsl
		if !strings.HasSuffix(name, ".dsl") {
			continue
		}

		// Skip fragment files (underscore prefix - used for !include)
		if strings.HasPrefix(name, "_") {
			continue
		}

		fullPath := filepath.Join(designDir, name)

		// Track if we have the main workspace.dsl
		if name == WorkspaceDSL {
			hasMainWorkspace = true
			// Prepend workspace.dsl so it's validated first
			files = append([]string{fullPath}, files...)
		} else {
			files = append(files, fullPath)
		}
	}

	// If no main workspace.dsl but other files exist, that's fine
	_ = hasMainWorkspace

	return files, nil
}

// EACConfigPath returns the path to the EAC configuration directory
// Example: .r2r/eac
func EACConfigPath(repoRoot string) string {
	return filepath.Join(repoRoot, R2RDir, EACDir)
}

// R2RPath returns the path to the .r2r directory
// Example: .r2r
func R2RPath(repoRoot string) string {
	return filepath.Join(repoRoot, R2RDir)
}

// ContractsVersionPath returns the path to a contracts version directory
// Example: contracts/eac-core/0.1.0
func ContractsVersionPath(repoRoot, module, version string) string {
	return filepath.Join(repoRoot, ContractsDir, module, version)
}

// LogsPath returns the path to the logs output directory
// Example: out/logs
func LogsPath(repoRoot string) string {
	return filepath.Join(repoRoot, OutDir, LogsDir)
}

// TestOutputPath returns the path to a test suite's output directory
// Example: out/test/acceptance
func TestOutputPath(repoRoot, suiteName string) string {
	return filepath.Join(repoRoot, OutDir, TestDir, suiteName)
}

// SecurityOutputPath returns the path to security scan output
// Example: out/security/trivy
func SecurityOutputPath(repoRoot, scanner string) string {
	return filepath.Join(repoRoot, OutDir, SecurityDir, scanner)
}

// RiskProfilePath returns the path to a module's OSCAL profile
// Example: specs/risk-controls/billing.profile.json
func RiskProfilePath(repoRoot, moduleName string) string {
	if moduleName == "" {
		moduleName = "common"
	}
	return filepath.Join(repoRoot, SpecsDir, RiskControlsDir, moduleName+".profile.json")
}

// RiskAssessmentResultsPath returns the path to a module's OSCAL assessment-results with timestamp
// Example: out/risk/billing/assessment-results-20251204-143025.json
func RiskAssessmentResultsPath(repoRoot, moduleName string) string {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("assessment-results-%s.json", timestamp)
	return filepath.Join(repoRoot, OutDir, RiskDir, moduleName, filename)
}

// RiskOutputPath returns the path to a module's risk output directory
// Example: out/risk/billing
func RiskOutputPath(repoRoot, moduleName string) string {
	return filepath.Join(repoRoot, OutDir, RiskDir, moduleName)
}

// TemplatePath returns the path to a template file or directory
// Example: templates/test-reports/suite-summary.md
func TemplatePath(repoRoot string, subpaths ...string) string {
	parts := append([]string{repoRoot, TemplatesDir}, subpaths...)
	return filepath.Join(parts...)
}

// CommandLogsPath returns the path to a command's log directory
// Example: out/logs/commit
func CommandLogsPath(repoRoot, command string) string {
	return filepath.Join(repoRoot, OutDir, LogsDir, command)
}

// StripSpecsPrefix removes the specs/ prefix from a path (handles both / and \)
func StripSpecsPrefix(path string) string {
	path = strings.TrimPrefix(path, SpecsDir+"/")
	path = strings.TrimPrefix(path, SpecsDir+"\\")
	return path
}

// IsGodogTestDir checks if a directory contains a godog_test.go file
func IsGodogTestDir(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, GodogTestFile))
	return err == nil
}

// ExtractMonikerFromSpecsPath extracts the module moniker from a specs path
// Input:  specs/eac-commands/templates/specification.feature
// Output: eac-commands
func ExtractMonikerFromSpecsPath(specsPath string) string {
	relPath := StripSpecsPrefix(specsPath)
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return ""
}
