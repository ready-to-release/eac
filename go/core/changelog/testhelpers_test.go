package changelog

import "time"

// --- shared test helpers for parser_test.go and conventional_test.go ---

// makeEntry creates a changelog Entry with the given commit type and description.
func makeEntry(commitType, description string) Entry {
	return Entry{
		CommitType:  commitType,
		Description: description,
	}
}

// makeBreakingEntry creates a breaking-change Entry.
func makeBreakingEntry(commitType, description string) Entry {
	return Entry{
		CommitType:  commitType,
		Description: description,
		Breaking:    true,
	}
}

// makeCommit creates a Commit with the given type and description.
func makeCommit(commitType, description string) *Commit {
	return &Commit{
		Type:        commitType,
		Description: description,
	}
}

// makeCommitWithFiles creates a Commit that touches the given files.
func makeCommitWithFiles(sha string, files []string) *Commit {
	return &Commit{SHA: sha, Files: files}
}

// makeVersion creates a Version with a number, date, and entries by category.
func makeVersion(number string, date time.Time, added, fixed []Entry) Version {
	return Version{
		Number: number,
		Date:   date,
		Added:  added,
		Fixed:  fixed,
	}
}

// testDate returns a fixed date useful for deterministic tests.
func testDate() time.Time {
	return time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
}

// minimalChangelog returns the smallest valid changelog string for testing.
func minimalChangelog() string {
	return `# Changelog

## [Unreleased]

## [1.0.0] - 2025-12-01

### Added

- Initial release
`
}
