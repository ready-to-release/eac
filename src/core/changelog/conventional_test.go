package changelog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseCommitMessage_ConventionalFormat(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantType string
		scope    string
		desc     string
		breaking bool
	}{
		{
			name:     "feat with scope",
			message:  "feat(api): add new endpoint",
			wantType: "feat",
			scope:    "api",
			desc:     "add new endpoint",
			breaking: false,
		},
		{
			name:     "fix without scope",
			message:  "fix: resolve memory leak",
			wantType: "fix",
			scope:    "",
			desc:     "resolve memory leak",
			breaking: false,
		},
		{
			name:     "breaking with exclamation",
			message:  "feat(api)!: change response format",
			wantType: "feat",
			scope:    "api",
			desc:     "change response format",
			breaking: true,
		},
		{
			name:     "breaking without scope",
			message:  "refactor!: restructure modules",
			wantType: "refactor",
			scope:    "",
			desc:     "restructure modules",
			breaking: true,
		},
		{
			name:     "all types - docs",
			message:  "docs(readme): update installation guide",
			wantType: "docs",
			scope:    "readme",
			desc:     "update installation guide",
			breaking: false,
		},
		{
			name:     "all types - chore",
			message:  "chore: update dependencies",
			wantType: "chore",
			scope:    "",
			desc:     "update dependencies",
			breaking: false,
		},
		{
			name:     "all types - test",
			message:  "test(unit): add coverage for parser",
			wantType: "test",
			scope:    "unit",
			desc:     "add coverage for parser",
			breaking: false,
		},
		{
			name:     "all types - perf",
			message:  "perf: optimize database queries",
			wantType: "perf",
			scope:    "",
			desc:     "optimize database queries",
			breaking: false,
		},
		{
			name:     "all types - style",
			message:  "style: format code with prettier",
			wantType: "style",
			scope:    "",
			desc:     "format code with prettier",
			breaking: false,
		},
		{
			name:     "all types - ci",
			message:  "ci: add GitHub Actions workflow",
			wantType: "ci",
			scope:    "",
			desc:     "add GitHub Actions workflow",
			breaking: false,
		},
		{
			name:     "all types - build",
			message:  "build(docker): optimize image size",
			wantType: "build",
			scope:    "docker",
			desc:     "optimize image size",
			breaking: false,
		},
		{
			name:     "all types - revert",
			message:  "revert: undo previous commit",
			wantType: "revert",
			scope:    "",
			desc:     "undo previous commit",
			breaking: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commit := ParseCommitMessage(tt.message)
			assert.Equal(t, tt.wantType, commit.Type)
			assert.Equal(t, tt.scope, commit.Scope)
			assert.Equal(t, tt.desc, commit.Description)
			assert.Equal(t, tt.breaking, commit.Breaking)
			assert.True(t, commit.IsConventionalCommit())
		})
	}
}

func TestParseCommitMessage_NonConventional(t *testing.T) {
	messages := []string{
		"Update readme",
		"Fix bug",
		"WIP: working on feature",
		"Merge branch 'main'",
	}

	for _, msg := range messages {
		t.Run(msg, func(t *testing.T) {
			commit := ParseCommitMessage(msg)
			assert.Equal(t, "", commit.Type)
			assert.Equal(t, msg, commit.Description)
			assert.False(t, commit.IsConventionalCommit())
		})
	}
}

func TestParseCommitMessage_BreakingInBody(t *testing.T) {
	message := `feat(api): add new feature

This is a detailed description.

BREAKING CHANGE: this changes the API behavior`

	commit := ParseCommitMessage(message)
	assert.Equal(t, "feat", commit.Type)
	assert.Equal(t, "api", commit.Scope)
	assert.True(t, commit.Breaking)
	assert.Contains(t, commit.Body, "BREAKING CHANGE:")
}

