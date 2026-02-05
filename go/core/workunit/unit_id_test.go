package workunit

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Context Constants Tests
// =============================================================================

func TestContext_Constants(t *testing.T) {
	tests := []struct {
		name     string
		context  Context
		expected string
	}{
		{
			name:     "build context",
			context:  ContextBuild,
			expected: "build",
		},
		{
			name:     "test context",
			context:  ContextTest,
			expected: "test",
		},
		{
			name:     "lint context",
			context:  ContextLint,
			expected: "lint",
		},
		{
			name:     "scan context",
			context:  ContextScan,
			expected: "scan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.context))
		})
	}
}

// =============================================================================
// DisplayName Tests
// =============================================================================

func TestUnitID_DisplayName(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "build - component equals tool",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: "core: go",
		},
		{
			name: "build - component differs from tool",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "docs",
				Component: "site",
				Tool:      "mkdocs",
			},
			expected: "docs: site: mkdocs",
		},
		{
			name: "test - BDD spec",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-cli",
				Component: "gherkin",
				Tool:      "godog",
				Spec:      "build-module",
			},
			expected: "build-module: godog",
		},
		{
			name: "test - unit with testname",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testname": "impl-build"},
			},
			expected: "impl-build: unit",
		},
		{
			name: "test - unit without testname uses component",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
			},
			expected: "go: unit",
		},
		{
			name: "lint",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "core",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: "lint:go:golangci-lint",
		},
		{
			name: "scan with category",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-sbom",
				Extra:     map[string]string{"category": "sbom"},
			},
			expected: "scan:go:sbom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.DisplayName())
		})
	}
}

// =============================================================================
// Longname Tests - Basic (No Extra)
// =============================================================================

func TestUnitID_Longname_Basic(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "build:core:go:go",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: "build:core:go:go",
		},
		{
			name: "lint:core:go:golangci-lint",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "core",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: "lint:core:go:golangci-lint",
		},
		{
			name: "scan:core:go:trivy-vuln",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expected: "scan:core:go:trivy-vuln",
		},
		{
			name: "build with empty extra returns base format",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "module-a",
				Component: "comp",
				Tool:      "tool",
				Extra:     map[string]string{},
			},
			expected: "build:module-a:comp:tool",
		},
		{
			name: "build with nil extra returns base format",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "module-a",
				Component: "comp",
				Tool:      "tool",
				Extra:     nil,
			},
			expected: "build:module-a:comp:tool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.Longname())
		})
	}
}

// =============================================================================
// Longname Tests - With Extra (Test Context)
// =============================================================================

func TestUnitID_Longname_WithTestset(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "test:core:go:gotest:unit",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: "test:core:go:gotest:unit",
		},
		{
			name: "test:core:go:gotest:integration",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "integration"},
			},
			expected: "test:core:go:gotest:integration",
		},
		{
			name: "test with gherkin component",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-cli",
				Component: "gherkin",
				Tool:      "godog",
				Extra:     map[string]string{"testset": "acceptance"},
			},
			expected: "test:eac-cli:gherkin:godog:acceptance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.Longname())
		})
	}
}

// =============================================================================
// String Tests
// =============================================================================

func TestUnitID_String(t *testing.T) {
	tests := []struct {
		name   string
		unitID UnitID
	}{
		{
			name: "build unit string equals longname",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
		},
		{
			name: "test unit string equals longname",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
		},
		{
			name: "lint unit string equals longname",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "shared-lib",
				Component: "go",
				Tool:      "golangci-lint",
			},
		},
		{
			name: "scan unit string equals longname",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.unitID.Longname(), tt.unitID.String(),
				"String() should return the same value as Longname()")
		})
	}
}

// =============================================================================
// OutDir Tests
// =============================================================================

func TestUnitID_OutDir(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "build output directory",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: filepath.Join("out", "build", "core", "go-go"),
		},
		{
			name: "lint output directory",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "core",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: filepath.Join("out", "lint", "core", "go-golangci-lint"),
		},
		{
			name: "scan output directory",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expected: filepath.Join("out", "scan", "core", "go-trivy-vuln"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.OutDir())
		})
	}
}

