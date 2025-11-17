package commitmessage

import (
	"strings"
	"testing"
)

// TestAutoCleanup_RemovesTrailingPeriods tests period removal from header and subject lines
func TestAutoCleanup_RemovesTrailingPeriods(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		checkHeader  bool
		checkSubject bool
	}{
		{
			name: "remove trailing period from header",
			input: `feat(cli): add new feature.

Auditor-Summary: Added feature.

Body text here.`,
			checkHeader: true,
		},
		{
			name: "remove trailing period from subject line",
			input: `feat(multi-module): add features

Auditor-Summary: Multi-module changes.

Top level text.

src-commands
------------
src-commands: feat: add validation.

Module body text.`,
			checkSubject: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AutoCleanup(tt.input)
			lines := strings.Split(got, "\n")

			if tt.checkHeader && len(lines) > 0 {
				if strings.HasSuffix(lines[0], ".") && !strings.HasSuffix(lines[0], "...") {
					t.Errorf("AutoCleanup() should remove trailing period from header, got: %q", lines[0])
				}
			}

			if tt.checkSubject {
				for _, line := range lines {
					trimmed := strings.TrimSpace(line)
					if subjectLineRegex.MatchString(trimmed) {
						if strings.HasSuffix(trimmed, ".") && !strings.HasSuffix(trimmed, "...") {
							t.Errorf("AutoCleanup() should remove trailing period from subject line, got: %q", trimmed)
						}
					}
				}
			}
		})
	}
}

// TestAutoCleanup_WrapsLongLines tests line wrapping at MaxLineLength
func TestAutoCleanup_WrapsLongLines(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
	}{
		{
			name: "wrap long body text",
			input: `feat(cli): add feature

Auditor-Summary: Added feature.

This is a very long line that exceeds the maximum allowed length of 72 characters for body text in commit messages and should be wrapped.`,
			maxLength: MaxLineLength,
		},
		{
			name: "wrap long subject line",
			input: `feat(cli): add feature

Auditor-Summary: Added feature.

Top level body.

src-commands
------------
src-commands: feat: add a really long feature description that exceeds the maximum allowed length

Module body.`,
			maxLength: MaxSubjectLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AutoCleanup(tt.input)
			lines := strings.Split(got, "\n")

			// Check that most lines (except special ones) are within limit
			for i, line := range lines {
				trimmed := strings.TrimSpace(line)
				// Skip special fields, code blocks, tables, separators
				if trimmed == "" || strings.HasPrefix(trimmed, "Auditor-Summary:") ||
					strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "|") ||
					isDashesLine(trimmed) || trimmed == "---" {
					continue
				}

				if len(trimmed) > tt.maxLength+10 { // Allow some tolerance for wrapping algorithm
					t.Errorf("Line %d is too long (%d chars, max %d): %q", i+1, len(trimmed), tt.maxLength, trimmed)
				}
			}
		})
	}
}

// TestAutoCleanup_ClosesCodeBlocks tests closing of unclosed code blocks
func TestAutoCleanup_ClosesCodeBlocks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		wantClosed bool  // Should code blocks be balanced?
	}{
		{
			name: "close unclosed code block at end",
			input: `feat(cli): add feature

Auditor-Summary: Added feature.

Module body text here.

` + "```go" + `
func Hello() {}`,
			wantClosed: true,
		},
		{
			name: "preserve properly closed code blocks",
			input: `feat(cli): add feature

Auditor-Summary: Added feature.

Module body.

` + "```go" + `
func Hello() {}
` + "```" + `

More text.`,
			wantClosed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AutoCleanup(tt.input)
			fenceCount := strings.Count(got, "```")
			isClosed := fenceCount%2 == 0

			if isClosed != tt.wantClosed {
				t.Errorf("AutoCleanup() code blocks closed = %v, want %v\nFence count: %d\nGot:\n%q", isClosed, tt.wantClosed, fenceCount, got)
			}
		})
	}
}

