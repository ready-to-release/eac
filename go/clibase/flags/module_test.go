package flags

import (
	"testing"

	"github.com/ready-to-release/eac/go/clibase/environment"
)

func TestModuleFlagSet_Parse(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantExclude   string
		wantSkipDepm  bool
		wantRemaining []string
	}{
		{
			name:          "no flags",
			args:          []string{"module1", "module2"},
			wantExclude:   "",
			wantSkipDepm:  false,
			wantRemaining: []string{"module1", "module2"},
		},
		{
			name:          "exclude with space",
			args:          []string{"--exclude", "eac-specs", "module1"},
			wantExclude:   "eac-specs",
			wantSkipDepm:  false,
			wantRemaining: []string{"module1"},
		},
		{
			name:          "exclude=value syntax",
			args:          []string{"--exclude=eac-specs", "module1"},
			wantExclude:   "eac-specs",
			wantSkipDepm:  false,
			wantRemaining: []string{"module1"},
		},
		{
			name:          "skip-depm flag",
			args:          []string{"--skip-depm", "module1"},
			wantExclude:   "",
			wantSkipDepm:  true,
			wantRemaining: []string{"module1"},
		},
		{
			name:          "both flags",
			args:          []string{"--exclude", "test-*", "--skip-depm"},
			wantExclude:   "test-*",
			wantSkipDepm:  true,
			wantRemaining: nil,
		},
		{
			name:          "exclude with pattern containing =",
			args:          []string{"--exclude=foo=bar"},
			wantExclude:   "foo=bar",
			wantSkipDepm:  false,
			wantRemaining: nil,
		},
		{
			name:          "other flags pass through",
			args:          []string{"--skip-depm", "--turbo", "--debug"},
			wantExclude:   "",
			wantSkipDepm:  true,
			wantRemaining: []string{"--turbo", "--debug"},
		},
		{
			name:          "exclude at end without value consumes nothing",
			args:          []string{"module1", "--exclude"},
			wantExclude:   "",
			wantSkipDepm:  false,
			wantRemaining: []string{"module1", "--exclude"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewModuleFlagSet()
			env := &environment.Env{IsLocalConsole: true}

			remaining, err := s.Parse(tt.args, env)
			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			flags := s.Values()
			if flags.Exclude != tt.wantExclude {
				t.Errorf("Exclude = %v, want %v", flags.Exclude, tt.wantExclude)
			}
			if flags.SkipDepm != tt.wantSkipDepm {
				t.Errorf("SkipDepm = %v, want %v", flags.SkipDepm, tt.wantSkipDepm)
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

func TestModuleFlagSet_Metadata(t *testing.T) {
	s := NewModuleFlagSet()

	if s.Name() != "module" {
		t.Errorf("Name() = %v, want module", s.Name())
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
	if !flagNames["exclude"] {
		t.Error("Flags() missing exclude flag")
	}
	if !flagNames["skip-depm"] {
		t.Error("Flags() missing skip-depm flag")
	}
}
