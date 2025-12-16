package reports

import (
	"os"
	"strings"
	"testing"
)

func TestGetSpecs(t *testing.T) {
	tests := []struct {
		name           string
		module         string
		version        string
		wantErr        bool
		errContains    string
		expectedModule string
	}{
		{
			name:           "valid module - unreleased",
			module:         "ext-eac",
			version:        "unreleased",
			wantErr:        false,
			expectedModule: "ext-eac",
		},
		{
			name:           "valid module - latest",
			module:         "ext-eac",
			version:        "latest",
			wantErr:        false,
			expectedModule: "ext-eac",
		},
		{
			name:           "valid module - empty (defaults to unreleased)",
			module:         "ext-eac",
			version:        "",
			wantErr:        false,
			expectedModule: "ext-eac",
		},
		{
			name:           "regular module without dependencies",
			module:         "eac-commands",
			version:        "unreleased",
			wantErr:        false,
			expectedModule: "eac-commands",
		},
		{
			name:        "invalid module",
			module:      "nonexistent-module-xyz",
			version:     "",
			wantErr:     true,
			errContains: "module not found",
		},
		{
			name:        "invalid version",
			module:      "ext-eac",
			version:     "99.99.99",
			wantErr:     true,
			errContains: "version not found",
		},
	}

	// Get repository root
	workspaceRoot := os.Getenv("WORKSPACE_ROOT")
	if workspaceRoot == "" {
		t.Skip("WORKSPACE_ROOT not set, skipping integration test")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := GetSpecs(workspaceRoot, tt.module, tt.version, "")

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetSpecs() expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetSpecs() error = %v, should contain %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("GetSpecs() unexpected error = %v", err)
				return
			}

			if report == nil {
				t.Errorf("GetSpecs() returned nil report")
				return
			}

			if report.Module != tt.expectedModule {
				t.Errorf("GetSpecs() module = %v, want %v", report.Module, tt.expectedModule)
			}

			if report.Version == "" {
				t.Errorf("GetSpecs() version is empty")
			}

			// Verify spec files structure
			if report.SpecFiles == nil {
				t.Errorf("GetSpecs() SpecFiles is nil")
			}

			// Verify counts are consistent
			calculatedTotal := report.AddedCount + report.ModifiedCount + report.DeletedCount
			actualTotal := len(report.SpecFiles)
			if calculatedTotal != actualTotal {
				t.Errorf("GetSpecs() count mismatch: added(%d) + modified(%d) + deleted(%d) = %d, but got %d spec files",
					report.AddedCount, report.ModifiedCount, report.DeletedCount, calculatedTotal, actualTotal)
			}

			// Verify scenario count is non-negative
			if report.TotalScenarios < 0 {
				t.Errorf("GetSpecs() TotalScenarios = %d, should be non-negative", report.TotalScenarios)
			}
		})
	}
}

// TestGetSpecs_BundleModuleAggregation verifies that bundle modules
// aggregate specs from all their dependencies
func TestGetSpecs_BundleModuleAggregation(t *testing.T) {
	// Get repository root
	workspaceRoot := os.Getenv("WORKSPACE_ROOT")
	if workspaceRoot == "" {
		t.Skip("WORKSPACE_ROOT not set, skipping integration test")
	}

	// Test ext-eac bundle module (depends on eac-commands and r2r-cli)
	t.Run("ext-eac aggregates specs from dependencies", func(t *testing.T) {
		bundleReport, err := GetSpecs(workspaceRoot, "ext-eac", "unreleased", "")
		if err != nil {
			t.Fatalf("GetSpecs(ext-eac) failed: %v", err)
		}

		// Get specs for dependency module directly
		depReport, err := GetSpecs(workspaceRoot, "eac-commands", "unreleased", "")
		if err != nil {
			t.Fatalf("GetSpecs(eac-commands) failed: %v", err)
		}

		// Bundle should have at least as many specs as the dependency
		// (unless dependency has no specs)
		if len(depReport.SpecFiles) > 0 && len(bundleReport.SpecFiles) < len(depReport.SpecFiles) {
			t.Errorf("Bundle ext-eac has fewer specs (%d) than dependency eac-commands (%d)",
				len(bundleReport.SpecFiles), len(depReport.SpecFiles))
		}

		// Verify that specs from dependency are included in bundle
		if len(depReport.SpecFiles) > 0 {
			foundDepSpec := false
			for _, spec := range bundleReport.SpecFiles {
				if strings.Contains(spec.RelativePath, "specs/eac-commands/") {
					foundDepSpec = true
					break
				}
			}
			if !foundDepSpec {
				t.Error("Bundle ext-eac should include specs from dependency eac-commands")
			}
		}
	})

	// Test regular module (no dependencies)
	t.Run("regular module only includes own specs", func(t *testing.T) {
		report, err := GetSpecs(workspaceRoot, "eac-commands", "unreleased", "")
		if err != nil {
			t.Fatalf("GetSpecs(eac-commands) failed: %v", err)
		}

		// All specs should be from eac-commands directory
		for _, spec := range report.SpecFiles {
			if !strings.Contains(spec.RelativePath, "specs/eac-commands/") {
				t.Errorf("Regular module eac-commands should only include own specs, found: %s",
					spec.RelativePath)
			}
		}
	})
}
