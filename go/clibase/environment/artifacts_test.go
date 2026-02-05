package environment

import (
	"runtime"
	"testing"
)

func TestArtifactsMode_IsValid(t *testing.T) {
	tests := []struct {
		mode  ArtifactsMode
		valid bool
	}{
		{ArtifactsModeAll, true},
		{ArtifactsModeReduced, true},
		{ArtifactsMode(""), false},
		{ArtifactsMode("invalid"), false},
		{ArtifactsMode("ALL"), false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			if got := tt.mode.IsValid(); got != tt.valid {
				t.Errorf("ArtifactsMode(%q).IsValid() = %v, want %v", tt.mode, got, tt.valid)
			}
		})
	}
}

func TestArtifactsMode_String(t *testing.T) {
	tests := []struct {
		mode     ArtifactsMode
		expected string
	}{
		{ArtifactsModeAll, "all"},
		{ArtifactsModeReduced, "reduced"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.expected {
				t.Errorf("ArtifactsMode.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEnv_DefaultArtifactsMode(t *testing.T) {
	tests := []struct {
		name     string
		env      Env
		expected ArtifactsMode
	}{
		{
			name:     "CI defaults to all",
			env:      Env{IsCI: true},
			expected: ArtifactsModeAll,
		},
		{
			name:     "LocalConsole defaults to reduced",
			env:      Env{IsLocalConsole: true},
			expected: ArtifactsModeReduced,
		},
		{
			name:     "Container defaults to reduced",
			env:      Env{IsContainer: true},
			expected: ArtifactsModeReduced,
		},
		{
			name:     "TestContext defaults to reduced",
			env:      Env{IsTestContext: true},
			expected: ArtifactsModeReduced,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.env.DefaultArtifactsMode(); got != tt.expected {
				t.Errorf("Env.DefaultArtifactsMode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestArtifactsMode_PDFPageLimit(t *testing.T) {
	tests := []struct {
		name     string
		mode     ArtifactsMode
		expected int
	}{
		{
			name:     "all mode has no limit",
			mode:     ArtifactsModeAll,
			expected: 0,
		},
		{
			name:     "reduced mode has limit",
			mode:     ArtifactsModeReduced,
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.PDFPageLimit(); got != tt.expected {
				t.Errorf("ArtifactsMode.PDFPageLimit() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestArtifactsMode_AllArtifactsRequested(t *testing.T) {
	tests := []struct {
		mode     ArtifactsMode
		expected bool
	}{
		{ArtifactsModeAll, true},
		{ArtifactsModeReduced, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			if got := tt.mode.AllArtifactsRequested(); got != tt.expected {
				t.Errorf("ArtifactsMode(%q).AllArtifactsRequested() = %v, want %v", tt.mode, got, tt.expected)
			}
		})
	}
}

func TestArtifactsMode_ShouldIncludeDerivedArtifacts(t *testing.T) {
	tests := []struct {
		mode     ArtifactsMode
		expected bool
	}{
		{ArtifactsModeAll, true},
		{ArtifactsModeReduced, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			if got := tt.mode.ShouldIncludeDerivedArtifacts(); got != tt.expected {
				t.Errorf("ArtifactsMode(%q).ShouldIncludeDerivedArtifacts() = %v, want %v", tt.mode, got, tt.expected)
			}
		})
	}
}

func TestArtifactsMode_ShouldBuildTarget(t *testing.T) {
	currentOS := runtime.GOOS
	currentArch := runtime.GOARCH

	tests := []struct {
		name        string
		mode        ArtifactsMode
		targetOS    string
		targetArch  string
		compression string
		expected    bool
	}{
		// All mode tests
		{
			name:        "all mode builds everything",
			mode:        ArtifactsModeAll,
			targetOS:    "linux",
			targetArch:  "amd64",
			compression: "",
			expected:    true,
		},
		{
			name:        "all mode builds UPX",
			mode:        ArtifactsModeAll,
			targetOS:    "linux",
			targetArch:  "amd64",
			compression: "upx",
			expected:    true,
		},
		// Reduced mode tests - current platform
		{
			name:        "reduced mode builds current platform",
			mode:        ArtifactsModeReduced,
			targetOS:    currentOS,
			targetArch:  currentArch,
			compression: "",
			expected:    true,
		},
		{
			name:        "reduced mode skips UPX",
			mode:        ArtifactsModeReduced,
			targetOS:    currentOS,
			targetArch:  currentArch,
			compression: "upx",
			expected:    false,
		},
		{
			name:        "reduced mode skips different platform",
			mode:        ArtifactsModeReduced,
			targetOS:    "freebsd",
			targetArch:  "386",
			compression: "",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mode.ShouldBuildTarget(tt.targetOS, tt.targetArch, tt.compression)
			if got != tt.expected {
				t.Errorf("ShouldBuildTarget(%s, %s, %q) = %v, want %v",
					tt.targetOS, tt.targetArch, tt.compression, got, tt.expected)
			}
		})
	}
}

func TestArtifactsMode_ShouldBuildTarget_WindowsWSL(t *testing.T) {
	// Test Windows WSL special case
	mode := ArtifactsModeReduced

	// On Windows reduced, linux-amd64 should be built for WSL
	got := mode.ShouldBuildTargetWithCurrentPlatform("linux", "amd64", "", "windows", "amd64")
	if !got {
		t.Error("reduced mode on Windows should build linux-amd64 for WSL")
	}

	// But not from other platforms
	got = mode.ShouldBuildTargetWithCurrentPlatform("linux", "amd64", "", "darwin", "arm64")
	if got {
		t.Error("reduced mode on macOS should not build linux-amd64")
	}
}

func TestParseArtifactsMode(t *testing.T) {
	tests := []struct {
		input   string
		want    ArtifactsMode
		wantErr bool
	}{
		{"all", ArtifactsModeAll, false},
		{"reduced", ArtifactsModeReduced, false},
		{"", ArtifactsMode(""), true},
		{"invalid", ArtifactsMode(""), true},
		{"ALL", ArtifactsMode(""), true}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseArtifactsMode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseArtifactsMode(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseArtifactsMode(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
