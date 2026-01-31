//go:build L1 && ov
// +build L1,ov

package domain

import (
	"testing"
)

func TestBaseContract_Getters(t *testing.T) {
	contract := BaseContract{
		Moniker:     "test-moniker",
		Name:        "Test Name",
		Description: "Test description",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "test/root"},
		},
	}

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"GetMoniker", contract.GetMoniker(), "test-moniker"},
		{"GetName", contract.GetName(), "Test Name"},
		{"GetDescription", contract.GetDescription(), "Test description"},
		{"GetComponentRoot", contract.GetComponentRoot("go"), "test/root"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: expected '%s', got '%s'", tt.name, tt.expected, tt.got)
			}
		})
	}
}

// TestBaseContract_GetChangelog validates the GetChangelog method behavior.
// Tests explicit changelog paths, default path fallback, and versioning scheme behavior.
func TestBaseContract_GetChangelog(t *testing.T) {
	tests := []struct {
		name     string
		contract BaseContract
		want     string
	}{
		{
			name: "explicit changelog path",
			contract: BaseContract{
				Moniker: "test-module",
				Versioning: &ModuleVersioning{
					Scheme:    "SemVer",
					Changelog: "custom/path/CHANGELOG.md",
				},
			},
			want: "custom/path/CHANGELOG.md",
		},
		{
			name: "SemVer module defaults to release folder",
			contract: BaseContract{
				Moniker: "test-module",
				Versioning: &ModuleVersioning{
					Scheme: "SemVer",
				},
			},
			want: "release/test-module/CHANGELOG.md",
		},
		{
			name: "CalVer module returns empty (auto-managed)",
			contract: BaseContract{
				Moniker: "my-module",
				Versioning: &ModuleVersioning{
					Scheme: "CalVer",
				},
			},
			want: "",
		},
		{
			name: "Implicit module returns empty (non-releasable)",
			contract: BaseContract{
				Moniker: "internal-module",
				Versioning: &ModuleVersioning{
					Scheme: "Implicit",
				},
			},
			want: "",
		},
		{
			name: "nil versioning returns empty",
			contract: BaseContract{
				Moniker:    "no-version-module",
				Versioning: nil,
			},
			want: "",
		},
		{
			name: "explicit changelog with CalVer module",
			contract: BaseContract{
				Moniker: "eac-mcp-commands",
				Versioning: &ModuleVersioning{
					Scheme:    "CalVer",
					Changelog: "go/eac/mcp/commands/CHANGELOG.md",
				},
			},
			want: "go/eac/mcp/commands/CHANGELOG.md",
		},
		{
			name: "SemVer module with special characters in moniker",
			contract: BaseContract{
				Moniker: "eac-mcp-commands",
				Versioning: &ModuleVersioning{
					Scheme: "SemVer",
				},
			},
			want: "release/eac-mcp-commands/CHANGELOG.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.contract.GetChangelog()
			if got != tt.want {
				t.Errorf("GetChangelog() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestBaseContract_GetChangelog_ConsistentDefault validates that the default path
// format is consistent across multiple invocations for SemVer modules.
func TestBaseContract_GetChangelog_ConsistentDefault(t *testing.T) {
	contract := BaseContract{
		Moniker: "test-module",
		Versioning: &ModuleVersioning{
			Scheme: "SemVer",
		},
	}

	// Call multiple times to ensure consistency
	path1 := contract.GetChangelog()
	path2 := contract.GetChangelog()
	path3 := contract.GetChangelog()

	if path1 != path2 || path2 != path3 {
		t.Errorf("GetChangelog() returned inconsistent results: %q, %q, %q", path1, path2, path3)
	}

	expectedPattern := "release/test-module/CHANGELOG.md"
	if path1 != expectedPattern {
		t.Errorf("GetChangelog() = %q, want %q", path1, expectedPattern)
	}
}

// TestBaseContract_GetChangelog_EdgeCases tests edge cases and unusual inputs.
func TestBaseContract_GetChangelog_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		contract BaseContract
		want     string
	}{
		{
			name: "empty moniker with explicit changelog",
			contract: BaseContract{
				Moniker: "",
				Versioning: &ModuleVersioning{
					Changelog: "path/CHANGELOG.md",
				},
			},
			want: "path/CHANGELOG.md",
		},
		{
			name: "empty moniker SemVer module",
			contract: BaseContract{
				Moniker: "",
				Versioning: &ModuleVersioning{
					Scheme: "SemVer",
				},
			},
			want: "release//CHANGELOG.md", // Edge case: double slash
		},
		{
			name: "moniker with slash for SemVer module",
			contract: BaseContract{
				Moniker: "module/submodule",
				Versioning: &ModuleVersioning{
					Scheme: "SemVer",
				},
			},
			want: "release/module/submodule/CHANGELOG.md",
		},
		{
			name: "absolute path in changelog",
			contract: BaseContract{
				Moniker: "test",
				Versioning: &ModuleVersioning{
					Changelog: "/absolute/path/CHANGELOG.md",
				},
			},
			want: "/absolute/path/CHANGELOG.md",
		},
		{
			name: "CalVer without explicit changelog returns empty",
			contract: BaseContract{
				Moniker: "test",
				Versioning: &ModuleVersioning{
					Scheme: "CalVer",
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.contract.GetChangelog()
			if got != tt.want {
				t.Errorf("GetChangelog() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestModuleVersioning_ReleaseType validates the ReleaseType field behavior.
func TestModuleVersioning_ReleaseType(t *testing.T) {
	tests := []struct {
		name        string
		versioning  *ModuleVersioning
		wantType    string
		description string
	}{
		{
			name: "published release type",
			versioning: &ModuleVersioning{
				Scheme:      "SemVer",
				Changelog:   "release/r2r-cli/CHANGELOG.md",
				ReleaseType: "published",
			},
			wantType:    "published",
			description: "module creates GitHub releases with downloadable artifacts",
		},
		{
			name: "internal release type",
			versioning: &ModuleVersioning{
				Scheme:      "SemVer",
				Changelog:   "go/eac/commands/CHANGELOG.md",
				ReleaseType: "internal",
			},
			wantType:    "internal",
			description: "module is version tracked for dependencies, not published",
		},
		{
			name: "bundle release type",
			versioning: &ModuleVersioning{
				Scheme:      "CalVer",
				Changelog:   "release/r2r-eac-bundle/CHANGELOG.md",
				ReleaseType: "bundle",
			},
			wantType:    "bundle",
			description: "module aggregates other releases",
		},
		{
			name: "none release type",
			versioning: &ModuleVersioning{
				Scheme:      "Implicit",
				ReleaseType: "none",
			},
			wantType:    "none",
			description: "module has no versioning or releases",
		},
		{
			name: "empty release type defaults to empty",
			versioning: &ModuleVersioning{
				Scheme:    "SemVer",
				Changelog: "release/test/CHANGELOG.md",
			},
			wantType:    "",
			description: "when not specified, release_type is empty string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.versioning.ReleaseType != tt.wantType {
				t.Errorf("ReleaseType = %q, want %q", tt.versioning.ReleaseType, tt.wantType)
			}
		})
	}
}

// TestBaseContract_ReleaseTypeConsistency validates that release_type aligns with changelog location.
// This test implements the validation rules from the architecture plan.
func TestBaseContract_ReleaseTypeConsistency(t *testing.T) {
	tests := []struct {
		name        string
		contract    BaseContract
		shouldBeOK  bool
		description string
	}{
		{
			name: "published with release/ changelog - OK",
			contract: BaseContract{
				Moniker: "r2r-cli",
				Versioning: &ModuleVersioning{
					Scheme:      "SemVer",
					Changelog:   "release/r2r-cli/CHANGELOG.md",
					ReleaseType: "published",
				},
			},
			shouldBeOK:  true,
			description: "published modules must have changelogs in release/",
		},
		{
			name: "published with non-release/ changelog - INVALID",
			contract: BaseContract{
				Moniker: "bad-published",
				Versioning: &ModuleVersioning{
					Scheme:      "SemVer",
					Changelog:   "go/bad/CHANGELOG.md",
					ReleaseType: "published",
				},
			},
			shouldBeOK:  false,
			description: "published modules cannot have changelogs outside release/",
		},
		{
			name: "internal with module root changelog - OK",
			contract: BaseContract{
				Moniker: "eac-commands",
				Versioning: &ModuleVersioning{
					Scheme:      "SemVer",
					Changelog:   "go/eac/commands/CHANGELOG.md",
					ReleaseType: "internal",
				},
			},
			shouldBeOK:  true,
			description: "internal modules should have changelogs in module root",
		},
		{
			name: "internal with release/ changelog - INVALID",
			contract: BaseContract{
				Moniker: "bad-internal",
				Versioning: &ModuleVersioning{
					Scheme:      "SemVer",
					Changelog:   "release/bad-internal/CHANGELOG.md",
					ReleaseType: "internal",
				},
			},
			shouldBeOK:  false,
			description: "internal modules should not have changelogs in release/",
		},
		{
			name: "bundle with release/ changelog - OK",
			contract: BaseContract{
				Moniker: "r2r-eac-bundle",
				Versioning: &ModuleVersioning{
					Scheme:      "CalVer",
					Changelog:   "release/r2r-eac-bundle/CHANGELOG.md",
					ReleaseType: "bundle",
				},
			},
			shouldBeOK:  true,
			description: "bundle modules are treated like published",
		},
		{
			name: "none release type - always OK",
			contract: BaseContract{
				Moniker: "eac-core",
				Versioning: &ModuleVersioning{
					Scheme:      "Implicit",
					ReleaseType: "none",
				},
			},
			shouldBeOK:  true,
			description: "none release type has no changelog requirements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isConsistent := validateReleaseTypeConsistency(tt.contract)
			if isConsistent != tt.shouldBeOK {
				t.Errorf("%s: consistency = %v, want %v", tt.description, isConsistent, tt.shouldBeOK)
			}
		})
	}
}

// validateReleaseTypeConsistency checks if a module's release_type aligns with its changelog location.
// This is a helper function that will be used by validation commands.
func validateReleaseTypeConsistency(contract BaseContract) bool {
	if contract.Versioning == nil {
		return true // No versioning = no requirements
	}

	releaseType := contract.Versioning.ReleaseType
	changelogPath := contract.GetChangelog()

	// Determine if changelog is in release/ folder
	isInReleaseFolder := len(changelogPath) >= 8 && changelogPath[:8] == "release/"

	switch releaseType {
	case "published", "bundle":
		// Must be in release/ folder
		return isInReleaseFolder
	case "internal":
		// Must NOT be in release/ folder
		return !isInReleaseFolder
	case "none", "":
		// No restrictions
		return true
	default:
		// Unknown release type
		return false
	}
}

