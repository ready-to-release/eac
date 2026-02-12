// Package repository contains godog step implementations for specs/repository.
package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ============================================================================
// Markdown Validation Steps
// ============================================================================

func (c *repositoryContext) discoverAllMarkdownFiles() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	// Use cached git ls-files instead of filepath.Walk (49s -> <1s)
	if err := c.sharedCtx.EnsureOriginalRepoCache(); err != nil {
		return err
	}

	// Get all .md files from cache and convert to absolute paths
	// Filter out files that no longer exist on disk (e.g., deleted but not yet staged)
	relPaths := c.sharedCtx.OriginalRepoCache.FilesByExtension(".md")
	c.markdownFiles = make([]string, 0, len(relPaths))
	for _, relPath := range relPaths {
		absPath := c.sharedCtx.OriginalRepoCache.AbsolutePath(relPath)
		if _, err := os.Stat(absPath); err == nil {
			c.markdownFiles = append(c.markdownFiles, absPath)
		}
	}
	return nil
}

func (c *repositoryContext) validateEachMarkdownFile() error {
	// Basic markdown validation - check for common issues
	for _, path := range c.markdownFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			c.markdownErrors[path] = append(c.markdownErrors[path], fmt.Sprintf("failed to read: %v", err))
			continue
		}

		// Check for malformed headers (e.g., #NoSpace)
		// Must track code blocks to avoid false positives on shebangs like #!/bin/bash
		lines := strings.Split(string(content), "\n")
		inCodeBlock := false
		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)

			// Track code block boundaries
			if strings.HasPrefix(trimmedLine, "```") {
				inCodeBlock = !inCodeBlock
				continue
			}

			// Skip lines inside code blocks
			if inCodeBlock {
				continue
			}

			if strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "# ") &&
				!strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") &&
				!strings.HasPrefix(line, "#### ") && !strings.HasPrefix(line, "##### ") &&
				!strings.HasPrefix(line, "###### ") && line != "#" {
				// Check if it's actually a header without space
				trimmed := strings.TrimLeft(line, "#")
				if len(trimmed) > 0 && trimmed[0] != ' ' && trimmed[0] != '\n' && trimmed[0] != '\r' {
					c.markdownErrors[path] = append(c.markdownErrors[path],
						fmt.Sprintf("line %d: malformed header (missing space after #)", i+1))
				}
			}
		}
	}
	return nil
}

func (c *repositoryContext) allFilesShouldHaveValidMarkdownSyntax() error {
	if len(c.markdownErrors) > 0 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("found %d file(s) with markdown errors:\n", len(c.markdownErrors)))
		for path, errors := range c.markdownErrors {
			sb.WriteString(fmt.Sprintf("\n  %s:\n", path))
			for _, err := range errors {
				sb.WriteString(fmt.Sprintf("    - %s\n", err))
			}
		}
		return fmt.Errorf("%s", sb.String())
	}
	return nil
}

func (c *repositoryContext) noFilesShouldHaveBrokenLinks() error {
	// Check for broken relative file links in markdown files.
	// Only validates local file references (not URLs), verifying that
	// referenced paths actually exist on disk.
	linkRegex := regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)

	var brokenLinks []string
	for mdPath, _ := range c.markdownErrors {
		// Already tracked files with errors
		_ = mdPath
	}

	for _, mdPath := range c.markdownFiles {
		contentBytes, err := os.ReadFile(mdPath)
		if err != nil {
			continue
		}

		// Strip fenced code blocks and inline code to avoid false positives
		content := stripCodeBlocks(string(contentBytes))

		dir := filepath.Dir(mdPath)
		matches := linkRegex.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			linkTarget := match[2]

			// Skip external URLs, anchors, and mailto links
			if strings.HasPrefix(linkTarget, "http://") ||
				strings.HasPrefix(linkTarget, "https://") ||
				strings.HasPrefix(linkTarget, "#") ||
				strings.HasPrefix(linkTarget, "mailto:") {
				continue
			}

			// Strip anchor from path (e.g., "file.md#section" -> "file.md")
			if idx := strings.Index(linkTarget, "#"); idx >= 0 {
				linkTarget = linkTarget[:idx]
			}
			if linkTarget == "" {
				continue
			}

			// Resolve relative to the markdown file's directory
			resolved := filepath.Join(dir, linkTarget)
			if _, err := os.Stat(resolved); os.IsNotExist(err) {
				relPath, _ := filepath.Rel(c.repoRoot, mdPath)
				brokenLinks = append(brokenLinks, fmt.Sprintf("%s: broken link to %s", relPath, linkTarget))
			}
		}
	}

	if len(brokenLinks) > 0 {
		return fmt.Errorf("found %d broken link(s):\n  %s", len(brokenLinks), strings.Join(brokenLinks, "\n  "))
	}
	return nil
}

// stripCodeBlocks removes fenced code blocks and inline code spans from markdown
// content so that link-like syntax inside code is not checked as real links.
func stripCodeBlocks(content string) string {
	// Remove fenced code blocks (```...```)
	fenced := regexp.MustCompile("(?s)```[^`]*```")
	content = fenced.ReplaceAllString(content, "")
	// Remove inline code spans (`...`)
	inline := regexp.MustCompile("`[^`]+`")
	content = inline.ReplaceAllString(content, "")
	return content
}

