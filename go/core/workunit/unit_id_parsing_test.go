package workunit

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

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
				Action:    core.ActionScan,
				Module:    "core",
				ComponentType: "go", ComponentName: "go",
				Tool:      "trivy",
				Extra:     map[string]string{"category": "sbom", "testset": "unit"},
			},
			// scan gets no qualifier — extras not appended
			expected: "scan:core:go:trivy",
		},
		{
			name: "scan context with category extra field",
			unitID: UnitID{
				Action:    core.ActionScan,
				Module:    "core",
				ComponentType: "go", ComponentName: "go",
				Tool:      "trivy-sbom",
				Extra:     map[string]string{"category": "sbom"},
			},
			expected: "scan:core:go:trivy-sbom",
		},
		{
			name: "empty values in extra are skipped",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "core",
				ComponentType: "go", ComponentName: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"category": "", "testset": "unit"},
			},
			// testset is not testname, so no qualifier for test action either
			expected: "test:core:go:gotest",
		},
		{
			name: "three extra fields sorted alphabetically",
			unitID: UnitID{
				Action:    core.ActionScan,
				Module:    "core",
				ComponentType: "go", ComponentName: "go",
				Tool:      "trivy",
				Extra:     map[string]string{"zone": "prod", "category": "sbom", "testset": "unit"},
			},
			// scan gets no qualifier — extras not appended
			expected: "scan:core:go:trivy",
		},
		{
			name: "single non-testset extra field",
			unitID: UnitID{
				Action:    core.ActionScan,
				Module:    "core",
				ComponentType: "go", ComponentName: "go",
				Tool:      "trivy",
				Extra:     map[string]string{"category": "vuln"},
			},
			expected: "scan:core:go:trivy",
		},
		{
			name: "all empty extra values produce base format",
			unitID: UnitID{
				Action:    core.ActionScan,
				Module:    "core",
				ComponentType: "go", ComponentName: "go",
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
				Action:    core.ActionScan,
				Module:    "core",
				ComponentType: "go", ComponentName: "go",
				Tool:      "trivy",
				Extra:     map[string]string{"category": "sbom", "testset": "unit"},
			},
			// Format: componentName_tool (no extras appended for scan)
			expected: filepath.Join("out", "scan", "core", "go_trivy"),
		},
		{
			name: "scan context with category extra field",
			unitID: UnitID{
				Action:    core.ActionScan,
				Module:    "core",
				ComponentType: "go", ComponentName: "go",
				Tool:      "trivy-sbom",
				Extra:     map[string]string{"category": "sbom"},
			},
			expected: filepath.Join("out", "scan", "core", "go_trivy-sbom"),
		},
		{
			name: "empty values in extra are skipped in path",
			unitID: UnitID{
				Action:    core.ActionTest,
				Module:    "core",
				ComponentType: "go", ComponentName: "go",
				Tool:      "gotest",
				Extra:     map[string]string{"category": "", "testset": "unit"},
			},
			// testset is not testname, so no qualifier appended
			expected: filepath.Join("out", "test", "core", "go_gotest"),
		},
		{
			name: "three extra fields sorted alphabetically in path",
			unitID: UnitID{
				Action:    core.ActionScan,
				Module:    "core",
				ComponentType: "go", ComponentName: "go",
				Tool:      "trivy",
				Extra:     map[string]string{"zone": "prod", "category": "sbom", "testset": "unit"},
			},
			// scan gets no qualifier — extras not appended
			expected: filepath.Join("out", "scan", "core", "go_trivy"),
		},
		{
			name: "single non-testset extra field in path",
			unitID: UnitID{
				Action:    core.ActionScan,
				Module:    "core",
				ComponentType: "go", ComponentName: "go",
				Tool:      "trivy",
				Extra:     map[string]string{"category": "vuln"},
			},
			expected: filepath.Join("out", "scan", "core", "go_trivy"),
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
	// Two scan units differing only by category now produce the SAME longname
	// because category is no longer appended to longname or dirname

	sbomScan := UnitID{
		Action:    core.ActionScan,
		Module:    "core",
		ComponentType: "go", ComponentName: "go",
		Tool:      "trivy",
		Extra:     map[string]string{"category": "sbom"},
	}

	vulnScan := UnitID{
		Action:    core.ActionScan,
		Module:    "core",
		ComponentType: "go", ComponentName: "go",
		Tool:      "trivy",
		Extra:     map[string]string{"category": "vuln"},
	}

	t.Run("different categories produce same longnames", func(t *testing.T) {
		assert.Equal(t, sbomScan.Longname(), vulnScan.Longname(),
			"Category is no longer appended — scan units with different categories have the same Longname")
	})

	t.Run("different categories produce same output directories", func(t *testing.T) {
		assert.Equal(t, sbomScan.OutDir(), vulnScan.OutDir(),
			"Category is no longer appended — scan units with different categories have the same OutDir")
	})

	t.Run("sbom scan has expected longname", func(t *testing.T) {
		assert.Equal(t, "scan:core:go:trivy", sbomScan.Longname())
	})

	t.Run("vuln scan has expected longname", func(t *testing.T) {
		assert.Equal(t, "scan:core:go:trivy", vulnScan.Longname())
	})

	t.Run("sbom scan has expected outdir", func(t *testing.T) {
		assert.Equal(t, filepath.Join("out", "scan", "core", "go_trivy"), sbomScan.OutDir())
	})

	t.Run("vuln scan has expected outdir", func(t *testing.T) {
		assert.Equal(t, filepath.Join("out", "scan", "core", "go_trivy"), vulnScan.OutDir())
	})
}

func TestUnitID_DerivedFiles_WithAllExtraFields(t *testing.T) {
	// Verify that all derived file paths use the full OutDir with all extra fields
	unitID := UnitID{
		Action:    core.ActionScan,
		Module:    "core",
		ComponentType: "go", ComponentName: "go",
		Tool:      "trivy",
		Extra:     map[string]string{"category": "sbom", "testset": "unit"},
	}

	// Expected base: out/scan/core/go_trivy (no extras appended for scan)
	expectedOutDir := filepath.Join("out", "scan", "core", "go_trivy")

	t.Run("OutDir includes all extra fields sorted", func(t *testing.T) {
		assert.Equal(t, expectedOutDir, unitID.OutDir())
	})

	t.Run("LockFile uses full OutDir", func(t *testing.T) {
		assert.Equal(t, filepath.Join(expectedOutDir, ".lock"), unitID.LockFile())
	})

	t.Run("StateFile uses StateCacheDir", func(t *testing.T) {
		expectedStateCacheDir := filepath.Join(".cache", "eac", "incremental", "scan", "core", "go_trivy")
		assert.Equal(t, filepath.Join(expectedStateCacheDir, "state.json"), unitID.StateFile())
	})

	t.Run("LogFile uses full OutDir", func(t *testing.T) {
		assert.Equal(t, filepath.Join(expectedOutDir, "execution.log"), unitID.LogFile())
	})

	t.Run("ResultsFile uses full OutDir", func(t *testing.T) {
		assert.Equal(t, filepath.Join(expectedOutDir, "results.json"), unitID.ResultsFile())
	})
}
