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
			// category comes before testset alphabetically
			expected: "scan:core:go:go:trivy:sbom:unit",
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
			expected: "scan:core:go:go:trivy-sbom:sbom",
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
			// empty category should be skipped, only testset appended
			expected: "test:core:go:go:gotest:unit",
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
			// category < testset < zone alphabetically
			expected: "scan:core:go:go:trivy:sbom:unit:prod",
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
			expected: "scan:core:go:go:trivy:vuln",
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
			expected: "scan:core:go:go:trivy",
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
			// Format: component-tool-extra1-extra2 (category before testset alphabetically)
			expected: filepath.Join("out", "scan", "core", "go-trivy-sbom-unit"),
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
			expected: filepath.Join("out", "scan", "core", "go-trivy-sbom-sbom"),
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
			// empty category should be skipped
			expected: filepath.Join("out", "test", "core", "go-gotest-unit"),
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
			// category < testset < zone alphabetically, joined with dashes
			expected: filepath.Join("out", "scan", "core", "go-trivy-sbom-unit-prod"),
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

	t.Run("different categories produce different longnames", func(t *testing.T) {
		assert.NotEqual(t, sbomScan.Longname(), vulnScan.Longname(),
			"Units with different category values must have different Longnames for uniqueness")
	})

	t.Run("different categories produce different output directories", func(t *testing.T) {
		assert.NotEqual(t, sbomScan.OutDir(), vulnScan.OutDir(),
			"Units with different category values must have different OutDirs for uniqueness")
	})

	t.Run("sbom scan has expected longname", func(t *testing.T) {
		assert.Equal(t, "scan:core:go:go:trivy:sbom", sbomScan.Longname())
	})

	t.Run("vuln scan has expected longname", func(t *testing.T) {
		assert.Equal(t, "scan:core:go:go:trivy:vuln", vulnScan.Longname())
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
		Action:    core.ActionScan,
		Module:    "core",
		ComponentType: "go", ComponentName: "go",
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

	t.Run("StateFile uses StateCacheDir", func(t *testing.T) {
		expectedStateCacheDir := filepath.Join(".cache", "eac", "incremental", "scan", "core", "go-trivy-sbom-unit")
		assert.Equal(t, filepath.Join(expectedStateCacheDir, "state.json"), unitID.StateFile())
	})

	t.Run("LogFile uses full OutDir", func(t *testing.T) {
		assert.Equal(t, filepath.Join(expectedOutDir, "execution.log"), unitID.LogFile())
	})

	t.Run("ResultsFile uses full OutDir", func(t *testing.T) {
		assert.Equal(t, filepath.Join(expectedOutDir, "results.json"), unitID.ResultsFile())
	})
}