func (c *repositoryContext) noFilesShouldHaveMalformedHeaders() error {
	for path, errors := range c.markdownErrors {
		for _, err := range errors {
			if strings.Contains(err, "malformed header") {
				return fmt.Errorf("file %s has malformed headers", path)
			}
		}
	}
	return nil
}

func (c *repositoryContext) ifAnyFileHasErrorsIShouldSeeDetails() error {
	// Passive assertion
	return nil
}

// ============================================================================
// Feature Tag Validation Steps
// ============================================================================

func (c *repositoryContext) discoverAllGherkinFeatureFiles() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	// Use cached git ls-files instead of filepath.Walk
	if err := c.sharedCtx.EnsureOriginalRepoCache(); err != nil {
		return err
	}

	// Get all .feature files in specs/ directory from cache
	relPaths := c.sharedCtx.OriginalRepoCache.FilesInDirWithExtension("specs", ".feature")
	c.featureFiles = make([]string, len(relPaths))
	for i, relPath := range relPaths {
		c.featureFiles[i] = c.sharedCtx.OriginalRepoCache.AbsolutePath(relPath)
	}
	return nil
}

func (c *repositoryContext) validateFeatureFilesForConflictingLLevelTags() error {
	for _, featureFile := range c.featureFiles {
		content, err := os.ReadFile(featureFile)
		if err != nil {
			continue
		}
		if conflicts := detectTagConflicts(string(content), featureFile); len(conflicts) > 0 {
			c.tagConflicts = append(c.tagConflicts, conflicts...)
		}
	}
	return nil
}

// detectTagConflicts parses a Gherkin feature file and detects conflicting
// L-level and verification tags. A conflict occurs when the Feature-level tag
// asserts a single level (e.g., @L2) but individual Scenarios carry different
// level tags.
func detectTagConflicts(content, filePath string) []string {
	var conflicts []string
	lines := strings.Split(content, "\n")

	var featureTags []string
	var currentScenarioName string
	var currentScenarioTags []string
	scenarioLevels := map[string][]string{} // tag prefix -> list of scenario-level values

	inFeatureHeader := true

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Collect tags
		if strings.HasPrefix(trimmed, "@") {
			tags := strings.Fields(trimmed)
			if inFeatureHeader {
				featureTags = append(featureTags, tags...)
			} else {
				currentScenarioTags = append(currentScenarioTags, tags...)
			}
			continue
		}

		// Feature keyword ends the header tag collection
		if strings.HasPrefix(trimmed, "Feature:") {
			inFeatureHeader = false
			continue
		}

		// Scenario or Scenario Outline
		if strings.HasPrefix(trimmed, "Scenario:") || strings.HasPrefix(trimmed, "Scenario Outline:") {
			currentScenarioName = strings.TrimPrefix(strings.TrimPrefix(trimmed, "Scenario Outline:"), "Scenario:")
			currentScenarioName = strings.TrimSpace(currentScenarioName)

			// Record L-level and verification tags for this scenario
			for _, tag := range currentScenarioTags {
				if strings.HasPrefix(tag, "@L") && len(tag) <= 4 {
					scenarioLevels["@L"] = append(scenarioLevels["@L"], tag+":"+currentScenarioName)
				}
				if strings.HasPrefix(tag, "@verification-") {
					scenarioLevels["@verification"] = append(scenarioLevels["@verification"], tag+":"+currentScenarioName)
				}
			}
			currentScenarioTags = nil
			continue
		}
	}

	// Check for conflicts: feature has @Lx but scenarios have different @Ly
	featureLLevel := ""
	featureVerification := ""
	for _, tag := range featureTags {
		if strings.HasPrefix(tag, "@L") && len(tag) <= 4 {
			featureLLevel = tag
		}
		if strings.HasPrefix(tag, "@verification-") {
			featureVerification = tag
		}
	}

	if featureLLevel != "" {
		for _, entry := range scenarioLevels["@L"] {
			parts := strings.SplitN(entry, ":", 2)
			scenarioTag := parts[0]
			scenarioName := parts[1]
			if scenarioTag != featureLLevel {
				conflicts = append(conflicts, fmt.Sprintf("%s: Feature has %s but scenario '%s' has %s",
					filePath, featureLLevel, scenarioName, scenarioTag))
			}
		}
	}

	if featureVerification != "" {
		for _, entry := range scenarioLevels["@verification"] {
			parts := strings.SplitN(entry, ":", 2)
			scenarioTag := parts[0]
			scenarioName := parts[1]
			if scenarioTag != featureVerification {
				conflicts = append(conflicts, fmt.Sprintf("%s: Feature has %s but scenario '%s' has %s",
					filePath, featureVerification, scenarioName, scenarioTag))
			}
		}
	}

	return conflicts
}

func (c *repositoryContext) noFeatureShouldHaveConflictingLTags() error {
	if len(c.tagConflicts) > 0 {
		return fmt.Errorf("found tag conflicts: %v", c.tagConflicts)
	}
	return nil
}

func (c *repositoryContext) noFeatureShouldHaveConflictingVerificationTags() error {
	// Same as L-tags for now
	return nil
}

func (c *repositoryContext) ifConflictsFoundIShouldSeeDetails() error {
	// Passive assertion
	return nil
}
