package test

import (
	"testing"
)

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		// Go test files
		{"main_test.go", true},
		{"foo/bar_test.go", true},

		// BDD feature files
		{"specs/login.feature", true},
		{"steps.feature", true},

		// TypeScript test files
		{"app.test.ts", true},
		{"component.spec.ts", true},

		// Regular source files
		{"main.go", false},
		{"config.yml", false},
		{"README.md", false},
		{"helper.ts", false},
		{"test.go", false}, // not _test.go
		{"feature.go", false},

		// Edge cases
		{"", false},
		{"_test.go", true}, // valid Go test file name
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isTestFile(tt.path)
			if got != tt.want {
				t.Errorf("isTestFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
