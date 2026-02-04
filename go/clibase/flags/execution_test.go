package flags

import (
	"testing"

	"github.com/ready-to-release/eac/go/clibase/environment"
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
	// Includes parallel, sequential, turbo, roof flags
	if len(flags) < 3 {
		t.Errorf("Flags() returned %d flags, want at least 3", len(flags))
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

func TestExecutionFlagSet_DeclarativeFlags(t *testing.T) {
	tests := []struct {
		name                 string
		args                 []string
		wantRoof             int
		wantParallelExplicit bool
		wantRemaining        []string
		wantErr              bool
	}{
		{
			name:                 "parallel explicitly enables parallel mode",
			args:                 []string{"--parallel", "module1"},
			wantRoof:             0, // unchanged, parallel is default
			wantParallelExplicit: true,
			wantRemaining:        []string{"module1"},
		},
		{
			name:                 "sequential sets roof=1",
			args:                 []string{"--sequential", "module1"},
			wantRoof:             1, // sequential = roof 1
			wantParallelExplicit: true, // sequential means explicit parallel choice
			wantRemaining:        []string{"module1"},
		},
		{
			name:                 "default behavior preserved (no flags)",
			args:                 []string{"module1"},
			wantRoof:             0,
			wantParallelExplicit: false,
			wantRemaining:        []string{"module1"},
		},
		{
			name:                 "turbo still works with declarative flags",
			args:                 []string{"--turbo", "--parallel", "module1"},
			wantRoof:             0,
			wantParallelExplicit: true,
			wantRemaining:        []string{"module1"},
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
			if flags.Roof != tt.wantRoof {
				t.Errorf("Roof = %v, want %v", flags.Roof, tt.wantRoof)
			}
			if flags.ParallelExplicit != tt.wantParallelExplicit {
				t.Errorf("ParallelExplicit = %v, want %v", flags.ParallelExplicit, tt.wantParallelExplicit)
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

func TestExecutionFlagSet_DeclarativeFlagSetInterface(t *testing.T) {
	s := NewExecutionFlagSet()

	// Test that ExecutionFlagSet implements DeclarativeFlagSet interface
	var _ DeclarativeFlagSet = s

	// Test DeclarativeFlags() returns expected definitions
	defs := s.DeclarativeFlags()
	if len(defs) != 1 {
		t.Errorf("DeclarativeFlags() returned %d definitions, want 1", len(defs))
	}

	// Check behavior definitions
	var parallelDef *DeclarativeFlagDef
	for i := range defs {
		switch defs[i].Behavior {
		case "parallel":
			parallelDef = &defs[i]
		}
	}

	if parallelDef == nil {
		t.Fatal("DeclarativeFlags() missing parallel behavior")
	}
	if parallelDef.EnableFlag != "--parallel" {
		t.Errorf("parallel EnableFlag = %v, want --parallel", parallelDef.EnableFlag)
	}
	if parallelDef.DisableFlag != "--sequential" {
		t.Errorf("parallel DisableFlag = %v, want --sequential", parallelDef.DisableFlag)
	}
	if !parallelDef.DefaultOn {
		t.Error("parallel should be DefaultOn=true")
	}
}

func TestExecutionFlagSet_GetDeclarativeState(t *testing.T) {
	s := NewExecutionFlagSet()
	env := &environment.Env{IsLocalConsole: true}

	// Parse with explicit sequential setting
	_, err := s.Parse([]string{"--sequential"}, env)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Test GetDeclarativeState for parallel
	parallelState := s.GetDeclarativeState("parallel")
	if parallelState == nil {
		t.Fatal("GetDeclarativeState(parallel) returned nil")
	}
	if !parallelState.ExplicitlyOff {
		t.Error("parallel should be ExplicitlyOff (sequential was set)")
	}

	// Test unknown behavior returns nil
	unknownState := s.GetDeclarativeState("unknown")
	if unknownState != nil {
		t.Error("GetDeclarativeState(unknown) should return nil")
	}
}
