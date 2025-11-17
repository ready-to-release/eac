package commit

import (
	"strings"
	"testing"

	"github.com/ready-to-release/eac/src/core/repository"
)

// TestStripAgentNoise_TopLevel tests noise removal for top-level agent output
// Linked to spec: specs\src-commands\ai-commit-generation\specification.feature
func TestStripAgentNoise_TopLevel(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		agentType string
		want      string
	}{
		{
			name: "removes initialization message before multi-module commit",
			input: `**Initialized and ready to assist**

# multi-module: feat: add new features

Body text here.`,
			agentType: "top-level",
			want: `# multi-module: feat: add new features

Body text here.`,
		},
		{
			name: "removes INITIALIZED banner before single-module commit",
			input: `**INITIALIZED** ✓

# src-commands: feat: add commit validation

Body text here.`,
			agentType: "top-level",
			want: `# src-commands: feat: add commit validation

Body text here.`,
		},
		{
			name: "removes 'ready to assist' message",
			input: `Loading project context from agent.md
System ready to assist with your tasks.

# cli: fix: resolve bug in parsing

Body text here.`,
			agentType: "top-level",
			want: `# cli: fix: resolve bug in parsing

Body text here.`,
		},
		{
			name: "handles commit with scope in parentheses",
			input: `# src-core: feat(ai): add new provider

Body text here.`,
			agentType: "top-level",
			want: `# src-core: feat(ai): add new provider

Body text here.`,
		},
		{
			name: "removes horizontal rule before content",
			input: `---

# multi-module: chore: update dependencies

Body text here.`,
			agentType: "top-level",
			want: `# multi-module: chore: update dependencies

Body text here.`,
		},
		{
			name: "removes empty lines before content",
			input: `


# src-commands: docs: update documentation

Body text here.`,
			agentType: "top-level",
			want: `# src-commands: docs: update documentation

Body text here.`,
		},
		{
			name: "handles multi-module format without # prefix",
			input: `Some noise here
multi-module: refactor: restructure modules

Body text here.`,
			agentType: "top-level",
			want: `multi-module: refactor: restructure modules

Body text here.`,
		},
		{
			name: "preserves content after valid heading",
			input: `# cli: test: add new test cases

This commit adds comprehensive test coverage for the
new validation logic.

## cli

cli: test: add validation tests`,
			agentType: "top-level",
			want: `# cli: test: add new test cases

This commit adds comprehensive test coverage for the
new validation logic.

## cli

cli: test: add validation tests`,
		},
		{
			name: "skips noise headers like Summary or Context",
			input: `# Summary

Here's what I did...

# src-commands: feat: add feature

Body text here.`,
			agentType: "top-level",
			want: `# src-commands: feat: add feature

Body text here.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripAgentNoise(tt.input, tt.agentType)
			if got != tt.want {
				t.Errorf("stripAgentNoise() failed\nInput:\n%s\n\nGot:\n%s\n\nWant:\n%s", tt.input, got, tt.want)
			}
		})
	}
}

// TestStripAgentNoise_Module tests noise removal for module agent output
// Linked to spec: specs\src-commands\ai-commit-generation\specification.feature
func TestStripAgentNoise_Module(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		agentType string
		want      string
	}{
		{
			name: "removes greeting before module section",
			input: `I'll generate the module section for you.

## src-commands

src-commands: feat: add validation

Body text here.`,
			agentType: "module",
			want: `## src-commands

src-commands: feat: add validation

Body text here.`,
		},
		{
			name: "removes horizontal rule before module section",
			input: `---

## cli

cli: fix: resolve parsing bug

Fixed the issue with argument parsing.`,
			agentType: "module",
			want: `## cli

cli: fix: resolve parsing bug

Fixed the issue with argument parsing.`,
		},
		{
			name: "preserves module section with code blocks",
			input: `## src-core

src-core: refactor: simplify logic

` + "```go" + `
+func NewValidator() *Validator {
+	return &Validator{}
+}
` + "```",
			agentType: "module",
			want: `## src-core

src-core: refactor: simplify logic

` + "```go" + `
+func NewValidator() *Validator {
+	return &Validator{}
+}
` + "```",
		},
		{
			name: "rejects module section with separator after ##",
			input: `## ---

Some content`,
			agentType: "module",
			want: `## ---

Some content`,
		},
		{
			name: "handles empty lines before module section",
			input: `


## src-commands

src-commands: chore: update deps`,
			agentType: "module",
			want: `## src-commands

src-commands: chore: update deps`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripAgentNoise(tt.input, tt.agentType)
			if got != tt.want {
				t.Errorf("stripAgentNoise() failed\nInput:\n%s\n\nGot:\n%s\n\nWant:\n%s", tt.input, got, tt.want)
			}
		})
	}
}

