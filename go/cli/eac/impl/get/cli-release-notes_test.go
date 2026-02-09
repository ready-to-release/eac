package get

import (
	"strings"
	"testing"
)

func TestGenerateCLIReleaseNotes(t *testing.T) {
	sizes := map[string]string{
		"linux-amd64":       "10.5",
		"linux-arm64":       "11.2",
		"darwin-amd64":      "12.0",
		"darwin-arm64":      "12.5",
		"windows-amd64":     "13.0",
		"linux-amd64-upx":   "4.2",
		"windows-amd64-upx": "5.0",
	}

	notes := generateCLIReleaseNotes(
		"clie",
		"clie",
		"1.2.3",
		"clie/1.2.3",
		"abc123def456",
		"owner/repo",
		"12345",
		sizes,
	)

	// Check required sections exist
	requiredSections := []string{
		"# clie v1.2.3",
		"## Installation",
		"### Standard Binaries",
		"### UPX-Compressed Binaries",
		"## Supply Chain Security",
		"## Release Information",
		"- **Version**: 1.2.3",
		"- **Commit**: abc123def456",
		"- **Tag**: clie/1.2.3",
		"gh attestation verify",
		"--repo owner/repo",
	}

	for _, section := range requiredSections {
		if !strings.Contains(notes, section) {
			t.Errorf("release notes missing section: %q", section)
		}
	}
}

func TestGenerateCLIReleaseNotes_BinarySizes(t *testing.T) {
	sizes := map[string]string{
		"linux-amd64":       "10.5",
		"linux-arm64":       "11.2",
		"darwin-amd64":      "12.0",
		"darwin-arm64":      "12.5",
		"windows-amd64":     "13.0",
		"linux-amd64-upx":   "4.2",
		"windows-amd64-upx": "5.0",
	}

	notes := generateCLIReleaseNotes(
		"clie",
		"clie",
		"1.0.0",
		"clie/1.0.0",
		"abc123",
		"owner/repo",
		"999",
		sizes,
	)

	// Check all binary sizes are included
	expectedSizes := []string{
		"10.5 MB",
		"11.2 MB",
		"12.0 MB",
		"12.5 MB",
		"13.0 MB",
		"4.2 MB",
		"5.0 MB",
	}

	for _, size := range expectedSizes {
		if !strings.Contains(notes, size) {
			t.Errorf("release notes missing size: %q", size)
		}
	}
}

func TestGenerateCLIReleaseNotes_BinaryNames(t *testing.T) {
	sizes := map[string]string{
		"linux-amd64":       "10.0",
		"linux-arm64":       "10.0",
		"darwin-amd64":      "10.0",
		"darwin-arm64":      "10.0",
		"windows-amd64":     "10.0",
		"linux-amd64-upx":   "4.0",
		"windows-amd64-upx": "4.0",
	}

	notes := generateCLIReleaseNotes(
		"test-cli",
		"test",
		"2.0.0",
		"test-cli/2.0.0",
		"def456",
		"",
		"",
		sizes,
	)

	// Check binary names with prefix
	expectedBinaries := []string{
		"`test-linux-amd64`",
		"`test-linux-arm64`",
		"`test-darwin-amd64`",
		"`test-darwin-arm64`",
		"`test-windows-amd64.exe`",
		"`test-linux-amd64-upx`",
		"`test-windows-amd64-upx.exe`",
	}

	for _, binary := range expectedBinaries {
		if !strings.Contains(notes, binary) {
			t.Errorf("release notes missing binary: %q", binary)
		}
	}
}

func TestGenerateCLIReleaseNotes_MinimalInputs(t *testing.T) {
	sizes := map[string]string{
		"linux-amd64": "10.0",
	}

	notes := generateCLIReleaseNotes(
		"test-cli",
		"test",
		"0.1.0",
		"test-cli/0.1.0",
		"def456",
		"", // no repo
		"", // no run id
		sizes,
	)

	if !strings.Contains(notes, "# test-cli v0.1.0") {
		t.Error("missing title")
	}
	if !strings.Contains(notes, "Generated with GitHub Actions") {
		t.Error("missing footer")
	}
	// Should use generic attestation command without specific repo
	if !strings.Contains(notes, "gh attestation verify <binary-file> --repo <owner>/<repo>") {
		t.Error("missing generic attestation command")
	}
}

