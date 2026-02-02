package docsmanifest

import (
	"testing"
)

func TestCleanImageRef(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"image.png", "image.png"},
		{"path/to/image.png", "path/to/image.png"},
		{"image.png?width=100", "image.png"},
		{"image.png#section", "image.png"},
		{"image.png?width=100#section", "image.png"},
		{"https://example.com/image.png", ""},
		{"http://example.com/image.png", ""},
		{"  image.png  ", "image.png"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := cleanImageRef(tt.input)
			if result != tt.expected {
				t.Errorf("cleanImageRef(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	if !contains(slice, "a") {
		t.Error("Expected contains(slice, 'a') to be true")
	}
	if !contains(slice, "b") {
		t.Error("Expected contains(slice, 'b') to be true")
	}
	if contains(slice, "d") {
		t.Error("Expected contains(slice, 'd') to be false")
	}
	if contains([]string{}, "a") {
		t.Error("Expected contains(empty, 'a') to be false")
	}
}

func TestImageRefPattern(t *testing.T) {
	tests := []struct {
		line     string
		expected []string
	}{
		{
			"![alt text](../assets/image.png)",
			[]string{"../assets/image.png"},
		},
		{
			"![](image.png)",
			[]string{"image.png"},
		},
		{
			`<img src="path/to/image.svg">`,
			[]string{"path/to/image.svg"},
		},
		{
			`<img src='path/to/image.svg'>`,
			[]string{"path/to/image.svg"},
		},
		{
			"Some text ![img1](a.png) and ![img2](b.png) more text",
			[]string{"a.png", "b.png"},
		},
		{
			"No images here",
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			matches := imageRefPattern.FindAllStringSubmatch(tt.line, -1)
			var results []string
			for _, match := range matches {
				ref := match[1]
				if ref == "" {
					ref = match[2]
				}
				results = append(results, ref)
			}

			if len(results) != len(tt.expected) {
				t.Errorf("Expected %d matches, got %d: %v", len(tt.expected), len(results), results)
				return
			}

			for i, exp := range tt.expected {
				if results[i] != exp {
					t.Errorf("Match %d: expected %q, got %q", i, exp, results[i])
				}
			}
		})
	}
}
