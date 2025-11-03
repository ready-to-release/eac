package modules

import (
	"testing"

	"github.com/ready-to-release/eac/src/contracts"
)

func TestNewModuleContract(t *testing.T) {
	base := contracts.BaseContract{
		Moniker: "test-module",
		Name:    "Test Module",
		Type:    "test-type",
		Source: contracts.Source{
			Root:     "test/root",
			Includes: []string{"**/*.go"},
		},
	}

	module := NewModuleContract(base, "/workspace")

	if module.GetMoniker() != "test-module" {
		t.Errorf("Expected moniker 'test-module', got '%s'", module.GetMoniker())
	}

	if module.workspaceRoot != "/workspace" {
		t.Errorf("Expected workspace root '/workspace', got '%s'", module.workspaceRoot)
	}
}

func TestModuleContract_GetGlobPatterns(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		includes []string
		expected []string
	}{
		{
			name:     "simple pattern",
			root:     "src/test",
			includes: []string{"**/*.go"},
			expected: []string{"src/test/**/*.go"},
		},
		{
			name:     "multiple patterns",
			root:     "src/mcp/vscode",
			includes: []string{"go.mod", "**.go"},
			expected: []string{"src/mcp/vscode/go.mod", "src/mcp/vscode/**.go"},
		},
		{
			name:     "pattern with root prefix",
			root:     "src/test",
			includes: []string{"src/test/*.go"},
			expected: []string{"src/test/*.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := contracts.BaseContract{
				Source: contracts.Source{
					Root:     tt.root,
					Includes: tt.includes,
				},
			}
			module := NewModuleContract(base, "")

			globs := module.GetGlobPatterns()

			if len(globs) != len(tt.expected) {
				t.Fatalf("Expected %d patterns, got %d", len(tt.expected), len(globs))
			}

			for i, expected := range tt.expected {
				if globs[i] != expected {
					t.Errorf("Pattern %d: expected '%s', got '%s'", i, expected, globs[i])
				}
			}
		})
	}
}

func TestModuleContract_MatchesFile(t *testing.T) {
	base := contracts.BaseContract{
		Source: contracts.Source{
			Root:     "src/mcp/vscode",
			Includes: []string{"go.mod", "**.go"},
		},
	}
	module := NewModuleContract(base, "")

	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{"matches go.mod", "src/mcp/vscode/go.mod", true},
		{"matches go file", "src/mcp/vscode/main.go", true},
		{"matches nested go file", "src/mcp/vscode/sub/test.go", true},
		{"different root", "src/mcp/pwsh/main.go", false},
		{"wrong extension", "src/mcp/vscode/README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := module.MatchesFile(tt.filePath)
			if got != tt.expected {
				t.Errorf("MatchesFile(%s) = %v, expected %v", tt.filePath, got, tt.expected)
			}
		})
	}
}