func TestUnitID_OutDir_WithTestset(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "test unit output directory includes testset",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: filepath.Join("out", "test", "core", "go-gotest-unit"),
		},
		{
			name: "test integration output directory",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "integration"},
			},
			expected: filepath.Join("out", "test", "core", "go-gotest-integration"),
		},
		{
			name: "test gherkin output directory",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-cli",
				Component: "gherkin",
				Tool:      "godog",
				Extra:     map[string]string{"testset": "acceptance"},
			},
			expected: filepath.Join("out", "test", "eac-cli", "gherkin-godog-acceptance"),
		},
		{
			name: "test without testset falls back to base path",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     nil,
			},
			expected: filepath.Join("out", "test", "core", "go-gotest"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.OutDir())
		})
	}
}

// =============================================================================
// LockFile Tests
// =============================================================================

func TestUnitID_LockFile(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "build lock file",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: filepath.Join("out", "build", "core", "go-go", ".lock"),
		},
		{
			name: "test lock file with testset",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: filepath.Join("out", "test", "core", "go-gotest-unit", ".lock"),
		},
		{
			name: "lint lock file",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "shared-lib",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: filepath.Join("out", "lint", "shared-lib", "go-golangci-lint", ".lock"),
		},
		{
			name: "scan lock file",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expected: filepath.Join("out", "scan", "core", "go-trivy-vuln", ".lock"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.LockFile())
		})
	}
}

// =============================================================================
// StateFile Tests
// =============================================================================

func TestUnitID_StateFile(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "build state file",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: filepath.Join("out", "build", "core", "go-go", "state.json"),
		},
		{
			name: "test state file with testset",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: filepath.Join("out", "test", "core", "go-gotest-unit", "state.json"),
		},
		{
			name: "lint state file",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "shared-lib",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: filepath.Join("out", "lint", "shared-lib", "go-golangci-lint", "state.json"),
		},
		{
			name: "scan state file",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expected: filepath.Join("out", "scan", "core", "go-trivy-vuln", "state.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.StateFile())
		})
	}
}

// =============================================================================
// LogFile Tests
// =============================================================================

func TestUnitID_LogFile(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "build log file",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: filepath.Join("out", "build", "core", "go-go", "execution.log"),
		},
		{
			name: "test log file with testset",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: filepath.Join("out", "test", "core", "go-gotest-unit", "execution.log"),
		},
		{
			name: "lint log file",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "shared-lib",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: filepath.Join("out", "lint", "shared-lib", "go-golangci-lint", "execution.log"),
		},
		{
			name: "scan log file",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expected: filepath.Join("out", "scan", "core", "go-trivy-vuln", "execution.log"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.LogFile())
		})
	}
}

// =============================================================================
// ResultsFile Tests
// =============================================================================

func TestUnitID_ResultsFile(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "test results file with testset",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: filepath.Join("out", "test", "core", "go-gotest-unit", "results.json"),
		},
		{
			name: "test integration results file",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "integration"},
			},
			expected: filepath.Join("out", "test", "core", "go-gotest-integration", "results.json"),
		},
		{
			name: "lint results file",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "shared-lib",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: filepath.Join("out", "lint", "shared-lib", "go-golangci-lint", "results.json"),
		},
		{
			name: "scan results file",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expected: filepath.Join("out", "scan", "core", "go-trivy-vuln", "results.json"),
		},
		{
			name: "build results file (if applicable)",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: filepath.Join("out", "build", "core", "go-go", "results.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.ResultsFile())
		})
	}
}

// =============================================================================
// Plan Examples Tests - Comprehensive Examples from Specification
// =============================================================================

