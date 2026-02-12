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
//
// versionHeaderRegex: Matches version headers.
//
//	Examples: "## [1.2.0] - 2025-12-01", "## [Unreleased]", "## [1.0.0] - 2025-01-01 [YANKED]"
//	Group 1: version number or "Unreleased"
//	Group 2: optional ISO date (YYYY-MM-DD)
//
// sectionHeaderRegex: Matches section headers like "### Added".
//
//	Group 1: section name
//
// listItemRegex: Matches list items like "- description" (allows leading whitespace).
//
//	Group 1: item text after "- "
//
// linkDefRegex: Matches link reference definitions like "[1.0.0]: https://..."
//
//	Group 1: link label
//	Group 2: URL
//
// conventionalEntryRegex: Matches conventional commit format in changelog entries.
//
//	Group 1: type (feat, fix, etc.)
//	Group 2: optional scope in parens
//	Group 3: optional breaking indicator "!"
//	Group 4: description text
//
// calverRegex: Matches calendar versioning format (YYYY.MM.DD or YYYY.MM.DD.N).
var (
	versionHeaderRegex     = regexp.MustCompile(`^##\s+\[([^\]]+)\](?:\s*-\s*(\d{4}-\d{2}-\d{2}))?(?:\s+\[YANKED\])?`)
	sectionHeaderRegex     = regexp.MustCompile(`^###\s+(.+)$`)
	listItemRegex          = regexp.MustCompile(`^\s*-\s+(.+)$`)
	linkDefRegex           = regexp.MustCompile(`^\[([^\]]+)\]:\s+(.+)$`)
	conventionalEntryRegex = regexp.MustCompile(`^(feat|fix|refactor|docs|chore|test|perf|style)(\([^)]+\))?(!)?:\s*(.+)$`)
	calverRegex            = regexp.MustCompile(`^\d{4}\.\d{2}\.\d{2}(\.\d+)?$`)
)

// parseState tracks parser state during line-by-line scanning.
// This replaces loose local variables with a struct so the state
// is explicit and easier to reason about.
type parseState struct {
	changelog        *Changelog
	currentVersion   *Version
	currentSection   ChangeType
	descriptionLines []string
	inHeader         bool
}

// newParseState creates a fresh parse state.
func newParseState() *parseState {
	return &parseState{
		changelog: &Changelog{
			Title:       "Changelog",
			VersionType: Semver,
		},
		inHeader: true,
	}
}

// saveCurrentVersion appends the current version to the appropriate slot.
func (ps *parseState) saveCurrentVersion() {
	if ps.currentVersion == nil {
		return
	}
	if ps.currentVersion.Number == "Unreleased" {
		ps.changelog.Unreleased = ps.currentVersion
	} else {
		ps.changelog.Versions = append(ps.changelog.Versions, *ps.currentVersion)
	}
}

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
	ps := newParseState()

	for scanner.Scan() {
		// Trim trailing \r to handle Windows line endings
		line := strings.TrimRight(scanner.Text(), "\r")

		// Parse main title
		if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## ") {
			ps.changelog.Title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			continue
		}

		// Parse version headers
		if matches := versionHeaderRegex.FindStringSubmatch(line); matches != nil {
			ps.inHeader = false

			// Save previous version
			ps.saveCurrentVersion()

			// Start new version
			versionNum := matches[1]
			ps.currentVersion = &Version{
				Number: versionNum,
				Yanked: strings.Contains(line, "[YANKED]"),
			}

			// Parse date if present
			if matches[2] != "" {
				if date, err := time.Parse("2006-01-02", matches[2]); err == nil {
					ps.currentVersion.Date = date
				}
			}

			// Detect calver
			if isCalverVersion(versionNum) {
				ps.changelog.VersionType = Calver
			}

			ps.currentSection = ""
			continue
		}

		// Collect description lines before first version
		if ps.inHeader && line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "[") {
			ps.descriptionLines = append(ps.descriptionLines, line)
			continue
		}

		// Parse section headers (### Added, ### Changed, etc.)
		if matches := sectionHeaderRegex.FindStringSubmatch(line); matches != nil {
			ps.currentSection = ChangeType(strings.TrimSpace(matches[1]))
			continue
		}

		// Parse list items
		if matches := listItemRegex.FindStringSubmatch(line); matches != nil && ps.currentVersion != nil {
			entry := parseEntry(matches[1])
			addEntryToVersion(ps.currentVersion, ps.currentSection, entry)
			continue
		}

		// Parse link definitions to extract repo URL
		if matches := linkDefRegex.FindStringSubmatch(line); matches != nil {
			if ps.changelog.RepoURL == "" {
				// Extract base repo URL from comparison link
				url := matches[2]
				if idx := strings.Index(url, "/compare/"); idx > 0 {
					ps.changelog.RepoURL = url[:idx]
				} else if idx := strings.Index(url, "/releases/"); idx > 0 {
					ps.changelog.RepoURL = url[:idx]
				}
			}
			continue
		}
	}

	// Save last version
	ps.saveCurrentVersion()

	// Set description
	if len(ps.descriptionLines) > 0 {
		ps.changelog.Description = strings.Join(ps.descriptionLines, "\n")
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading changelog: %w", err)
	}

	return ps.changelog, nil
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
	return calverRegex.MatchString(version)
}
