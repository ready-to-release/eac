// Package paths provides centralized path constants and utilities for the EAC repository.
// This package has NO dependencies (only stdlib) to avoid import cycles.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

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

	// R2RDir is the configuration directory (.r2r).
	R2RDir = ".r2r"

	// EACDir is the EAC configuration subdirectory under R2RDir.
	EACDir = "eac"

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

	// EACCoreModule is the contract module name for EAC core.
	EACCoreModule = "eac-core"

	// DefaultsVersion is the current defaults version.
	DefaultsVersion = "0.1.0"

	// DefaultsDir is the defaults subdirectory under a contract version.
	DefaultsDir = "defaults"

	// AIDir is the AI configuration subdirectory.
	AIDir = "ai"

	// AIConfigFilename is the AI configuration filename.
	AIConfigFilename = "ai-config.yml"

	// ContainerRootEnv is the environment variable for container root path.
	ContainerRootEnv = "R2R_CONTAINER_ROOT"

	// ContainerRepoRoot is the standard path where the repository is mounted inside containers.
	// This is the canonical mount point for Docker containers running eac commands.
	ContainerRepoRoot = "/var/task"

	// EACCacheRoot is the root directory for all EAC build caches.
	// This is separate from out/ to clearly distinguish working cache from build outputs.
	// Structure:
	//   - .cache/eac/mkdocs/          - MkDocs build state and staging
	//   - .cache/eac/build/           - Build acceleration caches (hashes, state)
	//   - .cache/eac/pdf-screenshots/ - Extracted PDF page images
	EACCacheRoot = ".cache/eac"
)

// Relative path constants.
const (
	// EACConfigRelPath is the relative path from repo root to EAC configuration.
	EACConfigRelPath = R2RDir + "/" + EACDir

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

	// EACCommandsModule is the module name for the EAC commands binary.
	EACCommandsModule = "eac-commands"
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
func LintOutputPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, OutDir, LintDir, moniker)
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

// StagingPath returns the path to a module's staging directory.
func StagingPath(repoRoot, name string) string {
	return filepath.Join(repoRoot, OutDir, StagingDir, name)
}

// CachePath returns the path to the cache directory.
// DEPRECATED: Use CacheRootPath for new code. This now returns .cache/eac.
func CachePath(repoRoot string) string {
	return CacheRootPath(repoRoot)
}

// CacheRootPath returns the root EAC cache directory (.cache/eac).
// All transient build caches should be stored under this directory.
func CacheRootPath(repoRoot string) string {
	return filepath.Join(repoRoot, EACCacheRoot)
}

// BuildCachePath returns the path to the build acceleration cache directory.
// Used for file hashes and build state tracking.
// Path: .cache/eac/build/
func BuildCachePath(repoRoot string) string {
	return filepath.Join(CacheRootPath(repoRoot), "build")
}

// FileHashCachePath returns the path to a book's file hash cache.
// Used for incremental preprocessing to track which files have changed.
// Path: .cache/eac/build/hashes/{bookName}.json
func FileHashCachePath(repoRoot, bookName string) string {
	return filepath.Join(BuildCachePath(repoRoot), "hashes", bookName+".json")
}

// BuildStateCachePath returns the path to the build state cache directory.
// Used for tracking incremental build state (PDF/site content hashes).
// Path: .cache/eac/build/state/
func BuildStateCachePath(repoRoot string) string {
	return filepath.Join(BuildCachePath(repoRoot), "state")
}

// StagingCachePath returns the path to the staging cache directory.
// Used for persistent staging areas that survive across builds.
// Path: .cache/eac/staging/
func StagingCachePath(repoRoot string) string {
	return filepath.Join(CacheRootPath(repoRoot), "staging")
}