// TestStripAgentNoiseHardcoded_Fallback tests the hardcoded fallback logic
// Linked to spec: specs\src-commands\ai-commit-generation\specification.feature
func TestStripAgentNoiseHardcoded_Fallback(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		agentType string
		want      string
	}{
		{
			name: "fallback handles standard noise patterns",
			input: `**Initialized and ready to assist**

# cli: feat: add feature

Body text.`,
			agentType: "top-level",
			want: `# cli: feat: add feature

Body text.`,
		},
		{
			name: "fallback removes horizontal rules",
			input: `---

## src-commands

Module content.`,
			agentType: "module",
			want: `## src-commands

Module content.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripAgentNoiseHardcoded(tt.input, tt.agentType)
			if got != tt.want {
				t.Errorf("stripAgentNoiseHardcoded() failed\nGot:\n%s\n\nWant:\n%s", got, tt.want)
			}
		})
	}
}

// TestFilterDiffForModule tests git diff filtering for specific modules
// Linked to spec: specs\src-commands\ai-commit-generation\specification.feature
func TestFilterDiffForModule(t *testing.T) {
	fullDiff := `diff --git a/src/commands/impl/commit/ai.go b/src/commands/impl/commit/ai.go
index 1234567..abcdefg 100644
--- a/src/commands/impl/commit/ai.go
+++ b/src/commands/impl/commit/ai.go
@@ -10,6 +10,7 @@ func CommitAI() int {
+	// New line added
 	return 0
 }
diff --git a/src/core/ai/executor.go b/src/core/ai/executor.go
index 9876543..fedcba9 100644
--- a/src/core/ai/executor.go
+++ b/src/core/ai/executor.go
@@ -5,4 +5,5 @@ func Execute() {
+	// Another change
 }
diff --git a/README.md b/README.md
index aaa1111..bbb2222 100644
--- a/README.md
+++ b/README.md
@@ -1,2 +1,3 @@
 # Project
+New line in readme`

	tests := []struct {
		name        string
		moduleFiles []repository.RepositoryFileWithModule
		want        string
	}{
		{
			name: "filters single file from module",
			moduleFiles: []repository.RepositoryFileWithModule{
				{Name: "src/commands/impl/commit/ai.go"},
			},
			want: `diff --git a/src/commands/impl/commit/ai.go b/src/commands/impl/commit/ai.go
index 1234567..abcdefg 100644
--- a/src/commands/impl/commit/ai.go
+++ b/src/commands/impl/commit/ai.go
@@ -10,6 +10,7 @@ func CommitAI() int {
+	// New line added
 	return 0
 }`,
		},
		{
			name: "filters multiple files from module",
			moduleFiles: []repository.RepositoryFileWithModule{
				{Name: "src/commands/impl/commit/ai.go"},
				{Name: "src/core/ai/executor.go"},
			},
			want: `diff --git a/src/commands/impl/commit/ai.go b/src/commands/impl/commit/ai.go
index 1234567..abcdefg 100644
--- a/src/commands/impl/commit/ai.go
+++ b/src/commands/impl/commit/ai.go
@@ -10,6 +10,7 @@ func CommitAI() int {
+	// New line added
 	return 0
 }
diff --git a/src/core/ai/executor.go b/src/core/ai/executor.go
index 9876543..fedcba9 100644
--- a/src/core/ai/executor.go
+++ b/src/core/ai/executor.go
@@ -5,4 +5,5 @@ func Execute() {
+	// Another change
 }`,
		},
		{
			name: "returns empty for non-matching files",
			moduleFiles: []repository.RepositoryFileWithModule{
				{Name: "src/other/file.go"},
			},
			want: "",
		},
		{
			name:        "returns empty for empty module files list",
			moduleFiles: []repository.RepositoryFileWithModule{},
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterDiffForModule(fullDiff, tt.moduleFiles)
			if got != tt.want {
				t.Errorf("filterDiffForModule() failed\nGot:\n%s\n\nWant:\n%s", got, tt.want)
			}
		})
	}
}

// TestExtractModelFromAgent tests frontmatter parsing for model configuration
// Linked to spec: specs\src-commands\ai-commit-generation\specification.feature
func TestExtractModelFromAgent(t *testing.T) {
	tests := []struct {
		name         string
		agentContent string
		want         string
	}{
		{
			name: "extracts model from valid frontmatter",
			agentContent: `---
model: claude-3-5-sonnet-20241022
temperature: 0.7
---

# Agent Instructions

Generate a commit message.`,
			want: "claude-3-5-sonnet-20241022",
		},
		{
			name: "extracts model with extra whitespace",
			agentContent: `---
model:   gpt-4
---

Instructions here.`,
			want: "gpt-4",
		},
		{
			name: "returns empty when no model field",
			agentContent: `---
temperature: 0.7
max_tokens: 1000
---

Instructions here.`,
			want: "",
		},
		{
			name: "returns empty when no frontmatter",
			agentContent: `# Agent Instructions

No frontmatter here.`,
			want: "",
		},
		{
			name: "ignores model field outside frontmatter",
			agentContent: `---
temperature: 0.7
---

model: should-not-match

Instructions here.`,
			want: "",
		},
		{
			name: "handles empty frontmatter",
			agentContent: `---
---

Instructions here.`,
			want: "",
		},
		{
			name: "handles malformed frontmatter",
			agentContent: `---
this is not yaml
model: should-still-extract
---

Instructions.`,
			want: "should-still-extract",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractModelFromAgent(tt.agentContent)
			if got != tt.want {
				t.Errorf("extractModelFromAgent() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCombineCommitSections tests combining top-level and module sections
// Linked to spec: specs\src-commands\ai-commit-generation\specification.feature
func TestCombineCommitSections(t *testing.T) {
	tests := []struct {
		name           string
		topLevel       string
		moduleSections []string
		want           string
	}{
		{
			name: "combines single-module commit (no module sections)",
			topLevel: `# src-commands: feat: add validation

This commit adds validation logic.`,
			moduleSections: []string{},
			want: `# src-commands: feat: add validation

This commit adds validation logic.`,
		},
		{
			name: "combines multi-module commit with sections",
			topLevel: `# multi-module: feat: add features

This affects multiple modules.`,
			moduleSections: []string{
				`## src-commands

src-commands: feat: add command validation`,
				`## src-core

src-core: feat: add core validation`,
			},
			want: `# multi-module: feat: add features

This affects multiple modules.

## src-commands

src-commands: feat: add command validation

---

## src-core

src-core: feat: add core validation`,
		},
		{
			name: "handles single module section",
			topLevel: `# multi-module: chore: update

Updates to modules.`,
			moduleSections: []string{
				`## src-commands

src-commands: chore: update deps`,
			},
			want: `# multi-module: chore: update

Updates to modules.

## src-commands

src-commands: chore: update deps`,
		},
		{
			name: "combines three module sections with separators",
			topLevel: `# multi-module: refactor: restructure

Restructuring multiple modules.`,
			moduleSections: []string{
				`## module-a

module-a: refactor: simplify`,
				`## module-b

module-b: refactor: extract`,
				`## module-c

module-c: refactor: rename`,
			},
			want: `# multi-module: refactor: restructure

Restructuring multiple modules.

## module-a

module-a: refactor: simplify

---

## module-b

module-b: refactor: extract

---

## module-c

module-c: refactor: rename`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := combineCommitSections(tt.topLevel, tt.moduleSections)
			if got != tt.want {
				t.Errorf("combineCommitSections() failed\nGot:\n%s\n\nWant:\n%s", got, tt.want)
			}
		})
	}
}

// TestBuildTopLevelContext tests context building for top-level agent
// Linked to spec: specs\src-commands\ai-commit-generation\specification.feature
func TestBuildTopLevelContext(t *testing.T) {
	tests := []struct {
		name             string
		stagedFilesTable string
		gitDiff          string
		affectedModules  []string
		wantContains     []string
	}{
		{
			name: "builds context for single-module commit",
			stagedFilesTable: `| File                  | Modules      |
|---------------------------|--------------|
| src/commands/ai.go        | src-commands |`,
			gitDiff: `diff --git a/src/commands/ai.go b/src/commands/ai.go
+func NewFunc() {}`,
			affectedModules: []string{"src-commands"},
			wantContains: []string{
				"## Module Count",
				"1 (single-module)",
				"## Affected Modules",
				"- src-commands",
				"## Staged Files",
				"## Git Diff",
				"```diff",
				"+func NewFunc() {}",
			},
		},
		{
			name: "builds context for multi-module commit",
			stagedFilesTable: `| File                  | Modules      |
|---------------------------|--------------|
| src/commands/ai.go        | src-commands |
| src/core/executor.go      | src-core     |`,
			gitDiff: `diff --git a/src/commands/ai.go b/src/commands/ai.go
+func NewFunc() {}
diff --git a/src/core/executor.go b/src/core/executor.go
+func Execute() {}`,
			affectedModules: []string{"src-commands", "src-core"},
			wantContains: []string{
				"## Module Count",
				"2 (multi-module)",
				"## Affected Modules",
				"- src-commands",
				"- src-core",
				"## Staged Files",
				"## Git Diff",
				"```diff",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTopLevelContext(tt.stagedFilesTable, tt.gitDiff, tt.affectedModules)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("buildTopLevelContext() missing expected content\nWant to contain: %q\nGot:\n%s", want, got)
				}
			}
		})
	}
}

