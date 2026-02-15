package get

import (
	"testing"
)

func TestFormatPlatformName(t *testing.T) {
	tests := []struct {
		variant string
		want    string
	}{
		{"linux-amd64", "Linux (AMD64)"},
		{"linux-arm64", "Linux (ARM64)"},
		{"darwin-amd64", "macOS (Intel)"},
		{"darwin-arm64", "macOS (Apple Silicon)"},
		{"windows-amd64", "Windows (AMD64)"},
		{"unknown", "unknown"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.variant, func(t *testing.T) {
			if got := formatPlatformName(tt.variant); got != tt.want {
				t.Errorf("formatPlatformName(%q) = %q, want %q", tt.variant, got, tt.want)
			}
		})
	}
}

func TestGetBinarySuffix(t *testing.T) {
	tests := []struct {
		variant string
		want    string
	}{
		{"linux-amd64", "-linux-amd64"},
		{"linux-amd64-upx", "-linux-amd64-upx"},
		{"linux-arm64", "-linux-arm64"},
		{"darwin-amd64", "-darwin-amd64"},
		{"darwin-arm64", "-darwin-arm64"},
		{"windows-amd64", "-windows-amd64.exe"},
		{"windows-amd64-upx", "-windows-amd64-upx.exe"},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.variant, func(t *testing.T) {
			if got := getBinarySuffix(tt.variant); got != tt.want {
				t.Errorf("getBinarySuffix(%q) = %q, want %q", tt.variant, got, tt.want)
			}
		})
	}
}

func TestBinarySize_Struct(t *testing.T) {
	size := BinarySize{
		Name:      "linux-amd64",
		Path:      "/path/to/binary",
		SizeBytes: 10485760, // 10 MB
		SizeMB:    10.0,
		Exists:    true,
	}

	if size.Name != "linux-amd64" {
		t.Errorf("Name = %q, want %q", size.Name, "linux-amd64")
	}
	if size.Path != "/path/to/binary" {
		t.Errorf("Path = %q, want %q", size.Path, "/path/to/binary")
	}
	if size.SizeBytes != 10485760 {
		t.Errorf("SizeBytes = %d, want %d", size.SizeBytes, 10485760)
	}
	if size.SizeMB != 10.0 {
		t.Errorf("SizeMB = %f, want %f", size.SizeMB, 10.0)
	}
	if !size.Exists {
		t.Error("Exists = false, want true")
	}
}

func TestBinarySize_NonExistent(t *testing.T) {
	size := BinarySize{
		Name:      "darwin-arm64",
		Path:      "/nonexistent/path",
		SizeBytes: 0,
		SizeMB:    0,
		Exists:    false,
	}

	if size.Exists {
		t.Error("Exists = true, want false")
	}
	if size.SizeBytes != 0 {
		t.Errorf("SizeBytes = %d, want 0", size.SizeBytes)
	}
}

func TestBinarySizesOutput_Struct(t *testing.T) {
	output := BinarySizesOutput{
		Module:   "clie",
		Binaries: make(map[string]BinarySize),
	}
	output.Binaries["linux-amd64"] = BinarySize{
		Name:      "linux-amd64",
		SizeBytes: 16252928, // ~15.5 MB
		SizeMB:    15.5,
		Exists:    true,
	}
	output.Binaries["darwin-arm64"] = BinarySize{
		Name:      "darwin-arm64",
		SizeBytes: 13107200, // ~12.5 MB
		SizeMB:    12.5,
		Exists:    true,
	}

	if output.Module != "clie" {
		t.Errorf("Module = %q, want %q", output.Module, "clie")
	}
	if len(output.Binaries) != 2 {
		t.Errorf("len(Binaries) = %d, want 2", len(output.Binaries))
	}

	linux := output.Binaries["linux-amd64"]
	if linux.SizeMB != 15.5 {
		t.Errorf("linux-amd64 SizeMB = %f, want 15.5", linux.SizeMB)
	}
}