func TestUnitID_PlanExamples(t *testing.T) {
	examples := []struct {
		name           string
		unitID         UnitID
		expectedLong   string
		expectedShort  string
		expectedOutDir string
	}{
		{
			name: "build:core:go:go",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expectedLong:   "build:core:go:go",
			expectedShort:  "core:go",
			expectedOutDir: filepath.Join("out", "build", "core", "go-go"),
		},
		{
			name: "test:core:go:gotest:unit",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expectedLong:   "test:core:go:gotest:unit",
			expectedShort:  "core:go",
			expectedOutDir: filepath.Join("out", "test", "core", "go-gotest-unit"),
		},
		{
			name: "test:core:go:gotest:integration",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "integration"},
			},
			expectedLong:   "test:core:go:gotest:integration",
			expectedShort:  "core:go",
			expectedOutDir: filepath.Join("out", "test", "core", "go-gotest-integration"),
		},
		{
			name: "lint:core:go:golangci-lint",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "core",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expectedLong:   "lint:core:go:golangci-lint",
			expectedShort:  "core:go",
			expectedOutDir: filepath.Join("out", "lint", "core", "go-golangci-lint"),
		},
		{
			name: "scan:core:go:trivy-vuln",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expectedLong:   "scan:core:go:trivy-vuln",
			expectedShort:  "core:go",
			expectedOutDir: filepath.Join("out", "scan", "core", "go-trivy-vuln"),
		},
	}

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			t.Run("Longname", func(t *testing.T) {
				assert.Equal(t, ex.expectedLong, ex.unitID.Longname())
			})

			t.Run("String equals Longname", func(t *testing.T) {
				assert.Equal(t, ex.unitID.Longname(), ex.unitID.String())
			})

			t.Run("Path", func(t *testing.T) {
				assert.Equal(t, ex.expectedShort, ex.unitID.Path())
			})

			t.Run("OutDir", func(t *testing.T) {
				assert.Equal(t, ex.expectedOutDir, ex.unitID.OutDir())
			})
		})
	}
}

// =============================================================================
// Edge Cases and Special Scenarios
// =============================================================================

func TestUnitID_EdgeCases(t *testing.T) {
	t.Run("empty extra map treated same as nil", func(t *testing.T) {
		nilExtra := UnitID{
			Context:   ContextBuild,
			Module:    "mod",
			Component: "comp",
			Tool:      "tool",
			Extra:     nil,
		}
		emptyExtra := UnitID{
			Context:   ContextBuild,
			Module:    "mod",
			Component: "comp",
			Tool:      "tool",
			Extra:     map[string]string{},
		}

		assert.Equal(t, nilExtra.Longname(), emptyExtra.Longname(),
			"nil and empty Extra should produce same Longname")
		assert.Equal(t, nilExtra.OutDir(), emptyExtra.OutDir(),
			"nil and empty Extra should produce same OutDir")
	})

	t.Run("module name with hyphens", func(t *testing.T) {
		uid := UnitID{
			Context:   ContextBuild,
			Module:    "my-complex-module-name",
			Component: "go",
			Tool:      "go",
		}

		assert.Equal(t, "build:my-complex-module-name:go:go", uid.Longname())
		assert.Equal(t, "my-complex-module-name:go", uid.Path())
	})

	t.Run("component with underscores", func(t *testing.T) {
		uid := UnitID{
			Context:   ContextTest,
			Module:    "mod",
			Component: "go_test",
			Tool:      "gotest",
			Extra:     map[string]string{"testset": "unit"},
		}

		assert.Equal(t, "test:mod:go_test:gotest:unit", uid.Longname())
	})

	t.Run("tool name with hyphens", func(t *testing.T) {
		uid := UnitID{
			Context:   ContextLint,
			Module:    "mod",
			Component: "go",
			Tool:      "golangci-lint",
		}

		assert.Equal(t, "lint:mod:go:golangci-lint", uid.Longname())
	})
}

// =============================================================================
// File Path Consistency Tests
// =============================================================================

func TestUnitID_FilePathConsistency(t *testing.T) {
	unitID := UnitID{
		Context:   ContextTest,
		Module:    "core",
		Component: "go",
		Tool:      "gotest",
		Extra:     map[string]string{"testset": "unit"},
	}

	outDir := unitID.OutDir()

	t.Run("LockFile is in OutDir", func(t *testing.T) {
		lockFile := unitID.LockFile()
		expectedLockFile := filepath.Join(outDir, ".lock")
		assert.Equal(t, expectedLockFile, lockFile)
	})

	t.Run("StateFile is in OutDir", func(t *testing.T) {
		stateFile := unitID.StateFile()
		expectedStateFile := filepath.Join(outDir, "state.json")
		assert.Equal(t, expectedStateFile, stateFile)
	})

	t.Run("LogFile is in OutDir", func(t *testing.T) {
		logFile := unitID.LogFile()
		expectedLogFile := filepath.Join(outDir, "execution.log")
		assert.Equal(t, expectedLogFile, logFile)
	})

	t.Run("ResultsFile is in OutDir", func(t *testing.T) {
		resultsFile := unitID.ResultsFile()
		expectedResultsFile := filepath.Join(outDir, "results.json")
		assert.Equal(t, expectedResultsFile, resultsFile)
	})
}

