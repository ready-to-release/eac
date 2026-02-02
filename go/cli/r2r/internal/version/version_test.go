//go:build L0
// +build L0

package version

import (
	"testing"
)

func TestEnsurePrefix(t *testing.T) {
	tests := []struct {
		version  string
		prefix   string
		expected string
	}{
		{"1.0.0", "v", "v1.0.0"},
		{"v1.0.0", "v", "v1.0.0"},
		{"", "v", "v"},
		{"1.0.0", "", "1.0.0"},
	}

	for _, tt := range tests {
		result := EnsurePrefix(tt.version, tt.prefix)
		if result != tt.expected {
			t.Errorf("EnsurePrefix(%q, %q) = %q; want %q", tt.version, tt.prefix, result, tt.expected)
		}
	}
}

func TestSplitVersion(t *testing.T) {
	tests := []struct {
		version    string
		main       string
		prerelease string
	}{
		{"1.0.0", "1.0.0", ""},
		{"1.0.0-alpha", "1.0.0", "alpha"},
		{"1.0.0+build", "1.0.0", ""},
		{"1.0.0-alpha+build", "1.0.0", "alpha"},
		{"2.1.3-beta.1", "2.1.3", "beta.1"},
	}

	for _, tt := range tests {
		main, prerelease := SplitVersion(tt.version)
		if main != tt.main || prerelease != tt.prerelease {
			t.Errorf("SplitVersion(%q) = (%q, %q); want (%q, %q)",
				tt.version, main, prerelease, tt.main, tt.prerelease)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"v1.0.0", "1.0.0", 0},
		{"1.0.0", "v1.0.0", 0},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0", "1.0.0-alpha", 1},
		{"2.0.0", "1.9.9", 1},
		{"1.9.9", "2.0.0", -1},
		{"1.2.3", "1.2.10", -1},
		{"invalid", "1.0.0", 0},
		{"1.0.0", "invalid", 0},
	}

	for _, tt := range tests {
		result := CompareVersions(tt.v1, tt.v2)
		if result != tt.expected {
			t.Errorf("CompareVersions(%q, %q) = %d; want %d", tt.v1, tt.v2, result, tt.expected)
		}
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"1.0.0", true},
		{"v1.0.0", true},
		{"1.0.0-alpha", true},
		{"1.0.0+build", true},
		{"1.0.0-alpha.1+build.2", true},
		{"1.0", false},
		{"1", false},
		{"1.0.0.0", false},
		{"invalid", false},
		{"", false},
		{"v", false},
	}

	for _, tt := range tests {
		result := IsValid(tt.version)
		if result != tt.valid {
			t.Errorf("IsValid(%q) = %t; want %t", tt.version, result, tt.valid)
		}
	}
}

func TestSetAndGetInfo(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalCommit := Commit

	// Test setting version info
	SetVersion("2.0.0", "2024-01-01T00:00:00Z", "abcd1234", "2024-01-01", "clean")

	info := GetInfo()
	if info.Version != "2.0.0" {
		t.Errorf("GetInfo().Version = %q; want %q", info.Version, "2.0.0")
	}
	if info.Commit != "abcd1234" {
		t.Errorf("GetInfo().Commit = %q; want %q", info.Commit, "abcd1234")
	}

	// Restore original values
	Version = originalVersion
	Commit = originalCommit
}

func TestValidate(t *testing.T) {
	// Save original version
	originalVersion := Version
	defer func() { Version = originalVersion }()

	tests := []struct {
		version     string
		forceUpdate bool
		shouldError bool
	}{
		{"undefined", false, false},
		{"dev", false, false},
		{"", false, false},
		{"1.0.0", false, false},
		{"v1.0.0", false, false},
		{"1.0.0", true, false},
		{"invalid", false, true},
		{"invalid", true, false},
	}

	for _, tt := range tests {
		Version = tt.version
		err := Validate(tt.forceUpdate)
		if tt.shouldError && err == nil {
			t.Errorf("Validate(%q, %t) expected error but got none", tt.version, tt.forceUpdate)
		}
		if !tt.shouldError && err != nil {
			t.Errorf("Validate(%q, %t) unexpected error: %v", tt.version, tt.forceUpdate, err)
		}
	}
}

func TestGetRange(t *testing.T) {
	min, max := GetRange()
	if min == "" || max == "" {
		t.Errorf("GetRange() returned empty values: min=%q, max=%q", min, max)
	}
}

func TestSetRange(t *testing.T) {
	// Save original range
	originalMin, originalMax := GetRange()
	defer SetRange(originalMin, originalMax)

	// Test setting new range
	SetRange("v1.0.0", "v10.0.0")
	min, max := GetRange()

	if min != "v1.0.0" {
		t.Errorf("After SetRange, min = %q; want %q", min, "v1.0.0")
	}
	if max != "v10.0.0" {
		t.Errorf("After SetRange, max = %q; want %q", max, "v10.0.0")
	}
}
