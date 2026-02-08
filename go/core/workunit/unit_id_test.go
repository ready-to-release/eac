package workunit

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

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
				Action:    core.ActionBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: "core: go",
		},
		{
			name: "build - component differs from tool",
			unitID: UnitID{
				Action:    core.ActionBuild,
				Module:    "docs",
				Component: "site",
				Tool:      "mkdocs",
			},
			expected: "docs: site: mkdocs",
		},
		{
			name: "test - BDD spec",
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
			name: "test - unit with testname",
			unitID: UnitID{
				Action:    core.ActionTest,
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
				Action:    core.ActionTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
			},
			expected: "go: unit",
		},
		{
			name: "lint",
			unitID: UnitID{
				Action:    core.ActionLint,
				Module:    "core",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: "lint:go:golangci-lint",
		},
		{
			name: "scan with category",
			unitID: UnitID{
				Action:    core.ActionScan,
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
				Action:    core.ActionBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: "build:core:go:go",
		},
		{
			name: "lint:core:go:golangci-lint",
			unitID: UnitID{
				Action:    core.ActionLint,
				Module:    "core",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: "lint:core:go:golangci-lint",
		},
		{
			name: "scan:core:go:trivy-vuln",
			unitID: UnitID{
				Action:    core.ActionScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expected: "scan:core:go:trivy-vuln",
		},
		{
			name: "build with empty extra returns base format",
			unitID: UnitID{
				Action:    core.ActionBuild,
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
				Action:    core.ActionBuild,
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
				Action:    core.ActionTest,
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
				Action:    core.ActionTest,
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
				Action:    core.ActionTest,
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
				Action:    core.ActionBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
		},
		{
			name: "test unit string equals longname",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
		},
		{
			name: "lint unit string equals longname",
			unitID: UnitID{
				Action:    core.ActionLint,
				Module:    "shared-lib",
				Component: "go",
				Tool:      "golangci-lint",
			},
		},
		{
			name: "scan unit string equals longname",
			unitID: UnitID{
				Action:    core.ActionScan,
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
				Action:    core.ActionBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: filepath.Join("out", "build", "core", "go-go"),
		},
		{
			name: "lint output directory",
			unitID: UnitID{
				Action:    core.ActionLint,
				Module:    "core",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: filepath.Join("out", "lint", "core", "go-golangci-lint"),
		},
		{
			name: "scan output directory",
			unitID: UnitID{
				Action:    core.ActionScan,
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
				Action:    core.ActionTest,
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
				Action:    core.ActionTest,
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
				Action:    core.ActionTest,
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
				Action:    core.ActionTest,
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
				Action:    core.ActionBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: filepath.Join("out", "build", "core", "go-go", ".lock"),
		},
		{
			name: "test lock file with testset",
			unitID: UnitID{
				Action:    core.ActionTest,
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
				Action:    core.ActionLint,
				Module:    "shared-lib",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: filepath.Join("out", "lint", "shared-lib", "go-golangci-lint", ".lock"),
		},
		{
			name: "scan lock file",
			unitID: UnitID{
				Action:    core.ActionScan,
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

func TestUnitID_StateCacheDir(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "build state cache dir",
			unitID: UnitID{
				Action:    core.ActionBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: filepath.Join(".cache", "eac", "incremental", "build", "core", "go-go"),
		},
		{
			name: "test state cache dir with testset",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: filepath.Join(".cache", "eac", "incremental", "test", "core", "go-gotest-unit"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.unitID.StateCacheDir())
		})
	}
}

func TestUnitID_StateFile(t *testing.T) {
	tests := []struct {
		name     string
		unitID   UnitID
		expected string
	}{
		{
			name: "build state file",
			unitID: UnitID{
				Action:    core.ActionBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: filepath.Join(".cache", "eac", "incremental", "build", "core", "go-go", "state.json"),
		},
		{
			name: "test state file with testset",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"testset": "unit"},
			},
			expected: filepath.Join(".cache", "eac", "incremental", "test", "core", "go-gotest-unit", "state.json"),
		},
		{
			name: "lint state file",
			unitID: UnitID{
				Action:    core.ActionLint,
				Module:    "shared-lib",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: filepath.Join(".cache", "eac", "incremental", "lint", "shared-lib", "go-golangci-lint", "state.json"),
		},
		{
			name: "scan state file",
			unitID: UnitID{
				Action:    core.ActionScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expected: filepath.Join(".cache", "eac", "incremental", "scan", "core", "go-trivy-vuln", "state.json"),
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
				Action:    core.ActionBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
			},
			expected: filepath.Join("out", "build", "core", "go-go", "execution.log"),
		},
		{
			name: "test log file with testset",
			unitID: UnitID{
				Action:    core.ActionTest,
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
				Action:    core.ActionLint,
				Module:    "shared-lib",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: filepath.Join("out", "lint", "shared-lib", "go-golangci-lint", "execution.log"),
		},
		{
			name: "scan log file",
			unitID: UnitID{
				Action:    core.ActionScan,
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
				Action:    core.ActionTest,
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
				Action:    core.ActionTest,
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
				Action:    core.ActionLint,
				Module:    "shared-lib",
				Component: "go",
				Tool:      "golangci-lint",
			},
			expected: filepath.Join("out", "lint", "shared-lib", "go-golangci-lint", "results.json"),
		},
		{
			name: "scan results file",
			unitID: UnitID{
				Action:    core.ActionScan,
				Module:    "core",
				Component: "go",
				Tool:      "trivy-vuln",
			},
			expected: filepath.Join("out", "scan", "core", "go-trivy-vuln", "results.json"),
		},
		{
			name: "build results file (if applicable)",
			unitID: UnitID{
				Action:    core.ActionBuild,
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
				Action:    core.ActionBuild,
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
				Action:    core.ActionTest,
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
				Action:    core.ActionTest,
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
				Action:    core.ActionLint,
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
				Action:    core.ActionScan,
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
			Action:    core.ActionBuild,
			Module:    "mod",
			Component: "comp",
			Tool:      "tool",
			Extra:     nil,
		}
		emptyExtra := UnitID{
			Action:    core.ActionBuild,
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
			Action:    core.ActionBuild,
			Module:    "my-complex-module-name",
			Component: "go",
			Tool:      "go",
		}

		assert.Equal(t, "build:my-complex-module-name:go:go", uid.Longname())
		assert.Equal(t, "my-complex-module-name:go", uid.Path())
	})

	t.Run("component with underscores", func(t *testing.T) {
		uid := UnitID{
			Action:    core.ActionTest,
			Module:    "mod",
			Component: "go_test",
			Tool:      "gotest",
			Extra:     map[string]string{"testset": "unit"},
		}

		assert.Equal(t, "test:mod:go_test:gotest:unit", uid.Longname())
	})

	t.Run("tool name with hyphens", func(t *testing.T) {
		uid := UnitID{
			Action:    core.ActionLint,
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
		Action:    core.ActionTest,
		Module:    "core",
		Component: "go",
		Tool:      "gotest",
		Extra:     map[string]string{"testset": "unit"},
	}

	outDir := unitID.OutDir()
	stateCacheDir := unitID.StateCacheDir()

	t.Run("LockFile is in OutDir", func(t *testing.T) {
		lockFile := unitID.LockFile()
		expectedLockFile := filepath.Join(outDir, ".lock")
		assert.Equal(t, expectedLockFile, lockFile)
	})

	t.Run("StateFile is in StateCacheDir", func(t *testing.T) {
		stateFile := unitID.StateFile()
		expectedStateFile := filepath.Join(stateCacheDir, "state.json")
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
		Action:    core.ActionTest,
		Module:    "core",
		Component: "go",
		Tool:      "gotest",
		Extra:     map[string]string{"testset": "unit"},
	}

	integrationTest := UnitID{
		Action:    core.ActionTest,
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
// Display Method Tests
// =============================================================================

func TestUnitID_Path(t *testing.T) {
	u := UnitID{
		Action:    core.ActionBuild,
		Module:    "core",
		Component: "go",
		Tool:      "go",
	}

	expected := "core:go"
	assert.Equal(t, expected, u.Path())
}

func TestUnitID_ComponentName(t *testing.T) {
	u := UnitID{
		Action:    core.ActionBuild,
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
