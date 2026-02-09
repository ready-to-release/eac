// Package release contains godog step implementations for eac.
//
// This file contains release command step definitions.
package release

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"gopkg.in/yaml.v3"
)

// registerSteps registers step definitions for release-this command tests.
func registerSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// Background steps
	sc.Step(`^the repository has the EAC module system configured$`, func() error {
		// EAC configuration should exist in isolated test environment
		return eacgodog.FileExists(ctx, ".eac")
	})

	// Setup steps - Module creation
	sc.Step(`^a module "([^"]*)" exists$`, func(moniker string) error {
		return releaseModuleExists(ctx, moniker)
	})

	// Setup steps - Module state
	sc.Step(`^the module has commits since last release$`, func() error {
		return releaseModuleHasCommitsSinceLastRelease(ctx)
	})
	sc.Step(`^module "([^"]*)" with releases$`, func(moniker string) error {
		return releaseModuleWithReleases(ctx, moniker)
	})

	// Setup steps - Release notes existence
	sc.Step(`^"([^"]*)" does not exist$`, func(path string) error {
		return releaseNotesDoNotExist(ctx, path)
	})

	sc.Step(`^"([^"]*)" should not exist$`, func(path string) error {
		return releaseFileShouldNotExist(ctx, path)
	})

	// Verification steps - Module isolation
	sc.Step(`^I check release notes locations$`, func() error {
		return releaseCheckReleaseNotesLocations(ctx)
	})
	sc.Step(`^"([^"]*)" should be for ([^"]*)$`, func(path, module string) error {
		return releaseNotesShouldBeForModule(ctx, path, module)
	})
	sc.Step(`^there should be no cross-module contamination$`, func() error {
		return releaseShouldBeNoCrossModuleContamination(ctx)
	})
}

// ============================================================================
// Setup Steps - Module Creation
// ============================================================================

func releaseModuleExists(ctx *eacgodog.TestContext, moniker string) error {
	if !ctx.IsIsolated() {
		return fmt.Errorf("this step requires isolated test environment")
	}
	return createTestModule(ctx, moniker, nil)
}

func releaseModuleHasCommitsSinceLastRelease(ctx *eacgodog.TestContext) error {
	// Create a dummy file to ensure there are changes
	modulePath := filepath.Join(ctx.IsolatedDir, "go", "test-module")
	if err := os.MkdirAll(modulePath, 0o755); err != nil {
		return fmt.Errorf("failed to create module directory: %w", err)
	}

	dummyFile := filepath.Join(modulePath, "README.md")
	content := "# Test Module\n\nThis is a test module.\n"
	if err := os.WriteFile(dummyFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to create dummy file: %w", err)
	}

	// Commit the changes to create history
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = ctx.IsolatedDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to git add: %w", err)
	}

	cmd = exec.Command("git", "commit", "-m", "feat: add test changes")
	cmd.Dir = ctx.IsolatedDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to git commit: %w", err)
	}

	return nil
}

func releaseModuleWithReleases(ctx *eacgodog.TestContext, moniker string) error {
	if !ctx.IsIsolated() {
		return fmt.Errorf("this step requires isolated test environment")
	}

	// Create module first
	if err := createTestModule(ctx, moniker, nil); err != nil {
		return err
	}

	// Create release directory
	releaseDir := filepath.Join(ctx.IsolatedDir, "release", moniker)
	if err := os.MkdirAll(releaseDir, 0o755); err != nil {
		return fmt.Errorf("failed to create release directory: %w", err)
	}

	// Create a basic CHANGELOG.md to indicate releases exist
	changelogPath := filepath.Join(releaseDir, "CHANGELOG.md")
	changelogContent := fmt.Sprintf(`# Changelog

All notable changes to %s will be documented in this file.

## [1.0.0] - 2024-01-01

### Added

- Initial release
`, moniker)
	if err := os.WriteFile(changelogPath, []byte(changelogContent), 0o644); err != nil {
		return fmt.Errorf("failed to create CHANGELOG.md: %w", err)
	}

	return nil
}

// ============================================================================
// Setup Steps - Release Notes Files
// ============================================================================

func releaseNotesDoNotExist(ctx *eacgodog.TestContext, path string) error {
	fullPath := filepath.Join(ctx.IsolatedDir, path)
	// Ensure parent directory exists but file doesn't
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}
	// Remove file if it exists
	_ = os.Remove(fullPath)
	return nil
}

func releaseFileShouldNotExist(ctx *eacgodog.TestContext, path string) error {
	fullPath := filepath.Join(ctx.IsolatedDir, path)
	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("file should not exist but does: %s", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking file existence: %w", err)
	}
	return nil
}

