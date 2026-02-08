package workunit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// =============================================================================
// Context Constants Tests
// =============================================================================

func TestContext_Constants(t *testing.T) {
	tests := []struct {
		name     string
		action   core.ActionType
		expected string
	}{
		{
			name:     "build context",
			action:   core.ActionBuild,
			expected: "build",
		},
		{
			name:     "test context",
			action:   core.ActionTest,
			expected: "test",
		},
		{
			name:     "lint context",
			action:   core.ActionLint,
			expected: "lint",
		},
		{
			name:     "scan context",
			action:   core.ActionScan,
			expected: "scan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.action))
		})
	}
}

// =============================================================================
// Context Type Behavior Tests
// =============================================================================

func TestContext_IsValidContext(t *testing.T) {
	validContexts := []core.ActionType{core.ActionBuild, core.ActionTest, core.ActionLint, core.ActionScan}

	for _, ctx := range validContexts {
		t.Run(string(ctx)+" is valid", func(t *testing.T) {
			// Context should be non-empty and one of the expected values
			assert.NotEmpty(t, string(ctx))
			assert.Contains(t, []string{"build", "test", "lint", "scan"}, string(ctx))
		})
	}
}

// =============================================================================
// Spec Field Tests (BDD Tests: godog, tscucumber)
// =============================================================================

func TestUnitID_Spec_DisplayName(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "spec test returns name: tool format",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "eac-cli",
				Component: "gherkin",
				Tool:      "godog",
				Spec:      "build-module",
			},
			expected: "build-module: godog",
		},
		{
			name: "non-spec test returns component: unit format",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
			},
			expected: "go: unit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.DisplayName())
		})
	}
}

func TestUnitID_Spec_Longname(t *testing.T) {
	// Spec tests use standard Longname format: context:module:component:tool
	// The Spec field is metadata for display only, not for identification
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "spec test uses standard format",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "eac-cli",
				Component: "build-module",
				Tool:      "godog",
				Spec:      "build-module", // Metadata only, doesn't affect Longname
			},
			expected: "test:eac-cli:build-module:godog",
		},
		{
			name: "spec test with different module",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "core",
				Component: "cache-invalidation",
				Tool:      "godog",
				Spec:      "cache-invalidation",
			},
			expected: "test:core:cache-invalidation:godog",
		},
		{
			name: "non-spec test returns standard format",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: "test:core:go:gotest:unit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.Longname())
		})
	}
}

func TestUnitID_Spec_TabLabel(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		maxWidth int
		expected string
	}{
		{
			name:     "short spec fits",
			spec:     "build",
			maxWidth: 10,
			expected: "build",
		},
		{
			name:     "long spec truncated",
			spec:     "build-module-name",
			maxWidth: 10,
			expected: "build-m...",
		},
		{
			name:     "exact fit",
			spec:     "build-mod",
			maxWidth: 9,
			expected: "build-mod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := UnitID{
				Component: "complex-component-path",
				Spec:      tt.spec,
			}
			assert.Equal(t, tt.expected, u.TabLabel(tt.maxWidth))
		})
	}
}

// =============================================================================
// NormalDisplay Tests - Human-Readable Sentences
// =============================================================================

func TestUnitID_NormalDisplay_BuildContext(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "go build - component equals tool",
			unitID: UnitID{
				Action:    core.ActionBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: "Building go in core",
		},
		{
			name: "site build with mkdocs",
			unitID: UnitID{
				Action:    core.ActionBuild,
				Module:    "docs",
				Component: "site",
				Tool:      "mkdocs",
			},
			expected: "Building site in docs with mkdocs",
		},
		{
			name: "build with empty tool",
			unitID: UnitID{
				Action:    core.ActionBuild,
				Module:    "core",
				Component: "go",
				Tool:      "",
			},
			expected: "Building go in core",
		},
		{
			name: "typescript build",
			unitID: UnitID{
				Action:    core.ActionBuild,
				Module:    "mcp",
				Component: "typescript",
				Tool:      "npm",
			},
			expected: "Building typescript in mcp with npm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.NormalDisplay())
		})
	}
}

func TestUnitID_NormalDisplay_TestContext(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "BDD test",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "eac-cli",
				Component: "gherkin",
				Tool:      "godog",
				Spec:      "build-module",
			},
			expected: "Testing spec build-module in eac-cli",
		},
		{
			name: "unit test with testname",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testname": "impl-build"},
			},
			expected: "Testing impl-build in core",
		},
		{
			name: "unit test without testname",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
			},
			expected: "Testing go in core",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.NormalDisplay())
		})
	}
}

func TestUnitID_NormalDisplay_LintContext(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "golangci-lint",
			unitID: UnitID{
				Action:    core.ActionLint,
				Module:    "core",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: "Linting go in core with golangci-lint",
		},
		{
			name: "eslint",
			unitID: UnitID{
				Action:    core.ActionLint,
				Module:    "mcp",
				Component: "typescript",
				Tool:      "eslint",
			},
			expected: "Linting typescript in mcp with eslint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.NormalDisplay())
		})
	}
}

func TestUnitID_NormalDisplay_ScanContext(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "scan with category",
			unitID: UnitID{
				Action:    core.ActionScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-sbom",
				Extra:     map[string]string{"category": "sbom"},
			},
			expected: "Scanning go in core for sbom",
		},
		{
			name: "scan without category",
			unitID: UnitID{
				Action:    core.ActionScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy",
			},
			expected: "Scanning go in core for trivy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.NormalDisplay())
		})
	}
}

