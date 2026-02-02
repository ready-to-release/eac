// Package repository contains godog step implementations for specs/repository.
package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// ============================================================================
// Script Location Validation Steps
// ============================================================================

func (c *repositoryContext) theFollowingScriptExtensionsAreTracked(table *godog.Table) error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	c.scriptExtensions = []string{}
	c.discoveredScripts = []string{}
	c.disallowedScripts = []string{}
	c.looseScriptsInType = []string{}

	// Parse table (skip header row)
	for i, row := range table.Rows {
		if i == 0 {
			continue // Skip header
		}
		if len(row.Cells) >= 1 {
			c.scriptExtensions = append(c.scriptExtensions, row.Cells[0].Value)
		}
	}
	return nil
}

func (c *repositoryContext) scanRepositoryForScriptFiles() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	// Use cached git ls-files instead of filepath.Walk (55s -> <1s)
	if err := c.sharedCtx.EnsureOriginalRepoCache(); err != nil {
		return err
	}

	// Get all files matching any script extension from cache
	c.discoveredScripts = c.sharedCtx.OriginalRepoCache.FilesMatchingAnyExtension(c.scriptExtensions)
	return nil
}

func (c *repositoryContext) allScriptsShouldBeInApprovedLocations(table *godog.Table) error {
	// Approved patterns from the table
	approvedPatterns := []string{
		".claude/hooks/",
		"scripts/pwsh/",
		"scripts/sh/",
		"scripts/cmd/",
	}
	approvedRootFiles := []string{
		"importer.sh",
		"importer.ps1",
	}

	for _, script := range c.discoveredScripts {
		approved := false

		// Check root-level importers
		for _, rootFile := range approvedRootFiles {
			if script == rootFile {
				approved = true
				break
			}
		}

		// Check approved directory patterns
		if !approved {
			for _, pattern := range approvedPatterns {
				if strings.HasPrefix(script, pattern) {
					approved = true
					break
				}
			}
		}

		// Allow container entrypoint scripts (standard Docker pattern)
		if !approved && strings.HasPrefix(script, "containers/") && strings.HasSuffix(script, "/entrypoint.sh") {
			approved = true
		}

		if !approved {
			c.disallowedScripts = append(c.disallowedScripts, script)
		}
	}

	if len(c.disallowedScripts) > 0 {
		return fmt.Errorf("found %d script(s) in unapproved locations:\n  %s",
			len(c.disallowedScripts), strings.Join(c.disallowedScripts, "\n  "))
	}
	return nil
}

func (c *repositoryContext) noScriptsShouldExistIn(table *godog.Table) error {
	// Parse disallowed locations from table (skip header)
	disallowedPrefixes := []string{}
	for i, row := range table.Rows {
		if i == 0 {
			continue
		}
		if len(row.Cells) >= 1 {
			disallowedPrefixes = append(disallowedPrefixes, row.Cells[0].Value)
		}
	}

	var violations []string
	for _, script := range c.discoveredScripts {
		for _, prefix := range disallowedPrefixes {
			if strings.HasPrefix(script, prefix) {
				violations = append(violations, script)
				break
			}
		}
	}

	if len(violations) > 0 {
		return fmt.Errorf("found %d script(s) in disallowed locations:\n  %s",
			len(violations), strings.Join(violations, "\n  "))
	}
	return nil
}

func (c *repositoryContext) theseLocationsAreExcludedFromValidation(table *godog.Table) error {
	// This is informational - excluded locations are already skipped during scanning
	return nil
}

func (c *repositoryContext) scanScriptsDirectoryStructure() error {
	// Already scanned in scanRepositoryForScriptFiles
	return nil
}

func (c *repositoryContext) eachScriptTypeDirectoryShouldContainOnlyPackageSubdirectories() error {
	// This is checked in scriptsTypeDirShouldContainOnlyDirectories
	return nil
}

func (c *repositoryContext) packageNamesShouldBeLowercaseWithHyphens() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	typesDirs := []string{"scripts/pwsh", "scripts/sh", "scripts/cmd"}
	var violations []string

	for _, typeDir := range typesDirs {
		fullPath := filepath.Join(c.repoRoot, typeDir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(fullPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				name := entry.Name()
				// Check lowercase with hyphens only
				for _, ch := range name {
					if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-') {
						violations = append(violations, fmt.Sprintf("%s/%s", typeDir, name))
						break
					}
				}
			}
		}
	}

	if len(violations) > 0 {
		return fmt.Errorf("found %d package(s) with invalid names (must be lowercase with hyphens):\n  %s",
			len(violations), strings.Join(violations, "\n  "))
	}
	return nil
}

func (c *repositoryContext) eachPackageShouldContainAtLeastOneScriptFile() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	// Use default extensions if not set by previous step
	extensions := c.scriptExtensions
	if len(extensions) == 0 {
		extensions = []string{".sh", ".ps1", ".psm1", ".bat", ".cmd"}
	}

	typesDirs := []string{"scripts/pwsh", "scripts/sh", "scripts/cmd"}
	var emptyPackages []string

	for _, typeDir := range typesDirs {
		fullPath := filepath.Join(c.repoRoot, typeDir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(fullPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				packagePath := filepath.Join(fullPath, entry.Name())
				hasScript := false

				// Check for any script file in this package
				_ = filepath.Walk(packagePath, func(path string, info os.FileInfo, err error) error {
					if err != nil || info.IsDir() {
						return nil
					}
					for _, ext := range extensions {
						if strings.HasSuffix(info.Name(), ext) {
							hasScript = true
							return filepath.SkipAll
						}
					}
					return nil
				})

				if !hasScript {
					relPath := fmt.Sprintf("%s/%s", typeDir, entry.Name())
					emptyPackages = append(emptyPackages, relPath)
				}
			}
		}
	}

	if len(emptyPackages) > 0 {
		return fmt.Errorf("found %d empty package(s) with no script files:\n  %s",
			len(emptyPackages), strings.Join(emptyPackages, "\n  "))
	}
	return nil
}

func (c *repositoryContext) checkScriptsDirectoryStructure() error {
	return c.ensureRepoRoot()
}

func (c *repositoryContext) scriptsTypeDirShouldContainOnlyDirectories(typeDir string) error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	fullPath := filepath.Join(c.repoRoot, typeDir)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// Directory doesn't exist, that's fine
		return nil
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", typeDir, err)
	}

	var looseFiles []string
	for _, entry := range entries {
		if !entry.IsDir() {
			looseFiles = append(looseFiles, fmt.Sprintf("%s/%s", typeDir, entry.Name()))
		}
	}

	if len(looseFiles) > 0 {
		c.looseScriptsInType = append(c.looseScriptsInType, looseFiles...)
		return fmt.Errorf("found %d loose file(s) in %s (should be in package subdirectory):\n  %s",
			len(looseFiles), typeDir, strings.Join(looseFiles, "\n  "))
	}
	return nil
}

func (c *repositoryContext) scriptsCmdDirShouldContainOnlyDirectoriesIfExists() error {
	return c.scriptsTypeDirShouldContainOnlyDirectories("scripts/cmd")
}
