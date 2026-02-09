// Package paths provides centralized path constants and utilities for the EAC repository.
package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// CommandsBinaryPath returns the full path to the pre-built eac binary.
// This is THE canonical way to locate the commands binary for execution.
//
// When running in a container (CLIE_CONTAINER_ROOT is set), uses the container's
// internal path (e.g., /app/out/tools/eac) where the binary
// was pre-built during container image creation.
//
// When running locally, uses the repo root's tools output directory
// (e.g., out/tools/eac.exe on Windows).
//
// The commands binary is stored in out/tools/ (not out/build/) because it's a
// CI tool used to create builds, not a build output itself. This separation
// ensures the tool binary isn't confused with or overwritten by module build outputs.
//
// Path Configuration:
// The tools directory path is defined in .eac/repository.yml under paths.out.tools.
// The default value "out/tools" is also defined as the ToolsDir constant in this package.
// GitHub Actions and Dockerfile must use this same path - they cannot read the config
// dynamically, so if the path changes, all locations must be updated together:
//   - .eac/repository.yml (paths.out.tools)
//   - go/core/paths/paths.go (ToolsDir constant)
//   - .github/actions/setup-commands/action.yaml
//   - containers/eac-ext/Dockerfile
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
	binaryName := "eac"
	if runtime.GOOS == "windows" {
		binaryName = "eac.exe"
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
	return filepath.Join(repoRoot, EACDir)
}

// CLIEPath returns the path to the .clie directory.
func CLIEPath(repoRoot string) string {
	return filepath.Join(repoRoot, CLIEDir)
}

// ContractsVersionPath returns the path to a contracts version directory.
func ContractsVersionPath(repoRoot, module, version string) string {
	return filepath.Join(repoRoot, ContractsDir, module, version)
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
	return filepath.Join(repoRoot, EACDir, AIDir, command)
}

// AIConfigFile returns the path to a specific AI config file.
func AIConfigFile(repoRoot, command, filename string) string {
	return filepath.Join(AIConfigPath(repoRoot, command), filename)
}

// AITestMockPath returns the path to the AI test mock response file.
func AITestMockPath(repoRoot string) string {
	return filepath.Join(repoRoot, EACDir, "test", "ai-mock.txt")
}

// AIPromptsPath returns path to AI prompts (team override or system default)
// promptType: "team" for .eac/templates/ai, "system" for templates/ai.
func AIPromptsPath(repoRoot, promptType, command, filename string) string {
	if promptType == "team" {
		return filepath.Join(repoRoot, EACDir, TemplatesDir, AIDir, command, filename)
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
