//go:build L1 && ov
// +build L1,ov

package modules

import (
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
)

func TestNewModuleContract(t *testing.T) {
	base := domain.BaseContract{
		Moniker: "test-module",
		Name:    "Test Module",
		Components: config.ModuleComponents{
			"go": &config.ComponentEntry{
				Root:     "test/root",
				Patterns: &config.ComponentPatterns{Source: []string{"**/*.go"}},
			},
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
		moniker  string
		root     string
		source   []string
		specs    string   // specs component root (repo-root relative)
		specsPat []string // specs component patterns
		expected []string
	}{
		{
			name:     "simple pattern",
			moniker:  "eac-test",
			root:     "go/eac/test",
			source:   []string{"**/*.go"},
			specs:    "specs/eac-test",
			specsPat: []string{"**"},
			expected: []string{"go/eac/test/**/*.go", "specs/eac-test/**"},
		},
		{
			name:     "multiple patterns",
			moniker:  "eac-mcp-vscode",
			root:     "go/eac/mcp/vscode",
			source:   []string{"go.mod", "**.go"},
			specs:    "specs/eac-mcp-vscode",
			specsPat: []string{"**"},
			expected: []string{"go/eac/mcp/vscode/go.mod", "go/eac/mcp/vscode/**.go", "specs/eac-mcp-vscode/**"},
		},
		{
			name:     "specs patterns are repo-root relative",
			moniker:  "clie",
			root:     "go/cli/clie",
			source:   []string{"go.mod", "**.go"},
			specs:    []string{"specs/clie/**"},
			expected: []string{"go/cli/clie/go.mod", "go/cli/clie/**.go", "specs/clie/**"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := config.ModuleComponents{
				"go": &config.ComponentEntry{
					Root:     tt.root,
					Patterns: &config.ComponentPatterns{Source: tt.source},
				},
			}
			if tt.specs != "" {
				components["specs"] = &config.ComponentEntry{
					Root:     tt.specs,
					Patterns: &config.ComponentPatterns{Source: tt.specsPat},
				}
			}

			base := domain.BaseContract{
				Moniker:    tt.moniker,
				Components: components,
			}
			module := NewModuleContract(base, "")

			globs := module.GetGlobPatterns()

			if len(globs) != len(tt.expected) {
				t.Fatalf("Expected %d patterns, got %d: %v", len(tt.expected), len(globs), globs)
			}

			// Check all expected patterns are present (map iteration order is non-deterministic)
			expectedSet := make(map[string]bool, len(tt.expected))
			for _, e := range tt.expected {
				expectedSet[e] = true
			}
			for _, g := range globs {
				if !expectedSet[g] {
					t.Errorf("Unexpected pattern %q in result %v", g, globs)
				}
			}
			for _, e := range tt.expected {
				found := false
				for _, g := range globs {
					if g == e {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected pattern %q not found in result %v", e, globs)
				}
			}
		})
	}
}

func TestModuleContract_MatchesFile(t *testing.T) {
	base := domain.BaseContract{
		Components: config.ModuleComponents{
			"go": &config.ComponentEntry{
				Root: "go/eac/mcp/vscode",
				Patterns: &config.ComponentPatterns{
					Source: []string{"go.mod", "**/*.go"},
				},
			},
		},
	}
	module := NewModuleContract(base, "")

	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{"matches go.mod", "go/eac/mcp/vscode/go.mod", true},
		{"matches go file", "go/eac/mcp/vscode/main.go", true},
		{"matches nested go file", "go/eac/mcp/vscode/sub/test.go", true},
		{"different root", "go/eac/mcp/pwsh/main.go", false},
		{"wrong extension", "go/eac/mcp/vscode/README.md", false},
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
		{"multi pattern exact", "go/eac/mcp/vscode", []string{"go.mod", "**/*.*"}, "go/eac/mcp/vscode/go.mod", true},
		{"multi pattern root file", "go/eac/mcp/vscode", []string{"go.mod", "**/*.*"}, "go/eac/mcp/vscode/main.go", true},
		{"multi pattern nested", "go/eac/mcp/vscode", []string{"go.mod", "**/*.*"}, "go/eac/mcp/vscode/sub/test.go", true},
		{"multi pattern no ext", "go/eac/mcp/vscode", []string{"go.mod", "**/*.*"}, "go/eac/mcp/vscode/Makefile", false},
		{"multi pattern wrong root", "go/eac/mcp/vscode", []string{"go.mod", "**/*.*"}, "go/eac/mcp/pwsh/main.go", false},

		// Edge case: empty root should work
		{"empty root with **/", "", []string{"**/*.md"}, "README.md", true},
		{"empty root nested", "", []string{"**/*.md"}, "docs/README.md", true},

		// Edge case: root="/" should work (repository root)
		{"root slash simple", "/", []string{"agent.md"}, "agent.md", true},
		{"root slash nested", "/", []string{"**/agent.md"}, "docs/agent.md", true},
		{"root slash pattern", "/", []string{".claude/*.json"}, ".claude/mcp.json", true},
		{"root slash no match", "/", []string{"agent.md"}, "OTHER.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := domain.BaseContract{
				Components: config.ModuleComponents{
					"source": &config.ComponentEntry{
						Root: tt.root,
						Patterns: &config.ComponentPatterns{
							Source: tt.includes,
						},
					},
				},
			}
			module := NewModuleContract(base, "")

			got := module.MatchesFile(tt.filePath)
			if got != tt.expected {
				t.Errorf("MatchesFile(%s) = %v, expected %v", tt.filePath, got, tt.expected)
			}
		})
	}
}

