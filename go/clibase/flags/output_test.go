package flags

import (
	"testing"

	"github.com/ready-to-release/eac/go/clibase/display"
	"github.com/ready-to-release/eac/go/clibase/environment"
)

func TestOutputFlagSet_Parse(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantUseTUI    bool
		wantNoTUI     bool
		wantHeight    int
		wantASCII     bool
		wantDebug     bool
		wantTimings   bool
		wantRemaining []string
		wantErr       bool
	}{
		{
			name:          "no flags",
			args:          []string{"module1"},
			wantUseTUI:    false,
			wantHeight:    display.DefaultHeight,
			wantRemaining: []string{"module1"},
		},
		{
			name:          "tui flag",
			args:          []string{"--tui"},
			wantUseTUI:    true,
			wantHeight:    display.DefaultHeight,
			wantRemaining: nil,
		},
		{
			name:          "no-tui flag",
			args:          []string{"--no-tui"},
			wantUseTUI:    false,
			wantNoTUI:     true,
			wantHeight:    display.DefaultHeight,
			wantRemaining: nil,
		},
		{
			name:          "debug long form",
			args:          []string{"--debug", "module1"},
			wantDebug:     true,
			wantHeight:    display.DefaultHeight,
			wantRemaining: []string{"module1"},
		},
		{
			name:          "debug short form",
			args:          []string{"-d", "module1"},
			wantDebug:     true,
			wantHeight:    display.DefaultHeight,
			wantRemaining: []string{"module1"},
		},
		{
			name:          "timings flag",
			args:          []string{"--timings"},
			wantTimings:   true,
			wantHeight:    display.DefaultHeight,
			wantRemaining: nil,
		},
		{
			name:          "ascii flag",
			args:          []string{"--ascii"},
			wantASCII:     true,
			wantHeight:    display.DefaultHeight,
			wantRemaining: nil,
		},
		{
			name:          "tui-height with space",
			args:          []string{"--tui-height", "10"},
			wantHeight:    10,
			wantRemaining: nil,
		},
		{
			name:          "tui-height=value syntax",
			args:          []string{"--tui-height=15"},
			wantHeight:    15,
			wantRemaining: nil,
		},
		{
			name:    "tui-height too low",
			args:    []string{"--tui-height", "2"},
			wantErr: true,
		},
		{
			name:    "tui-height too high",
			args:    []string{"--tui-height", "21"},
			wantErr: true,
		},
		{
			name:    "tui-height missing value",
			args:    []string{"--tui-height"},
			wantErr: true,
		},
		{
			name:    "tui-height=invalid",
			args:    []string{"--tui-height=abc"},
			wantErr: true,
		},
		{
			name:          "multiple flags",
			args:          []string{"--tui", "--debug", "--timings", "--ascii", "--tui-height", "8"},
			wantUseTUI:    true,
			wantDebug:     true,
			wantTimings:   true,
			wantASCII:     true,
			wantHeight:    8,
			wantRemaining: nil,
		},
		{
			name:          "other flags pass through",
			args:          []string{"--debug", "--turbo", "--skip-cache"},
			wantDebug:     true,
			wantHeight:    display.DefaultHeight,
			wantRemaining: []string{"--turbo", "--skip-cache"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewOutputFlagSet()
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
			if flags.UseTUI != tt.wantUseTUI {
				t.Errorf("UseTUI = %v, want %v", flags.UseTUI, tt.wantUseTUI)
			}
			if flags.NoTUI != tt.wantNoTUI {
				t.Errorf("NoTUI = %v, want %v", flags.NoTUI, tt.wantNoTUI)
			}
			if flags.TUIHeight != tt.wantHeight {
				t.Errorf("TUIHeight = %v, want %v", flags.TUIHeight, tt.wantHeight)
			}
			if flags.TUIASCIIMode != tt.wantASCII {
				t.Errorf("TUIASCIIMode = %v, want %v", flags.TUIASCIIMode, tt.wantASCII)
			}
			if flags.Debug != tt.wantDebug {
				t.Errorf("Debug = %v, want %v", flags.Debug, tt.wantDebug)
			}
			if flags.ShowTimings != tt.wantTimings {
				t.Errorf("ShowTimings = %v, want %v", flags.ShowTimings, tt.wantTimings)
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

func TestOutputFlagSet_ApplyDefaults(t *testing.T) {
	tests := []struct {
		name          string
		env           *environment.Env
		explicitlySet bool
		wantUseTUI    bool
	}{
		{
			name:          "local console defaults to TUI",
			env:           &environment.Env{IsLocalConsole: true},
			explicitlySet: false,
			wantUseTUI:    true,
		},
		{
			name:          "CI defaults to no TUI",
			env:           &environment.Env{IsCI: true},
			explicitlySet: false,
			wantUseTUI:    false,
		},
		{
			name:          "explicit --tui overrides",
			env:           &environment.Env{IsCI: true},
			explicitlySet: true,
			wantUseTUI:    true, // Will be set by parse, not defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewOutputFlagSet()

			if tt.explicitlySet {
				// Simulate parsing --tui
				s.flags.UseTUI = true
				s.flags.TUIExplicitlySet = true
			}

			s.ApplyDefaults(tt.env)

			if !tt.explicitlySet && s.Values().UseTUI != tt.wantUseTUI {
				t.Errorf("UseTUI = %v, want %v", s.Values().UseTUI, tt.wantUseTUI)
			}
		})
	}
}

func TestOutputFlagSet_ValidateTUI(t *testing.T) {
	tests := []struct {
		name          string
		env           *environment.Env
		explicitlySet bool
		useTUI        bool
		wantErr       bool
	}{
		{
			name:          "local console allows TUI",
			env:           &environment.Env{IsLocalConsole: true},
			explicitlySet: true,
			useTUI:        true,
			wantErr:       false,
		},
		{
			name:          "CI rejects explicit --tui",
			env:           &environment.Env{IsCI: true},
			explicitlySet: true,
			useTUI:        true,
			wantErr:       true,
		},
		{
			name:          "CI allows no TUI",
			env:           &environment.Env{IsCI: true},
			explicitlySet: false,
			useTUI:        false,
			wantErr:       false,
		},
		{
			name:          "container rejects explicit --tui",
			env:           &environment.Env{IsContainer: true},
			explicitlySet: true,
			useTUI:        true,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewOutputFlagSet()
			s.flags.TUIExplicitlySet = tt.explicitlySet
			s.flags.UseTUI = tt.useTUI

			err := s.ValidateTUI(tt.env)

			if tt.wantErr && err == nil {
				t.Error("ValidateTUI() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateTUI() unexpected error: %v", err)
			}
		})
	}
}

func TestOutputFlagSet_Metadata(t *testing.T) {
	s := NewOutputFlagSet()

	if s.Name() != "output" {
		t.Errorf("Name() = %v, want output", s.Name())
	}

	if s.Description() == "" {
		t.Error("Description() should not be empty")
	}

	flags := s.Flags()
	// Now includes with-tui in addition to existing flags
	if len(flags) < 8 {
		t.Errorf("Flags() returned %d flags, want at least 8", len(flags))
	}

	flagNames := make(map[string]bool)
	for _, f := range flags {
		flagNames[f.Name] = true
	}

	expected := []string{"tui", "no-tui", "tui-height", "ascii", "skip-tui-delay", "debug", "timings", "demo"}
	for _, name := range expected {
		if !flagNames[name] {
			t.Errorf("Flags() missing %s flag", name)
		}
	}
}

func TestOutputFlagSet_DeclarativeFlags(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		wantUseTUI         bool
		wantTUIExplicit    bool
		wantRemaining      []string
		wantErr            bool
	}{
		{
			name:            "with-tui explicitly enables TUI",
			args:            []string{"--with-tui", "module1"},
			wantUseTUI:      true,
			wantTUIExplicit: true,
			wantRemaining:   []string{"module1"},
		},
		{
			name:            "tui still works as before (backward compat)",
			args:            []string{"--tui", "module1"},
			wantUseTUI:      true,
			wantTUIExplicit: true,
			wantRemaining:   []string{"module1"},
		},
		{
			name:            "no-tui still works",
			args:            []string{"--no-tui", "module1"},
			wantUseTUI:      false,
			wantTUIExplicit: true,
			wantRemaining:   []string{"module1"},
		},
		{
			name:            "default behavior preserved (no flags)",
			args:            []string{"module1"},
			wantUseTUI:      false, // Not yet applied defaults
			wantTUIExplicit: false,
			wantRemaining:   []string{"module1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewOutputFlagSet()
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
			if flags.UseTUI != tt.wantUseTUI {
				t.Errorf("UseTUI = %v, want %v", flags.UseTUI, tt.wantUseTUI)
			}
			if flags.TUIExplicitlySet != tt.wantTUIExplicit {
				t.Errorf("TUIExplicitlySet = %v, want %v", flags.TUIExplicitlySet, tt.wantTUIExplicit)
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

func TestOutputFlagSet_WithTUI_ValidateTUI(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		env     *environment.Env
		wantErr bool
	}{
		{
			name:    "with-tui errors in CI",
			args:    []string{"--with-tui"},
			env:     &environment.Env{IsCI: true},
			wantErr: true,
		},
		{
			name:    "with-tui errors in container",
			args:    []string{"--with-tui"},
			env:     &environment.Env{IsContainer: true},
			wantErr: true,
		},
		{
			name:    "with-tui OK in local console",
			args:    []string{"--with-tui"},
			env:     &environment.Env{IsLocalConsole: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewOutputFlagSet()
			_, err := s.Parse(tt.args, tt.env)
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}

			err = s.ValidateTUI(tt.env)
			if tt.wantErr && err == nil {
				t.Error("ValidateTUI() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateTUI() unexpected error: %v", err)
			}
		})
	}
}

func TestOutputFlagSet_DeclarativeFlagSetInterface(t *testing.T) {
	s := NewOutputFlagSet()

	// Test that OutputFlagSet implements DeclarativeFlagSet interface
	var _ DeclarativeFlagSet = s

	// Test DeclarativeFlags() returns expected definitions
	defs := s.DeclarativeFlags()
	if len(defs) != 1 {
		t.Errorf("DeclarativeFlags() returned %d definitions, want 1", len(defs))
	}

	// Check tui behavior definition
	var tuiDef *DeclarativeFlagDef
	for i := range defs {
		if defs[i].Behavior == "tui" {
			tuiDef = &defs[i]
		}
	}

	if tuiDef == nil {
		t.Fatal("DeclarativeFlags() missing tui behavior")
	}
	if tuiDef.EnableFlag != "--with-tui" {
		t.Errorf("tui EnableFlag = %v, want --with-tui", tuiDef.EnableFlag)
	}
	if tuiDef.DisableFlag != "--no-tui" {
		t.Errorf("tui DisableFlag = %v, want --no-tui", tuiDef.DisableFlag)
	}
	if !tuiDef.EnvAware {
		t.Error("tui should be EnvAware=true")
	}
}

func TestOutputFlagSet_GetDeclarativeState(t *testing.T) {
	s := NewOutputFlagSet()
	env := &environment.Env{IsLocalConsole: true}

	// Parse with explicit --no-tui
	_, err := s.Parse([]string{"--no-tui"}, env)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Test GetDeclarativeState for tui
	tuiState := s.GetDeclarativeState("tui")
	if tuiState == nil {
		t.Fatal("GetDeclarativeState(tui) returned nil")
	}
	if !tuiState.ExplicitlyOff {
		t.Error("tui should be ExplicitlyOff")
	}
	if tuiState.ExplicitlyOn {
		t.Error("tui should not be ExplicitlyOn")
	}

	// Test unknown behavior returns nil
	unknownState := s.GetDeclarativeState("unknown")
	if unknownState != nil {
		t.Error("GetDeclarativeState(unknown) should return nil")
	}
}
