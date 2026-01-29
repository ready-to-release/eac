package flags

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
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
			wantHeight:    tui.DefaultHeight,
			wantRemaining: []string{"module1"},
		},
		{
			name:          "tui flag",
			args:          []string{"--tui"},
			wantUseTUI:    true,
			wantHeight:    tui.DefaultHeight,
			wantRemaining: nil,
		},
		{
			name:          "no-tui flag",
			args:          []string{"--no-tui"},
			wantUseTUI:    false,
			wantNoTUI:     true,
			wantHeight:    tui.DefaultHeight,
			wantRemaining: nil,
		},
		{
			name:          "debug long form",
			args:          []string{"--debug", "module1"},
			wantDebug:     true,
			wantHeight:    tui.DefaultHeight,
			wantRemaining: []string{"module1"},
		},
		{
			name:          "debug short form",
			args:          []string{"-d", "module1"},
			wantDebug:     true,
			wantHeight:    tui.DefaultHeight,
			wantRemaining: []string{"module1"},
		},
		{
			name:          "timings flag",
			args:          []string{"--timings"},
			wantTimings:   true,
			wantHeight:    tui.DefaultHeight,
			wantRemaining: nil,
		},
		{
			name:          "ascii flag",
			args:          []string{"--ascii"},
			wantASCII:     true,
			wantHeight:    tui.DefaultHeight,
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
			wantHeight:    tui.DefaultHeight,
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
	if len(flags) != 7 {
		t.Errorf("Flags() returned %d flags, want 7", len(flags))
	}

	flagNames := make(map[string]bool)
	for _, f := range flags {
		flagNames[f.Name] = true
	}

	expected := []string{"tui", "no-tui", "tui-height", "ascii", "skip-tui-delay", "debug", "timings"}
	for _, name := range expected {
		if !flagNames[name] {
			t.Errorf("Flags() missing %s flag", name)
		}
	}
}