// BookStagingCachePath returns the path to a book's staging directory.
// The staging directory persists across builds for incremental preprocessing.
// Path: .cache/eac/staging/{moniker}/{bookName}
func BookStagingCachePath(repoRoot, moniker, bookName string) string {
	return filepath.Join(StagingCachePath(repoRoot), moniker, bookName)
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
//   - go/eac/core/paths/paths.go (ToolsDir constant)
//   - .github/actions/setup-commands/action.yaml
//   - containers/ext-eac/Dockerfile
//
// Usage:
//
//	binaryPath := paths.CommandsBinaryPath(repoRoot)
//	cmd := exec.Command(binaryPath, "show", "modules")
func CommandsBinaryPath(repoRoot string) string {
	return CommandsBinaryPathWithToolsDir(repoRoot, "")
}

// CommandsBinaryPathWithToolsDir returns the full path to the commands binary,
// allowing the tools directory to be specified explicitly.
// If toolsDir is empty, uses the default ToolsDir constant.
// This variant is useful when the caller has access to configuration.
func CommandsBinaryPathWithToolsDir(repoRoot, toolsDir string) string {
	binaryName := "commands"
	if runtime.GOOS == "windows" {
		binaryName = "commands.exe"
	}

	if toolsDir == "" {
		toolsDir = filepath.Join(OutDir, ToolsDir)
	}

	// Use distribution root for tool binaries (container root if running in container)
	// Note: Can't import repository package here to avoid cycles, so inline the check
	// See repository.GetDistRoot() for the canonical implementation
	distRoot := repoRoot
	if containerRoot := os.Getenv(ContainerRootEnv); containerRoot != "" {
		distRoot = containerRoot
	}

	return filepath.Join(distRoot, toolsDir, binaryName)
}

// CommandsBinaryExists checks if the commands binary exists at the expected path.
func CommandsBinaryExists(repoRoot string) (string, bool) {
	binaryPath := CommandsBinaryPath(repoRoot)
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, true
	}
	return "", false
}

// SpecsPath returns the path to a module's specs directory.
func SpecsPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, SpecsDir, moniker)
}

// DesignPath returns the path to a module's design workspace directory.
func DesignPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, SpecsDir, moniker, DesignDir)
}

// WorkspaceDSLPath returns the path to a module's workspace.dsl file.
func WorkspaceDSLPath(repoRoot, moniker string) string {
	return filepath.Join(repoRoot, SpecsDir, moniker, DesignDir, WorkspaceDSL)
}

// WorkspaceDSLFiles returns all validatable DSL files in a module's .design folder.
func WorkspaceDSLFiles(repoRoot, moniker string) ([]string, error) {
	designDir := filepath.Join(repoRoot, SpecsDir, moniker, DesignDir)

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

		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(name, ".dsl") {
			continue
		}

		if strings.HasPrefix(name, "_") {
			continue
		}

		fullPath := filepath.Join(designDir, name)

		if name == WorkspaceDSL {
			hasMainWorkspace = true
			files = append([]string{fullPath}, files...)
		} else {
			files = append(files, fullPath)
		}
	}

	_ = hasMainWorkspace

	return files, nil
}

// EACConfigPath returns the path to the EAC configuration directory.
func EACConfigPath(repoRoot string) string {
	return filepath.Join(repoRoot, R2RDir, EACDir)
}

// R2RPath returns the path to the .r2r directory.
func R2RPath(repoRoot string) string {
	return filepath.Join(repoRoot, R2RDir)
}

// ContractsVersionPath returns the path to a contracts version directory.
func ContractsVersionPath(repoRoot, module, version string) string {
	return filepath.Join(repoRoot, ContractsDir, module, version)
}

// LoggingDefaultsPath returns the path to the logging defaults file
// Container-aware: uses R2R_CONTAINER_ROOT when running in container.
func LoggingDefaultsPath(repoRoot string) string {
	root := defaultsRoot(repoRoot)
	return filepath.Join(root, ContractsDir, EACCoreModule, DefaultsVersion, DefaultsDir, "logging.yml")
}