func TestStandardBinaryVariants(t *testing.T) {
	// Verify standard variants are defined correctly
	if len(standardBinaryVariants) < 7 {
		t.Errorf("expected at least 7 standard variants, got %d", len(standardBinaryVariants))
	}

	// Check specific variants exist
	variantNames := make(map[string]bool)
	for _, v := range standardBinaryVariants {
		variantNames[v.Name] = true
	}

	required := []string{
		"linux-amd64",
		"linux-amd64-upx",
		"linux-arm64",
		"darwin-amd64",
		"darwin-arm64",
		"windows-amd64",
		"windows-amd64-upx",
	}
	for _, name := range required {
		if !variantNames[name] {
			t.Errorf("missing required variant: %s", name)
		}
	}
}

func TestStandardBinaryVariants_EnvVars(t *testing.T) {
	// Verify env var names are set correctly
	envVars := make(map[string]string)
	for _, v := range standardBinaryVariants {
		envVars[v.Name] = v.EnvVar
	}

	expected := map[string]string{
		"linux-amd64":       "SIZE_LINUX_AMD64",
		"linux-amd64-upx":   "SIZE_LINUX_AMD64_UPX",
		"linux-arm64":       "SIZE_LINUX_ARM64",
		"darwin-amd64":      "SIZE_DARWIN_AMD64",
		"darwin-arm64":      "SIZE_DARWIN_ARM64",
		"windows-amd64":     "SIZE_WINDOWS_AMD64",
		"windows-amd64-upx": "SIZE_WINDOWS_AMD64_UPX",
	}

	for name, wantEnv := range expected {
		if gotEnv, ok := envVars[name]; !ok {
			t.Errorf("missing variant %s", name)
		} else if gotEnv != wantEnv {
			t.Errorf("variant %s EnvVar = %q, want %q", name, gotEnv, wantEnv)
		}
	}
}

func TestStandardBinaryVariants_Suffixes(t *testing.T) {
	// Verify suffixes include .exe for Windows
	suffixes := make(map[string]string)
	for _, v := range standardBinaryVariants {
		suffixes[v.Name] = v.Suffix
	}

	// Windows should have .exe
	if !hasExeSuffix(suffixes["windows-amd64"]) {
		t.Errorf("windows-amd64 suffix %q should end with .exe", suffixes["windows-amd64"])
	}
	if !hasExeSuffix(suffixes["windows-amd64-upx"]) {
		t.Errorf("windows-amd64-upx suffix %q should end with .exe", suffixes["windows-amd64-upx"])
	}

	// Non-Windows should not have .exe
	if hasExeSuffix(suffixes["linux-amd64"]) {
		t.Errorf("linux-amd64 suffix %q should not end with .exe", suffixes["linux-amd64"])
	}
	if hasExeSuffix(suffixes["darwin-arm64"]) {
		t.Errorf("darwin-arm64 suffix %q should not end with .exe", suffixes["darwin-arm64"])
	}
}

func hasExeSuffix(s string) bool {
	return len(s) >= 4 && s[len(s)-4:] == ".exe"
}

func TestSizeCalculation(t *testing.T) {
	// Test that 1 MB = 1048576 bytes
	bytes := int64(1048576)
	mb := float64(bytes) / 1048576.0

	if mb != 1.0 {
		t.Errorf("1048576 bytes = %f MB, want 1.0", mb)
	}

	// Test 10 MB
	bytes = 10485760
	mb = float64(bytes) / 1048576.0
	if mb != 10.0 {
		t.Errorf("10485760 bytes = %f MB, want 10.0", mb)
	}

	// Test fractional MB
	bytes = 5242880 // 5 MB
	mb = float64(bytes) / 1048576.0
	if mb != 5.0 {
		t.Errorf("5242880 bytes = %f MB, want 5.0", mb)
	}
}
