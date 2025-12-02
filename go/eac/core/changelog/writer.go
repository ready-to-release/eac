package changelog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Write serializes the changelog to a file
func (c *Changelog) Write(path string) error {
	content := c.String()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write changelog: %w", err)
	}

	return nil
}

// String renders the changelog to a markdown string
func (c *Changelog) String() string {
	var sb strings.Builder

	// Title
	title := c.Title
	if title == "" {
		title = "Changelog"
	}
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))

	// Description
	if c.Description != "" {
		sb.WriteString(c.Description)
		sb.WriteString("\n\n")
	} else {
		// Default description
		sb.WriteString(fmt.Sprintf("All notable changes to **%s** will be documented in this file.\n\n", c.Module))
		sb.WriteString("The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),\n")
		if c.VersionType == Calver {
			sb.WriteString("and this project adheres to [Calendar Versioning](https://calver.org/) (YYYY.MM.DD).\n\n")
		} else {
			sb.WriteString("and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).\n\n")
		}
	}

	// Unreleased section (placeholder for manual additions between releases)
	// Entries here will be merged with commit-generated entries by `release this`
	sb.WriteString("## [Unreleased]\n")
	if c.Unreleased != nil && c.Unreleased.HasEntries() {
		writeVersionEntries(&sb, c.Unreleased)
	}
	sb.WriteString("\n")

	// Version sections
	for _, v := range c.Versions {
		writeVersionHeader(&sb, &v)
		writeVersionEntries(&sb, &v)
		sb.WriteString("\n")
	}

	// Link definitions
	writeLinkDefinitions(&sb, c)

	return sb.String()
}

// writeVersionHeader writes the version header line
func writeVersionHeader(sb *strings.Builder, v *Version) {
	if v.Date.IsZero() {
		sb.WriteString(fmt.Sprintf("## [%s]\n", v.Number))
	} else {
		dateStr := v.Date.Format("2006-01-02")
		if v.Yanked {
			sb.WriteString(fmt.Sprintf("## [%s] - %s [YANKED]\n", v.Number, dateStr))
		} else {
			sb.WriteString(fmt.Sprintf("## [%s] - %s\n", v.Number, dateStr))
		}
	}
}

// writeVersionEntries writes all entries for a version
func writeVersionEntries(sb *strings.Builder, v *Version) {
	if len(v.Added) > 0 {
		sb.WriteString("\n### Added\n\n")
		for _, e := range v.Added {
			writeEntry(sb, e)
		}
	}

	if len(v.Changed) > 0 {
		sb.WriteString("\n### Changed\n\n")
		for _, e := range v.Changed {
			writeEntry(sb, e)
		}
	}

	if len(v.Deprecated) > 0 {
		sb.WriteString("\n### Deprecated\n\n")
		for _, e := range v.Deprecated {
			writeEntry(sb, e)
		}
	}

	if len(v.Removed) > 0 {
		sb.WriteString("\n### Removed\n\n")
		for _, e := range v.Removed {
			writeEntry(sb, e)
		}
	}

	if len(v.Fixed) > 0 {
		sb.WriteString("\n### Fixed\n\n")
		for _, e := range v.Fixed {
			writeEntry(sb, e)
		}
	}

	if len(v.Security) > 0 {
		sb.WriteString("\n### Security\n\n")
		for _, e := range v.Security {
			writeEntry(sb, e)
		}
	}
}

// writeEntry writes a single changelog entry
func writeEntry(sb *strings.Builder, e Entry) {
	// Format with conventional commit style if available
	if e.CommitType != "" {
		breaking := ""
		if e.Breaking {
			breaking = "!"
		}
		if e.Scope != "" {
			sb.WriteString(fmt.Sprintf("- %s(%s)%s: %s\n", e.CommitType, e.Scope, breaking, e.Description))
		} else {
			sb.WriteString(fmt.Sprintf("- %s%s: %s\n", e.CommitType, breaking, e.Description))
		}
	} else {
		sb.WriteString(fmt.Sprintf("- %s\n", e.Description))
	}
}

// writeLinkDefinitions writes the reference links at the bottom
func writeLinkDefinitions(sb *strings.Builder, c *Changelog) {
	if c.RepoURL == "" {
		return
	}

	// Unreleased link (compare latest version to HEAD)
	if len(c.Versions) > 0 {
		latestTag := formatTag(c.Module, c.Versions[0].Number)
		sb.WriteString(fmt.Sprintf("[Unreleased]: %s/compare/%s...HEAD\n", c.RepoURL, latestTag))
	}

	// Version links
	for i, v := range c.Versions {
		currentTag := formatTag(c.Module, v.Number)
		if i < len(c.Versions)-1 {
			// Compare to previous version
			prevTag := formatTag(c.Module, c.Versions[i+1].Number)
			sb.WriteString(fmt.Sprintf("[%s]: %s/compare/%s...%s\n", v.Number, c.RepoURL, prevTag, currentTag))
		} else {
			// First version - link to release
			sb.WriteString(fmt.Sprintf("[%s]: %s/releases/tag/%s\n", v.Number, c.RepoURL, currentTag))
		}
	}
}

// formatTag creates a git tag name from module and version
func formatTag(module, version string) string {
	if module == "" {
		return version
	}
	return fmt.Sprintf("%s/%s", module, version)
}
