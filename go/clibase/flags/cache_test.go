package flags

import (
	"testing"

	"github.com/ready-to-release/eac/go/clibase/environment"
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
	// Updated: now includes with-cache, no-cache, with-deps
	if len(flags) < 2 {
		t.Errorf("Flags() returned %d flags, want at least 2", len(flags))
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

func TestCacheFlagSet_DeclarativeFlags(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		wantSkipCache     bool
		wantCacheExplicit bool
		wantSkipDeps      bool
		wantDepsExplicit  bool
		wantRemaining     []string
	}{
		{
			name:              "with-cache explicitly enables caching",
			args:              []string{"--with-cache", "module1"},
			wantSkipCache:     false,
			wantCacheExplicit: true,
			wantSkipDeps:      false,
			wantDepsExplicit:  false,
			wantRemaining:     []string{"module1"},
		},
		{
			name:              "no-cache explicitly disables caching",
			args:              []string{"--no-cache", "module1"},
			wantSkipCache:     true,
			wantCacheExplicit: true,
			wantSkipDeps:      false,
			wantDepsExplicit:  false,
			wantRemaining:     []string{"module1"},
		},
		{
			name:              "skip-cache is backward compatible alias for no-cache",
			args:              []string{"--skip-cache", "module1"},
			wantSkipCache:     true,
			wantCacheExplicit: true,
			wantSkipDeps:      false,
			wantDepsExplicit:  false,
			wantRemaining:     []string{"module1"},
		},
		{
			name:              "with-deps explicitly enables deps",
			args:              []string{"--with-deps", "module1"},
			wantSkipCache:     false,
			wantCacheExplicit: false,
			wantSkipDeps:      false,
			wantDepsExplicit:  true,
			wantRemaining:     []string{"module1"},
		},
		{
			name:              "no-deps explicitly disables deps",
			args:              []string{"--no-deps", "module1"},
			wantSkipCache:     false,
			wantCacheExplicit: false,
			wantSkipDeps:      true,
			wantDepsExplicit:  true,
			wantRemaining:     []string{"module1"},
		},
		{
			name:              "skip-deps remains compatible",
			args:              []string{"--skip-deps", "module1"},
			wantSkipCache:     false,
			wantCacheExplicit: false,
			wantSkipDeps:      true,
			wantDepsExplicit:  true,
			wantRemaining:     []string{"module1"},
		},
		{
			name:              "default behavior preserved (no explicit flags)",
			args:              []string{"module1"},
			wantSkipCache:     false,
			wantCacheExplicit: false,
			wantSkipDeps:      false,
			wantDepsExplicit:  false,
			wantRemaining:     []string{"module1"},
		},
		{
			name:              "all declarative cache flags together",
			args:              []string{"--with-cache", "--with-deps", "module1"},
			wantSkipCache:     false,
			wantCacheExplicit: true,
			wantSkipDeps:      false,
			wantDepsExplicit:  true,
			wantRemaining:     []string{"module1"},
		},
		{
			name:              "mixed legacy and declarative flags",
			args:              []string{"--no-cache", "--skip-deps", "module1"},
			wantSkipCache:     true,
			wantCacheExplicit: true,
			wantSkipDeps:      true,
			wantDepsExplicit:  true,
			wantRemaining:     []string{"module1"},
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
			if flags.CacheExplicit != tt.wantCacheExplicit {
				t.Errorf("CacheExplicit = %v, want %v", flags.CacheExplicit, tt.wantCacheExplicit)
			}
			if flags.SkipDeps != tt.wantSkipDeps {
				t.Errorf("SkipDeps = %v, want %v", flags.SkipDeps, tt.wantSkipDeps)
			}
			if flags.DepsExplicit != tt.wantDepsExplicit {
				t.Errorf("DepsExplicit = %v, want %v", flags.DepsExplicit, tt.wantDepsExplicit)
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

func TestCacheFlagSet_DeclarativeFlagSetInterface(t *testing.T) {
	s := NewCacheFlagSet()

	// Test that CacheFlagSet implements DeclarativeFlagSet interface
	var _ DeclarativeFlagSet = s

	// Test DeclarativeFlags() returns expected definitions
	defs := s.DeclarativeFlags()
	if len(defs) != 2 {
		t.Errorf("DeclarativeFlags() returned %d definitions, want 2", len(defs))
	}

	// Check cache behavior definition
	var cacheDef *DeclarativeFlagDef
	var depsDef *DeclarativeFlagDef
	for i := range defs {
		switch defs[i].Behavior {
		case "cache":
			cacheDef = &defs[i]
		case "deps":
			depsDef = &defs[i]
		}
	}

	if cacheDef == nil {
		t.Fatal("DeclarativeFlags() missing cache behavior")
	}
	if cacheDef.EnableFlag != "--with-cache" {
		t.Errorf("cache EnableFlag = %v, want --with-cache", cacheDef.EnableFlag)
	}
	if cacheDef.DisableFlag != "--no-cache" {
		t.Errorf("cache DisableFlag = %v, want --no-cache", cacheDef.DisableFlag)
	}
	if !cacheDef.DefaultOn {
		t.Error("cache should be DefaultOn=true")
	}

	if depsDef == nil {
		t.Fatal("DeclarativeFlags() missing deps behavior")
	}
	if depsDef.EnableFlag != "--with-deps" {
		t.Errorf("deps EnableFlag = %v, want --with-deps", depsDef.EnableFlag)
	}
	if depsDef.DisableFlag != "--no-deps" {
		t.Errorf("deps DisableFlag = %v, want --no-deps", depsDef.DisableFlag)
	}
}

func TestCacheFlagSet_GetDeclarativeState(t *testing.T) {
	s := NewCacheFlagSet()
	env := &environment.Env{IsLocalConsole: true}

	// Parse with explicit cache setting
	_, err := s.Parse([]string{"--no-cache"}, env)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Test GetDeclarativeState for cache
	cacheState := s.GetDeclarativeState("cache")
	if cacheState == nil {
		t.Fatal("GetDeclarativeState(cache) returned nil")
	}
	if !cacheState.ExplicitlyOff {
		t.Error("cache should be ExplicitlyOff")
	}
	if cacheState.ExplicitlyOn {
		t.Error("cache should not be ExplicitlyOn")
	}

	// Test GetDeclarativeState for deps (not set)
	depsState := s.GetDeclarativeState("deps")
	if depsState == nil {
		t.Fatal("GetDeclarativeState(deps) returned nil")
	}
	if depsState.ExplicitlyOff || depsState.ExplicitlyOn {
		t.Error("deps should not be explicitly set")
	}

	// Test unknown behavior returns nil
	unknownState := s.GetDeclarativeState("unknown")
	if unknownState != nil {
		t.Error("GetDeclarativeState(unknown) should return nil")
	}
}

func TestCacheFlagSet_AllDeclarativeStates(t *testing.T) {
	s := NewCacheFlagSet()
	env := &environment.Env{IsLocalConsole: true}

	_, err := s.Parse([]string{"--with-cache", "--skip-deps"}, env)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	states := s.AllDeclarativeStates()
	if len(states) != 2 {
		t.Errorf("AllDeclarativeStates() returned %d states, want 2", len(states))
	}

	// Verify states
	stateMap := make(map[string]*DeclarativeState)
	for _, s := range states {
		stateMap[s.Behavior] = s
	}

	cacheState := stateMap["cache"]
	if cacheState == nil {
		t.Fatal("missing cache state")
	}
	if !cacheState.ExplicitlyOn {
		t.Error("cache should be ExplicitlyOn")
	}

	depsState := stateMap["deps"]
	if depsState == nil {
		t.Fatal("missing deps state")
	}
	if !depsState.ExplicitlyOff {
		t.Error("deps should be ExplicitlyOff")
	}
}