// =============================================================================
// Multiple Test Sets for Same Module
// =============================================================================

func TestUnitID_MultipleTestSetsForSameModule(t *testing.T) {
	// Verify that different testsets produce unique paths
	unitTest := UnitID{
		Context:   ContextTest,
		Module:    "core",
		Component: "go",
		Tool:      "gotest",
		Extra:     map[string]string{"testset": "unit"},
	}

	integrationTest := UnitID{
		Context:   ContextTest,
		Module:    "core",
		Component: "go",
		Tool:      "gotest",
		Extra:     map[string]string{"testset": "integration"},
	}

	t.Run("different longnames", func(t *testing.T) {
		assert.NotEqual(t, unitTest.Longname(), integrationTest.Longname())
	})

	t.Run("same path", func(t *testing.T) {
		assert.Equal(t, unitTest.Path(), integrationTest.Path())
	})

	t.Run("different output directories", func(t *testing.T) {
		assert.NotEqual(t, unitTest.OutDir(), integrationTest.OutDir())
	})

	t.Run("different lock files", func(t *testing.T) {
		assert.NotEqual(t, unitTest.LockFile(), integrationTest.LockFile())
	})

	t.Run("different state files", func(t *testing.T) {
		assert.NotEqual(t, unitTest.StateFile(), integrationTest.StateFile())
	})

	t.Run("different log files", func(t *testing.T) {
		assert.NotEqual(t, unitTest.LogFile(), integrationTest.LogFile())
	})

	t.Run("different results files", func(t *testing.T) {
		assert.NotEqual(t, unitTest.ResultsFile(), integrationTest.ResultsFile())
	})
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
				Context:   ContextTest,
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
				Context:   ContextTest,
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
				Context:   ContextTest,
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
				Context:   ContextTest,
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
				Context:   ContextTest,
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
		name      string
		spec      string
		maxWidth  int
		expected  string
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
// Context Type Behavior Tests
// =============================================================================

func TestContext_IsValidContext(t *testing.T) {
	validContexts := []Context{ContextBuild, ContextTest, ContextLint, ContextScan}

	for _, ctx := range validContexts {
		t.Run(string(ctx)+" is valid", func(t *testing.T) {
			// Context should be non-empty and one of the expected values
			assert.NotEmpty(t, string(ctx))
			assert.Contains(t, []string{"build", "test", "lint", "scan"}, string(ctx))
		})
	}
}

// =============================================================================
// Display Method Tests
// =============================================================================

func TestUnitID_Path(t *testing.T) {
	u := UnitID{
		Context:   ContextBuild,
		Module:    "core",
		Component: "go",
		Tool:      "go",
	}

	expected := "core:go"
	assert.Equal(t, expected, u.Path())
}

func TestUnitID_ComponentName(t *testing.T) {
	u := UnitID{
		Context:   ContextBuild,
		Module:    "core",
		Component: "impl/build",
		Tool:      "go",
	}

	assert.Equal(t, "impl/build", u.ComponentName())
}

func TestUnitID_TabLabel(t *testing.T) {
	tests := []struct {
		name      string
		component string
		maxWidth  int
		expected  string
	}{
		{
			name:      "short component fits",
			component: "go",
			maxWidth:  10,
			expected:  "go",
		},
		{
			name:      "long component truncated",
			component: "component-name",
			maxWidth:  10,
			expected:  "compone...",
		},
		{
			name:      "exact fit",
			component: "component",
			maxWidth:  9,
			expected:  "component",
		},
		{
			name:      "very short maxWidth",
			component: "component",
			maxWidth:  2,
			expected:  "co",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := UnitID{Component: tt.component}
			assert.Equal(t, tt.expected, u.TabLabel(tt.maxWidth))
		})
	}
}

