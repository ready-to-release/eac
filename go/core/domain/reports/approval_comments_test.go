package reports

import (
	"strings"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/ready-to-release/eac/go/core/git"
	coretesting "github.com/ready-to-release/eac/go/core/testing"
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

// mockGitRepo implements git.GitRepository for testing.
// It returns minimal/empty data to avoid expensive git operations.
type mockGitRepo struct {
	commits []git.CommitInfo
	tags    []string
}

func (m *mockGitRepo) RootPath() string                                  { return "" }
func (m *mockGitRepo) RemoteURL(remoteName string) (string, error)       { return "", nil }
func (m *mockGitRepo) AddRemote(name, url string) error                  { return nil }
func (m *mockGitRepo) CurrentBranch() (string, error)                    { return "main", nil }
func (m *mockGitRepo) HeadShortSHA() (string, error)                     { return "abc1234", nil }
func (m *mockGitRepo) HeadCommit() (string, error)                       { return "abc1234567890", nil }
func (m *mockGitRepo) UncommittedFiles() ([]string, error)               { return nil, nil }
func (m *mockGitRepo) TrackedFiles() ([]string, error)                   { return nil, nil }
func (m *mockGitRepo) StagedFiles() ([]string, error)                    { return nil, nil }
func (m *mockGitRepo) IsFileTracked(relPath string) bool                 { return true }
func (m *mockGitRepo) IsFileIgnored(relPath string) bool                 { return false }
func (m *mockGitRepo) Add(path string) error                             { return nil }
func (m *mockGitRepo) Commit(msg, name, email string) (string, error)    { return "abc1234", nil }
func (m *mockGitRepo) StagedDiff() (string, error)                       { return "", nil }
func (m *mockGitRepo) StagedDiffStats() (string, error)                  { return "", nil }
func (m *mockGitRepo) ConfigSet(section, key, value string) error        { return nil }
func (m *mockGitRepo) GoGitRepo() *gogit.Repository                      { return nil }
func (m *mockGitRepo) CommitsSince(ref string) ([]git.CommitInfo, error) { return m.commits, nil }
func (m *mockGitRepo) TagsMatching(pattern string) ([]string, error)     { return m.tags, nil }
func (m *mockGitRepo) LatestTag(pattern string) (string, error)          { return "", nil }
func (m *mockGitRepo) TagCommit(tagName string) (string, error)          { return "", nil }
func (m *mockGitRepo) TagDate(tagName string) (time.Time, error)         { return time.Time{}, nil }
func (m *mockGitRepo) TagExists(tagName string) (bool, error)            { return false, nil }
func (m *mockGitRepo) GetBranchCommits(base string) ([]git.CommitInfo, error) {
	return m.commits, nil
}
func (m *mockGitRepo) GetBranchDiff(baseBranch string) (string, error)      { return "", nil }
func (m *mockGitRepo) GetBranchDiffStats(baseBranch string) (string, error) { return "", nil }
func (m *mockGitRepo) GetBranchFiles(baseBranch string) ([]string, error)   { return nil, nil }

// CommitsBetween is the key method - returns mock commits to avoid expensive git operations
func (m *mockGitRepo) CommitsBetween(fromRef, toRef string) ([]git.CommitInfo, error) {
	return m.commits, nil
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

	// Set up mock GitHub CLI
	mockCLI := &mockGitHubCLI{
		prs: map[int]*PRData{},
	}
	SetGitHubCLI(mockCLI)
	defer SetGitHubCLI(nil)

	// Set up mock git repo to avoid expensive git operations (rename detection etc)
	mockRepo := &mockGitRepo{
		commits: []git.CommitInfo{}, // No commits = no expensive diff operations
		tags:    []string{},
	}
	SetGitRepo(mockRepo)
	defer SetGitRepo(nil)

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
	workspaceRoot := coretesting.SetupWorkspaceIsolation(t)

	// Set up mock GitHub CLI with sample PR data
	mockCLI := &mockGitHubCLI{
		prs: map[int]*PRData{
			123: {
				Number:             123,
				Title:              "Add new feature to eac",
				Author:             "developer1",
				Body:               "This PR adds a new feature to eac module",
				MergedAt:           time.Now(),
				MergeCommitMessage: "Merge pull request #123\n\nAdded new feature implementation",
				Files:              []string{"specs/eac/new-feature.feature", "go/cli/eac/impl/test.go"},
				Reviews: []ReviewData{
					{Author: "reviewer1", State: "APPROVED", SubmittedAt: time.Now()},
					{Author: "reviewer2", State: "APPROVED", SubmittedAt: time.Now()},
				},
			},
			456: {
				Number:             456,
				Title:              "Update clie spec",
				Author:             "developer2",
				Body:               "Updates the clie specification",
				MergedAt:           time.Now(),
				MergeCommitMessage: "Merge pull request #456\n\nUpdated clie spec",
				Files:              []string{"specs/clie/update.feature"},
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

	// Set up mock git repo with sample commits containing PR references
	mockRepo := &mockGitRepo{
		commits: []git.CommitInfo{
			{
				SHA:      "abc123",
				ShortSHA: "abc123",
				Message:  "feat: add new feature (#123)\n\nThis adds the feature.",
				Subject:  "feat: add new feature (#123)",
				Author:   "developer1",
				Date:     time.Now(),
				Files:    []string{"specs/eac/new-feature.feature"},
			},
			{
				SHA:      "def456",
				ShortSHA: "def456",
				Message:  "Merge pull request #456 from user/branch\n\nUpdate specs",
				Subject:  "Merge pull request #456 from user/branch",
				Author:   "developer2",
				Date:     time.Now(),
				Files:    []string{"specs/clie/update.feature"},
			},
		},
		tags: []string{},
	}
	SetGitRepo(mockRepo)
	defer SetGitRepo(nil)

	// Test eac-ext bundle module (depends on eac-cli and clie)
	t.Run("eac-ext returns valid report structure", func(t *testing.T) {
		bundleReport, err := GetApprovalComments(workspaceRoot, "eac-ext", "unreleased", false, "")
		if err != nil {
			t.Fatalf("GetApprovalComments(eac-ext) failed: %v", err)
		}

		if bundleReport == nil {
			t.Fatal("GetApprovalComments(eac-ext) returned nil")
		}

		// Verify structure
		if bundleReport.Module != "eac-ext" {
			t.Errorf("Module = %v, want eac-ext", bundleReport.Module)
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
				for _, prefix := range []string{"specs/eac/", "specs/clie/", "specs/eac-ext/"} {
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
		report, err := GetApprovalComments(workspaceRoot, "eac", "unreleased", false, "")
		if err != nil {
			t.Fatalf("GetApprovalComments(eac) failed: %v", err)
		}

		// All spec files should be from eac directory
		for _, approval := range report.Approvals {
			for _, specFile := range approval.SpecFiles {
				if !strings.HasPrefix(specFile, "specs/eac/") {
					t.Errorf("Regular module eac should only include own approvals, found: %s",
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
		"specs/eac/feature1.feature",
		"specs/eac/feature2.feature",
		"specs/clie/feature3.feature",
		"specs/other-module/feature4.feature",
		"go/cli/eac/test.go",
		"README.md",
	}

	tests := []struct {
		name     string
		modules  []string
		expected int
	}{
		{
			name:     "single module",
			modules:  []string{"eac"},
			expected: 2,
		},
		{
			name:     "multiple modules",
			modules:  []string{"eac", "clie"},
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
