package changelog

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// Parser regular expressions.
var (
	// Matches version headers like "## [1.2.0] - 2025-12-01" or "## [Unreleased]".
	versionHeaderRegex = regexp.MustCompile(`^##\s+\[([^\]]+)\](?:\s*-\s*(\d{4}-\d{2}-\d{2}))?(?:\s+\[YANKED\])?`)

	// Matches section headers like "### Added".
	sectionHeaderRegex = regexp.MustCompile(`^###\s+(.+)$`)

	// Matches list items like "- description".
	listItemRegex = regexp.MustCompile(`^-\s+(.+)$`)

	// Matches link definitions like "[1.0.0]: https://..."
	linkDefRegex = regexp.MustCompile(`^\[([^\]]+)\]:\s+(.+)$`)

	// Matches conventional commit format in entry: "feat(scope): description".
	conventionalEntryRegex = regexp.MustCompile(`^(feat|fix|refactor|docs|chore|test|perf|style)(\([^)]+\))?(!)?:\s*(.+)$`)
)

// Parse reads and parses a CHANGELOG.md file.
func Parse(path string) (*Changelog, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open changelog: %w", err)
	}
	defer file.Close()

	return ParseReader(bufio.NewScanner(file))
}

// ParseString parses changelog content from a string.
func ParseString(content string) (*Changelog, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	return ParseReader(scanner)
}

// ParseReader parses changelog content from a scanner.
func ParseReader(scanner *bufio.Scanner) (*Changelog, error) {
	changelog := &Changelog{
		Title:       "Changelog",
		VersionType: Semver,
	}

	var currentVersion *Version
	var currentSection ChangeType
	var descriptionLines []string
	inHeader := true

	for scanner.Scan() {
		line := scanner.Text()

		// Parse main title
		if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## ") {
			changelog.Title = strings.TrimPrefix(line, "# ")
			continue
		}

		// Parse version headers
		if matches := versionHeaderRegex.FindStringSubmatch(line); matches != nil {
			inHeader = false

			// Save previous version
			if currentVersion != nil {
				if currentVersion.Number == "Unreleased" {
					changelog.Unreleased = currentVersion
				} else {
					changelog.Versions = append(changelog.Versions, *currentVersion)
				}
			}

			// Start new version
			versionNum := matches[1]
			currentVersion = &Version{
				Number: versionNum,
				Yanked: strings.Contains(line, "[YANKED]"),
			}

			// Parse date if present
			if matches[2] != "" {
				if date, err := time.Parse("2006-01-02", matches[2]); err == nil {
					currentVersion.Date = date
				}
			}

			// Detect calver
			if isCalverVersion(versionNum) {
				changelog.VersionType = Calver
			}

			currentSection = ""
			continue
		}

		// Collect description lines before first version
		if inHeader && line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "[") {
			descriptionLines = append(descriptionLines, line)
			continue
		}

		// Parse section headers (### Added, ### Changed, etc.)
		if matches := sectionHeaderRegex.FindStringSubmatch(line); matches != nil {
			currentSection = ChangeType(matches[1])
			continue
		}

		// Parse list items
		if matches := listItemRegex.FindStringSubmatch(line); matches != nil && currentVersion != nil {
			entry := parseEntry(matches[1])
			addEntryToVersion(currentVersion, currentSection, entry)
			continue
		}

		// Parse link definitions to extract repo URL
		if matches := linkDefRegex.FindStringSubmatch(line); matches != nil {
			if changelog.RepoURL == "" {
				// Extract base repo URL from comparison link
				url := matches[2]
				if idx := strings.Index(url, "/compare/"); idx > 0 {
					changelog.RepoURL = url[:idx]
				} else if idx := strings.Index(url, "/releases/"); idx > 0 {
					changelog.RepoURL = url[:idx]
				}
			}
			continue
		}
	}

	// Save last version
	if currentVersion != nil {
		if currentVersion.Number == "Unreleased" {
			changelog.Unreleased = currentVersion
		} else {
			changelog.Versions = append(changelog.Versions, *currentVersion)
		}
	}

	// Set description
	if len(descriptionLines) > 0 {
		changelog.Description = strings.Join(descriptionLines, "\n")
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading changelog: %w", err)
	}

	return changelog, nil
}

// parseEntry parses a list item into an Entry.
func parseEntry(text string) Entry {
	entry := Entry{
		Description: text,
	}

	// Try to parse conventional commit format
	if matches := conventionalEntryRegex.FindStringSubmatch(text); matches != nil {
		entry.CommitType = matches[1]
		if matches[2] != "" {
			// Remove parentheses from scope
			entry.Scope = strings.Trim(matches[2], "()")
		}
		entry.Breaking = matches[3] == "!"
		entry.Description = matches[4]
	}

	return entry
}

// addEntryToVersion adds an entry to the appropriate section.
func addEntryToVersion(v *Version, section ChangeType, entry Entry) {
	switch section {
	case Added:
		v.Added = append(v.Added, entry)
	case Changed:
		v.Changed = append(v.Changed, entry)
	case Deprecated:
		v.Deprecated = append(v.Deprecated, entry)
	case Removed:
		v.Removed = append(v.Removed, entry)
	case Fixed:
		v.Fixed = append(v.Fixed, entry)
	case Security:
		v.Security = append(v.Security, entry)
	}
}

// isCalverVersion checks if a version string looks like calendar versioning.
func isCalverVersion(version string) bool {
	// Calver format: YYYY.MM.DD or YYYY.MM.DD.N
	calverRegex := regexp.MustCompile(`^\d{4}\.\d{2}\.\d{2}(\.\d+)?$`)
	return calverRegex.MatchString(version)
}