// TestBuildModuleContext tests context building for module agent
// Linked to spec: specs\src-commands\ai-commit-generation\specification.feature
func TestBuildModuleContext(t *testing.T) {
	fullDiff := `diff --git a/src/commands/ai.go b/src/commands/ai.go
+func NewFunc() {}
diff --git a/src/core/executor.go b/src/core/executor.go
+func Execute() {}`

	tests := []struct {
		name        string
		moduleName  string
		moduleFiles []repository.RepositoryFileWithModule
		wantContains []string
	}{
		{
			name:       "builds context for single file module",
			moduleName: "src-commands",
			moduleFiles: []repository.RepositoryFileWithModule{
				{Name: "src/commands/ai.go", Modules: []string{"src-commands"}},
			},
			wantContains: []string{
				"## Module Name",
				"src-commands",
				"## Files",
				"src/commands/ai.go",
				"## Git Diff",
				"```diff",
				"+func NewFunc() {}",
			},
		},
		{
			name:       "builds context for multi-file module",
			moduleName: "src-commands",
			moduleFiles: []repository.RepositoryFileWithModule{
				{Name: "src/commands/ai.go", Modules: []string{"src-commands"}},
				{Name: "src/commands/init.go", Modules: []string{"src-commands"}},
			},
			wantContains: []string{
				"## Module Name",
				"src-commands",
				"## Files",
				"src/commands/ai.go",
				"src/commands/init.go",
				"## Git Diff",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildModuleContext(tt.moduleName, tt.moduleFiles, fullDiff)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("buildModuleContext() missing expected content\nWant to contain: %q\nGot:\n%s", want, got)
				}
			}
		})
	}
}

