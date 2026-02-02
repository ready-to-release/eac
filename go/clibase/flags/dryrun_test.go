package flags

import (
	"testing"

	"github.com/ready-to-release/eac/go/clibase/environment"
)

func TestDryRunFlagSet_Parse(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantDryRun    bool
		wantRemaining []string
	}{
		{
			name:          "no flags",
			args:          []string{"module1", "module2"},
			wantDryRun:    false,
			wantRemaining: []string{"module1", "module2"},
		},
		{
			name:          "dry-run flag",
			args:          []string{"--dry-run", "module1"},
			wantDryRun:    true,
			wantRemaining: []string{"module1"},
		},
		{
			name:          "dry-run only",
			args:          []string{"--dry-run"},
			wantDryRun:    true,
			wantRemaining: nil,
		},
		{
			name:          "other flags pass through",
			args:          []string{"--dry-run", "--turbo", "--debug"},
			wantDryRun:    true,
			wantRemaining: []string{"--turbo", "--debug"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDryRunFlagSet()
			env := &environment.Env{IsLocalConsole: true}

			remaining, err := s.Parse(tt.args, env)
			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			flags := s.Values()
			if flags.DryRun != tt.wantDryRun {
				t.Errorf("DryRun = %v, want %v", flags.DryRun, tt.wantDryRun)
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

func TestDryRunFlagSet_Metadata(t *testing.T) {
	s := NewDryRunFlagSet()

	if s.Name() != "dryrun" {
		t.Errorf("Name() = %v, want dryrun", s.Name())
	}

	if s.Description() == "" {
		t.Error("Description() should not be empty")
	}

	flags := s.Flags()
	if len(flags) != 1 {
		t.Errorf("Flags() returned %d flags, want 1", len(flags))
	}

	if flags[0].Name != "dry-run" {
		t.Errorf("Flags()[0].Name = %v, want dry-run", flags[0].Name)
	}
}
