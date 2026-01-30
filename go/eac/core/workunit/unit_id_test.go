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
// Shortname Tests
// =============================================================================

func TestUnitID_Shortname(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "build unit shortname",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "eac-core",
				Component: "go",
				Tool:      "go",
			},
			expected: "eac-core:go",
		},
		{
			name: "test unit shortname",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-cli",
				Component: "gherkin",
				Tool:      "godog",
			},
			expected: "eac-cli:gherkin",
		},
		{
			name: "lint unit shortname",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "shared-lib",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: "shared-lib:go",
		},
		{
			name: "shortname ignores extra fields",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: "eac-core:go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.Shortname())
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
			name: "build:eac-core:go:go",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "eac-core",
				Component: "go",
				Tool:      "go",
			},
			expected: "build:eac-core:go:go",
		},
		{
			name: "lint:eac-core:go:golangci-lint",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "eac-core",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: "lint:eac-core:go:golangci-lint",
		},
		{
			name: "scan:eac-core:go:trivy-vuln",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "eac-core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expected: "scan:eac-core:go:trivy-vuln",
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
			name: "test:eac-core:go:gotest:unit",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: "test:eac-core:go:gotest:unit",
		},
		{
			name: "test:eac-core:go:gotest:integration",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "integration"},
			},
			expected: "test:eac-core:go:gotest:integration",
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
				Module:    "eac-core",
				Component: "go",
				Tool:      "go",
			},
		},
		{
			name: "test unit string equals longname",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-core",
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
				Module:    "eac-core",
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
				Module:    "eac-core",
				Component: "go",
				Tool:      "go",
			},
			expected: filepath.Join("out", "build", "eac-core", "go"),
		},
		{
			name: "lint output directory",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "eac-core",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: filepath.Join("out", "lint", "eac-core", "go"),
		},
		{
			name: "scan output directory",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "eac-core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expected: filepath.Join("out", "scan", "eac-core", "go"),
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
				Module:    "eac-core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: filepath.Join("out", "test", "eac-core", "go", "unit"),
		},
		{
			name: "test integration output directory",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "integration"},
			},
			expected: filepath.Join("out", "test", "eac-core", "go", "integration"),
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
			expected: filepath.Join("out", "test", "eac-cli", "gherkin", "acceptance"),
		},
		{
			name: "test without testset falls back to base path",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-core",
				Component: "go",
				Tool:      "gotest",
				Extra:     nil,
			},
			expected: filepath.Join("out", "test", "eac-core", "go"),
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
				Module:    "eac-core",
				Component: "go",
				Tool:      "go",
			},
			expected: filepath.Join("out", "build", "eac-core", "go", ".lock"),
		},
		{
			name: "test lock file with testset",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: filepath.Join("out", "test", "eac-core", "go", "unit", ".lock"),
		},
		{
			name: "lint lock file",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "shared-lib",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: filepath.Join("out", "lint", "shared-lib", "go", ".lock"),
		},
		{
			name: "scan lock file",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "eac-core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expected: filepath.Join("out", "scan", "eac-core", "go", ".lock"),
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
				Module:    "eac-core",
				Component: "go",
				Tool:      "go",
			},
			expected: filepath.Join("out", "build", "eac-core", "go", "state.json"),
		},
		{
			name: "test state file with testset",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: filepath.Join("out", "test", "eac-core", "go", "unit", "state.json"),
		},
		{
			name: "lint state file",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "shared-lib",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: filepath.Join("out", "lint", "shared-lib", "go", "state.json"),
		},
		{
			name: "scan state file",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "eac-core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expected: filepath.Join("out", "scan", "eac-core", "go", "state.json"),
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
				Module:    "eac-core",
				Component: "go",
				Tool:      "go",
			},
			expected: filepath.Join("out", "build", "eac-core", "go", "execution.log"),
		},
		{
			name: "test log file with testset",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: filepath.Join("out", "test", "eac-core", "go", "unit", "execution.log"),
		},
		{
			name: "lint log file",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "shared-lib",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: filepath.Join("out", "lint", "shared-lib", "go", "execution.log"),
		},
		{
			name: "scan log file",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "eac-core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expected: filepath.Join("out", "scan", "eac-core", "go", "execution.log"),
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
				Module:    "eac-core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: filepath.Join("out", "test", "eac-core", "go", "unit", "results.json"),
		},
		{
			name: "test integration results file",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "integration"},
			},
			expected: filepath.Join("out", "test", "eac-core", "go", "integration", "results.json"),
		},
		{
			name: "lint results file",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "shared-lib",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: filepath.Join("out", "lint", "shared-lib", "go", "results.json"),
		},
		{
			name: "scan results file",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "eac-core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expected: filepath.Join("out", "scan", "eac-core", "go", "results.json"),
		},
		{
			name: "build results file (if applicable)",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "eac-core",
				Component: "go",
				Tool:      "go",
			},
			expected: filepath.Join("out", "build", "eac-core", "go", "results.json"),
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
			name: "build:eac-core:go:go",
			unitID: UnitID{
				Context:   ContextBuild,
				Module:    "eac-core",
				Component: "go",
				Tool:      "go",
			},
			expectedLong:   "build:eac-core:go:go",
			expectedShort:  "eac-core:go",
			expectedOutDir: filepath.Join("out", "build", "eac-core", "go"),
		},
		{
			name: "test:eac-core:go:gotest:unit",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expectedLong:   "test:eac-core:go:gotest:unit",
			expectedShort:  "eac-core:go",
			expectedOutDir: filepath.Join("out", "test", "eac-core", "go", "unit"),
		},
		{
			name: "test:eac-core:go:gotest:integration",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "integration"},
			},
			expectedLong:   "test:eac-core:go:gotest:integration",
			expectedShort:  "eac-core:go",
			expectedOutDir: filepath.Join("out", "test", "eac-core", "go", "integration"),
		},
		{
			name: "lint:eac-core:go:golangci-lint",
			unitID: UnitID{
				Context:   ContextLint,
				Module:    "eac-core",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expectedLong:   "lint:eac-core:go:golangci-lint",
			expectedShort:  "eac-core:go",
			expectedOutDir: filepath.Join("out", "lint", "eac-core", "go"),
		},
		{
			name: "scan:eac-core:go:trivy-vuln",
			unitID: UnitID{
				Context:   ContextScan,
				Module:    "eac-core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expectedLong:   "scan:eac-core:go:trivy-vuln",
			expectedShort:  "eac-core:go",
			expectedOutDir: filepath.Join("out", "scan", "eac-core", "go"),
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

			t.Run("Shortname", func(t *testing.T) {
				assert.Equal(t, ex.expectedShort, ex.unitID.Shortname())
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
		assert.Equal(t, "my-complex-module-name:go", uid.Shortname())
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
		Module:    "eac-core",
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
		Module:    "eac-core",
		Component: "go",
		Tool:      "gotest",
		Extra:     map[string]string{"testset": "unit"},
	}

	integrationTest := UnitID{
		Context:   ContextTest,
		Module:    "eac-core",
		Component: "go",
		Tool:      "gotest",
		Extra:     map[string]string{"testset": "integration"},
	}

	t.Run("different longnames", func(t *testing.T) {
		assert.NotEqual(t, unitTest.Longname(), integrationTest.Longname())
	})

	t.Run("same shortname", func(t *testing.T) {
		assert.Equal(t, unitTest.Shortname(), integrationTest.Shortname())
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

func TestUnitID_Spec_Shortname(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "spec test returns spec name",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-commands",
				Component: "build-module:go/eac/specs/impl/eac-commands:specs/.../specification.feature:godog",
				Tool:      "godog",
				Spec:      "build-module",
			},
			expected: "build-module",
		},
		{
			name: "non-spec test returns module:component",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-core",
				Component: "go",
				Tool:      "gotest",
			},
			expected: "eac-core:go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.Shortname())
		})
	}
}

func TestUnitID_Spec_Longname(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "spec test returns module:spec:specname format",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-commands",
				Component: "build-module:go/eac/specs/impl/eac-commands:specs/.../specification.feature:godog",
				Tool:      "godog",
				Spec:      "build-module",
			},
			expected: "eac-commands:spec:build-module",
		},
		{
			name: "spec test with different module",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-core",
				Component: "cache-invalidation:go/eac/specs/impl/eac-core:specs/.../specification.feature:godog",
				Tool:      "godog",
				Spec:      "cache-invalidation",
			},
			expected: "eac-core:spec:cache-invalidation",
		},
		{
			name: "non-spec test returns standard format",
			unitID: UnitID{
				Context:   ContextTest,
				Module:    "eac-core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: "test:eac-core:go:gotest:unit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.Longname())
		})
	}
}