// =============================================================================
// Longname Tests - All Extra Fields (TDD: Expect these to fail initially)
// =============================================================================

func TestUnitID_Longname_AllExtraFields(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "multiple extra fields sorted alphabetically by key",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy",
				Extra:     map[string]string{"category": "sbom", "testset": "unit"},
			},
			// category comes before testset alphabetically
			expected: "scan:core:go:trivy:sbom:unit",
		},
		{
			name: "scan context with category extra field",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-sbom",
				Extra:     map[string]string{"category": "sbom"},
			},
			expected: "scan:core:go:trivy-sbom:sbom",
		},
		{
			name: "empty values in extra are skipped",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"category": "", "testset": "unit"},
			},
			// empty category should be skipped, only testset appended
			expected: "test:core:go:gotest:unit",
		},
		{
			name: "three extra fields sorted alphabetically",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy",
				Extra:     map[string]string{"zone": "prod", "category": "sbom", "testset": "unit"},
			},
			// category < testset < zone alphabetically
			expected: "scan:core:go:trivy:sbom:unit:prod",
		},
		{
			name: "single non-testset extra field",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy",
				Extra:     map[string]string{"category": "vuln"},
			},
			expected: "scan:core:go:trivy:vuln",
		},
		{
			name: "all empty extra values produce base format",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy",
				Extra:     map[string]string{"category": "", "testset": ""},
			},
			expected: "scan:core:go:trivy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.Longname(),
				"Longname should include all non-empty Extra fields in sorted key order")
		})
	}
}

func TestUnitID_OutDir_AllExtraFields(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "multiple extra fields sorted alphabetically in path",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy",
				Extra:     map[string]string{"category": "sbom", "testset": "unit"},
			},
			// Format: component-tool-extra1-extra2 (category before testset alphabetically)
			expected: filepath.Join("out", "scan", "core", "go-trivy-sbom-unit"),
		},
		{
			name: "scan context with category extra field",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-sbom",
				Extra:     map[string]string{"category": "sbom"},
			},
			expected: filepath.Join("out", "scan", "core", "go-trivy-sbom-sbom"),
		},
		{
			name: "empty values in extra are skipped in path",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"category": "", "testset": "unit"},
			},
			// empty category should be skipped
			expected: filepath.Join("out", "test", "core", "go-gotest-unit"),
		},
		{
			name: "three extra fields sorted alphabetically in path",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy",
				Extra:     map[string]string{"zone": "prod", "category": "sbom", "testset": "unit"},
			},
			// category < testset < zone alphabetically, joined with dashes
			expected: filepath.Join("out", "scan", "core", "go-trivy-sbom-unit-prod"),
		},
		{
			name: "single non-testset extra field in path",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy",
				Extra:     map[string]string{"category": "vuln"},
			},
			expected: filepath.Join("out", "scan", "core", "go-trivy-vuln"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.OutDir(),
				"OutDir should include all non-empty Extra fields in sorted key order")
		})
	}
}

func TestUnitID_Uniqueness_WithDifferentExtraFields(t *testing.T) {
	// Two units differing only by category should have different Longnames
	// This proves uniqueness for scan context where category distinguishes work units

	sbomScan := UnitID{
		Context:   ContextScan,
		Module:    "core",
		Component: "go",
		Tool:      "trivy",
		Extra:     map[string]string{"category": "sbom"},
	}

	vulnScan := UnitID{
		Context:   ContextScan,
		Module:    "core",
		Component: "go",
		Tool:      "trivy",
		Extra:     map[string]string{"category": "vuln"},
	}

	t.Run("different categories produce different longnames", func(t *testing.T) {
		assert.NotEqual(t, sbomScan.Longname(), vulnScan.Longname(),
			"Units with different category values must have different Longnames for uniqueness")
	})

	t.Run("different categories produce different output directories", func(t *testing.T) {
		assert.NotEqual(t, sbomScan.OutDir(), vulnScan.OutDir(),
			"Units with different category values must have different OutDirs for uniqueness")
	})

	t.Run("sbom scan has expected longname", func(t *testing.T) {
		assert.Equal(t, "scan:core:go:trivy:sbom", sbomScan.Longname())
	})

	t.Run("vuln scan has expected longname", func(t *testing.T) {
		assert.Equal(t, "scan:core:go:trivy:vuln", vulnScan.Longname())
	})

	t.Run("sbom scan has expected outdir", func(t *testing.T) {
		assert.Equal(t, filepath.Join("out", "scan", "core", "go-trivy-sbom"), sbomScan.OutDir())
	})

	t.Run("vuln scan has expected outdir", func(t *testing.T) {
		assert.Equal(t, filepath.Join("out", "scan", "core", "go-trivy-vuln"), vulnScan.OutDir())
	})
}

