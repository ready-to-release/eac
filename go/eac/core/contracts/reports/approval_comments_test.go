package reports

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/git"
)

// mockGitHubCLI implements GitHubCLI for testing.
type mockGitHubCLI struct {
	prs map[int]*PRData
}

func (m *mockGitHubCLI) GetPR(workspaceRoot string, prNumber int) (*PRData, error) {
	pr, ok := m.prs[prNumber]
	if !ok {
		return nil, ErrPRNotFound
	}
	return pr, nil
}

func TestGetApprovalComments(t *testing.T) {
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

	// Set up mock GitHub CLI
	mockCLI := &mockGitHubCLI{
		prs: map[int]*PRData{},
	}
	SetGitHubCLI(mockCLI)
	defer SetGitHubCLI(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := GetApprovalComments(workspaceRoot, tt.module, tt.version, false, "")

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetApprovalComments() expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetApprovalComments() error = %v, should contain %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("GetApprovalComments() unexpected error = %v", err)
				return
			}

			if report == nil {
				t.Errorf("GetApprovalComments() returned nil report")
				return
			}

			if report.Module != tt.expectedModule {
				t.Errorf("GetApprovalComments() module = %v, want %v", report.Module, tt.expectedModule)
			}

			if report.Version == "" {
				t.Errorf("GetApprovalComments() version is empty")
			}

			// Verify approvals structure
			if report.Approvals == nil {
				t.Errorf("GetApprovalComments() Approvals is nil")
			}

			// Verify counts are consistent
			if report.TotalApprovals != len(report.Approvals) {
				t.Errorf("GetApprovalComments() TotalApprovals = %d, but got %d approvals",
					report.TotalApprovals, len(report.Approvals))
			}
		})
	}
}

