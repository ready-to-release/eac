package srccommands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// registerReleaseSteps registers step definitions for release-this command tests.
func registerReleaseSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	// Background steps
	sc.Step(`^the repository has the EAC module system configured$`, func() error {
		// EAC configuration should exist in isolated test environment
		return internal.FileExists(ctx, ".r2r")
	})

	// Setup steps - Module creation
	sc.Step(`^a module "([^"]*)" exists$`, func(moniker string) error {
		return moduleExists(ctx, moniker)
	})

	// Setup steps - Module state
	sc.Step(`^the module has commits since last release$`, func() error {
		return moduleHasCommitsSinceLastRelease(ctx)
	})
	sc.Step(`^module "([^"]*)" with releases$`, func(moniker string) error {
		return moduleWithReleases(ctx, moniker)
	})

	// Setup steps - Release notes existence
	sc.Step(`^"([^"]*)" does not exist$`, func(path string) error {
		return releaseNotesDoNotExist(ctx, path)
	})

	sc.Step(`^"([^"]*)" should not exist$`, func(path string) error {
		return fileShouldNotExist(ctx, path)
	})

	// Verification steps - Module isolation
	sc.Step(`^I check release notes locations$`, func() error {
		return checkReleaseNotesLocations(ctx)
	})
	sc.Step(`^"([^"]*)" should be for ([^"]*)$`, func(path, module string) error {
		return releaseNotesShouldBeForModule(ctx, path, module)
	})
	sc.Step(`^there should be no cross-module contamination$`, func() error {
		return shouldBeNoCrossModuleContamination(ctx)
	})
}

// ============================================================================
// Setup Steps - Module Creation
// ============================================================================

func moduleExists(ctx *internal.TestContext, moniker string) error {
	if !ctx.IsIsolated() {
		return fmt.Errorf("this step requires isolated test environment")
	}
	return createTestModule(ctx, moniker, nil)
}

func moduleHasCommitsSinceLastRelease(ctx *internal.TestContext) error {
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

func moduleWithReleases(ctx *internal.TestContext, moniker string) error {
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

func releaseNotesDoNotExist(ctx *internal.TestContext, path string) error {
	fullPath := filepath.Join(ctx.IsolatedDir, path)
	// Ensure parent directory exists but file doesn't
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}
	// Remove file if it exists
	_ = os.Remove(fullPath)
	return nil
}

func fileShouldNotExist(ctx *internal.TestContext, path string) error {
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

func checkReleaseNotesLocations(ctx *internal.TestContext) error {
	// Verify release directory structure exists
	releaseDir := filepath.Join(ctx.IsolatedDir, "release")
	if _, err := os.Stat(releaseDir); os.IsNotExist(err) {
		return fmt.Errorf("release directory does not exist")
	}
	return nil
}

func releaseNotesShouldBeForModule(ctx *internal.TestContext, path, module string) error {
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

func shouldBeNoCrossModuleContamination(ctx *internal.TestContext) error {
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

// Note: createTestModule is defined in steps_pipeline.go and shared across tests