func TestUnitID_DerivedFiles_WithAllExtraFields(t *testing.T) {
	// Verify that all derived file paths use the full OutDir with all extra fields
	unitID := UnitID{
		Context:   ContextScan,
		Module:    "core",
		Component: "go",
		Tool:      "trivy",
		Extra:     map[string]string{"category": "sbom", "testset": "unit"},
	}

	// Expected base: out/scan/core/go-trivy-sbom-unit (component-tool-extra1-extra2, sorted alphabetically)
	expectedOutDir := filepath.Join("out", "scan", "core", "go-trivy-sbom-unit")

	t.Run("OutDir includes all extra fields sorted", func(t *testing.T) {
		assert.Equal(t, expectedOutDir, unitID.OutDir())
	})

	t.Run("LockFile uses full OutDir", func(t *testing.T) {
		assert.Equal(t, filepath.Join(expectedOutDir, ".lock"), unitID.LockFile())
	})

	t.Run("StateFile uses full OutDir", func(t *testing.T) {
		assert.Equal(t, filepath.Join(expectedOutDir, "state.json"), unitID.StateFile())
	})

	t.Run("LogFile uses full OutDir", func(t *testing.T) {
		assert.Equal(t, filepath.Join(expectedOutDir, "execution.log"), unitID.LogFile())
	})

	t.Run("ResultsFile uses full OutDir", func(t *testing.T) {
		assert.Equal(t, filepath.Join(expectedOutDir, "results.json"), unitID.ResultsFile())
	})
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
				Context:   ContextBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: "Building go in core",
		},
		{
			name: "site build with mkdocs",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "docs",
				Component: "site",
				Tool:      "mkdocs",
			},
			expected: "Building site in docs with mkdocs",
		},
		{
			name: "build with empty tool",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "core",
				Component: "go",
				Tool:      "",
			},
			expected: "Building go in core",
		},
		{
			name: "typescript build",
			unitID: UnitID{
				Context:   ContextBuild,
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
				Context:   ContextTest,
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
				Context:   ContextTest,
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
				Context:   ContextTest,
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
				Context:   ContextLint,
				Module:    "core",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: "Linting go in core with golangci-lint",
		},
		{
			name: "eslint",
			unitID: UnitID{
				Context:   ContextLint,
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
				Context:   ContextScan,
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
				Context:   ContextScan,
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
		Context:   "unknown",
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
				Context:   ContextBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: "go",
		},
		{
			name: "test context with spec uses spec",
			unitID: UnitID{
				Context:   ContextTest,
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
				Context:   ContextTest,
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
				Context:   ContextTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
			},
			expected: "go",
		},
		{
			name: "lint context uses component",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "core",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: "go",
		},
		{
			name: "scan context uses component",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy",
			},
			expected: "go",
		},
		{
			name: "spec takes priority over testname",
			unitID: UnitID{
				Context:   ContextTest,
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
				Context:   ContextTest,
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
				Context:   ContextTest,
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
				Context:   ContextTest,
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
				Context:   ContextTest,
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
				Context:   ContextTest,
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
				Context:   ContextBuild,
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
				Context:   ContextBuild,
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
				Context:   ContextBuild,
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
				Context:   ContextBuild,
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
				Context:   ContextLint,
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
				Context:   ContextLint,
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
				Context:   ContextScan,
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
				Context:   ContextScan,
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
