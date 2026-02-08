// Package paths provides centralized path constants and utilities for the EAC repository.
// This package has NO dependencies (only stdlib) to avoid import cycles.
package paths

// Repository directory conventions.
const (
	// OutDir is the root output directory for all generated artifacts.
	OutDir = "out"

	// DocsDir is the root directory for documentation.
	DocsDir = "docs"

	// AssetsDir is the assets subdirectory under docs.
	AssetsDir = "assets"

	// DocsCacheDir is the cache subdirectory under docs/assets (git-tracked for CI optimization).
	DocsCacheDir = "cache"

	// BuildDir is the subdirectory under OutDir for build outputs.
	BuildDir = "build"

	// TestDir is the subdirectory under OutDir for test outputs.
	TestDir = "test"

	// LogsDir is the subdirectory under OutDir for log files.
	LogsDir = "logs"

	// SecurityDir is the subdirectory under OutDir for scan outputs.
	SecurityDir = "scan"

	// LintDir is the subdirectory under OutDir for lint outputs.
	LintDir = "lint"

	// AISummaryDir is the subdirectory under OutDir for AI summary outputs.
	AISummaryDir = "ai-summary"

	// RiskDir is the subdirectory under OutDir for risk assessment outputs.
	RiskDir = "risk"

	// EvidenceDir is the subdirectory under OutDir for evidence outputs.
	EvidenceDir = "evidence"

	// RiskControlsDir is the subdirectory under SpecsDir for OSCAL profiles.
	RiskControlsDir = ".risk-controls"

	// ToolsDir is the subdirectory under OutDir for CI tools (not build outputs).
	ToolsDir = "tools"

	// StagingDir is the subdirectory under OutDir for build staging areas.
	StagingDir = "staging"

	// SpecsDir is the root directory for specifications (Gherkin, Structurizr).
	SpecsDir = "specs"

	// ContractsDir is the root directory for contract definitions.
	ContractsDir = "contracts"

	// SrcDir is the root directory for source code.
	SrcDir = "src"

	// GoDir is the root directory for Go source code.
	GoDir = "go"

	// TemplatesDir is the root directory for templates.
	TemplatesDir = "templates"

	// CLIEDir is the clie CLI configuration directory (.clie).
	CLIEDir = ".clie"

	// EACDir is the eac configuration directory (.eac) - sibling to CLIEDir.
	EACDir = ".eac"

	// DesignDir is the design workspace subdirectory under a module's specs.
	DesignDir = ".design"

	// WorkspaceDSL is the standard name for Structurizr workspace files.
	WorkspaceDSL = "workspace.dsl"

	// ReleaseDir is the directory for release metadata and changelogs.
	ReleaseDir = "release"

	// GitHubDir is the GitHub configuration directory.
	GitHubDir = ".github"

	// WorkflowsDir is the workflows subdirectory under GitHubDir.
	WorkflowsDir = "workflows"

	// EACCoreModule is the contract module name for core.
	EACCoreModule = "core"

	// DefaultsVersion is the current defaults version.
	DefaultsVersion = "0.1.0"

	// DefaultsDir is the defaults subdirectory under a contract version.
	DefaultsDir = "defaults"

	// SchemasDir is the standard subdirectory for schemas and defaults in contract packages.
	SchemasDir = "schemas"

	// AIDir is the AI configuration subdirectory.
	AIDir = "ai"

	// AIConfigFilename is the AI configuration filename.
	AIConfigFilename = "ai-config.yml"

	// ContainerRootEnv is the environment variable for container root path.
	ContainerRootEnv = "CLIE_CONTAINER_ROOT"

	// ContainerRepoRoot is the standard path where the repository is mounted inside containers.
	// This is the canonical mount point for Docker containers running eac commands.
	ContainerRepoRoot = "/var/task"

	// EACCacheRoot is the root directory for all EAC caches.
	// This is separate from out/ to clearly distinguish cache state from build outputs.
	// Deleting .cache/eac/ is guaranteed safe — all ephemeral caches are here.
	// Git-tracked assets in docs/assets/cache/ are NOT under this root.
	//
	// Structure:
	//   - .cache/eac/build/           - Build acceleration caches (hashes, state)
	//   - .cache/eac/staging/         - Persistent doc staging areas
	//   - .cache/eac/structurizr/     - Structurizr acceleration cache
	//   - .cache/eac/drawio/          - DrawIO acceleration cache
	//   - .cache/eac/mermaid/         - Mermaid acceleration cache
	//   - .cache/eac/pdf-screenshots/ - Extracted PDF page images
	//   - .cache/eac/incremental/     - UoW incremental state (build/test/lint/scan)
	//   - .cache/eac/semaphores/      - Cross-process capacity coordination
	//   - .cache/eac/npm/             - NPM isolation work dirs and download cache
	//   - .cache/eac/preprocess/      - Book preprocessing state
	EACCacheRoot = ".cache/eac"
)

// Relative path constants.
const (
	// EACConfigRelPath is the relative path from repo root to EAC configuration.
	EACConfigRelPath = EACDir

	// OutBuildRelPath is the relative path from repo root to build output.
	OutBuildRelPath = OutDir + "/" + BuildDir

	// OutTestRelPath is the relative path from repo root to test output.
	OutTestRelPath = OutDir + "/" + TestDir

	// OutLogsRelPath is the relative path from repo root to logs output.
	OutLogsRelPath = OutDir + "/" + LogsDir

	// OutSecurityRelPath is the relative path from repo root to scan output.
	OutSecurityRelPath = OutDir + "/" + SecurityDir

	// OutLintRelPath is the relative path from repo root to lint output.
	OutLintRelPath = OutDir + "/" + LintDir

	// OutStagingRelPath is the relative path from repo root to staging area.
	OutStagingRelPath = OutDir + "/" + StagingDir

	// OutAISummaryRelPath is the relative path from repo root to AI summary output.
	OutAISummaryRelPath = OutDir + "/" + AISummaryDir

	// EACCommandsModule is the module name for the EAC commands binary.
	EACCommandsModule = "eac-cli"
)
