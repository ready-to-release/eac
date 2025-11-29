package repository

import "path/filepath"

// Repository directory conventions
// These are the standard directory names used throughout the repository.
const (
	// OutDir is the root output directory for all generated artifacts
	OutDir = "out"

	// BuildDir is the subdirectory under OutDir for build outputs
	BuildDir = "build"

	// LogsDir is the subdirectory under OutDir for log files
	LogsDir = "logs"

	// SpecsDir is the root directory for specifications (Gherkin, Structurizr)
	SpecsDir = "specs"

	// ContractsDir is the root directory for contract definitions
	ContractsDir = "contracts"

	// SrcDir is the root directory for source code
	SrcDir = "src"

	// R2RDir is the configuration directory (.r2r)
	R2RDir = ".r2r"

	// EACDir is the EAC configuration subdirectory under R2RDir
	EACDir = "eac"

	// RepositoryConfigDir is the repository configuration subdirectory under EACDir
	RepositoryConfigDir = "repository"

	// DesignDir is the design workspace subdirectory under a module's specs
	DesignDir = ".design"

	// WorkspaceDSL is the standard name for Structurizr workspace files
	WorkspaceDSL = "workspace.dsl"
)

// Relative path constants (combinations of directory names)
const (
	// EACConfigRelPath is the relative path from repo root to EAC repository configuration
	EACConfigRelPath = R2RDir + "/" + EACDir + "/" + RepositoryConfigDir

	// OutBuildRelPath is the relative path from repo root to build output
	OutBuildRelPath = OutDir + "/" + BuildDir

	// OutLogsRelPath is the relative path from repo root to logs output
	OutLogsRelPath = OutDir + "/" + LogsDir
)

// Path builder functions for common paths

// BuildOutputPath returns the path to a module's build output directory
// Example: out/build/src-cli
func BuildOutputPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, OutDir, BuildDir, moniker)
}

// SpecsPath returns the path to a module's specs directory
// Example: specs/src-cli
func SpecsPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, SpecsDir, moniker)
}

// DesignPath returns the path to a module's design workspace directory
// Example: specs/src-cli/.design
func DesignPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, SpecsDir, moniker, DesignDir)
}

// WorkspaceDSLPath returns the path to a module's workspace.dsl file
// Example: specs/src-cli/.design/workspace.dsl
func WorkspaceDSLPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, SpecsDir, moniker, DesignDir, WorkspaceDSL)
}

// EACConfigPath returns the path to the EAC repository configuration directory
// Example: .r2r/eac/repository
func EACConfigPath(repoRoot string) string {
	return filepath.Join(repoRoot, R2RDir, EACDir, RepositoryConfigDir)
}

// R2RPath returns the path to the .r2r directory
// Example: .r2r
func R2RPath(repoRoot string) string {
	return filepath.Join(repoRoot, R2RDir)
}

// ContractsVersionPath returns the path to a contracts version directory
// Example: contracts/src-core/0.1.0
func ContractsVersionPath(repoRoot, module, version string) string {
	return filepath.Join(repoRoot, ContractsDir, module, version)
}

// LogsPath returns the path to the logs output directory
// Example: out/logs
func LogsPath(repoRoot string) string {
	return filepath.Join(repoRoot, OutDir, LogsDir)
}
