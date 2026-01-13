//go:build L1 && ov
// +build L1,ov

package commitmessage

import (
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/repository"
)

// TestFilterDiffForModule tests git diff filtering for specific modules
// Linked to spec: specs\eac-commands\ai-commit-generation\specification.feature
func TestFilterDiffForModule(t *testing.T) {
	fullDiff := `diff --git a/go/eac/commands/impl/commit/ai.go b/go/eac/commands/impl/commit/ai.go
index 1234567..abcdefg 100644
--- a/go/eac/commands/impl/commit/ai.go
+++ b/go/eac/commands/impl/commit/ai.go
@@ -10,6 +10,7 @@ func CommitAI() int {
+	// New line added
 	return 0
 }
diff --git a/go/eac/ai/executor.go b/go/eac/ai/executor.go
index 9876543..fedcba9 100644
--- a/go/eac/ai/executor.go
+++ b/go/eac/ai/executor.go
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
				{Name: "go/eac/commands/impl/commit/ai.go"},
			},
			want: `diff --git a/go/eac/commands/impl/commit/ai.go b/go/eac/commands/impl/commit/ai.go
index 1234567..abcdefg 100644
--- a/go/eac/commands/impl/commit/ai.go
+++ b/go/eac/commands/impl/commit/ai.go
@@ -10,6 +10,7 @@ func CommitAI() int {
+	// New line added
 	return 0
 }`,
		},
		{
			name: "filters multiple files from module",
			moduleFiles: []repository.RepositoryFileWithModule{
				{Name: "go/eac/commands/impl/commit/ai.go"},
				{Name: "go/eac/ai/executor.go"},
			},
			want: `diff --git a/go/eac/commands/impl/commit/ai.go b/go/eac/commands/impl/commit/ai.go
index 1234567..abcdefg 100644
--- a/go/eac/commands/impl/commit/ai.go
+++ b/go/eac/commands/impl/commit/ai.go
@@ -10,6 +10,7 @@ func CommitAI() int {
+	// New line added
 	return 0
 }
diff --git a/go/eac/ai/executor.go b/go/eac/ai/executor.go
index 9876543..fedcba9 100644
--- a/go/eac/ai/executor.go
+++ b/go/eac/ai/executor.go
@@ -5,4 +5,5 @@ func Execute() {
+	// Another change
 }`,
		},
		{
			name: "returns empty for non-matching files",
			moduleFiles: []repository.RepositoryFileWithModule{
				{Name: "go/eac/other/file.go"},
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
// Linked to spec: specs\eac-commands\ai-commit-generation\specification.feature
func TestExtractModelFromAgent(t *testing.T) {
	tests := []struct {
		name         string
		agentContent string
		want         string
	}{
		{
			name: "extracts model from valid frontmatter",
			agentContent: `---
model: claude-3-haiku-20240307
temperature: 0.7
---

# Agent Instructions

Generate a commit message.`,
			want: "claude-3-haiku-20240307",
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
// Linked to spec: specs\eac-commands\ai-commit-generation\specification.feature
func TestCombineCommitSections(t *testing.T) {
	tests := []struct {
		name           string
		topLevel       string
		moduleSections []string
		want           string
	}{
		{
			name: "combines single-module commit (no module sections)",
			topLevel: `feat(eac-commands): add validation

This commit adds validation logic.`,
			moduleSections: []string{},
			want: `feat(eac-commands): add validation

This commit adds validation logic.`,
		},
		{
			name: "combines multi-module commit with sections",
			topLevel: `feat(multi-module): add features

This affects multiple modules.`,
			moduleSections: []string{
				`eac-commands
------------

eac-commands: feat: add command validation`,
				`eac-core
--------

eac-core: feat: add core validation`,
			},
			want: `feat(multi-module): add features

This affects multiple modules.

eac-commands
------------

eac-commands: feat: add command validation

---

eac-core
--------

eac-core: feat: add core validation`,
		},
		{
			name: "handles single module section",
			topLevel: `chore(multi-module): update

Updates to modules.`,
			moduleSections: []string{
				`eac-commands
------------

eac-commands: chore: update deps`,
			},
			want: `chore(multi-module): update

Updates to modules.

eac-commands
------------

eac-commands: chore: update deps`,
		},
		{
			name: "combines three module sections with separators",
			topLevel: `refactor(multi-module): restructure

Restructuring multiple modules.`,
			moduleSections: []string{
				`module-a
--------

module-a: refactor: simplify`,
				`module-b
--------

module-b: refactor: extract`,
				`module-c
--------

module-c: refactor: rename`,
			},
			want: `refactor(multi-module): restructure

Restructuring multiple modules.

module-a
--------

module-a: refactor: simplify

---

module-b
--------

module-b: refactor: extract

---

module-c
--------

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
// Linked to spec: specs\eac-commands\ai-commit-generation\specification.feature
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
| go/eac/commands/ai.go     | eac-commands |`,
			gitDiff: `diff --git a/go/eac/commands/ai.go b/go/eac/commands/ai.go
+func NewFunc() {}`,
			affectedModules: []string{"eac-commands"},
			wantContains: []string{
				"## Module Count",
				"1 (single-module)",
				"## Affected Modules",
				"- eac-commands",
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
| go/eac/commands/ai.go     | eac-commands |
| go/eac/ai/executor.go     | eac-core     |`,
			gitDiff: `diff --git a/go/eac/commands/ai.go b/go/eac/commands/ai.go
+func NewFunc() {}
diff --git a/go/eac/ai/executor.go b/go/eac/ai/executor.go
+func Execute() {}`,
			affectedModules: []string{"eac-commands", "eac-core"},
			wantContains: []string{
				"## Module Count",
				"2 (multi-module)",
				"## Affected Modules",
				"- eac-commands",
				"- eac-core",
				"## Staged Files",
				"## Git Diff",
				"```diff",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTopLevelContext(tt.stagedFilesTable, tt.gitDiff, "", tt.affectedModules)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("buildTopLevelContext() missing expected content\nWant to contain: %q\nGot:\n%s", want, got)
				}
			}
		})
	}
}

