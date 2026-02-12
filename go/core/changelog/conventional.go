package changelog

import (
	"regexp"
	"strings"
	"time"
)

// Commit represents a parsed conventional commit.
type Commit struct {
	// SHA is the commit hash (short or full)
	SHA string

	// Type is the conventional commit type (feat, fix, etc.)
	Type string

	// Scope is the optional scope
	Scope string

	// Description is the commit subject/summary
	Description string

	// Body is the full commit body
	Body string

	// Breaking indicates if this is a breaking change
	Breaking bool

	// Date is the commit timestamp
	Date time.Time

	// Files is the list of files changed in this commit
	Files []string
}

// conventionalCommitRegex matches the Conventional Commits specification format.
//
// The pattern matches messages like:
//
//	feat(api)!: add breaking endpoint
//	fix: resolve memory leak
//	docs(readme): update install guide
//
// Capture groups:
//
//	Group 1 -- type: one of the allowed commit types
//	           (feat|fix|refactor|docs|chore|test|perf|style|ci|build|revert)
//	Group 2 -- scope: optional parenthesized scope, e.g. "(api)" captures "api"
//	Group 3 -- breaking indicator: optional "!" before the colon
//	Group 4 -- description: the commit subject after ": "
var conventionalCommitRegex = regexp.MustCompile(
	`^(feat|fix|refactor|docs|chore|test|perf|style|ci|build|revert)` + // Group 1: type
		`(?:\(([^)]+)\))?` + // Group 2: optional scope in parens
		`(!)?` + // Group 3: optional breaking indicator
		`:\s*(.+)$`) // Group 4: colon and description

// ParseCommitMessage parses a commit message into a Commit struct.
func ParseCommitMessage(message string) *Commit {
	lines := strings.SplitN(message, "\n", 2)
	subject := strings.TrimSpace(lines[0])

	commit := &Commit{
		Description: subject,
	}

	// Parse body if present
	if len(lines) > 1 {
		commit.Body = strings.TrimSpace(lines[1])

		// Check for BREAKING CHANGE in body
		if strings.Contains(commit.Body, "BREAKING CHANGE:") ||
			strings.Contains(commit.Body, "BREAKING-CHANGE:") {
			commit.Breaking = true
		}
	}

	// Parse conventional commit format
	matches := conventionalCommitRegex.FindStringSubmatch(subject)
	if matches != nil {
		commit.Type = matches[1]
		commit.Scope = matches[2]
		commit.Breaking = commit.Breaking || matches[3] == "!"
		commit.Description = matches[4]
	}

	return commit
}

// IsConventionalCommit returns true if the commit follows conventional commit format.
func (c *Commit) IsConventionalCommit() bool {
	return c.Type != ""
}

// ToEntry converts a Commit to a changelog Entry.
func (c *Commit) ToEntry() Entry {
	return Entry{
		Description: c.Description,
		CommitType:  c.Type,
		Scope:       c.Scope,
		CommitSHA:   c.SHA,
		Breaking:    c.Breaking,
	}
}

// CommitsToEntries converts a slice of commits to changelog entries
// grouped by change type.
func CommitsToEntries(commits []*Commit) map[ChangeType][]Entry {
	result := make(map[ChangeType][]Entry)

	for _, c := range commits {
		if !c.IsConventionalCommit() {
			continue
		}

		entry := c.ToEntry()
		changeType := CommitTypeToChangeType(c.Type)
		result[changeType] = append(result[changeType], entry)
	}

	return result
}

// CommitsToVersion creates a Version from commits.
func CommitsToVersion(commits []*Commit, versionNumber string, date time.Time) Version {
	entriesByType := CommitsToEntries(commits)

	return Version{
		Number:     versionNumber,
		Date:       date,
		Added:      entriesByType[Added],
		Changed:    entriesByType[Changed],
		Deprecated: entriesByType[Deprecated],
		Removed:    entriesByType[Removed],
		Fixed:      entriesByType[Fixed],
		Security:   entriesByType[Security],
	}
}

// FilterCommitsByModule filters commits that affect files matching module patterns.
func FilterCommitsByModule(commits []*Commit, patterns []string) []*Commit {
	if len(patterns) == 0 {
		return commits
	}

	var filtered []*Commit
	for _, c := range commits {
		if commitMatchesPatterns(c, patterns) {
			filtered = append(filtered, c)
		}
	}

	return filtered
}

// commitMatchesPatterns checks if any of the commit's files match the patterns.
func commitMatchesPatterns(c *Commit, patterns []string) bool {
	for _, file := range c.Files {
		for _, pattern := range patterns {
			if matchGlobPattern(file, pattern) {
				return true
			}
		}
	}
	return false
}

// matchGlobPattern is a simple glob matcher for common patterns
// Supports: *, **, and literal matches.
func matchGlobPattern(path, pattern string) bool {
	// Normalize separators
	path = strings.ReplaceAll(path, "\\", "/")
	pattern = strings.ReplaceAll(pattern, "\\", "/")

	// Handle ** pattern (any depth)
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := strings.TrimPrefix(parts[1], "/")

			// Check prefix match
			if prefix != "" && !strings.HasPrefix(path, prefix) {
				return false
			}

			// Check suffix match
			if suffix != "" {
				remainingPath := strings.TrimPrefix(path, prefix)
				return matchSimpleGlob(remainingPath, suffix) || strings.HasSuffix(remainingPath, suffix)
			}

			return true
		}
	}

	return matchSimpleGlob(path, pattern)
}

// matchSimpleGlob matches a path against a pattern with * wildcards.
func matchSimpleGlob(path, pattern string) bool {
	// Simple case: no wildcards
	if !strings.Contains(pattern, "*") {
		return path == pattern || strings.HasPrefix(path, pattern+"/")
	}

	// Handle single * (matches within one directory level)
	parts := strings.Split(pattern, "*")
	if len(parts) == 2 {
		prefix := parts[0]
		suffix := parts[1]

		if !strings.HasPrefix(path, prefix) {
			return false
		}

		remaining := strings.TrimPrefix(path, prefix)

		// * doesn't match /
		if strings.Contains(remaining, "/") {
			idx := strings.Index(remaining, "/")
			remaining = remaining[:idx]
		}

		return strings.HasSuffix(remaining, suffix) || suffix == ""
	}

	return false
}
