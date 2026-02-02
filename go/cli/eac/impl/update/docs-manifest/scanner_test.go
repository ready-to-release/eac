package docsmanifest

import (
	"testing"
)

func TestDetermineAssetType(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		// DrawIO exported PNG (primary type)
		{"diagram.drawio.png", "drawio"},
		{"DIAGRAM.DRAWIO.PNG", "drawio"},
		{"complex.name.drawio.png", "drawio"},

		// DrawIO source files
		{"diagram.drawio", "drawio-source"},
		{"DIAGRAM.DRAWIO", "drawio-source"},

		// Regular PNG images
		{"photo.png", "image"},
		{"PHOTO.PNG", "image"},
		{"icon.png", "image"},

		// Not tracked (returns empty string)
		{"icon.svg", ""},
		{"photo.jpg", ""},
		{"photo.jpeg", ""},
		{"animation.gif", ""},
		{"document.md", ""},
		{"script.js", ""},
		{"manifest.json", ""},
		{"readme.txt", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineAssetType(tt.name)
			if result != tt.expected {
				t.Errorf("determineAssetType(%q) = %q, want %q", tt.name, result, tt.expected)
			}
		})
	}
}

func TestDetermineAssetType_DrawioPrecedence(t *testing.T) {
	// Verify that .drawio.png takes precedence over .png
	// The order of checks matters: .drawio.png must be checked before .png
	result := determineAssetType("test.drawio.png")
	if result != "drawio" {
		t.Errorf("Expected 'drawio' for .drawio.png, got %q", result)
	}

	// Regular .png should be "image"
	result = determineAssetType("test.png")
	if result != "image" {
		t.Errorf("Expected 'image' for .png, got %q", result)
	}
}