// TestAddMissingModules tests stub generation for missing module sections
// Linked to spec: specs\src-commands\ai-commit-generation\specification.feature
func TestAddMissingModules(t *testing.T) {
	tests := []struct {
		name            string
		commitMessage   string
		affectedModules []string
		allFiles        []repository.RepositoryFileWithModule
		gitDiff         string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "returns original when no modules missing",
			commitMessage: `# multi-module: feat: add features

Body text.

## src-commands

src-commands: feat: add validation

## src-core

src-core: feat: add executor`,
			affectedModules: []string{"src-commands", "src-core"},
			allFiles: []repository.RepositoryFileWithModule{
				{Name: "src/commands/ai.go", Modules: []string{"src-commands"}},
				{Name: "src/core/executor.go", Modules: []string{"src-core"}},
			},
			gitDiff: "",
			wantContains: []string{
				"## src-commands",
				"## src-core",
			},
			wantNotContains: []string{
				"---\n\n## src-commands", // Should not add duplicate
			},
		},
		{
			name: "adds missing module section",
			commitMessage: `# multi-module: feat: add features

Body text.

## src-commands

src-commands: feat: add validation`,
			affectedModules: []string{"src-commands", "src-core"},
			allFiles: []repository.RepositoryFileWithModule{
				{Name: "src/commands/ai.go", Modules: []string{"src-commands"}},
				{Name: "src/core/executor.go", Modules: []string{"src-core"}},
			},
			gitDiff: `diff --git a/src/core/executor.go b/src/core/executor.go
+func Execute() {}`,
			wantContains: []string{
				"## src-commands",
				"---\n\n## src-core",
				"src-core: chore:",
			},
		},
		{
			name: "truncates long file list in stub",
			commitMessage: `# multi-module: feat: add features

Body text.`,
			affectedModules: []string{"src-commands"},
			allFiles: []repository.RepositoryFileWithModule{
				{Name: "src/commands/very/long/path/to/file/that/exceeds/limit.go", Modules: []string{"src-commands"}},
			},
			gitDiff: "",
			wantContains: []string{
				"## src-commands",
				"...", // Truncated file name
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addMissingModules(tt.commitMessage, tt.affectedModules, tt.allFiles, tt.gitDiff)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("addMissingModules() missing expected content\nWant to contain: %q\nGot:\n%s", want, got)
				}
			}

			for _, wantNot := range tt.wantNotContains {
				if strings.Contains(got, wantNot) {
					t.Errorf("addMissingModules() contains unexpected content\nShould not contain: %q\nGot:\n%s", wantNot, got)
				}
			}
		})
	}
}
