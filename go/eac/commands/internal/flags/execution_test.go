package flags

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
)

func TestExecutionFlagSet_Parse(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantTurbo     bool
		wantRoof      int
		wantRemaining []string
		wantErr       bool
	}{
		{
			name:          "no flags",
			args:          []string{"module1", "module2"},
			wantTurbo:     false,
			wantRoof:      0,
			wantRemaining: []string{"module1", "module2"},
		},
		{
			name:          "turbo flag",
			args:          []string{"--turbo", "module1"},
			wantTurbo:     true,
			wantRoof:      0,
			wantRemaining: []string{"module1"},
		},
		{
			name:          "roof with space",
			args:          []string{"--roof", "4", "module1"},
			wantTurbo:     false,
			wantRoof:      4,
			wantRemaining: []string{"module1"},
		},
		{
			name:          "roof=value syntax",
			args:          []string{"--roof=8", "module1"},
			wantTurbo:     false,
			wantRoof:      8,
			wantRemaining: []string{"module1"},
		},
		{
			name:          "turbo and roof together",
			args:          []string{"--turbo", "--roof", "2"},
			wantTurbo:     true,
			wantRoof:      2,
			wantRemaining: nil,
		},
		{
			name:          "roof=1 for sequential",
			args:          []string{"--roof=1"},
			wantTurbo:     false,
			wantRoof:      1,
			wantRemaining: nil,
		},
		{
			name:    "roof missing value",
			args:    []string{"--roof"},
			wantErr: true,
		},
		{
			name:    "roof negative value",
			args:    []string{"--roof", "-1"},
			wantErr: true,
		},
		{
			name:    "roof non-integer",
			args:    []string{"--roof", "abc"},
			wantErr: true,
		},
		{
			name:    "roof=negative",
			args:    []string{"--roof=-5"},
			wantErr: true,
		},
		{
			name:          "other flags pass through",
			args:          []string{"--turbo", "--debug", "--skip-cache", "module1"},
			wantTurbo:     true,
			wantRoof:      0,
			wantRemaining: []string{"--debug", "--skip-cache", "module1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewExecutionFlagSet()
			env := &environment.Env{IsLocalConsole: true}

			remaining, err := s.Parse(tt.args, env)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			flags := s.Values()
			if flags.Turbo != tt.wantTurbo {
				t.Errorf("Turbo = %v, want %v", flags.Turbo, tt.wantTurbo)
			}
			if flags.Roof != tt.wantRoof {
				t.Errorf("Roof = %v, want %v", flags.Roof, tt.wantRoof)
			}

			if len(remaining) != len(tt.wantRemaining) {
				t.Errorf("remaining = %v, want %v", remaining, tt.wantRemaining)
				return
			}
			for i, r := range remaining {
				if r != tt.wantRemaining[i] {
					t.Errorf("remaining[%d] = %v, want %v", i, r, tt.wantRemaining[i])
				}
			}
		})
	}
}

func TestExecutionFlagSet_Metadata(t *testing.T) {
	s := NewExecutionFlagSet()

	if s.Name() != "execution" {
		t.Errorf("Name() = %v, want execution", s.Name())
	}

	if s.Description() == "" {
		t.Error("Description() should not be empty")
	}

	flags := s.Flags()
	if len(flags) != 2 {
		t.Errorf("Flags() returned %d flags, want 2", len(flags))
	}

	flagNames := make(map[string]bool)
	for _, f := range flags {
		flagNames[f.Name] = true
	}
	if !flagNames["turbo"] {
		t.Error("Flags() missing turbo flag")
	}
	if !flagNames["roof"] {
		t.Error("Flags() missing roof flag")
	}
}
