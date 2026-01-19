// Package repository contains godog step implementations for specs/repository.
package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ============================================================================
// Release Folder Validation Steps
// ============================================================================

func (c *repositoryContext) scanReleaseDirectoryForSubdirectories() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	c.releaseSubdirs = []string{}
	c.orphanReleaseDirs = []string{}

	releaseDir := filepath.Join(c.repoRoot, "release")
	entries, err := os.ReadDir(releaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No release directory is fine
		}
		return fmt.Errorf("failed to read release directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			c.releaseSubdirs = append(c.releaseSubdirs, entry.Name())
		}
	}
	return nil
}

func (c *repositoryContext) everyReleaseSubdirShouldCorrespondToModule() error {
	if c.moduleReport == nil {
		return fmt.Errorf("module contracts not loaded")
	}

	for _, subdir := range c.releaseSubdirs {
		// Check if this subdirectory corresponds to a registered module
		if _, exists := c.moduleReport.Registry.Get(subdir); !exists {
			c.orphanReleaseDirs = append(c.orphanReleaseDirs, subdir)
		}
	}

	if len(c.orphanReleaseDirs) > 0 {
		return fmt.Errorf("found %d orphan release director(ies) with no corresponding module:\n  %s\n\nThese directories will cause 'release tag-pending --all' to fail in CI.\nEither create module contracts for these modules or delete the orphan directories.",
			len(c.orphanReleaseDirs), strings.Join(c.orphanReleaseDirs, "\n  "))
	}
	return nil
}

func (c *repositoryContext) ifOrphanReleaseDirsFoundShowNames() error {
	// Passive assertion - error in everyReleaseSubdirShouldCorrespondToModule provides details
	return nil
}

func (c *repositoryContext) scanReleaseDirectoryForChangelogFiles() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	c.releaseChangelogErrs = make(map[string][]string)
	c.unexpectedReleaseFiles = []string{}

	releaseDir := filepath.Join(c.repoRoot, "release")
	entries, err := os.ReadDir(releaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read release directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			// Files in release root (like README.md) are allowed
			continue
		}

		subdirPath := filepath.Join(releaseDir, entry.Name())
		subEntries, err := os.ReadDir(subdirPath)
		if err != nil {
			continue
		}

		hasChangelog := false
		for _, subEntry := range subEntries {
			name := subEntry.Name()
			if name == "CHANGELOG.md" {
				hasChangelog = true
			} else if name == "RELEASE-NOTES.md" {
				// RELEASE-NOTES.md is allowed
				continue
			} else if !subEntry.IsDir() {
				// Unexpected file in release subdirectory
				c.unexpectedReleaseFiles = append(c.unexpectedReleaseFiles,
					fmt.Sprintf("release/%s/%s", entry.Name(), name))
			}
		}

		if !hasChangelog {
			c.releaseChangelogErrs[entry.Name()] = append(c.releaseChangelogErrs[entry.Name()],
				"missing CHANGELOG.md")
		}
	}
	return nil
}

func (c *repositoryContext) everyChangelogShouldHaveValidFormat() error {
	if len(c.releaseChangelogErrs) > 0 {
		var details strings.Builder
		details.WriteString(fmt.Sprintf("Found %d release director(ies) with changelog issues:\n", len(c.releaseChangelogErrs)))
		for dir, errs := range c.releaseChangelogErrs {
			details.WriteString(fmt.Sprintf("  release/%s:\n", dir))
			for _, err := range errs {
				details.WriteString(fmt.Sprintf("    - %s\n", err))
			}
		}
		return fmt.Errorf("%s", details.String())
	}
	return nil
}

func (c *repositoryContext) noUnexpectedFilesInReleaseSubdirs() error {
	if len(c.unexpectedReleaseFiles) > 0 {
		return fmt.Errorf("found %d unexpected file(s) in release subdirectories:\n  %s\n\nOnly CHANGELOG.md and RELEASE-NOTES.md are expected in release/<module>/ directories.",
			len(c.unexpectedReleaseFiles), strings.Join(c.unexpectedReleaseFiles, "\n  "))
	}
	return nil
}

// Helper functions.
func indent(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}
