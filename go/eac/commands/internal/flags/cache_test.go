package flags

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
)

func TestCacheFlagSet_Parse(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantSkipCache bool
		wantSkipDeps  bool
		wantRemaining []string
	}{
		{
			name:          "no flags",
			args:          []string{"module1", "module2"},
			wantSkipCache: false,
			wantSkipDeps:  false,
			wantRemaining: []string{"module1", "module2"},
		},
		{
			name:          "skip-cache flag",
			args:          []string{"--skip-cache", "module1"},
			wantSkipCache: true,
			wantSkipDeps:  false,
			wantRemaining: []string{"module1"},
		},
		{
			name:          "skip-deps flag",
			args:          []string{"--skip-deps", "module1"},
			wantSkipCache: false,
			wantSkipDeps:  true,
			wantRemaining: []string{"module1"},
		},
		{
			name:          "both flags",
			args:          []string{"--skip-cache", "--skip-deps"},
			wantSkipCache: true,
			wantSkipDeps:  true,
			wantRemaining: nil,
		},
		{
			name:          "other flags pass through",
			args:          []string{"--skip-cache", "--turbo", "--debug"},
			wantSkipCache: true,
			wantSkipDeps:  false,
			wantRemaining: []string{"--turbo", "--debug"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewCacheFlagSet()
			env := &environment.Env{IsLocalConsole: true}

			remaining, err := s.Parse(tt.args, env)
			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			flags := s.Values()
			if flags.SkipCache != tt.wantSkipCache {
				t.Errorf("SkipCache = %v, want %v", flags.SkipCache, tt.wantSkipCache)
			}
			if flags.SkipDeps != tt.wantSkipDeps {
				t.Errorf("SkipDeps = %v, want %v", flags.SkipDeps, tt.wantSkipDeps)
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

func TestCacheFlagSet_Metadata(t *testing.T) {
	s := NewCacheFlagSet()

	if s.Name() != "cache" {
		t.Errorf("Name() = %v, want cache", s.Name())
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
	if !flagNames["skip-cache"] {
		t.Error("Flags() missing skip-cache flag")
	}
	if !flagNames["skip-deps"] {
		t.Error("Flags() missing skip-deps flag")
	}
}
