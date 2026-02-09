package reports

import (
	"strings"
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/core/git"
	coretesting "github.com/ready-to-release/eac/go/core/testing"
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
			module:         "eac-ext",
			version:        "unreleased",
			wantErr:        false,
			expectedModule: "eac-ext",
		},
		{
			name:           "valid module - latest",
			module:         "eac-ext",
			version:        "latest",
			wantErr:        false,
			expectedModule: "eac-ext",
		},
		{
			name:           "valid module - empty (defaults to unreleased)",
			module:         "eac-ext",
			version:        "",
			wantErr:        false,
			expectedModule: "eac-ext",
		},
		{
			name:           "regular module without dependencies",
			module:         "eac",
			version:        "unreleased",
			wantErr:        false,
			expectedModule: "eac",
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
			module:      "eac-ext",
			version:     "99.99.99",
			wantErr:     true,
			errContains: "version not found",
		},
	}

	workspaceRoot := coretesting.SetupWorkspaceIsolation(t)

	// Set up mock git repo to avoid expensive git operations
	mockRepo := &mockGitRepo{
		commits: []git.CommitInfo{}, // No commits = no expensive diff operations
		tags:    []string{},
	}
	SetGitRepo(mockRepo)
	defer SetGitRepo(nil)

	// Set up version resolver mock with tags so "latest" version validation passes
	versionMock := git.NewMockRepository(workspaceRoot).
		WithTag("eac-ext/0.0.9", "abc123", time.Now())
	SetVersionResolverRepo(versionMock)
	defer SetVersionResolverRepo(nil)

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
// aggregate specs from all their dependencies.
func TestGetSpecs_BundleModuleAggregation(t *testing.T) {
	workspaceRoot := coretesting.SetupWorkspaceIsolation(t)

	// Set up mock git repo to avoid expensive git operations
	mockRepo := &mockGitRepo{
		commits: []git.CommitInfo{}, // No commits = no expensive diff operations
		tags:    []string{},
	}
	SetGitRepo(mockRepo)
	defer SetGitRepo(nil)

	// Test eac-ext bundle module (depends on eac and clie)
	t.Run("eac-ext aggregates specs from dependencies", func(t *testing.T) {
		bundleReport, err := GetSpecs(workspaceRoot, "eac-ext", "unreleased", "")
		if err != nil {
			t.Fatalf("GetSpecs(eac-ext) failed: %v", err)
		}

		// Get specs for dependency module directly
		depReport, err := GetSpecs(workspaceRoot, "eac", "unreleased", "")
		if err != nil {
			t.Fatalf("GetSpecs(eac) failed: %v", err)
		}

		// Bundle should have at least as many specs as the dependency
		// (unless dependency has no specs)
		if len(depReport.SpecFiles) > 0 && len(bundleReport.SpecFiles) < len(depReport.SpecFiles) {
			t.Errorf("Bundle eac-ext has fewer specs (%d) than dependency eac (%d)",
				len(bundleReport.SpecFiles), len(depReport.SpecFiles))
		}

		// Verify that specs from dependency are included in bundle
		if len(depReport.SpecFiles) > 0 {
			foundDepSpec := false
			for _, spec := range bundleReport.SpecFiles {
				if strings.Contains(spec.RelativePath, "specs/eac/") {
					foundDepSpec = true
					break
				}
			}
			if !foundDepSpec {
				t.Error("Bundle eac-ext should include specs from dependency eac")
			}
		}
	})

	// Test regular module (no dependencies)
	t.Run("regular module only includes own specs", func(t *testing.T) {
		report, err := GetSpecs(workspaceRoot, "eac", "unreleased", "")
		if err != nil {
			t.Fatalf("GetSpecs(eac) failed: %v", err)
		}

		// All specs should be from eac directory
		for _, spec := range report.SpecFiles {
			if !strings.Contains(spec.RelativePath, "specs/eac/") {
				t.Errorf("Regular module eac should only include own specs, found: %s",
					spec.RelativePath)
			}
		}
	})
}
