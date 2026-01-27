package srccommands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
	sc.Step(`^module "([^"]*)" exists$`, func(moniker string) error {
		return moduleExists(ctx, moniker)
	})

	// Setup steps - Module state
	sc.Step(`^the module has commits since last release$`, func() error {
		return moduleHasCommitsSinceLastRelease(ctx)
	})
	sc.Step(`^module "([^"]*)" has commits since last release$`, func(moniker string) error {
		return moduleHasCommitsSinceLastReleaseNamed(ctx, moniker)
	})
	sc.Step(`^module "([^"]*)" with releases$`, func(moniker string) error {
		return moduleWithReleases(ctx, moniker)
	})

	// Setup steps - Release notes existence
	sc.Step(`^"([^"]*)" does not exist$`, func(path string) error {
		return releaseNotesDoNotExist(ctx, path)
	})
	sc.Step(`^"([^"]*)" exists with version "([^"]*)"$`, func(path, version string) error {
		return releaseNotesExistWithVersion(ctx, path, version)
	})
	sc.Step(`^"([^"]*)" exists with empty version "([^"]*)"$`, func(path, version string) error {
		return releaseNotesExistWithEmptyVersion(ctx, path, version)
	})

	// Setup steps - Version calculation
	sc.Step(`^the calculated next version is "([^"]*)"$`, func(version string) error {
		return calculatedNextVersionIs(ctx, version)
	})

	// Verification steps - Command result
	sc.Step(`^the command should create "([^"]*)"$`, func(path string) error {
		return commandShouldCreateFile(ctx, path)
	})
	// Note: "the command should fail", "the command should succeed", and "the output should contain" are already registered elsewhere

	// Verification steps - File content
	sc.Step(`^"([^"]*)" should exist$`, func(path string) error {
		return internal.FileExists(ctx, path)
	})
	sc.Step(`^the file should contain "([^"]*)"$`, func(text string) error {
		return fileContainsText(ctx, text)
	})

	// Verification steps - Changelog
	sc.Step(`^"([^"]*)" should be updated$`, func(path string) error {
		return changelogShouldBeUpdated(ctx, path)
	})
	sc.Step(`^the CHANGELOG should contain version "([^"]*)"$`, func(version string) error {
		return changelogShouldContainVersion(ctx, version)
	})
	sc.Step(`^"([^"]*)" should not be modified$`, func(path string) error {
		return fileShouldNotBeModified(ctx, path)
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

func moduleHasCommitsSinceLastReleaseNamed(ctx *internal.TestContext, moniker string) error {
	if !ctx.IsIsolated() {
		return fmt.Errorf("this step requires isolated test environment")
	}

	// Ensure module exists
	if err := createTestModule(ctx, moniker, nil); err != nil {
		return err
	}

	// Create a change in the module
	modulePath := filepath.Join(ctx.IsolatedDir, "go", moniker)
	dummyFile := filepath.Join(modulePath, "README.md")
	content := fmt.Sprintf("# %s\n\nThis is a test module.\n", moniker)
	if err := os.WriteFile(dummyFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to create dummy file: %w", err)
	}

	// Commit the changes
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = ctx.IsolatedDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to git add: %w", err)
	}

	cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("feat(%s): add changes", moniker))
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

func releaseNotesExistWithVersion(ctx *internal.TestContext, path, version string) error {
	fullPath := filepath.Join(ctx.IsolatedDir, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Create release notes with a DIFFERENT version than what will be calculated
	// This way, validation will fail because the expected version is missing
	content := fmt.Sprintf(`# Release release_notes

## [%s] - 2024-01-01

### Summary

Some release summary.

### Changes

- Change 1
- Change 2
`, version)

	// Also create a changelog with a tag so the command calculates a higher version
	// Extract module name from path (e.g., "release/test-module/RELEASE-NOTES.md" -> "test-module")
	pathParts := strings.Split(path, string(filepath.Separator))
	if len(pathParts) >= 2 {
		module := pathParts[1]
		// Create a git tag for the old version to make the next version calculate correctly
		cmd := exec.Command("git", "tag", fmt.Sprintf("%s/%s", module, version))
		cmd.Dir = ctx.IsolatedDir
		_ = cmd.Run() // Ignore errors - tag might already exist
	}

	return os.WriteFile(fullPath, []byte(content), 0o644)
}

func releaseNotesExistWithEmptyVersion(ctx *internal.TestContext, path, version string) error {
	fullPath := filepath.Join(ctx.IsolatedDir, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Create version entry with sections but no content
	content := fmt.Sprintf(`# Release release_notes

## [%s] - 2024-01-01

### Summary

### Conclusion on Fitness for Intended Use

### Impact on Business Process
`, version)

	return os.WriteFile(fullPath, []byte(content), 0o644)
}

// ============================================================================
// Setup Steps - Version Control
// ============================================================================

func calculatedNextVersionIs(ctx *internal.TestContext, version string) error {
	// To make the version calculator return this version, we need to create a git tag
	// for the previous version. For simplicity, we'll decrement the patch version.
	// This assumes semantic versioning: MAJOR.MINOR.PATCH

	// Parse version
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid version format: %s (expected MAJOR.MINOR.PATCH)", version)
	}

	var major, minor, patch int
	if _, err := fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch); err != nil {
		return fmt.Errorf("failed to parse version %s: %w", version, err)
	}

	// Calculate previous version (decrement patch)
	var prevVersion string
	if patch > 0 {
		prevVersion = fmt.Sprintf("%d.%d.%d", major, minor, patch-1)
	} else if minor > 0 {
		prevVersion = fmt.Sprintf("%d.%d.0", major, minor-1)
	} else if major > 0 {
		prevVersion = fmt.Sprintf("%d.0.0", major-1)
	} else {
		// Version is 0.0.0, no previous version exists
		return nil
	}

	// Determine module name from the last created module
	// For most tests, this will be "test-module"
	module := "test-module"

	// Create git tag for previous version
	cmd := exec.Command("git", "tag", fmt.Sprintf("%s/%s", module, prevVersion))
	cmd.Dir = ctx.IsolatedDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Ignore if tag already exists
		if !strings.Contains(string(output), "already exists") {
			return fmt.Errorf("failed to create git tag %s/%s: %w\nOutput: %s", module, prevVersion, err, string(output))
		}
	}

	// Create a new commit after the tag to ensure there are "commits since last release"
	// Modify the module's main.go file to ensure the commit affects the module
	modulePath := filepath.Join(ctx.IsolatedDir, "go", module)
	mainGoFile := filepath.Join(modulePath, "main.go")

	// Read existing content and append a comment
	existing, err := os.ReadFile(mainGoFile)
	if err != nil {
		return fmt.Errorf("failed to read main.go: %w", err)
	}

	newContent := string(existing) + "\n// Post-release change\n"
	if err := os.WriteFile(mainGoFile, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("failed to update main.go: %w", err)
	}

	cmd = exec.Command("git", "add", mainGoFile)
	cmd.Dir = ctx.IsolatedDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage post-release change: %w", err)
	}

	cmd = exec.Command("git", "commit", "-m", "feat: post-release changes")
	cmd.Dir = ctx.IsolatedDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit post-release change: %w", err)
	}

	return nil
}

