package flags

import (
	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
)

// CacheFlags holds cache and system dependency control flags.
type CacheFlags struct {
	SkipCache bool // --skip-cache: Bypass incremental detection
	SkipDeps  bool // --skip-deps: Skip system dependency verification
}

// CacheFlagSet implements FlagSet for cache control flags.
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
			Name:    "skip-cache",
			Type:    "bool",
			Default: "false",
			Usage:   "Skip incremental cache, force full execution",
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
		case "--skip-cache":
			s.flags.SkipCache = true
		case "--skip-deps":
			s.flags.SkipDeps = true
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
