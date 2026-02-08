package flags

import (
	"github.com/ready-to-release/eac/go/clibase/environment"
)

// DeclarativeFlagDef defines a behavior that can be explicitly enabled/disabled.
// This enables self-documenting CI pipelines where every significant behavior
// is visible in the command invocation.
type DeclarativeFlagDef struct {
	// Behavior is the logical behavior name (e.g., "cache", "parallel", "tui")
	Behavior string

	// EnableFlag is the flag to explicitly enable (e.g., "--with-cache")
	EnableFlag string

	// DisableFlag is the flag to explicitly disable (e.g., "--no-cache")
	DisableFlag string

	// DefaultOn indicates the default state when neither flag is specified
	DefaultOn bool

	// EnvAware indicates if the default depends on execution environment
	EnvAware bool

	// EnvDefaults maps environment context names to default values (when EnvAware=true)
	// Keys: "local", "CI", "container", "test"
	EnvDefaults map[string]bool

	// Description is the human-readable description of this behavior
	Description string
}

// ResolveDefault returns the default value for this behavior based on environment.
func (d *DeclarativeFlagDef) ResolveDefault(env *environment.Env) bool {
	if !d.EnvAware || d.EnvDefaults == nil {
		return d.DefaultOn
	}

	// Check environment-specific defaults using ContextName()
	contextName := env.ContextName()
	if val, ok := d.EnvDefaults[contextName]; ok {
		return val
	}

	// Fall back to DefaultOn if context not found in EnvDefaults
	return d.DefaultOn
}

// DeclarativeState tracks the explicit state of a behavior flag pair.
type DeclarativeState struct {
	// Behavior is the name of the behavior being tracked
	Behavior string

	// ExplicitlyOn is true if user passed the enable flag (e.g., --with-cache)
	ExplicitlyOn bool

	// ExplicitlyOff is true if user passed the disable flag (e.g., --no-cache)
	ExplicitlyOff bool
}

// IsEnabled returns whether the behavior is enabled, considering explicit
// settings and environment-aware defaults.
func (s *DeclarativeState) IsEnabled(def *DeclarativeFlagDef, env *environment.Env) bool {
	// If explicitly disabled, return false (safety: off wins in conflicts)
	if s.ExplicitlyOff {
		return false
	}

	// If explicitly enabled, return true
	if s.ExplicitlyOn {
		return true
	}

	// Use environment-aware default
	return def.ResolveDefault(env)
}

// WasExplicitlySet returns true if user specified either the enable or disable flag.
func (s *DeclarativeState) WasExplicitlySet() bool {
	return s.ExplicitlyOn || s.ExplicitlyOff
}

// DeclarativeFlagSet extends FlagSet with declarative flag support.
// Flag sets that support declarative flags should implement this interface.
type DeclarativeFlagSet interface {
	FlagSet

	// DeclarativeFlags returns all declarative flag definitions for this set.
	DeclarativeFlags() []DeclarativeFlagDef

	// GetDeclarativeState returns the state of a specific behavior.
	// Returns nil if the behavior is not tracked by this set.
	GetDeclarativeState(behavior string) *DeclarativeState

	// AllDeclarativeStates returns states for all behaviors tracked by this set.
	AllDeclarativeStates() []*DeclarativeState
}
