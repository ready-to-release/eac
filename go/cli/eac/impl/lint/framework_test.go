package lint

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
)

func TestContainsString(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		s     string
		want  bool
	}{
		{
			name:  "found",
			slice: []string{"a", "b", "c"},
			s:     "b",
			want:  true,
		},
		{
			name:  "not found",
			slice: []string{"a", "b", "c"},
			s:     "d",
			want:  false,
		},
		{
			name:  "empty slice",
			slice: nil,
			s:     "a",
			want:  false,
		},
		{
			name:  "empty string match",
			slice: []string{"", "a"},
			s:     "",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsString(tt.slice, tt.s)
			if got != tt.want {
				t.Errorf("containsString(%v, %q) = %v, want %v", tt.slice, tt.s, got, tt.want)
			}
		})
	}
}

func TestIsProviderEnabledForModule(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		linting  *config.ModuleLinting
		want     bool
	}{
		{
			name:     "nil linting returns true",
			provider: "golangci-lint",
			linting:  nil,
			want:     true,
		},
		{
			name:     "empty linting returns true",
			provider: "golangci-lint",
			linting:  &config.ModuleLinting{},
			want:     true,
		},
		{
			name:     "provider in disabled list",
			provider: "golangci-lint",
			linting:  &config.ModuleLinting{Disabled: []string{"golangci-lint"}},
			want:     false,
		},
		{
			name:     "provider not in disabled list",
			provider: "golangci-lint",
			linting:  &config.ModuleLinting{Disabled: []string{"markdownlint"}},
			want:     true,
		},
		{
			name:     "provider in enabled list",
			provider: "golangci-lint",
			linting:  &config.ModuleLinting{Enabled: []string{"golangci-lint", "markdownlint"}},
			want:     true,
		},
		{
			name:     "provider not in enabled list",
			provider: "golangci-lint",
			linting:  &config.ModuleLinting{Enabled: []string{"markdownlint"}},
			want:     false,
		},
		{
			name:     "disabled takes precedence over enabled",
			provider: "golangci-lint",
			linting: &config.ModuleLinting{
				Enabled:  []string{"golangci-lint"},
				Disabled: []string{"golangci-lint"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isProviderEnabledForModule(tt.provider, tt.linting)
			if got != tt.want {
				t.Errorf("isProviderEnabledForModule(%q, %+v) = %v, want %v",
					tt.provider, tt.linting, got, tt.want)
			}
		})
	}
}
