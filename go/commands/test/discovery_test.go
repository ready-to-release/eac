package test

import (
	"testing"
)

func TestSanitizePathForLog(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "forward slashes unchanged",
			input: "go/core/config",
			want:  "go/core/config",
		},
		{
			name:  "backslashes converted",
			input: "go\\core\\config",
			want:  "go/core/config",
		},
		{
			name:  "colons converted to slashes",
			input: "test:godog:specs",
			want:  "test/godog/specs",
		},
		{
			name:  "mixed separators",
			input: "go\\core:config/test",
			want:  "go/core/config/test",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "no separators",
			input: "simple",
			want:  "simple",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizePathForLog(tt.input)
			if got != tt.want {
				t.Errorf("sanitizePathForLog(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