func TestParseCommitMessage_BreakingInBodyAltFormat(t *testing.T) {
	message := `fix: update handler

BREAKING-CHANGE: old handlers no longer work`

	commit := ParseCommitMessage(message)
	assert.True(t, commit.Breaking)
}

func TestCommit_ToEntry(t *testing.T) {
	commit := &Commit{
		SHA:         "abc123",
		Type:        "feat",
		Scope:       "api",
		Description: "add endpoint",
		Breaking:    true,
	}

	entry := commit.ToEntry()
	assert.Equal(t, "add endpoint", entry.Description)
	assert.Equal(t, "feat", entry.CommitType)
	assert.Equal(t, "api", entry.Scope)
	assert.Equal(t, "abc123", entry.CommitSHA)
	assert.True(t, entry.Breaking)
}

func TestCommitsToEntries(t *testing.T) {
	commits := []*Commit{
		{Type: "feat", Scope: "api", Description: "new feature"},
		{Type: "fix", Scope: "core", Description: "bug fix"},
		{Type: "feat", Description: "another feature"},
		{Type: "docs", Description: "update docs"},
		{Type: "", Description: "non-conventional commit"}, // Should be skipped
	}

	result := CommitsToEntries(commits)

	assert.Len(t, result[Added], 2)
	assert.Len(t, result[Fixed], 1)
	assert.Len(t, result[Changed], 1) // docs -> Changed
}

func TestCommitsToVersion(t *testing.T) {
	commits := []*Commit{
		{Type: "feat", Description: "new feature"},
		{Type: "fix", Description: "bug fix"},
	}

	releaseDate := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	v := CommitsToVersion(commits, "1.0.0", releaseDate)

	assert.Equal(t, "1.0.0", v.Number)
	assert.Len(t, v.Added, 1)
	assert.Len(t, v.Fixed, 1)
}

func TestFilterCommitsByModule(t *testing.T) {
	commits := []*Commit{
		{SHA: "1", Files: []string{"src/cli/main.go"}},
		{SHA: "2", Files: []string{"src/core/types.go"}},
		{SHA: "3", Files: []string{"docs/readme.md"}},
		{SHA: "4", Files: []string{"src/cli/config.go", "src/core/utils.go"}},
	}

	// Filter for src/cli/**
	patterns := []string{"src/cli/**"}
	filtered := FilterCommitsByModule(commits, patterns)

	assert.Len(t, filtered, 2)
	assert.Equal(t, "1", filtered[0].SHA)
	assert.Equal(t, "4", filtered[1].SHA)
}

func TestFilterCommitsByModule_EmptyPatterns(t *testing.T) {
	commits := []*Commit{
		{SHA: "1", Files: []string{"src/cli/main.go"}},
		{SHA: "2", Files: []string{"src/core/types.go"}},
	}

	// Empty patterns returns all commits
	filtered := FilterCommitsByModule(commits, []string{})
	assert.Len(t, filtered, 2)
}

func TestMatchGlobPattern(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		// Literal matches
		{"src/cli/main.go", "src/cli/main.go", true},
		{"src/cli/main.go", "src/core/main.go", false},

		// ** patterns (prefix/**)
		{"src/cli/main.go", "src/cli/**", true},
		{"src/cli/internal/config.go", "src/cli/**", true},
		{"src/core/main.go", "src/cli/**", false},

		// Note: **/*.go pattern is not supported by our simple matcher
		// Use the full gobwas/glob library if needed

		// * patterns
		{"src/cli/main.go", "src/cli/*.go", true},
		{"src/cli/internal/config.go", "src/cli/*.go", false}, // * doesn't match /

		// Prefix patterns
		{"src/cli/main.go", "src/cli", true},
		{"src/cli/internal/config.go", "src/cli", true},
	}

	for _, tt := range tests {
		t.Run(tt.path+"_"+tt.pattern, func(t *testing.T) {
			result := matchGlobPattern(tt.path, tt.pattern)
			assert.Equal(t, tt.want, result)
		})
	}
}