func TestGenerateCLIReleaseNotes_WithRepoAndRunID(t *testing.T) {
	sizes := map[string]string{
		"linux-amd64": "10.0",
	}

	notes := generateCLIReleaseNotes(
		"clie",
		"clie",
		"1.0.0",
		"clie/1.0.0",
		"abc123",
		"ready-to-release/eac",
		"67890",
		sizes,
	)

	// Should have specific repo in attestation
	if !strings.Contains(notes, "--repo ready-to-release/eac") {
		t.Error("missing specific repo in attestation command")
	}

	// Should have GitHub Actions link with run ID
	if !strings.Contains(notes, "https://github.com/ready-to-release/eac/actions/runs/67890") {
		t.Error("missing GitHub Actions run link")
	}
}

func TestGenerateCLIReleaseNotes_MissingSizes(t *testing.T) {
	// Test with ? for unknown sizes
	sizes := map[string]string{
		"linux-amd64":   "10.0",
		"linux-arm64":   "?",
		"darwin-amd64":  "?",
		"darwin-arm64":  "?",
		"windows-amd64": "?",
	}

	notes := generateCLIReleaseNotes(
		"clie",
		"clie",
		"1.0.0",
		"clie/1.0.0",
		"abc123",
		"",
		"",
		sizes,
	)

	// Should handle ? sizes gracefully
	if !strings.Contains(notes, "10.0 MB") {
		t.Error("missing known size")
	}
	if !strings.Contains(notes, "? MB") {
		t.Error("missing unknown size placeholder")
	}
}

func TestGenerateCLIReleaseNotes_Structure(t *testing.T) {
	sizes := map[string]string{
		"linux-amd64":       "10.0",
		"linux-arm64":       "10.0",
		"darwin-amd64":      "10.0",
		"darwin-arm64":      "10.0",
		"windows-amd64":     "10.0",
		"linux-amd64-upx":   "4.0",
		"windows-amd64-upx": "4.0",
	}

	notes := generateCLIReleaseNotes(
		"clie",
		"clie",
		"1.0.0",
		"clie/1.0.0",
		"abc123",
		"owner/repo",
		"12345",
		sizes,
	)

	// Verify section ordering
	sections := []string{
		"# clie v1.0.0",
		"## Installation",
		"### Standard Binaries",
		"### UPX-Compressed Binaries",
		"## Supply Chain Security",
		"## Release Information",
		"---",
	}

	prevIndex := -1
	for _, section := range sections {
		index := strings.Index(notes, section)
		if index == -1 {
			t.Errorf("missing section: %q", section)
			continue
		}
		if index <= prevIndex {
			t.Errorf("section %q appears before expected position", section)
		}
		prevIndex = index
	}
}

func TestGenerateCLIReleaseNotes_PlatformNames(t *testing.T) {
	sizes := map[string]string{
		"linux-amd64":   "10.0",
		"linux-arm64":   "10.0",
		"darwin-amd64":  "10.0",
		"darwin-arm64":  "10.0",
		"windows-amd64": "10.0",
	}

	notes := generateCLIReleaseNotes(
		"clie",
		"clie",
		"1.0.0",
		"clie/1.0.0",
		"abc123",
		"",
		"",
		sizes,
	)

	// Check platform names are human-readable
	expectedPlatforms := []string{
		"Linux (AMD64)",
		"Linux (ARM64)",
		"macOS (Intel)",
		"macOS (Apple Silicon)",
		"Windows (AMD64)",
	}

	for _, platform := range expectedPlatforms {
		if !strings.Contains(notes, platform) {
			t.Errorf("missing platform name: %q", platform)
		}
	}
}