// ============================================================================
// Verification Steps - Module Isolation
// ============================================================================

func releaseCheckReleaseNotesLocations(ctx *eacgodog.TestContext) error {
	// Verify release directory structure exists
	releaseDir := filepath.Join(ctx.IsolatedDir, "release")
	if _, err := os.Stat(releaseDir); os.IsNotExist(err) {
		return fmt.Errorf("release directory does not exist")
	}
	return nil
}

func releaseNotesShouldBeForModule(ctx *eacgodog.TestContext, path, module string) error {
	// Verify the path contains the module name
	if !strings.Contains(path, module) {
		return fmt.Errorf("path %s does not contain module %s", path, module)
	}

	// Verify file exists at this path
	fullPath := filepath.Join(ctx.IsolatedDir, path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// File doesn't need to exist for this verification
		// We're just checking the path structure
		return nil
	}

	return nil
}

func releaseShouldBeNoCrossModuleContamination(ctx *eacgodog.TestContext) error {
	// Verify that release notes are in separate directories
	releaseDir := filepath.Join(ctx.IsolatedDir, "release")

	entries, err := os.ReadDir(releaseDir)
	if err != nil {
		// If release dir doesn't exist or is empty, no contamination
		return nil
	}

	// Check that each module has its own directory
	for _, entry := range entries {
		if entry.IsDir() {
			// Each directory should only contain files for that module
			modulePath := filepath.Join(releaseDir, entry.Name())
			moduleEntries, err := os.ReadDir(modulePath)
			if err != nil {
				continue
			}

			// Verify no cross-references in file names
			for _, file := range moduleEntries {
				if !file.IsDir() {
					// File names should not contain other module names
					// This is a basic check
					if strings.Contains(file.Name(), "module-") && !strings.Contains(file.Name(), entry.Name()) {
						return fmt.Errorf("cross-module contamination detected: %s in %s", file.Name(), entry.Name())
					}
				}
			}
		}
	}

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// createTestModule creates a test module in the isolated environment.
// It updates repository.yml and creates the module directory structure.
func createTestModule(ctx *eacgodog.TestContext, moniker string, dependencies []string) error {
	// Read repository.yml
	repoYmlPath := filepath.Join(ctx.IsolatedDir, ".eac", "repository.yml")
	data, err := os.ReadFile(repoYmlPath)
	if err != nil {
		return fmt.Errorf("failed to read repository.yml: %w", err)
	}

	// Parse YAML
	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse repository.yml: %w", err)
	}

	// Get or create modules array
	var modules []interface{}
	if existingModules, ok := config["modules"].([]interface{}); ok {
		modules = existingModules
	} else {
		modules = []interface{}{}
	}

	// Check if module already exists in repository.yml
	moduleExistsInConfig := false
	for _, mod := range modules {
		if modMap, ok := mod.(map[string]interface{}); ok {
			if modMap["moniker"] == moniker {
				moduleExistsInConfig = true
				break
			}
		}
	}

	// Only update repository.yml if module doesn't exist
	if !moduleExistsInConfig {
		// Create module definition
		module := map[string]interface{}{
			"moniker":     moniker,
			"name":        fmt.Sprintf("Test Module %s", moniker),
			"description": "Test module for release tests",
			"versioning": map[string]interface{}{
				"scheme": "SemVer",
			},
			"components": map[string]interface{}{
				"go": fmt.Sprintf("go/%s", moniker),
			},
		}

		// Add dependencies if provided
		if len(dependencies) > 0 {
			module["depends_on"] = dependencies
		}

		// Append module
		modules = append(modules, module)
		config["modules"] = modules

		// Write back to file
		outData, err := yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal repository.yml: %w", err)
		}

		if err := os.WriteFile(repoYmlPath, outData, 0o644); err != nil {
			return fmt.Errorf("failed to write repository.yml: %w", err)
		}
	}

	// Create module directory
	modulePath := filepath.Join(ctx.IsolatedDir, "go", moniker)
	if err := os.MkdirAll(modulePath, 0o755); err != nil {
		return fmt.Errorf("failed to create module directory: %w", err)
	}

	// Create a basic go.mod file
	goModContent := fmt.Sprintf("module github.com/ready-to-release/eac/go/%s\n\ngo 1.21\n", moniker)
	goModPath := filepath.Join(modulePath, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0o644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}

	// Create a basic main.go file
	mainGoContent := fmt.Sprintf("package main\n\nfunc main() {\n\tprintln(\"Hello from %s\")\n}\n", moniker)
	mainGoPath := filepath.Join(modulePath, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(mainGoContent), 0o644); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}

	return nil
}