// defaultsRoot returns the root for loading defaults (container-aware).
func defaultsRoot(repoRoot string) string {
	if containerRoot := os.Getenv(ContainerRootEnv); containerRoot != "" {
		return containerRoot
	}
	return repoRoot
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
//	CommandLogsPath(root, "build", "eac-core") → out/build/eac-core/
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

// StripSpecsPrefix removes the specs/ prefix from a path.
func StripSpecsPrefix(path string) string {
	path = strings.TrimPrefix(path, SpecsDir+"/")
	path = strings.TrimPrefix(path, SpecsDir+"\\")
	return path
}

// ExtractMonikerFromSpecsPath extracts the module moniker from a specs path.
func ExtractMonikerFromSpecsPath(specsPath string) string {
	relPath := StripSpecsPrefix(specsPath)
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

// AIConfigPath returns the path to AI configuration for a command.
func AIConfigPath(repoRoot, command string) string {
	return filepath.Join(repoRoot, R2RDir, EACDir, AIDir, command)
}

// AIConfigFile returns the path to a specific AI config file.
func AIConfigFile(repoRoot, command, filename string) string {
	return filepath.Join(AIConfigPath(repoRoot, command), filename)
}

// AITestMockPath returns the path to the AI test mock response file.
func AITestMockPath(repoRoot string) string {
	return filepath.Join(repoRoot, R2RDir, "test", "ai-mock.txt")
}

// AIPromptsPath returns path to AI prompts (team override or system default)
// promptType: "team" for .r2r/eac/templates/ai, "system" for templates/ai.
func AIPromptsPath(repoRoot, promptType, command, filename string) string {
	if promptType == "team" {
		return filepath.Join(repoRoot, R2RDir, EACDir, TemplatesDir, AIDir, command, filename)
	}
	return filepath.Join(repoRoot, TemplatesDir, AIDir, command, filename)
}

// EACConfigFilePath returns the path to the main EAC configuration file.
func EACConfigFilePath(repoRoot string) string {
	return filepath.Join(EACConfigPath(repoRoot), "ai-provider.yml")
}

// EACConfigPersonalFilePath returns the path to the personal EAC configuration file.
func EACConfigPersonalFilePath(repoRoot string) string {
	return filepath.Join(EACConfigPath(repoRoot), "ai-provider.personal.yml")
}

// EACLoggingConfigPath returns the path to the EAC logging configuration.
func EACLoggingConfigPath(repoRoot string) string {
	return filepath.Join(EACConfigPath(repoRoot), "logging.yml")
}

// ChangelogPath returns the path to a module's CHANGELOG.md file.
func ChangelogPath(repoRoot, module string) string {
	return filepath.Join(repoRoot, ReleaseDir, module, "CHANGELOG.md")
}

// ReleaseNotesPath returns the path to a module's RELEASE-NOTES.md file.
func ReleaseNotesPath(repoRoot, module string) string {
	return filepath.Join(repoRoot, ReleaseDir, module, "RELEASE-NOTES.md")
}

// WorkflowPath returns the path to a GitHub workflow file.
func WorkflowPath(repoRoot, workflow string) string {
	return filepath.Join(repoRoot, GitHubDir, WorkflowsDir, workflow)
}

// ============================================================================
// Common File Pattern Helpers
// ============================================================================

// GoModPath returns the path to a go.mod file in a module directory.
func GoModPath(moduleRoot string) string {
	return filepath.Join(moduleRoot, "go.mod")
}

// GitDir returns the path to the .git directory.
func GitDir(repoRoot string) string {
	return filepath.Join(repoRoot, ".git")
}

// GitConfigPath returns the path to the .git/config file.
func GitConfigPath(repoRoot string) string {
	return filepath.Join(repoRoot, ".git", "config")
}

// GitHeadPath returns the path to the .git/HEAD file.
func GitHeadPath(repoRoot string) string {
	return filepath.Join(repoRoot, ".git", "HEAD")
}

// IndexMarkdownPath returns the path to index.md in a directory.
func IndexMarkdownPath(dir string) string {
	return filepath.Join(dir, "index.md")
}

// NavigationConfigPath returns the path to .nav.yml navigation config.
func NavigationConfigPath(dir string) string {
	return filepath.Join(dir, ".nav.yml")
}

// ChecksumFilePath returns the path to checksums.txt in an output directory.
func ChecksumFilePath(outputDir string) string {
	return filepath.Join(outputDir, "checksums.txt")
}

// ============================================================================
// Container/Build Helpers
// ============================================================================

// ContainersPath returns the path to a container directory.
func ContainersPath(repoRoot, container string) string {
	return filepath.Join(repoRoot, "containers", container)
}

// ContainerDockerfilePath returns the path to a container's Dockerfile.
func ContainerDockerfilePath(repoRoot, container string) string {
	return filepath.Join(repoRoot, "containers", container, "Dockerfile")
}

// MkDocsConfigPath returns the path to mkdocs.yml in an output directory.
func MkDocsConfigPath(outputDir string) string {
	return filepath.Join(outputDir, "mkdocs.yml")
}

// MkDocsSiteTemplatePath returns the path to the mkdocs-site template.
func MkDocsSiteTemplatePath(repoRoot string) string {
	return filepath.Join(repoRoot, "containers", "mkdocs-site", "mkdocs.yml")
}

// MkDocsPdfTemplatePath returns the path to the mkdocs-pdf template.
func MkDocsPdfTemplatePath(repoRoot string) string {
	return filepath.Join(repoRoot, "containers", "mkdocs-pdf", "mkdocs.yml")
}

// SiteOutputPath returns the path to the site output directory.
func SiteOutputPath(outputDir string) string {
	return filepath.Join(outputDir, "site")
}

// ServeOutputPath returns the path to the serve output directory (for live docs serving).
func ServeOutputPath(repoRoot string) string {
	return filepath.Join(repoRoot, OutDir, "serve")
}

// BuildLockPath returns the path to a build lock file.
func BuildLockPath(buildDir, moniker string) string {
	return filepath.Join(buildDir, fmt.Sprintf(".lock-%s", moniker))
}

// ============================================================================
// Staging/Asset Helpers (for book building)
// ============================================================================

// StagingAssetsPath returns the path to assets directory within staging.
func StagingAssetsPath(stagingDir string) string {
	return filepath.Join(stagingDir, "assets")
}

// RenderedAssetsPath returns the path to rendered assets for a specific renderer
// Examples: renderer="mermaid" -> staging/assets/rendered/mermaid
//
//	renderer="drawio" -> staging/assets/rendered/drawio
func RenderedAssetsPath(stagingDir, renderer string) string {
	return filepath.Join(stagingDir, "assets", "rendered", renderer)
}

// DocsCachePath returns the path to the docs cache directory (git-tracked)
// This is used for assets that should be committed to the repository
// for CI optimization (e.g., pre-rendered mermaid diagrams).
func DocsCachePath(repoRoot string) string {
	return filepath.Join(repoRoot, DocsDir, AssetsDir, DocsCacheDir)
}

// StructurizrCachePath returns the path to the Structurizr cache directory in docs/assets/cache
// This is the git-tracked cache location for CI optimization (pre-rendered Structurizr diagrams).
func StructurizrCachePath(repoRoot string) string {
	return filepath.Join(DocsCachePath(repoRoot), "structurizr")
}

// StructurizrDocsCachePath returns the path to a cached Structurizr SVG in docs/assets/cache
// Filename format: {module}_{viewKey}_{dslHash}.svg
// The dslHash is the first 8 characters of the SHA256 hash of the workspace.dsl content
// This ensures all views from the same DSL file share the same hash for cache invalidation.
func StructurizrDocsCachePath(repoRoot, module, viewKey, dslHash string) string {
	filename := fmt.Sprintf("%s_%s_%s.svg", module, viewKey, dslHash)
	return filepath.Join(StructurizrCachePath(repoRoot), filename)
}

// PDFScreenshotsCachePath returns the path to the PDF screenshots cache directory.
// Located in .cache/eac/pdf-screenshots/ (not git-tracked).
func PDFScreenshotsCachePath(repoRoot string) string {
	return filepath.Join(CacheRootPath(repoRoot), "pdf-screenshots")
}

// PDFScreenshotsDirPath returns the path to a specific PDF's screenshot directory
// Each PDF gets its own directory named by the file's SHA256 hash (first 12 chars).
func PDFScreenshotsDirPath(repoRoot, hash string) string {
	return filepath.Join(PDFScreenshotsCachePath(repoRoot), hash)
}

// DocsSourcePath returns the path to docs directory within a source root.
func DocsSourcePath(sourceRoot string) string {
	return filepath.Join(sourceRoot, DocsDir)
}

// ============================================================================
// Specs/Test Asset Helpers
// ============================================================================

// SpecsImplPath returns the path to a spec implementation directory.
func SpecsImplPath(repoRoot, module string) string {
	return filepath.Join(repoRoot, GoDir, EACDir, "specs", "impl", module)
}

// SpecsAssetsPath returns the path to spec assets for a module.
func SpecsAssetsPath(repoRoot, module string) string {
	return filepath.Join(SpecsImplPath(repoRoot, module), "assets")
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

// ============================================================================
// Traceable Cache Path Helpers (V2 - human-readable filenames)
// ============================================================================

// SanitizeForCacheName converts a source path to a cache-friendly identifier.
// The result contains the last 2 meaningful path components for traceability.
//
// Examples:
//
//	"docs/assets/architecture/modules-overview.drawio.png" -> "architecture_modules-overview"
//	"docs/explanation/cd-model/overview.md" -> "cd-model_overview"
//	"docs/reference/eac/architecture/index.md" -> "eac_architecture_index"
//
// Rules:
//  1. Strip common prefixes (docs/assets/, docs/, assets/)
//  2. Remove file extensions (.drawio.png, .md, .png)
//  3. Take last 2 path components (parent_basename)
//  4. Convert path separators to underscores
//  5. Lowercase for consistency
//  6. Limit total length to 64 characters
func SanitizeForCacheName(sourcePath string) string {
	// Normalize path separators to forward slashes
	path := strings.ReplaceAll(sourcePath, "\\", "/")

	// Strip common prefixes
	for _, prefix := range []string{"docs/assets/", "docs/", "assets/"} {
		if strings.HasPrefix(path, prefix) {
			path = strings.TrimPrefix(path, prefix)
			break // Only strip one prefix
		}
	}

	// Remove extensions (order matters - check longest first)
	for _, ext := range []string{".drawio.png", ".drawio", ".md", ".png"} {
		if strings.HasSuffix(path, ext) {
			path = strings.TrimSuffix(path, ext)
			break // Only strip one extension
		}
	}

	// Split into path components
	parts := strings.Split(path, "/")

	// Take last 2 components (parent_basename)
	var result string
	if len(parts) >= 2 {
		result = parts[len(parts)-2] + "_" + parts[len(parts)-1]
	} else if len(parts) == 1 {
		// Single component - prefix with parent directory name from original path
		// If original was "docs/index.md", we want "docs_index"
		originalParts := strings.Split(strings.ReplaceAll(sourcePath, "\\", "/"), "/")
		if len(originalParts) >= 2 {
			parentIdx := len(originalParts) - 2
			parent := originalParts[parentIdx]
			// Remove extension from parent if it has one (shouldn't happen but be safe)
			result = parent + "_" + parts[0]
		} else {
			result = parts[0]
		}
	}

	// Lowercase for consistency
	result = strings.ToLower(result)

	// Limit length to 64 characters (truncate from the left to keep the end which is more specific)
	if len(result) > 64 {
		result = result[len(result)-64:]
	}

	return result
}

// CacheHashLength is the number of characters from a hash to include in cache filenames.
// 8 hex chars = 32 bits, enough to avoid collisions within a repository.
const CacheHashLength = 8

// truncateHash returns first n characters of hash, or full hash if shorter.
func truncateHash(hash string, n int) string {
	if len(hash) > n {
		return hash[:n]
	}
	return hash
}

// DrawioCachePath returns path with traceable filename for drawio cache.
// Format: {cacheRoot}/drawio/{identifier}_{hash8}.png
//
// Example:
//
//	DrawioCachePath("/cache", "docs/assets/arch/overview.drawio.png", "abc123def456")
//	-> "/cache/drawio/arch_overview_abc123de.png"
func DrawioCachePath(cacheRoot, sourcePath, contentHash string) string {
	identifier := SanitizeForCacheName(sourcePath)
	shortHash := truncateHash(contentHash, CacheHashLength)
	filename := fmt.Sprintf("%s_%s.png", identifier, shortHash)
	return filepath.Join(cacheRoot, "drawio", filename)
}

// MermaidCachePath returns path with traceable filename for mermaid cache.
// Format: {cacheRoot}/mermaid/{identifier}_{blockIndex}_{hash8}.svg
//
// Example:
//
//	MermaidCachePath("/cache", "docs/cd-model/overview.md", 0, "abc123def456")
//	-> "/cache/mermaid/cd-model_overview_0_abc123de.svg"
func MermaidCachePath(cacheRoot, sourcePath string, blockIndex int, contentHash string) string {
	identifier := SanitizeForCacheName(sourcePath)
	shortHash := truncateHash(contentHash, CacheHashLength)
	filename := fmt.Sprintf("%s_%d_%s.svg", identifier, blockIndex, shortHash)
	return filepath.Join(cacheRoot, "mermaid", filename)
}