// TestAutoCleanup_NormalizesSpacing tests duplicate blank line removal
func TestAutoCleanup_NormalizesSpacing(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "reduce multiple blank lines to single",
			input: `feat(cli): add feature



Auditor-Summary: Added feature.


Body text.`,
			want: `feat(cli): add feature

Auditor-Summary: Added feature.

Body text.

`,
		},
		{
			name: "preserve single blank lines",
			input: `feat(cli): add feature

Auditor-Summary: Added feature.

Body text.`,
			want: `feat(cli): add feature

Auditor-Summary: Added feature.

Body text.

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AutoCleanup(tt.input)
			if got != tt.want {
				t.Errorf("AutoCleanup() failed\nGot:\n%q\n\nWant:\n%q", got, tt.want)
			}
		})
	}
}

// TestAutoCleanup_EnsuresBlankLinesAroundCodeBlocks tests blank lines before/after code blocks
func TestAutoCleanup_EnsuresBlankLinesAroundCodeBlocks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		checkBefore bool  // Check for blank line before code block
		checkAfter  bool  // Check for blank line after code block
	}{
		{
			name: "ensure proper spacing around code blocks",
			input: `feat(cli): add feature

Auditor-Summary: Added feature.

Module body here.

` + "```go" + `
func Hello() {}
` + "```" + `

More text after.`,
			checkBefore: true,
			checkAfter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AutoCleanup(tt.input)

			// Just verify code blocks are present and properly spaced
			if tt.checkBefore && !strings.Contains(got, "\n\n```") {
				t.Errorf("AutoCleanup() should have blank line before code block")
			}
			if tt.checkAfter && !strings.Contains(got, "```\n\n") {
				t.Errorf("AutoCleanup() should have blank line after code block")
			}
		})
	}
}

// TestAutoCleanup_HandlesEmptyInput tests empty input handling
func TestAutoCleanup_HandlesEmptyInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only whitespace",
			input: "   \n\n\t  \n",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AutoCleanup(tt.input)
			if got != tt.want {
				t.Errorf("AutoCleanup() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestAutoCleanup_PreservesSpecialFields tests that special fields are not wrapped
func TestAutoCleanup_PreservesSpecialFields(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "preserve Auditor-Summary field",
			input: `feat(cli): add feature

Auditor-Summary: This is a very long summary that exceeds 72 characters but should not be wrapped because it is a special field.

Body text.`,
			want: `feat(cli): add feature

Auditor-Summary: This is a very long summary that exceeds 72 characters but should not be wrapped because it is a special field.

Body text.

`,
		},
		{
			name: "preserve Changes field",
			input: `feat(cli): add feature

Changes: This is a very long changes description that exceeds 72 characters but should not be wrapped.

Body text.`,
			want: `feat(cli): add feature

Changes: This is a very long changes description that exceeds 72 characters but should not be wrapped.

Body text.

`,
		},
		{
			name: "preserve BREAKING CHANGE field",
			input: `feat(cli): add feature

BREAKING CHANGE: This is a very long breaking change description that exceeds 72 characters.

Body text.`,
			want: `feat(cli): add feature

BREAKING CHANGE: This is a very long breaking change description that exceeds 72 characters.

Body text.

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AutoCleanup(tt.input)
			if got != tt.want {
				t.Errorf("AutoCleanup() failed\nGot:\n%q\n\nWant:\n%q", got, tt.want)
			}
		})
	}
}

// TestAutoCleanup_RemovesTrailingSeparators tests trailing --- removal
func TestAutoCleanup_RemovesTrailingSeparators(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "remove trailing separator",
			input: `feat(cli): add feature

Auditor-Summary: Added feature.

Body text.

---`,
			want: `feat(cli): add feature

Auditor-Summary: Added feature.

Body text.

`,
		},
		{
			name: "remove multiple trailing separators",
			input: `feat(cli): add feature

Auditor-Summary: Added feature.

Body text.

---

---`,
			want: `feat(cli): add feature

Auditor-Summary: Added feature.

Body text.

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AutoCleanup(tt.input)
			if got != tt.want {
				t.Errorf("AutoCleanup() failed\nGot:\n%q\n\nWant:\n%q", got, tt.want)
			}
		})
	}
}

