package flags

import (
	"github.com/ready-to-release/eac/go/clibase/environment"
)

// CacheFlags holds cache and system dependency control flags.
type CacheFlags struct {
	SkipCache bool // --skip-cache, --no-cache: Bypass incremental detection
	SkipDeps  bool // --skip-deps, --no-deps: Skip system dependency verification

	// Declarative tracking fields
	CacheExplicit bool // True if --with-cache or --no-cache was used
	DepsExplicit  bool // True if --with-deps, --no-deps, or --skip-deps was used
}

// CacheFlagSet implements FlagSet and DeclarativeFlagSet for cache control flags.
type CacheFlagSet struct {
	flags *CacheFlags
}

// NewCacheFlagSet creates a new cache flag set.
func NewCacheFlagSet() *CacheFlagSet {
	return &CacheFlagSet{
		flags: &CacheFlags{},
	}
}

// Name returns the unique identifier for this flag set.
func (s *CacheFlagSet) Name() string {
	return "cache"
}

// Description returns a human-readable description for documentation.
func (s *CacheFlagSet) Description() string {
	return "Control caching and system dependencies"
}

// Flags returns the flag definitions for this set.
func (s *CacheFlagSet) Flags() []FlagDef {
	return []FlagDef{
		{
			Name:    "with-cache",
			Type:    "bool",
			Default: "true",
			Usage:   "Enable incremental caching (default, for self-documenting CI)",
		},
		{
			Name:              "no-cache",
			Type:              "bool",
			Default:           "false",
			Usage:             "Disable incremental caching, force full execution",
			MutuallyExclusive: []string{"with-cache"},
		},
		{
			Name:    "skip-cache",
			Type:    "bool",
			Default: "false",
			Usage:   "Skip incremental cache (legacy alias for --no-cache)",
		},
		{
			Name:    "with-deps",
			Type:    "bool",
			Default: "true",
			Usage:   "Enable system dependency verification (default, for self-documenting CI)",
		},
		{
			Name:              "no-deps",
			Type:              "bool",
			Default:           "false",
			Usage:             "Skip system dependency verification",
			MutuallyExclusive: []string{"with-deps"},
		},
		{
			Name:    "skip-deps",
			Type:    "bool",
			Default: "false",
			Usage:   "Skip system dependency verification (go, docker, etc.)",
		},
	}
}

// Parse processes arguments and extracts flags belonging to this set.
func (s *CacheFlagSet) Parse(args []string, env *environment.Env) ([]string, error) {
	var remaining []string

	for _, arg := range args {
		switch arg {
		// Declarative cache flags
		case "--with-cache":
			s.flags.SkipCache = false
			s.flags.CacheExplicit = true
		case "--no-cache":
			s.flags.SkipCache = true
			s.flags.CacheExplicit = true

		// Legacy cache flag (backward compat)
		case "--skip-cache":
			s.flags.SkipCache = true
			s.flags.CacheExplicit = true

		// Declarative deps flags
		case "--with-deps":
			s.flags.SkipDeps = false
			s.flags.DepsExplicit = true
		case "--no-deps":
			s.flags.SkipDeps = true
			s.flags.DepsExplicit = true

		// Legacy deps flag (backward compat)
		case "--skip-deps":
			s.flags.SkipDeps = true
			s.flags.DepsExplicit = true

		default:
			remaining = append(remaining, arg)
		}
	}

	return remaining, nil
}

// Validate checks for invalid flag combinations within this set.
func (s *CacheFlagSet) Validate() error {
	return nil
}

// ApplyDefaults sets environment-aware default values.
func (s *CacheFlagSet) ApplyDefaults(env *environment.Env) {}

// Values returns the parsed flag values.
func (s *CacheFlagSet) Values() *CacheFlags {
	return s.flags
}

// DeclarativeFlags returns the declarative flag definitions for this set.
func (s *CacheFlagSet) DeclarativeFlags() []DeclarativeFlagDef {
	return []DeclarativeFlagDef{
		{
			Behavior:    "cache",
			EnableFlag:  "--with-cache",
			DisableFlag: "--no-cache",
			LegacyFlags: []LegacyFlagMapping{
				{LegacyFlag: "--skip-cache", MapsTo: "disable"},
			},
			DefaultOn:   true,
			EnvAware:    false,
			Description: "Incremental caching for faster rebuilds",
		},
		{
			Behavior:    "deps",
			EnableFlag:  "--with-deps",
			DisableFlag: "--no-deps",
			LegacyFlags: []LegacyFlagMapping{
				{LegacyFlag: "--skip-deps", MapsTo: "disable"},
			},
			DefaultOn:   true,
			EnvAware:    false,
			Description: "System dependency verification (go, docker, etc.)",
		},
	}
}

// GetDeclarativeState returns the state of a specific behavior.
// Returns nil if the behavior is not tracked by this set.
func (s *CacheFlagSet) GetDeclarativeState(behavior string) *DeclarativeState {
	switch behavior {
	case "cache":
		return &DeclarativeState{
			Behavior:      "cache",
			ExplicitlyOn:  s.flags.CacheExplicit && !s.flags.SkipCache,
			ExplicitlyOff: s.flags.CacheExplicit && s.flags.SkipCache,
		}
	case "deps":
		return &DeclarativeState{
			Behavior:      "deps",
			ExplicitlyOn:  s.flags.DepsExplicit && !s.flags.SkipDeps,
			ExplicitlyOff: s.flags.DepsExplicit && s.flags.SkipDeps,
		}
	default:
		return nil
	}
}

// AllDeclarativeStates returns states for all behaviors tracked by this set.
func (s *CacheFlagSet) AllDeclarativeStates() []*DeclarativeState {
	return []*DeclarativeState{
		s.GetDeclarativeState("cache"),
		s.GetDeclarativeState("deps"),
	}
}