// TestModuleContract_MatchesFile_RepoAbsolutePatterns tests that components with
// repo-root-relative paths match files outside the main component root.
func TestModuleContract_MatchesFile_RepoAbsolutePatterns(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		source   []string
		repoRoot string // root of the "other" component (repo-root-relative)
		repoPat  []string
		filePath string
		expected bool
	}{
		{"specs matches file", "go/cli/clie", []string{"specs/clie/**"}, []string{"go.mod"}, "specs/clie/test.feature", true},
		{"specs matches nested", "go/cli/clie", []string{"specs/clie/**"}, []string{"go.mod"}, "specs/clie/design/workspace.dsl", true},
		{"specs doesn't match other specs", "go/cli/clie", []string{"specs/clie/**"}, []string{"go.mod"}, "specs/core/test.feature", false},
		{"source pattern with specs", "go/cli/clie", []string{"specs/clie/**"}, []string{"go.mod"}, "go/cli/clie/go.mod", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := config.ModuleComponents{
				"go": &config.ComponentEntry{
					Root: tt.root,
					Patterns: &config.ComponentPatterns{
						Source: tt.source,
					},
				},
			}
			if tt.repoRoot != "" {
				components["other"] = &config.ComponentEntry{
					Root: tt.repoRoot,
					Patterns: &config.ComponentPatterns{
						Source: tt.repoPat,
					},
				}
			}

			base := domain.BaseContract{
				Components: components,
			}
			module := NewModuleContract(base, "")

			got := module.MatchesFile(tt.filePath)
			if got != tt.expected {
				t.Errorf("MatchesFile(%s) = %v, expected %v", tt.filePath, got, tt.expected)
			}
		})
	}
}

// TestModuleContract_MatchesFile_RepoSpecs tests that specs component patterns match spec files
func TestModuleContract_MatchesFile_RepoSpecs(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		specs    string   // specs component root
		specsPat []string // specs component patterns
		source   []string
		filePath string
		expected bool
	}{
		{"specs matches file", "go/cli/clie", "specs/clie", []string{"**"}, []string{"go.mod"}, "specs/clie/test.feature", true},
		{"specs matches nested", "go/cli/clie", "specs/clie", []string{"**"}, []string{"go.mod"}, "specs/clie/design/workspace.dsl", true},
		{"specs doesn't match other specs", "go/cli/clie", "specs/clie", []string{"**"}, []string{"go.mod"}, "specs/core/test.feature", false},
		{"source pattern with specs", "go/cli/clie", "specs/clie", []string{"**"}, []string{"go.mod"}, "go/cli/clie/go.mod", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := config.ModuleComponents{
				"go": &config.ComponentEntry{
					Root: tt.root,
					Patterns: &config.ComponentPatterns{
						Source: tt.source,
					},
				},
			}
			if tt.specs != "" {
				components["specs"] = &config.ComponentEntry{
					Root: tt.specs,
					Patterns: &config.ComponentPatterns{
						Source: tt.specsPat,
					},
				}
			}

			base := domain.BaseContract{
				Components: components,
			}
			module := NewModuleContract(base, "")

			got := module.MatchesFile(tt.filePath)
			if got != tt.expected {
				t.Errorf("MatchesFile(%s) with root=%q specs=%q source=%v = %v, expected %v",
					tt.filePath, tt.root, tt.specs, tt.source, got, tt.expected)
			}
		})
	}
}

