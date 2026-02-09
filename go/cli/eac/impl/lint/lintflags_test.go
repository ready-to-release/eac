package lint

import (
	"testing"
)

func TestParseLintSpecificFlags(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantFix       bool
		wantConfig    string
		wantRemaining []string
		wantErr       bool
	}{
		{
			name:          "empty args",
			args:          nil,
			wantRemaining: nil,
		},
		{
			name:    "fix flag",
			args:    []string{"--fix"},
			wantFix: true,
		},
		{
			name:       "config with separate value",
			args:       []string{"--config", "/path/to/config.yml"},
			wantConfig: "/path/to/config.yml",
		},
		{
			name:       "config with equals syntax",
			args:       []string{"--config=/path/to/config.yml"},
			wantConfig: "/path/to/config.yml",
		},
		{
			name:       "fix and config combined",
			args:       []string{"--fix", "--config", "lint.yml"},
			wantFix:    true,
			wantConfig: "lint.yml",
		},
		{
			name:          "unknown flags passed through",
			args:          []string{"--verbose", "--fix", "--output", "json"},
			wantFix:       true,
			wantRemaining: []string{"--verbose", "--output", "json"},
		},
		{
			name:    "config missing value",
			args:    []string{"--config"},
			wantErr: true,
		},
		{
			name:          "positional args passed through",
			args:          []string{"module-a", "--fix"},
			wantFix:       true,
			wantRemaining: []string{"module-a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags, remaining, err := ParseLintSpecificFlags(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if flags.Fix != tt.wantFix {
				t.Errorf("Fix = %v, want %v", flags.Fix, tt.wantFix)
			}
			if flags.ConfigPath != tt.wantConfig {
				t.Errorf("ConfigPath = %q, want %q", flags.ConfigPath, tt.wantConfig)
			}
			if len(remaining) != len(tt.wantRemaining) {
				t.Fatalf("remaining = %v, want %v", remaining, tt.wantRemaining)
			}
			for i := range remaining {
				if remaining[i] != tt.wantRemaining[i] {
					t.Errorf("remaining[%d] = %q, want %q", i, remaining[i], tt.wantRemaining[i])
				}
			}
		})
	}
}
