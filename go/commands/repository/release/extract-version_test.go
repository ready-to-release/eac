package release

import (
	"testing"
)

func TestExtractVersionLogic_TagRef(t *testing.T) {
	tests := []struct {
		name        string
		module      string
		versionType string
		ref         string
		version     string
		wantVersion string
		wantTag     string
		wantValid   bool
	}{
		{
			name:        "semver from tag ref",
			module:      "clie",
			versionType: "semver",
			ref:         "refs/tags/clie/1.2.3",
			wantVersion: "1.2.3",
			wantTag:     "clie/1.2.3",
			wantValid:   true,
		},
		{
			name:        "calver from tag ref",
			module:      "docs",
			versionType: "calver",
			ref:         "refs/tags/docs/2024.1217.1430",
			wantVersion: "2024.1217.1430",
			wantTag:     "docs/2024.1217.1430",
			wantValid:   true,
		},
		{
			name:        "invalid semver with v prefix",
			module:      "clie",
			versionType: "semver",
			ref:         "refs/tags/clie/v1.2.3",
			wantVersion: "v1.2.3",
			wantTag:     "clie/v1.2.3",
			wantValid:   false,
		},
		{
			name:        "explicit version semver",
			module:      "clie",
			versionType: "semver",
			ref:         "",
			version:     "2.0.0",
			wantVersion: "2.0.0",
			wantTag:     "clie/2.0.0",
			wantValid:   true,
		},
		{
			name:        "no version provided semver",
			module:      "clie",
			versionType: "semver",
			ref:         "",
			version:     "",
			wantVersion: "",
			wantTag:     "",
			wantValid:   false,
		},
		{
			name:        "explicit version calver",
			module:      "docs",
			versionType: "calver",
			ref:         "",
			version:     "2024.0101.0000",
			wantVersion: "2024.0101.0000",
			wantTag:     "docs/2024.0101.0000",
			wantValid:   true,
		},
		{
			name:        "eac-ext semver from tag",
			module:      "eac-ext",
			versionType: "semver",
			ref:         "refs/tags/eac-ext/0.5.0",
			wantVersion: "0.5.0",
			wantTag:     "eac-ext/0.5.0",
			wantValid:   true,
		},
		{
			name:        "invalid calver format from tag",
			module:      "docs",
			versionType: "calver",
			ref:         "refs/tags/docs/24.1217.143",
			wantVersion: "24.1217.143",
			wantTag:     "docs/24.1217.143",
			wantValid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVersionLogic(tt.module, tt.versionType, tt.ref, tt.version)
			if result.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", result.Version, tt.wantVersion)
			}
			if result.TagName != tt.wantTag {
				t.Errorf("TagName = %q, want %q", result.TagName, tt.wantTag)
			}
			if result.IsValid != tt.wantValid {
				t.Errorf("IsValid = %v, want %v", result.IsValid, tt.wantValid)
			}
		})
	}
}

func TestExtractVersionLogic_CalverAutoGenerate(t *testing.T) {
	// When calver type and no version provided, should auto-generate
	result := extractVersionLogic("docs", "calver", "", "")

	if result.Version == "" {
		t.Error("expected auto-generated calver version, got empty")
	}
	if !result.IsValid {
		t.Errorf("expected valid calver, got invalid: %s", result.Message)
	}
	if result.TagName == "" {
		t.Error("expected tag name to be set")
	}
	if result.Message != "Auto-generated calver" {
		t.Errorf("expected auto-generate message, got: %s", result.Message)
	}
}

func TestIsValidSemver(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"1.0.0", true},
		{"0.1.0", true},
		{"10.20.30", true},
		{"0.0.0", true},
		{"123.456.789", true},
		{"v1.0.0", false},     // No v prefix
		{"V1.0.0", false},     // No V prefix
		{"1.0", false},        // Missing patch
		{"1", false},          // Missing minor and patch
		{"1.0.0.0", false},    // Too many parts
		{"01.0.0", false},     // Leading zero in major
		{"1.00.0", false},     // Leading zero in minor
		{"1.0.00", false},     // Leading zero in patch
		{"", false},           // Empty
		{"a.b.c", false},      // Non-numeric
		{"1.0.0-beta", false}, // Pre-release not supported in this format
		{" 1.0.0", false},     // Leading space
		{"1.0.0 ", false},     // Trailing space
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := isValidSemver(tt.version); got != tt.want {
				t.Errorf("isValidSemver(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestIsValidCalver(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"2024.1217.1430", true},
		{"2024.0101.0000", true},
		{"2025.1231.2359", true},
		{"1999.0101.0000", true},
		{"2024.1201.0930", true},
		{"24.1217.1430", false},    // Year not 4 digits
		{"2024.117.1430", false},   // Month/day not 4 digits
		{"2024.1217.143", false},   // Time not 4 digits
		{"2024.1217.14300", false}, // Time too many digits
		{"2024.12171430", false},   // Missing second dot
		{"20241217.1430", false},   // Missing first dot
		{"", false},                // Empty
		{"2024-1217-1430", false},  // Wrong separator
		{"2024/1217/1430", false},  // Wrong separator
		{" 2024.1217.1430", false}, // Leading space
		{"2024.1217.1430 ", false}, // Trailing space
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := isValidCalver(tt.version); got != tt.want {
				t.Errorf("isValidCalver(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestGenerateCalver(t *testing.T) {
	calver := generateCalver()

	// Should match YYYY.MMDD.HHMM format
	if !isValidCalver(calver) {
		t.Errorf("generateCalver() = %q, which is not valid calver format", calver)
	}

	// Should start with current year (reasonable sanity check)
	if len(calver) < 4 {
		t.Errorf("generateCalver() = %q, too short", calver)
	}
	year := calver[:4]
	if year < "2024" || year > "2100" {
		t.Errorf("generateCalver() year = %q, expected reasonable year", year)
	}
}

func TestExtractVersionOutput_Struct(t *testing.T) {
	output := ExtractVersionOutput{
		Version: "1.2.3",
		TagName: "clie/1.2.3",
		IsValid: true,
		Message: "",
	}

	if output.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", output.Version, "1.2.3")
	}
	if output.TagName != "clie/1.2.3" {
		t.Errorf("TagName = %q, want %q", output.TagName, "clie/1.2.3")
	}
	if !output.IsValid {
		t.Error("IsValid = false, want true")
	}
}

func TestExtractVersionLogic_ErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		module      string
		versionType string
		ref         string
		version     string
		wantMsg     string
	}{
		{
			name:        "no version provided semver",
			module:      "clie",
			versionType: "semver",
			ref:         "",
			version:     "",
			wantMsg:     "No version provided",
		},
		{
			name:        "invalid semver format",
			module:      "clie",
			versionType: "semver",
			ref:         "",
			version:     "v1.0.0",
			wantMsg:     "Invalid semver format",
		},
		{
			name:        "invalid calver format",
			module:      "docs",
			versionType: "calver",
			ref:         "",
			version:     "bad-calver",
			wantMsg:     "Invalid calver format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVersionLogic(tt.module, tt.versionType, tt.ref, tt.version)
			if result.Message == "" {
				t.Error("expected error message, got empty")
			}
			if len(tt.wantMsg) > 0 && !contains(result.Message, tt.wantMsg) {
				t.Errorf("Message = %q, want to contain %q", result.Message, tt.wantMsg)
			}
		})
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