// TestModuleContract_MatchesFile_RootLevel tests that **/ patterns correctly match root-level files
func TestModuleContract_MatchesFile_RootLevel(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		includes []string
		filePath string
		expected bool
	}{
		// **/*.* patterns should match files with extensions at any level including root
		{"**/*.* matches root file", "docs", []string{"**/*.*"}, "docs/README.md", true},
		{"**/*.* matches nested file", "docs", []string{"**/*.*"}, "docs/guide/getting-started.md", true},
		{"**/*.* matches deep nested", "docs", []string{"**/*.*"}, "docs/sub/deep/nested/file.txt", true},
		{"**/*.* rejects no extension", "docs", []string{"**/*.*"}, "docs/noextension", false},
		{"**/*.* rejects different root", "docs", []string{"**/*.*"}, "src/test.go", false},

		// **/* patterns should match all files at any level including root
		{"**/* matches root file", "src", []string{"**/*"}, "src/main.go", true},
		{"**/* matches nested file", "src", []string{"**/*"}, "src/sub/test.go", true},
		{"**/* matches no extension", "src", []string{"**/*"}, "src/Makefile", true},
		{"**/* matches deep nested", "src", []string{"**/*"}, "src/deep/very/deep/file", true},
		{"**/* rejects different root", "src", []string{"**/*"}, "docs/test.md", false},

		// **/*.ext patterns should match specific extensions at any level
		{"**/*.go matches root", "src", []string{"**/*.go"}, "src/main.go", true},
		{"**/*.go matches nested", "src", []string{"**/*.go"}, "src/sub/test.go", true},
		{"**/*.go rejects wrong ext", "src", []string{"**/*.go"}, "src/README.md", false},

		// *.* patterns should only match root-level files with extensions
		{"*.* matches root file", "contracts", []string{"*.*"}, "contracts/README.md", true},
		{"*.* rejects nested file", "contracts", []string{"*.*"}, "contracts/sub/file.md", false},
		{"*.* rejects no extension", "contracts", []string{"*.*"}, "contracts/noextension", false},

		// Multiple patterns including **/ should work
		{"multi pattern exact", "src/mcp/vscode", []string{"go.mod", "**/*.*"}, "src/mcp/vscode/go.mod", true},
		{"multi pattern root file", "src/mcp/vscode", []string{"go.mod", "**/*.*"}, "src/mcp/vscode/main.go", true},
		{"multi pattern nested", "src/mcp/vscode", []string{"go.mod", "**/*.*"}, "src/mcp/vscode/sub/test.go", true},
		{"multi pattern no ext", "src/mcp/vscode", []string{"go.mod", "**/*.*"}, "src/mcp/vscode/Makefile", false},
		{"multi pattern wrong root", "src/mcp/vscode", []string{"go.mod", "**/*.*"}, "src/mcp/pwsh/main.go", false},

		// Edge case: empty root should work
		{"empty root with **/", "", []string{"**/*.md"}, "README.md", true},
		{"empty root nested", "", []string{"**/*.md"}, "docs/README.md", true},

		// Edge case: root="/" should work (repository root)
		{"root slash simple", "/", []string{"CLAUDE.md"}, "CLAUDE.md", true},
		{"root slash nested", "/", []string{"**/CLAUDE.md"}, "docs/CLAUDE.md", true},
		{"root slash pattern", "/", []string{".claude/*.json"}, ".claude/mcp.json", true},
		{"root slash no match", "/", []string{"CLAUDE.md"}, "OTHER.md", false},

		// Edge case: absolute patterns (leading /) should match from repository root
		{"absolute simple", "src/mcp/vscode", []string{"/requirements/spec.md"}, "requirements/spec.md", true},
		{"absolute with **", "src/mcp/vscode", []string{"/requirements/**/*.md"}, "requirements/api/spec.md", true},
		{"absolute with *", "src/mcp/vscode", []string{"/requirements/*"}, "requirements/spec.md", true},
		{"absolute no match", "src/mcp/vscode", []string{"/requirements/*.md"}, "other/spec.md", false},
		{"absolute and relative", "src/mcp/vscode", []string{"go.mod", "/requirements/*.md"}, "src/mcp/vscode/go.mod", true},
		{"absolute and relative 2", "src/mcp/vscode", []string{"go.mod", "/requirements/*.md"}, "requirements/spec.md", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := contracts.BaseContract{
				Source: contracts.Source{
					Root:     tt.root,
					Includes: tt.includes,
				},
			}
			module := NewModuleContract(base, "")

			got := module.MatchesFile(tt.filePath)
			if got != tt.expected {
				t.Errorf("MatchesFile(%s) with root=%q includes=%v = %v, expected %v",
					tt.filePath, tt.root, tt.includes, got, tt.expected)
			}
		})
	}
}

func TestModuleContract_GetDependencies(t *testing.T) {
	base := contracts.BaseContract{
		DependsOn: []string{"dep1", "dep2"},
	}
	module := NewModuleContract(base, "")

	deps := module.GetDependencies()

	if len(deps) != 2 {
		t.Fatalf("Expected 2 dependencies, got %d", len(deps))
	}

	if deps[0] != "dep1" || deps[1] != "dep2" {
		t.Error("Dependencies do not match expected values")
	}
}

func TestModuleContract_GetUsedBy(t *testing.T) {
	base := contracts.BaseContract{
		UsedBy: []string{"user1", "user2"},
	}
	module := NewModuleContract(base, "")

	users := module.GetUsedBy()

	if len(users) != 2 {
		t.Fatalf("Expected 2 users, got %d", len(users))
	}

	if users[0] != "user1" || users[1] != "user2" {
		t.Error("UsedBy does not match expected values")
	}
}