// TestBuildModuleContext tests context building for module agent
// Linked to spec: specs\eac-commands\ai-commit-generation\specification.feature
func TestBuildModuleContext(t *testing.T) {
	fullDiff := `diff --git a/go/eac/commands/ai.go b/go/eac/commands/ai.go
+func NewFunc() {}
diff --git a/go/eac/ai/executor.go b/go/eac/ai/executor.go
+func Execute() {}`

	tests := []struct {
		name         string
		moduleName   string
		moduleFiles  []repository.RepositoryFileWithModule
		wantContains []string
	}{
		{
			name:       "builds context for single file module",
			moduleName: "eac-commands",
			moduleFiles: []repository.RepositoryFileWithModule{
				{Name: "go/eac/commands/ai.go", Modules: []string{"eac-commands"}},
			},
			wantContains: []string{
				"## Module Name",
				"eac-commands",
				"## Files",
				"go/eac/commands/ai.go",
				"## Git Diff",
				"```diff",
				"+func NewFunc() {}",
			},
		},
		{
			name:       "builds context for multi-file module",
			moduleName: "eac-commands",
			moduleFiles: []repository.RepositoryFileWithModule{
				{Name: "go/eac/commands/ai.go", Modules: []string{"eac-commands"}},
				{Name: "go/eac/commands/init.go", Modules: []string{"eac-commands"}},
			},
			wantContains: []string{
				"## Module Name",
				"eac-commands",
				"## Files",
				"go/eac/commands/ai.go",
				"go/eac/commands/init.go",
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
// Linked to spec: specs\eac-commands\ai-commit-generation\specification.feature
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
			commitMessage: `feat(multi-module): add features

Body text.

eac-commands
------------

eac-commands: feat: add validation

eac-core
--------

eac-core: feat: add executor`,
			affectedModules: []string{"eac-commands", "eac-core"},
			allFiles: []repository.RepositoryFileWithModule{
				{Name: "go/eac/commands/ai.go", Modules: []string{"eac-commands"}},
				{Name: "go/eac/ai/executor.go", Modules: []string{"eac-core"}},
			},
			gitDiff: "",
			wantContains: []string{
				"eac-commands\n------------",
				"eac-core\n--------",
			},
			wantNotContains: []string{
				"---\n\neac-commands\n------------", // Should not add duplicate
			},
		},
		{
			name: "adds missing module section",
			commitMessage: `feat(multi-module): add features

Body text.

eac-commands
------------

eac-commands: feat: add validation`,
			affectedModules: []string{"eac-commands", "eac-core"},
			allFiles: []repository.RepositoryFileWithModule{
				{Name: "go/eac/commands/ai.go", Modules: []string{"eac-commands"}},
				{Name: "go/eac/ai/executor.go", Modules: []string{"eac-core"}},
			},
			gitDiff: `diff --git a/go/eac/ai/executor.go b/go/eac/ai/executor.go
+func Execute() {}`,
			wantContains: []string{
				"eac-commands\n------------",
				"eac-core\n---------",
				"eac-core: chore:",
			},
		},
		{
			name: "truncates long file list in stub",
			commitMessage: `feat(multi-module): add features

Body text.`,
			affectedModules: []string{"eac-commands"},
			allFiles: []repository.RepositoryFileWithModule{
				{Name: "go/eac/commands/very/long/path/to/file/that/exceeds/limit.go", Modules: []string{"eac-commands"}},
			},
			gitDiff: "",
			wantContains: []string{
				"eac-commands\n------------",
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
