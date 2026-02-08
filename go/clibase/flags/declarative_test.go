package flags

import (
	"testing"

	"github.com/ready-to-release/eac/go/clibase/environment"
)

func TestDeclarativeFlagDef_ResolveDefault(t *testing.T) {
	tests := []struct {
		name        string
		def         DeclarativeFlagDef
		env         *environment.Env
		wantDefault bool
	}{
		{
			name: "non-env-aware flag returns DefaultOn when true",
			def: DeclarativeFlagDef{
				Behavior:  "cache",
				DefaultOn: true,
				EnvAware:  false,
			},
			env:         &environment.Env{IsLocalConsole: true},
			wantDefault: true,
		},
		{
			name: "non-env-aware flag returns DefaultOn when false",
			def: DeclarativeFlagDef{
				Behavior:  "fix",
				DefaultOn: false,
				EnvAware:  false,
			},
			env:         &environment.Env{IsLocalConsole: true},
			wantDefault: false,
		},
		{
			name: "env-aware flag uses local default",
			def: DeclarativeFlagDef{
				Behavior:    "tui",
				DefaultOn:   false,
				EnvAware:    true,
				EnvDefaults: map[string]bool{"local": true, "CI": false, "container": false},
			},
			env:         &environment.Env{IsLocalConsole: true},
			wantDefault: true,
		},
		{
			name: "env-aware flag uses CI default",
			def: DeclarativeFlagDef{
				Behavior:    "tui",
				DefaultOn:   false,
				EnvAware:    true,
				EnvDefaults: map[string]bool{"local": true, "CI": false, "container": false},
			},
			env:         &environment.Env{IsCI: true},
			wantDefault: false,
		},
		{
			name: "env-aware flag uses container default",
			def: DeclarativeFlagDef{
				Behavior:    "tui",
				DefaultOn:   false,
				EnvAware:    true,
				EnvDefaults: map[string]bool{"local": true, "CI": false, "container": false},
			},
			env:         &environment.Env{IsContainer: true},
			wantDefault: false,
		},
		{
			name: "env-aware flag falls back to DefaultOn when context not in EnvDefaults",
			def: DeclarativeFlagDef{
				Behavior:    "tidy",
				DefaultOn:   true,
				EnvAware:    true,
				EnvDefaults: map[string]bool{"local": true},
			},
			env:         &environment.Env{IsCI: true},
			wantDefault: true, // falls back to DefaultOn
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.def.ResolveDefault(tt.env)
			if got != tt.wantDefault {
				t.Errorf("ResolveDefault() = %v, want %v", got, tt.wantDefault)
			}
		})
	}
}

func TestDeclarativeState_IsEnabled(t *testing.T) {
	tests := []struct {
		name          string
		state         DeclarativeState
		def           DeclarativeFlagDef
		env           *environment.Env
		wantEnabled   bool
		wantExplicit  bool
	}{
		{
			name: "explicitly on overrides default off",
			state: DeclarativeState{
				ExplicitlyOn:  true,
				ExplicitlyOff: false,
			},
			def: DeclarativeFlagDef{
				Behavior:  "fix",
				DefaultOn: false,
				EnvAware:  false,
			},
			env:          &environment.Env{IsLocalConsole: true},
			wantEnabled:  true,
			wantExplicit: true,
		},
		{
			name: "explicitly off overrides default on",
			state: DeclarativeState{
				ExplicitlyOn:  false,
				ExplicitlyOff: true,
			},
			def: DeclarativeFlagDef{
				Behavior:  "cache",
				DefaultOn: true,
				EnvAware:  false,
			},
			env:          &environment.Env{IsLocalConsole: true},
			wantEnabled:  false,
			wantExplicit: true,
		},
		{
			name: "no explicit setting uses default",
			state: DeclarativeState{
				ExplicitlyOn:  false,
				ExplicitlyOff: false,
			},
			def: DeclarativeFlagDef{
				Behavior:  "cache",
				DefaultOn: true,
				EnvAware:  false,
			},
			env:          &environment.Env{IsLocalConsole: true},
			wantEnabled:  true,
			wantExplicit: false,
		},
		{
			name: "explicitly off takes priority over explicitly on (safety)",
			state: DeclarativeState{
				ExplicitlyOn:  true,
				ExplicitlyOff: true, // conflicting - off wins for safety
			},
			def: DeclarativeFlagDef{
				Behavior:  "cache",
				DefaultOn: true,
				EnvAware:  false,
			},
			env:          &environment.Env{IsLocalConsole: true},
			wantEnabled:  false,
			wantExplicit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEnabled := tt.state.IsEnabled(&tt.def, tt.env)
			if gotEnabled != tt.wantEnabled {
				t.Errorf("IsEnabled() = %v, want %v", gotEnabled, tt.wantEnabled)
			}

			gotExplicit := tt.state.WasExplicitlySet()
			if gotExplicit != tt.wantExplicit {
				t.Errorf("WasExplicitlySet() = %v, want %v", gotExplicit, tt.wantExplicit)
			}
		})
	}
}

func TestDeclarativeFlagDef_Fields(t *testing.T) {
	def := DeclarativeFlagDef{
		Behavior:    "cache",
		EnableFlag:  "--with-cache",
		DisableFlag: "--no-cache",
		DefaultOn:   true,
		EnvAware:    false,
		EnvDefaults: nil,
		Description: "Incremental caching for faster rebuilds",
	}

	if def.Behavior != "cache" {
		t.Errorf("Behavior = %v, want cache", def.Behavior)
	}
	if def.EnableFlag != "--with-cache" {
		t.Errorf("EnableFlag = %v, want --with-cache", def.EnableFlag)
	}
	if def.DisableFlag != "--no-cache" {
		t.Errorf("DisableFlag = %v, want --no-cache", def.DisableFlag)
	}
	if !def.DefaultOn {
		t.Error("DefaultOn should be true")
	}
	if def.EnvAware {
		t.Error("EnvAware should be false")
	}
	if def.Description != "Incremental caching for faster rebuilds" {
		t.Errorf("Description = %v, want 'Incremental caching for faster rebuilds'", def.Description)
	}
}
