package reports

import (
	"strings"
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/core/git"
	coretesting "github.com/ready-to-release/eac/go/core/testing"
)

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name               string
		module             string
		versionStr         string
		expectedVersion    string
		expectedUnreleased bool
		expectedLatest     bool
		wantErr            bool
		errContains        string
	}{
		{
			name:               "unreleased - explicit",
			module:             "eac-ext",
			versionStr:         "unreleased",
			expectedVersion:    "Unreleased",
			expectedUnreleased: true,
			expectedLatest:     false,
			wantErr:            false,
		},
		{
			name:               "unreleased - implicit (empty)",
			module:             "eac-ext",
			versionStr:         "",
			expectedVersion:    "Unreleased",
			expectedUnreleased: true,
			expectedLatest:     false,
			wantErr:            false,
		},
		{
			name:               "latest version",
			module:             "eac-ext",
			versionStr:         "latest",
			expectedVersion:    "", // Actual version from changelog
			expectedUnreleased: false,
			expectedLatest:     true,
			wantErr:            false,
		},
		{
			name:        "invalid module",
			module:      "nonexistent-module-xyz",
			versionStr:  "latest",
			wantErr:     true,
			errContains: "module not found",
		},
		{
			name:        "version not found",
			module:      "eac-ext",
			versionStr:  "99.99.99",
			wantErr:     true,
			errContains: "version not found",
		},
		// Implicit versioned modules (no changelog)
		{
			name:               "implicit module - unreleased explicit",
			module:             "core",
			versionStr:         "unreleased",
			expectedVersion:    "Unreleased",
			expectedUnreleased: true,
			expectedLatest:     false,
			wantErr:            false,
		},
		{
			name:               "implicit module - unreleased implicit (empty)",
			module:             "core",
			versionStr:         "",
			expectedVersion:    "Unreleased",
			expectedUnreleased: true,
			expectedLatest:     false,
			wantErr:            false,
		},
		{
			name:        "implicit module - latest not supported",
			module:      "core",
			versionStr:  "latest",
			wantErr:     true,
			errContains: "does not support version",
		},
		{
			name:        "implicit module - specific version not supported",
			module:      "core",
			versionStr:  "1.0.0",
			wantErr:     true,
			errContains: "does not support version",
		},
	}

	workspaceRoot := coretesting.SetupWorkspaceIsolation(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ResolveVersion(workspaceRoot, tt.module, tt.versionStr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolveVersion() expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ResolveVersion() error = %v, should contain %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ResolveVersion() unexpected error = %v", err)
				return
			}

			if info.Module != tt.module {
				t.Errorf("Module = %v, want %v", info.Module, tt.module)
			}

			if tt.expectedVersion != "" && info.VersionNumber != tt.expectedVersion {
				t.Errorf("VersionNumber = %v, want %v", info.VersionNumber, tt.expectedVersion)
			}

			if info.IsUnreleased != tt.expectedUnreleased {
				t.Errorf("IsUnreleased = %v, want %v", info.IsUnreleased, tt.expectedUnreleased)
			}

			if info.IsLatest != tt.expectedLatest {
				t.Errorf("IsLatest = %v, want %v", info.IsLatest, tt.expectedLatest)
			}

			// Verify git tags are properly formatted when set
			if info.GitTag != "" && !strings.Contains(info.GitTag, "/") {
				t.Errorf("GitTag = %v, should contain module/version format", info.GitTag)
			}
		})
	}
}

func TestResolveBranchEndRef(t *testing.T) {
	tests := []struct {
		name        string
		branch      string
		trunkBranch string
		expected    string
	}{
		{
			name:        "empty branch with trunk configured",
			branch:      "",
			trunkBranch: "master",
			expected:    "master",
		},
		{
			name:        "empty branch with empty trunk defaults to main",
			branch:      "",
			trunkBranch: "",
			expected:    "main",
		},
		{
			name:        "explicit branch used as-is",
			branch:      "develop",
			trunkBranch: "main",
			expected:    "develop",
		},
		{
			name:        "HEAD resolves to HEAD",
			branch:      "HEAD",
			trunkBranch: "main",
			expected:    "HEAD",
		},
		{
			name:        "current resolves to HEAD",
			branch:      "current",
			trunkBranch: "main",
			expected:    "HEAD",
		},
		{
			name:        "explicit main branch",
			branch:      "main",
			trunkBranch: "master",
			expected:    "main",
		},
		{
			name:        "feature branch",
			branch:      "feature/my-feature",
			trunkBranch: "main",
			expected:    "feature/my-feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveBranchEndRef(tt.branch, tt.trunkBranch)
			if result != tt.expected {
				t.Errorf("ResolveBranchEndRef(%q, %q) = %q, want %q",
					tt.branch, tt.trunkBranch, result, tt.expected)
			}
		})
	}
}