func TestUnitID_NormalDisplay_UnknownContext(t *testing.T) {
	unitID := UnitID{
		Action:    "unknown",
		Module:    "mod",
		Component: "comp",
		Tool:      "tool",
	}

	assert.Equal(t, "unknown comp in mod", unitID.NormalDisplay())
}

// =============================================================================
// DisplayKey Tests - Key for Disambiguation
// =============================================================================

func TestUnitID_DisplayKey(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "build context uses component",
			unitID: UnitID{
				Action:    core.ActionBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: "go",
		},
		{
			name: "test context with spec uses spec",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "eac-cli",
				Component: "gherkin",
				Tool:      "godog",
				Spec:      "build-module",
			},
			expected: "build-module",
		},
		{
			name: "test context with testname uses testname",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testname": "impl-build"},
			},
			expected: "impl-build",
		},
		{
			name: "test context without spec or testname uses component",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
			},
			expected: "go",
		},
		{
			name: "lint context uses component",
			unitID: UnitID{
				Action:    core.ActionLint,
				Module:    "core",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: "go",
		},
		{
			name: "scan context uses component",
			unitID: UnitID{
				Action:    core.ActionScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy",
			},
			expected: "go",
		},
		{
			name: "spec takes priority over testname",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "eac-cli",
				Component: "gherkin",
				Tool:      "godog",
				Spec:      "build-module",
				Extra:     map[string]string{"testname": "ignored"},
			},
			expected: "build-module",
		},
		{
			name: "empty testname falls back to component",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testname": ""},
			},
			expected: "go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.DisplayKey())
		})
	}
}

// =============================================================================
// Plan Examples - Comprehensive Display Format Examples
// =============================================================================

func TestUnitID_DisplayFormats_PlanExamples(t *testing.T) {
	examples := []struct {
		description string
		unitID      UnitID
		displayName string // Context-aware display name
		normal      string // Human-readable sentence
	}{
		// Test Context - BDD
		{
			description: "BDD godog test",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "eac-cli",
				Component: "gherkin",
				Tool:      "godog",
				Spec:      "build-module",
			},
			displayName: "build-module: godog",
			normal:      "Testing spec build-module in eac-cli",
		},
		{
			description: "BDD cucumber test",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "core",
				Component: "typescript",
				Tool:      "tscucumber",
				Spec:      "cache-test",
			},
			displayName: "cache-test: tscucumber",
			normal:      "Testing spec cache-test in core",
		},
		// Test Context - Unit
		{
			description: "Unit gotest",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "eac-cli",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testname": "impl-build"},
			},
			displayName: "impl-build: unit",
			normal:      "Testing impl-build in eac-cli",
		},
		{
			description: "Unit mocha",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "mcp",
				Component: "typescript",
				Tool:      "mocha",
				Extra:     map[string]string{"testname": "client"},
			},
			displayName: "client: unit",
			normal:      "Testing client in mcp",
		},
		// Build Context
		{
			description: "Go binary",
			unitID: UnitID{
				Action:    core.ActionBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			displayName: "core: go",
			normal:      "Building go in core",
		},
		{
			description: "Site build",
			unitID: UnitID{
				Action:    core.ActionBuild,
				Module:    "docs",
				Component: "site",
				Tool:      "mkdocs",
			},
			displayName: "docs: site: mkdocs",
			normal:      "Building site in docs with mkdocs",
		},
		{
			description: "PDF build",
			unitID: UnitID{
				Action:    core.ActionBuild,
				Module:    "books",
				Component: "pdf",
				Tool:      "pdf",
			},
			displayName: "books: pdf",
			normal:      "Building pdf in books",
		},
		{
			description: "TypeScript build",
			unitID: UnitID{
				Action:    core.ActionBuild,
				Module:    "mcp",
				Component: "typescript",
				Tool:      "npm",
			},
			displayName: "mcp: typescript: npm",
			normal:      "Building typescript in mcp with npm",
		},
		// Lint Context
		{
			description: "golangci-lint",
			unitID: UnitID{
				Action:    core.ActionLint,
				Module:    "core",
				Component: "go",
				Tool:      "golangci-lint",
			},
			displayName: "lint:go:golangci-lint",
			normal:      "Linting go in core with golangci-lint",
		},
		{
			description: "eslint",
			unitID: UnitID{
				Action:    core.ActionLint,
				Module:    "mcp",
				Component: "typescript",
				Tool:      "eslint",
			},
			displayName: "lint:typescript:eslint",
			normal:      "Linting typescript in mcp with eslint",
		},
		// Scan Context
		{
			description: "SBOM scan",
			unitID: UnitID{
				Action:    core.ActionScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-sbom",
				Extra:     map[string]string{"category": "sbom"},
			},
			displayName: "scan:go:sbom",
			normal:      "Scanning go in core for sbom",
		},
		{
			description: "Vuln scan",
			unitID: UnitID{
				Action:    core.ActionScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-vuln",
				Extra:     map[string]string{"category": "vuln"},
			},
			displayName: "scan:go:vuln",
			normal:      "Scanning go in core for vuln",
		},
	}

	for _, ex := range examples {
		t.Run(ex.description, func(t *testing.T) {
			t.Run("DisplayName", func(t *testing.T) {
				assert.Equal(t, ex.displayName, ex.unitID.DisplayName())
			})
			t.Run("NormalDisplay", func(t *testing.T) {
				assert.Equal(t, ex.normal, ex.unitID.NormalDisplay())
			})
		})
	}
}
