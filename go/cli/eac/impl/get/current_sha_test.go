package get

import (
	"os"
	"testing"
)

func TestShortSHA(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abc123def456789", "abc123d"},
		{"abc123d", "abc123d"},
		{"abc", "abc"},
		{"", ""},
		{"1234567", "1234567"},
		{"12345678", "1234567"},
		{"abcdefghijklmnopqrstuvwxyz1234567890", "abcdefg"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := shortSHA(tt.input)
			if got != tt.want {
				t.Errorf("shortSHA(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSHASource_Constants(t *testing.T) {
	// Verify constant values are correct
	if SHASourceExplicit != "explicit" {
		t.Errorf("SHASourceExplicit = %q, want %q", SHASourceExplicit, "explicit")
	}
	if SHASourceCI != "ci" {
		t.Errorf("SHASourceCI = %q, want %q", SHASourceCI, "ci")
	}
	if SHASourceDevbox != "devbox" {
		t.Errorf("SHASourceDevbox = %q, want %q", SHASourceDevbox, "devbox")
	}
}

func TestSHAResult_Struct(t *testing.T) {
	result := SHAResult{
		SHA:    "abc123def456789",
		Source: SHASourceCI,
	}

	if result.SHA != "abc123def456789" {
		t.Errorf("SHA = %q, want %q", result.SHA, "abc123def456789")
	}
	if result.Source != SHASourceCI {
		t.Errorf("Source = %q, want %q", result.Source, SHASourceCI)
	}
}

func TestDetectCurrentSHA_ExplicitSHA(t *testing.T) {
	// Explicit SHA should always win regardless of env
	result, err := DetectCurrentSHA("", "explicit123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.SHA != "explicit123" {
		t.Errorf("SHA = %q, want %q", result.SHA, "explicit123")
	}
	if result.Source != SHASourceExplicit {
		t.Errorf("Source = %q, want %q", result.Source, SHASourceExplicit)
	}
}

func TestDetectCurrentSHA_GitHubSHA(t *testing.T) {
	// Save and restore env
	originalSHA := os.Getenv("GITHUB_SHA")
	defer os.Setenv("GITHUB_SHA", originalSHA)

	// Set GitHub SHA
	os.Setenv("GITHUB_SHA", "github123456789")

	result, err := DetectCurrentSHA("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.SHA != "github123456789" {
		t.Errorf("SHA = %q, want %q", result.SHA, "github123456789")
	}
	if result.Source != SHASourceCI {
		t.Errorf("Source = %q, want %q", result.Source, SHASourceCI)
	}
}

func TestDetectCurrentSHA_ExplicitOverridesEnv(t *testing.T) {
	// Save and restore env
	originalSHA := os.Getenv("GITHUB_SHA")
	defer os.Setenv("GITHUB_SHA", originalSHA)

	// Set GitHub SHA
	os.Setenv("GITHUB_SHA", "github123456789")

	// Explicit should still win
	result, err := DetectCurrentSHA("", "explicit_override")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.SHA != "explicit_override" {
		t.Errorf("SHA = %q, want %q", result.SHA, "explicit_override")
	}
	if result.Source != SHASourceExplicit {
		t.Errorf("Source = %q, want %q", result.Source, SHASourceExplicit)
	}
}

func TestSHASource_StringValues(t *testing.T) {
	// Test that SHASource can be used as string
	sources := []SHASource{SHASourceExplicit, SHASourceCI, SHASourceDevbox}
	expected := []string{"explicit", "ci", "devbox"}

	for i, src := range sources {
		if string(src) != expected[i] {
			t.Errorf("SHASource %d = %q, want %q", i, string(src), expected[i])
		}
	}
}