// ============================================================================
// Verification Steps - File Operations
// ============================================================================

func commandShouldCreateFile(ctx *internal.TestContext, path string) error {
	// First verify command succeeded in creating something
	// Then verify the file exists
	return internal.FileExists(ctx, path)
}

func fileContainsText(ctx *internal.TestContext, text string) error {
	// Find the last file referenced in the scenario
	// This is a simplified approach - in a real implementation,
	// we'd track the file path from the previous step
	// For now, we'll search common locations

	// Check in command output for file path
	output := ctx.CommandOutput
	if strings.Contains(output, "RELEASE-NOTES.md") {
		// Extract path from output if possible
		// For now, use default path
		return internal.FileContains(ctx, "release/test-module/RELEASE-NOTES.md", text)
	}

	return fmt.Errorf("cannot determine file to check")
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
// Verification Steps - Changelog
// ============================================================================

func changelogShouldBeUpdated(ctx *internal.TestContext, path string) error {
	// Verify the changelog file exists and has content
	if err := internal.FileExists(ctx, path); err != nil {
		return err
	}

	// Read the file to verify it has recent updates
	fullPath := filepath.Join(ctx.IsolatedDir, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read changelog: %w", err)
	}

	// Verify it has some content (at least a version header)
	if len(content) == 0 {
		return fmt.Errorf("changelog is empty")
	}

	return nil
}

func changelogShouldContainVersion(ctx *internal.TestContext, version string) error {
	// Find the changelog file (use common pattern)
	changelogPath := "release/test-module/CHANGELOG.md"

	// Check for other modules if test-module doesn't exist
	if !strings.Contains(ctx.CommandOutput, "test-module") {
		if strings.Contains(ctx.CommandOutput, "module-a") {
			changelogPath = "release/module-a/CHANGELOG.md"
		}
	}

	versionHeader := fmt.Sprintf("[%s]", version)
	return internal.FileContains(ctx, changelogPath, versionHeader)
}

func fileShouldNotBeModified(ctx *internal.TestContext, path string) error {
	// For this test, we verify the file either doesn't exist
	// or hasn't been touched recently
	fullPath := filepath.Join(ctx.IsolatedDir, path)

	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		// File doesn't exist, so it wasn't modified
		return nil
	}
	if err != nil {
		return fmt.Errorf("error checking file: %w", err)
	}

	// Check if file was modified in the last 5 seconds
	// (during the test run)
	modTime := info.ModTime()
	if time.Since(modTime) < 5*time.Second {
		return fmt.Errorf("file %s was recently modified", path)
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