func TestModuleContract_GetDependencies(t *testing.T) {
	base := domain.BaseContract{
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

// Note: UsedBy is no longer stored on ModuleContract.
// It is computed dynamically via Registry.GetUsedBy() from the reverse dependency graph.
// See registry.go GetUsedBy() and GetReverseDependencyGraph() methods.

func TestModuleContract_IsDefinitionsFile(t *testing.T) {
	tests := []struct {
		name     string
		moniker  string
		expected bool
	}{
		{"definitions moniker", "definitions", true},
		{"non-definitions moniker", "test", false},
		{"empty moniker", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := domain.BaseContract{
				Moniker: tt.moniker,
				Components: config.ModuleComponents{
					"go": &config.ComponentEntry{Root: "test"},
				},
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
		{"double star slash wildcard .go", "go/eac/mcp/vscode/main.go", "**/*.go", true},
		{"double star slash wildcard .yml", "contracts/modules/0.1.0/clie.yml", "**/*.yml", true},
		{"double star slash wildcard .md", "docs/guide/getting-started.md", "**/*.md", true},
		{"double star slash wildcard no match", "go/eac/main.txt", "**/*.go", false},
		{"prefix double star slash wildcard", "go/eac/mcp/vscode/main.go", "go/**/*.go", true},
		{"prefix double star slash wildcard no match", "docs/README.go", "go/**/*.go", false},

		// * patterns
		{"single star match", "go/eac/test.go", "go/eac/*.go", true},
		{"single star with path segments", "go/eac/sub/test.go", "go/eac/*/test.go", true},
		{"single star doesn't cross boundaries", "go/eac/sub/deep/test.go", "go/eac/*/test.go", false},

		// Multiple ** segments (now supported)
		{"multiple double star segments", "go/eac/mcp/vscode/test/unit_test.go", "go/**/test/*.go", true},
		{"multiple double star no match", "go/eac/mcp/main.go", "go/**/test/*.go", false},

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
		{"exact match", "go/eac/mcp/vscode/go.mod", "go/eac/mcp/vscode/go.mod", true},
		{"exact no match", "go/eac/mcp/vscode/main.go", "go/eac/mcp/vscode/go.mod", false},

		// Complex real-world patterns
		{"complex: test files", "go/eac/mcp/vscode/module_test.go", "**/*_test.go", true},
		{"complex: specific test dir", "go/eac/mcp/test/integration.go", "go/**/test/*.go", true},
		{"complex: yaml in specific dir", "contracts/modules/0.1.0/clie.yml", "contracts/**/*.yml", true},
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

func TestModuleContract_GetTestImplementationPath(t *testing.T) {
	tests := []struct {
		name         string
		gherkinSteps string
		expected     string
	}{
		{
			name:         "godog component defined",
			gherkinSteps: "go/eac/specs/repository",
			expected:     "go/eac/specs/repository",
		},
		{
			name:         "no godog returns empty",
			gherkinSteps: "",
			expected:     "",
		},
		{
			name:         "nested path",
			gherkinSteps: "go/cli/eac/specs",
			expected:     "go/cli/eac/specs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := config.ModuleComponents{}
			if tt.gherkinSteps != "" {
				components["godog"] = &config.ComponentEntry{Root: tt.gherkinSteps}
			}
			base := domain.BaseContract{
				Moniker:    "test-module",
				Components: components,
			}
			module := NewModuleContract(base, "/workspace")

			got := module.GetTestImplementationPath()
			if got != tt.expected {
				t.Errorf("GetTestImplementationPath() = %q, expected %q", got, tt.expected)
			}
		})
	}
}

func TestModuleContract_GetDesignPath(t *testing.T) {
	tests := []struct {
		name            string
		moniker         string
		design          string // design component root
		expected        string
		useFilepathJoin bool // if true, expected is built with filepath.Join
	}{
		{
			name:     "explicit design path",
			moniker:  "eac",
			design:   "specs/eac/.design",
			expected: "specs/eac/.design",
		},
		{
			name:            "empty design uses default",
			moniker:         "clie",
			design:          "",
			expected:        "", // will be computed with filepath.Join
			useFilepathJoin: true,
		},
		{
			name:     "custom design path",
			moniker:  "my-module",
			design:   "custom/design/path",
			expected: "custom/design/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := config.ModuleComponents{
				"go": &config.ComponentEntry{Root: "src/test"},
			}
			if tt.design != "" {
				components["design"] = &config.ComponentEntry{Root: tt.design}
			}

			base := domain.BaseContract{
				Moniker:    tt.moniker,
				Components: components,
			}
			module := NewModuleContract(base, "/workspace")

			got := module.GetDesignPath()

			expected := tt.expected
			if tt.useFilepathJoin {
				expected = filepath.Join("specs", tt.moniker, ".design")
			}

			if got != expected {
				t.Errorf("GetDesignPath() = %q, expected %q", got, expected)
			}
		})
	}
}
