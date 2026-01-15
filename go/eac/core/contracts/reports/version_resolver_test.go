package reports

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/git"
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
			module:             "ext-eac",
			versionStr:         "unreleased",
			expectedVersion:    "Unreleased",
			expectedUnreleased: true,
			expectedLatest:     false,
			wantErr:            false,
		},
		{
			name:               "unreleased - implicit (empty)",
			module:             "ext-eac",
			versionStr:         "",
			expectedVersion:    "Unreleased",
			expectedUnreleased: true,
			expectedLatest:     false,
			wantErr:            false,
		},
		{
			name:               "latest version",
			module:             "ext-eac",
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
			module:      "ext-eac",
			versionStr:  "99.99.99",
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

func TestResolveVersionWithValidation(t *testing.T) {
	// Get repository root for integration tests
	workspaceRoot := os.Getenv("WORKSPACE_ROOT")
	if workspaceRoot == "" {
		t.Skip("WORKSPACE_ROOT not set, skipping integration test")
	}

	// Test with mock repository to simulate missing tags
	t.Run("tag exists - validation passes", func(t *testing.T) {
		// Create mock with existing tag
		mockRepo := git.NewMockRepository(workspaceRoot).
			WithTag("ext-eac/0.0.8", "abc123", time.Now())

		// Inject mock
		SetVersionResolverRepo(mockRepo)
		defer SetVersionResolverRepo(nil)

		info, err := ResolveVersionWithValidation(workspaceRoot, "ext-eac", "0.0.8")
		if err != nil {
			t.Errorf("ResolveVersionWithValidation() unexpected error: %v", err)
			return
		}
		if info.VersionNumber != "0.0.8" {
			t.Errorf("VersionNumber = %v, want 0.0.8", info.VersionNumber)
		}
	})

	t.Run("tag missing - validation fails with diagnostic message", func(t *testing.T) {
		// Create mock WITHOUT the expected tag but WITH some other tags
		mockRepo := git.NewMockRepository(workspaceRoot).
			WithTag("ext-eac/0.0.7", "abc123", time.Now()).
			WithTag("ext-eac/0.0.6", "def456", time.Now())

		// Inject mock
		SetVersionResolverRepo(mockRepo)
		defer SetVersionResolverRepo(nil)

		_, err := ResolveVersionWithValidation(workspaceRoot, "ext-eac", "0.0.8")
		if err == nil {
			t.Error("ResolveVersionWithValidation() expected error for missing tag, got none")
			return
		}

		// Verify error message contains diagnostic information
		errStr := err.Error()
		if !strings.Contains(errStr, "ext-eac/0.0.8") {
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
		// Create mock with no tags at all
		mockRepo := git.NewMockRepository(workspaceRoot)

		// Inject mock
		SetVersionResolverRepo(mockRepo)
		defer SetVersionResolverRepo(nil)

		_, err := ResolveVersionWithValidation(workspaceRoot, "ext-eac", "0.0.8")
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
		// Create mock with no tags - unreleased should still work
		mockRepo := git.NewMockRepository(workspaceRoot)

		// Inject mock
		SetVersionResolverRepo(mockRepo)
		defer SetVersionResolverRepo(nil)

		info, err := ResolveVersionWithValidation(workspaceRoot, "ext-eac", "unreleased")
		if err != nil {
			t.Errorf("ResolveVersionWithValidation(unreleased) should not fail: %v", err)
			return
		}
		if !info.IsUnreleased {
			t.Error("Expected IsUnreleased to be true")
		}
	})

	t.Run("empty version - treated as unreleased, bypasses validation", func(t *testing.T) {
		// Create mock with no tags
		mockRepo := git.NewMockRepository(workspaceRoot)

		// Inject mock
		SetVersionResolverRepo(mockRepo)
		defer SetVersionResolverRepo(nil)

		info, err := ResolveVersionWithValidation(workspaceRoot, "ext-eac", "")
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
		info, err := ResolveVersion(workspaceRoot, "ext-eac", "latest")
		if err != nil {
			t.Skipf("Could not resolve latest version: %v", err)
		}

		// Create mock with the expected tag
		mockRepo := git.NewMockRepository(workspaceRoot).
			WithTag(info.GitTag, "abc123", time.Now())

		// Inject mock
		SetVersionResolverRepo(mockRepo)
		defer SetVersionResolverRepo(nil)

		validatedInfo, err := ResolveVersionWithValidation(workspaceRoot, "ext-eac", "latest")
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
		info, err := ResolveVersion(workspaceRoot, "ext-eac", "latest")
		if err != nil {
			t.Skipf("Could not resolve latest version: %v", err)
		}

		// Create mock WITHOUT the tag
		mockRepo := git.NewMockRepository(workspaceRoot)

		// Inject mock
		SetVersionResolverRepo(mockRepo)
		defer SetVersionResolverRepo(nil)

		_, err = ResolveVersionWithValidation(workspaceRoot, "ext-eac", "latest")
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