func TestModuleContract_IsDefinitionsFile(t *testing.T) {
	tests := []struct {
		name     string
		moniker  string
		typ      string
		expected bool
	}{
		{"definitions moniker", "definitions", "test", true},
		{"definitions type", "test", "definitions-type", true},
		{"neither", "test", "test-type", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := contracts.BaseContract{
				Moniker: tt.moniker,
				Type:    tt.typ,
			}
			module := NewModuleContract(base, "")

			got := module.IsDefinitionsFile()
			if got != tt.expected {
				t.Errorf("IsDefinitionsFile() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func Test_matchGlobPattern(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		// ** patterns (basic)
		{"double star prefix only", "src/test/file.go", "src/**", true},
		{"double star no match prefix", "docs/README.md", "src/**", false},
		{"double star suffix .go", "src/test/file.go", "**.go", true},
		{"double star suffix .md", "src/test/file.md", "**.go", false},

		// **/*.ext patterns (previously broken, now fixed)
		{"double star slash wildcard .go", "src/mcp/vscode/main.go", "**/*.go", true},
		{"double star slash wildcard .yml", "contracts/modules/0.1.0/src-cli.yml", "**/*.yml", true},
		{"double star slash wildcard .md", "docs/guide/getting-started.md", "**/*.md", true},
		{"double star slash wildcard no match", "src/main.txt", "**/*.go", false},
		{"prefix double star slash wildcard", "src/mcp/vscode/main.go", "src/**/*.go", true},
		{"prefix double star slash wildcard no match", "docs/README.go", "src/**/*.go", false},

		// * patterns
		{"single star match", "src/test.go", "src/*.go", true},
		{"single star with path segments", "src/sub/test.go", "src/*/test.go", true},
		{"single star doesn't cross boundaries", "src/sub/deep/test.go", "src/*/test.go", false},

		// Multiple ** segments (now supported)
		{"multiple double star segments", "src/mcp/vscode/test/unit_test.go", "src/**/test/*.go", true},
		{"multiple double star no match", "src/mcp/main.go", "src/**/test/*.go", false},

		// ? single character wildcard (now supported)
		{"single char wildcard match", "file1.go", "file?.go", true},
		{"single char wildcard match letter", "fileA.go", "file?.go", true},
		{"single char wildcard no match multiple", "file12.go", "file?.go", false},
		{"single char wildcard no match", "file.go", "file?.go", false},

		// Character classes (now supported)
		{"character class match first", "file1.go", "file[123].go", true},
		{"character class match middle", "file2.go", "file[123].go", true},
		{"character class match last", "file3.go", "file[123].go", true},
		{"character class no match", "file4.go", "file[123].go", false},
		{"character range match", "fileA.go", "file[A-Z].go", true},
		{"character range no match", "filea.go", "file[A-Z].go", false},

		// Negation in character classes (now supported)
		{"character class negation no match", "test.go", "[!t]*.go", false},
		{"character class negation match", "main.go", "[!t]*.go", true},

		// Exact match
		{"exact match", "src/mcp/vscode/go.mod", "src/mcp/vscode/go.mod", true},
		{"exact no match", "src/mcp/vscode/main.go", "src/mcp/vscode/go.mod", false},

		// Complex real-world patterns
		{"complex: test files", "src/mcp/vscode/module_test.go", "**/*_test.go", true},
		{"complex: specific test dir", "src/mcp/test/integration.go", "src/**/test/*.go", true},
		{"complex: yaml in specific dir", "contracts/modules/0.1.0/src-cli.yml", "contracts/**/*.yml", true},
		{"complex: markdown docs", ".claude/agents/boot.md", ".claude/**/*.md", true},

		// Edge cases
		{"empty pattern", "src/test.go", "", false},
		{"root level wildcard", "test.go", "*.go", true},
		{"deep nesting", "a/b/c/d/e/f/test.go", "a/**/test.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchGlobPattern(tt.path, tt.pattern)
			if got != tt.expected {
				t.Errorf("matchGlobPattern(%q, %q) = %v, expected %v",
					tt.path, tt.pattern, got, tt.expected)
			}
		})
	}
}

func Test_normalizePathSeparators(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"forward slashes", "src/test/file.go", "src/test/file.go"},
		{"backslashes", "src\\test\\file.go", "src/test/file.go"},
		{"mixed", "src/test\\sub/file.go", "src/test/sub/file.go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizePathSeparators(tt.path)
			if got != tt.expected {
				t.Errorf("normalizePathSeparators(%q) = %q, expected %q",
					tt.path, got, tt.expected)
			}
		})
	}
}