// TestGetApprovalComments_BundleModuleAggregation verifies that bundle modules
// aggregate approvals from all their dependencies.
func TestGetApprovalComments_BundleModuleAggregation(t *testing.T) {
	// Get repository root
	workspaceRoot := os.Getenv("WORKSPACE_ROOT")
	if workspaceRoot == "" {
		t.Skip("WORKSPACE_ROOT not set, skipping integration test")
	}

	// Set up mock GitHub CLI with sample PR data
	mockCLI := &mockGitHubCLI{
		prs: map[int]*PRData{
			123: {
				Number:             123,
				Title:              "Add new feature to eac-commands",
				Author:             "developer1",
				Body:               "This PR adds a new feature to eac-commands module",
				MergedAt:           time.Now(),
				MergeCommitMessage: "Merge pull request #123\n\nAdded new feature implementation",
				Files:              []string{"specs/eac-commands/new-feature.feature", "go/eac/commands/impl/test.go"},
				Reviews: []ReviewData{
					{Author: "reviewer1", State: "APPROVED", SubmittedAt: time.Now()},
					{Author: "reviewer2", State: "APPROVED", SubmittedAt: time.Now()},
				},
			},
			456: {
				Number:             456,
				Title:              "Update r2r-cli spec",
				Author:             "developer2",
				Body:               "Updates the r2r-cli specification",
				MergedAt:           time.Now(),
				MergeCommitMessage: "Merge pull request #456\n\nUpdated r2r-cli spec",
				Files:              []string{"specs/r2r-cli/update.feature"},
				Reviews: []ReviewData{
					{Author: "reviewer3", State: "APPROVED", SubmittedAt: time.Now()},
				},
			},
			789: {
				Number:             789,
				Title:              "Non-spec PR",
				Author:             "developer3",
				Body:               "This PR contains no spec files",
				MergedAt:           time.Now(),
				MergeCommitMessage: "Merge pull request #789\n\nNon-spec changes",
				Files:              []string{"go/eac/core/test.go"},
				Reviews: []ReviewData{
					{Author: "reviewer4", State: "APPROVED", SubmittedAt: time.Now()},
				},
			},
		},
	}
	SetGitHubCLI(mockCLI)
	defer SetGitHubCLI(nil)

	// Note: This test will work with actual git commits in the repo
	// The mock only provides PR data when PR numbers are found in commits

	// Test ext-eac bundle module (depends on eac-commands and r2r-cli)
	t.Run("ext-eac returns valid report structure", func(t *testing.T) {
		bundleReport, err := GetApprovalComments(workspaceRoot, "ext-eac", "unreleased", false, "")
		if err != nil {
			t.Fatalf("GetApprovalComments(ext-eac) failed: %v", err)
		}

		if bundleReport == nil {
			t.Fatal("GetApprovalComments(ext-eac) returned nil")
		}

		// Verify structure
		if bundleReport.Module != "ext-eac" {
			t.Errorf("Module = %v, want ext-eac", bundleReport.Module)
		}

		if bundleReport.Approvals == nil {
			t.Error("Approvals should not be nil")
		}

		// Verify all approvals are from spec PRs
		for _, approval := range bundleReport.Approvals {
			if len(approval.SpecFiles) == 0 {
				t.Errorf("Approval for PR #%d has no spec files", approval.PRNumber)
			}

			// Verify spec files are from queried modules
			for _, specFile := range approval.SpecFiles {
				hasValidPrefix := false
				for _, prefix := range []string{"specs/eac-commands/", "specs/r2r-cli/", "specs/ext-eac/"} {
					if strings.HasPrefix(specFile, prefix) {
						hasValidPrefix = true
						break
					}
				}
				if !hasValidPrefix {
					t.Errorf("Spec file %s doesn't match expected module prefixes", specFile)
				}
			}
		}
	})

	// Test regular module (no dependencies)
	t.Run("regular module only includes own approvals", func(t *testing.T) {
		report, err := GetApprovalComments(workspaceRoot, "eac-commands", "unreleased", false, "")
		if err != nil {
			t.Fatalf("GetApprovalComments(eac-commands) failed: %v", err)
		}

		// All spec files should be from eac-commands directory
		for _, approval := range report.Approvals {
			for _, specFile := range approval.SpecFiles {
				if !strings.HasPrefix(specFile, "specs/eac-commands/") {
					t.Errorf("Regular module eac-commands should only include own approvals, found: %s",
						specFile)
				}
			}
		}
	})
}

func TestExtractPRNumbers(t *testing.T) {
	commits := []git.CommitInfo{
		{Message: "feat: add new feature (#123)"},
		{Message: "Merge pull request #456 from user/branch"},
		{Message: "fix: bug fix (#789)"},
		{Message: "Regular commit without PR"},
		{Message: "Another feature (#123)"}, // Duplicate
	}

	prNumbers := extractPRNumbers(commits)

	expected := []int{123, 456, 789}
	if len(prNumbers) != len(expected) {
		t.Errorf("extractPRNumbers() returned %d PRs, want %d", len(prNumbers), len(expected))
	}

	for i, num := range prNumbers {
		if num != expected[i] {
			t.Errorf("extractPRNumbers()[%d] = %d, want %d", i, num, expected[i])
		}
	}
}

func TestFilterSpecFiles(t *testing.T) {
	files := []string{
		"specs/eac-commands/feature1.feature",
		"specs/eac-commands/feature2.feature",
		"specs/r2r-cli/feature3.feature",
		"specs/other-module/feature4.feature",
		"go/eac/commands/test.go",
		"README.md",
	}

	tests := []struct {
		name     string
		modules  []string
		expected int
	}{
		{
			name:     "single module",
			modules:  []string{"eac-commands"},
			expected: 2,
		},
		{
			name:     "multiple modules",
			modules:  []string{"eac-commands", "r2r-cli"},
			expected: 3,
		},
		{
			name:     "non-matching module",
			modules:  []string{"nonexistent"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterSpecFiles(files, tt.modules)
			if len(result) != tt.expected {
				t.Errorf("filterSpecFiles() returned %d files, want %d", len(result), tt.expected)
			}
		})
	}
}