func TestUnitID_Spec_DisplayName(t *testing.T) {
	specUnit := UnitID{
		Context:   ContextTest,
		Module:    "eac-commands",
		Component: "build-module:go/eac/specs/impl/eac-commands:specs/.../specification.feature:godog",
		Tool:      "godog",
		Spec:      "build-module",
	}

	t.Run("disambiguate false returns spec only", func(t *testing.T) {
		assert.Equal(t, "build-module", specUnit.DisplayName(false))
	})

	t.Run("disambiguate true returns module:spec:specname", func(t *testing.T) {
		assert.Equal(t, "eac-commands:spec:build-module", specUnit.DisplayName(true))
	})
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
		Module:    "eac-core",
		Component: "go",
		Tool:      "go",
	}

	expected := "eac-core:go"
	assert.Equal(t, expected, u.Path())
}

func TestUnitID_ComponentName(t *testing.T) {
	u := UnitID{
		Context:   ContextBuild,
		Module:    "eac-core",
		Component: "impl/build",
		Tool:      "go",
	}

	assert.Equal(t, "impl/build", u.ComponentName())
}

func TestUnitID_DisplayName(t *testing.T) {
	u := UnitID{
		Context:   ContextBuild,
		Module:    "eac-core",
		Component: "go",
		Tool:      "go",
	}

	t.Run("disambiguate false returns component only", func(t *testing.T) {
		assert.Equal(t, "go", u.DisplayName(false))
	})

	t.Run("disambiguate true returns full path", func(t *testing.T) {
		assert.Equal(t, "eac-core:go", u.DisplayName(true))
	})
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
