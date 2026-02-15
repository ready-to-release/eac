// go_version.go - Version injection and changelog parsing for Go builds.
//
// Builds ldflags for injecting version strings into Go binaries via -X flags.
// Reads version from CHANGELOG.md or uses dev placeholder for local builds.
package builders

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/environments"
)

// buildLdflags builds ldflags for version injection.
func buildLdflags(module *modules.ModuleContract, moduleRoot, workspaceRoot, explicitVersion string, logWriter io.Writer) string {
	ldflags := ""

	// Detect if running in CI (for version detection)
	isCI := environments.IsCI()

	// Determine version to inject
	version := explicitVersion
	if version == "" {
		if isCI {
			// CI: Auto-detect from changelog for release builds
			version = getVersionFromChangelog(moduleRoot, workspaceRoot, module.Moniker)
		} else {
			// Local dev: Use high version number to always be "newer" than releases
			version = "666.666.666-local"
		}
	}

	// Inject version if available
	if version != "" {
		// Get module import path for correct ldflags
		modulePath := getGoModulePath(moduleRoot)
		if modulePath != "" {
			// Use module/cmd.Version pattern (standard for CLI tools)
			versionFlag := fmt.Sprintf("-X %s/cmd.Version=%s", modulePath, version)
			ldflags = versionFlag
			Logln(logWriter, "Injecting version: %s", version)
		}
	}

	return ldflags
}

// getGoModulePath reads the module path from go.mod file.
func getGoModulePath(moduleRoot string) string {
	goModPath := filepath.Join(moduleRoot, "go.mod")
	f, err := os.Open(goModPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module ")
		}
	}
	return ""
}

// getVersionFromChangelog attempts to extract version from CHANGELOG.md
// Checks both module root and release/{moniker}/ directories.
func getVersionFromChangelog(moduleRoot, workspaceRoot, moniker string) string {
	// Try multiple locations for changelog
	// NOTE: config.RepositoryConfig.ReleaseModulePathAbs() is now available but this
	// function doesn't have access to config. Uses the same conventional "release/" path.
	paths := []string{
		filepath.Join(moduleRoot, "CHANGELOG.md"),
		filepath.Join(workspaceRoot, "release", moniker, "CHANGELOG.md"),
	}

	for _, changelogPath := range paths {
		if version := extractVersionFromFile(changelogPath); version != "" {
			return version
		}
	}
	return ""
}

// extractVersionFromFile extracts the first version from a changelog file.
func extractVersionFromFile(changelogPath string) string {
	f, err := os.Open(changelogPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Look for version headers like "## [0.0.17]" or "## 0.0.17"
		if strings.HasPrefix(line, "## ") {
			version := strings.TrimPrefix(line, "## ")
			// Remove brackets if present: [0.0.17] -> 0.0.17
			version = strings.TrimPrefix(version, "[")
			version = strings.Split(version, "]")[0]
			// Skip [Unreleased]
			if version != "Unreleased" && version != "" {
				return version
			}
		}
	}
	return ""
}