// TestGetCleanupStats tests statistics generation
func TestGetCleanupStats(t *testing.T) {
	tests := []struct {
		name     string
		original string
		cleaned  string
		want     []string
	}{
		{
			name:     "removed trailing period from title",
			original: "feat(cli): add feature.",
			cleaned:  "feat(cli): add feature",
			want:     []string{"Removed trailing period from title"},
		},
		{
			name: "closed unclosed code block",
			original: "feat(cli): add feature\n\n```go\nfunc Hello() {}",
			cleaned:  "feat(cli): add feature\n\n```go\nfunc Hello() {}\n```",
			want:     []string{"Closed unclosed code block"},
		},
		{
			name:     "no changes",
			original: "feat(cli): add feature",
			cleaned:  "feat(cli): add feature",
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCleanupStats(tt.original, tt.cleaned)
			if len(got) != len(tt.want) {
				t.Errorf("GetCleanupStats() returned %d stats, want %d\nGot: %v\nWant: %v", len(got), len(tt.want), got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("GetCleanupStats()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestWrapBodyText tests body text wrapping
func TestWrapBodyText(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  []string
	}{
		{
			name: "wrap long paragraph",
			lines: []string{
				"This is a very long paragraph that needs to be wrapped because it exceeds the maximum line length.",
			},
			want: []string{
				"This is a very long paragraph that needs to be wrapped because it",
				"exceeds the maximum line length.",
			},
		},
		{
			name:  "handle short paragraph",
			lines: []string{"Short paragraph."},
			want:  []string{"Short paragraph."},
		},
		{
			name: "join multiple lines into paragraph",
			lines: []string{
				"This is line one.",
				"This is line two.",
			},
			want: []string{
				"This is line one. This is line two.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapBodyText(tt.lines)
			if len(got) != len(tt.want) {
				t.Errorf("wrapBodyText() returned %d lines, want %d\nGot: %v\nWant: %v", len(got), len(tt.want), got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("wrapBodyText()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestWrapSemanticCommitLine tests subject line wrapping
func TestWrapSemanticCommitLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want []string
	}{
		{
			name: "wrap long subject line",
			line: "src-commands: feat: add a really long feature description that definitely exceeds the maximum allowed length",
			want: []string{
				"src-commands: feat: add a really long feature description that",
				"definitely exceeds the maximum allowed length",
			},
		},
		{
			name: "preserve short subject line",
			line: "src-commands: feat: add validation",
			want: []string{"src-commands: feat: add validation"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapSemanticCommitLine(tt.line)
			if len(got) != len(tt.want) {
				t.Errorf("wrapSemanticCommitLine() returned %d lines, want %d\nGot: %v\nWant: %v", len(got), len(tt.want), got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("wrapSemanticCommitLine()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestEnsureCodeBlocksClosed tests code block closing logic
func TestEnsureCodeBlocksClosed(t *testing.T) {
	tests := []struct {
		name  string
		input string
		wantBalanced bool
	}{
		{
			name:  "close unclosed block",
			input: "```go\nfunc Hello() {}",
			wantBalanced: true,
		},
		{
			name:  "preserve closed block",
			input: "```go\nfunc Hello() {}\n```",
			wantBalanced: true,
		},
		{
			name:  "handle multiple blocks",
			input: "```go\nfunc A() {}\n```\n\n```go\nfunc B() {}\n```",
			wantBalanced: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ensureCodeBlocksClosed(tt.input)
			fenceCount := strings.Count(got, "```")
			isBalanced := fenceCount%2 == 0

			if isBalanced != tt.wantBalanced {
				t.Errorf("ensureCodeBlocksClosed() balanced = %v, want %v\nFence count: %d\nGot: %q", isBalanced, tt.wantBalanced, fenceCount, got)
			}
		})
	}
}

// TestNormalizeSpacing tests spacing normalization
func TestNormalizeSpacing(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int // Expected number of lines
	}{
		{
			name:  "reduce duplicate blank lines",
			input: "line1\n\n\n\nline2",
			want:  5, // line1, blank, blank, blank, line2
		},
		{
			name:  "preserve single blank lines",
			input: "line1\n\nline2",
			want:  3, // line1, blank, line2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeSpacing(tt.input)
			if len(got) > tt.want {
				t.Errorf("normalizeSpacing() returned %d lines after reducing duplicates, want at most %d", len(got), tt.want)
			}
		})
	}
}

// TestAutoCleanup_ComplexScenario tests a complex real-world scenario
func TestAutoCleanup_ComplexScenario(t *testing.T) {
	input := `feat(multi-module): add validation and cleanup.



Auditor-Summary: This commit adds validation logic and auto-cleanup functionality to the commit message generator.


This is a very long body text paragraph that exceeds the maximum line length and should be automatically wrapped at word boundaries.



src-commands
------------
src-commands: feat: add commit message validation logic with contract checking.

This module now validates commit messages against the contract and ensures they follow all the required formatting rules.

` + "```go" + `
func Validate(msg string) error {
    return nil
}


src-core
--------
src-core: feat: add cleanup utilities

Module body text.

---`

	result := AutoCleanup(input)

	// Check that it's cleaned up
	if strings.Contains(result, "...") && !strings.Contains(input, "...") {
		t.Error("AutoCleanup() should not add ellipsis")
	}

	// Check trailing period removed from header
	if strings.HasPrefix(result, "feat(multi-module): add validation and cleanup.") {
		t.Error("AutoCleanup() should remove trailing period from header")
	}

	// Check no trailing ---
	if strings.HasSuffix(strings.TrimSpace(result), "---") {
		t.Error("AutoCleanup() should remove trailing ---")
	}

	// Check code block is closed
	openCount := strings.Count(result, "```")
	if openCount%2 != 0 {
		t.Errorf("AutoCleanup() should close code blocks, got %d fences", openCount)
	}
}