func TestResolveVersionWithValidation(t *testing.T) {
	workspaceRoot := coretesting.SetupWorkspaceIsolation(t)

	// Test with mock repository to simulate missing tags
	t.Run("tag exists - validation passes", func(t *testing.T) {
		deps := &ReportDeps{
			VersionResolverRepo: git.NewMockRepository(workspaceRoot).
				WithTag("eac-ext/0.0.1", "abc123", time.Now()),
		}

		info, err := ResolveVersionWithValidation(deps, workspaceRoot, "eac-ext", "0.0.1")
		if err != nil {
			t.Errorf("ResolveVersionWithValidation() unexpected error: %v", err)
			return
		}
		if info.VersionNumber != "0.0.1" {
			t.Errorf("VersionNumber = %v, want 0.0.1", info.VersionNumber)
		}
	})

	t.Run("tag missing - validation fails with diagnostic message", func(t *testing.T) {
		deps := &ReportDeps{
			VersionResolverRepo: git.NewMockRepository(workspaceRoot).
				WithTag("eac-ext/0.0.7", "abc123", time.Now()).
				WithTag("eac-ext/0.0.6", "def456", time.Now()),
		}

		_, err := ResolveVersionWithValidation(deps, workspaceRoot, "eac-ext", "0.0.1")
		if err == nil {
			t.Error("ResolveVersionWithValidation() expected error for missing tag, got none")
			return
		}

		// Verify error message contains diagnostic information
		errStr := err.Error()
		if !strings.Contains(errStr, "eac-ext/0.0.1") {
			t.Errorf("Error should mention the missing tag, got: %v", errStr)
		}
		if !strings.Contains(errStr, "not found") {
			t.Errorf("Error should say tag not found, got: %v", errStr)
		}
		if !strings.Contains(errStr, "Available") {
			t.Errorf("Error should list available tags, got: %v", errStr)
		}
		if !strings.Contains(errStr, "Possible causes") {
			t.Errorf("Error should suggest possible causes, got: %v", errStr)
		}
		if !strings.Contains(errStr, "fetch-depth") {
			t.Errorf("Error should mention fetch-depth as possible cause, got: %v", errStr)
		}
	})

	t.Run("tag missing - no other tags available", func(t *testing.T) {
		deps := &ReportDeps{
			VersionResolverRepo: git.NewMockRepository(workspaceRoot),
		}

		_, err := ResolveVersionWithValidation(deps, workspaceRoot, "eac-ext", "0.0.1")
		if err == nil {
			t.Error("ResolveVersionWithValidation() expected error for missing tag, got none")
			return
		}

		// Verify error mentions no tags found
		if !strings.Contains(err.Error(), "none found") {
			t.Errorf("Error should mention no tags found, got: %v", err)
		}
	})

	t.Run("unreleased version - bypasses tag validation", func(t *testing.T) {
		deps := &ReportDeps{
			VersionResolverRepo: git.NewMockRepository(workspaceRoot),
		}

		info, err := ResolveVersionWithValidation(deps, workspaceRoot, "eac-ext", "unreleased")
		if err != nil {
			t.Errorf("ResolveVersionWithValidation(unreleased) should not fail: %v", err)
			return
		}
		if !info.IsUnreleased {
			t.Error("Expected IsUnreleased to be true")
		}
	})

	t.Run("empty version - treated as unreleased, bypasses validation", func(t *testing.T) {
		deps := &ReportDeps{
			VersionResolverRepo: git.NewMockRepository(workspaceRoot),
		}

		info, err := ResolveVersionWithValidation(deps, workspaceRoot, "eac-ext", "")
		if err != nil {
			t.Errorf("ResolveVersionWithValidation('') should not fail: %v", err)
			return
		}
		if !info.IsUnreleased {
			t.Error("Expected IsUnreleased to be true for empty version")
		}
	})

	t.Run("latest version with tag - validation passes", func(t *testing.T) {
		// Get the actual latest version from changelog
		info, err := ResolveVersion(workspaceRoot, "eac-ext", "latest")
		if err != nil {
			t.Skipf("Could not resolve latest version: %v", err)
		}

		deps := &ReportDeps{
			VersionResolverRepo: git.NewMockRepository(workspaceRoot).
				WithTag(info.GitTag, "abc123", time.Now()),
		}

		validatedInfo, err := ResolveVersionWithValidation(deps, workspaceRoot, "eac-ext", "latest")
		if err != nil {
			t.Errorf("ResolveVersionWithValidation(latest) unexpected error: %v", err)
			return
		}
		if validatedInfo.VersionNumber != info.VersionNumber {
			t.Errorf("VersionNumber = %v, want %v", validatedInfo.VersionNumber, info.VersionNumber)
		}
	})

	t.Run("latest version without tag - validation fails", func(t *testing.T) {
		// Get the actual latest version from changelog
		info, err := ResolveVersion(workspaceRoot, "eac-ext", "latest")
		if err != nil {
			t.Skipf("Could not resolve latest version: %v", err)
		}

		deps := &ReportDeps{
			VersionResolverRepo: git.NewMockRepository(workspaceRoot),
		}

		_, err = ResolveVersionWithValidation(deps, workspaceRoot, "eac-ext", "latest")
		if err == nil {
			t.Error("ResolveVersionWithValidation(latest) should fail when tag is missing")
			return
		}

		// Verify error mentions the expected tag
		if !strings.Contains(err.Error(), info.GitTag) {
			t.Errorf("Error should mention missing tag %s, got: %v", info.GitTag, err)
		}
	})
}
