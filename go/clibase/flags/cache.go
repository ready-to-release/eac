package flags

import (
	"strings"

	"github.com/ready-to-release/eac/go/clibase/environment"
	"github.com/ready-to-release/eac/go/core/cache"
)

// CacheFlags holds cache and system dependency control flags.
type CacheFlags struct {
	// CacheConfig controls fine-grained cache skipping via --skip-cache=<spec>
	CacheConfig *cache.Config

	SkipDeps bool // --no-deps: Skip system dependency verification

	// Declarative tracking fields
	CacheExplicit bool // True if --with-cache or --skip-cache was used
	DepsExplicit  bool // True if --with-deps or --no-deps was used
}

// CacheFlagSet implements FlagSet and DeclarativeFlagSet for cache control flags.
type CacheFlagSet struct {
	flags *CacheFlags
}

// NewCacheFlagSet creates a new cache flag set.
func NewCacheFlagSet() *CacheFlagSet {
	return &CacheFlagSet{
		flags: &CacheFlags{
			CacheConfig: cache.NewConfig(),
		},
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
			Usage:             "Disable incremental caching (alias for --skip-cache=state)",
			MutuallyExclusive: []string{"with-cache"},
		},
		{
			Name:    "skip-cache",
			Type:    "string",
			Default: "",
			Usage:   "Skip cache by spec: state, asset, registry, layer, work, local, remote, all, or level:type (e.g., local:registry)",
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
	}
}

// Parse processes arguments and extracts flags belonging to this set.
func (s *CacheFlagSet) Parse(args []string, env *environment.Env) ([]string, error) {
	var remaining []string

	for _, arg := range args {
		switch {
		// --skip-cache=<spec> with value
		case strings.HasPrefix(arg, "--skip-cache="):
			value := strings.TrimPrefix(arg, "--skip-cache=")
			specs, err := cache.ParseSpecs(value)
			if err != nil {
				return nil, err
			}
			for _, spec := range specs {
				s.flags.CacheConfig.Skip(spec)
			}
			s.flags.CacheExplicit = true

		// --skip-cache without value: use default skip specs (state + work)
		case arg == "--skip-cache":
			for _, spec := range cache.DefaultSkipSpecs {
				s.flags.CacheConfig.Skip(spec)
			}
			s.flags.CacheExplicit = true

		// --no-cache (alias for --skip-cache=state)
		case arg == "--no-cache":
			s.flags.CacheConfig.Skip(cache.Spec{Level: cache.LevelAll, Type: cache.TypeState})
			s.flags.CacheExplicit = true

		// --with-cache (declarative, currently no-op but marks explicit)
		case arg == "--with-cache":
			s.flags.CacheExplicit = true

		// Declarative deps flags
		case arg == "--with-deps":
			s.flags.SkipDeps = false
			s.flags.DepsExplicit = true
		case arg == "--no-deps":
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
			DefaultOn:   true,
			EnvAware:    false,
			Description: "Incremental caching for faster rebuilds",
		},
		{
			Behavior:    "deps",
			EnableFlag:  "--with-deps",
			DisableFlag: "--no-deps",
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
		// Cache is "off" if any skip specs are present
		hasSkips := len(s.flags.CacheConfig.SkipSpecs) > 0
		return &DeclarativeState{
			Behavior:      "cache",
			ExplicitlyOn:  s.flags.CacheExplicit && !hasSkips,
			ExplicitlyOff: s.flags.CacheExplicit && hasSkips,
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
