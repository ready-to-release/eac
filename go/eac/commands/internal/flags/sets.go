// Package flags provides composable flag sets for command-line parsing.
// Commands subscribe to specific flag sets, enabling fine-grained control
// over which flags each command accepts while ensuring consistency.
package flags

import (
	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
)

// FlagSet represents a composable group of related flags.
// Each set is self-contained with its own parsing and validation.
type FlagSet interface {
	// Name returns the unique identifier for this flag set.
	Name() string

	// Description returns a human-readable description for documentation.
	Description() string

	// Flags returns the flag definitions for this set.
	Flags() []FlagDef

	// Parse processes arguments and extracts flags belonging to this set.
	// Returns remaining arguments that weren't consumed.
	Parse(args []string, env *environment.Env) (remaining []string, err error)

	// Validate checks for invalid flag combinations within this set.
	Validate() error

	// ApplyDefaults sets environment-aware default values.
	ApplyDefaults(env *environment.Env)
}

// FlagDef defines metadata for a single flag.
type FlagDef struct {
	Name              string   // Long name without "--" (e.g., "tui-height")
	Shorthand         string   // Short name without "-" (e.g., "d" for debug)
	Type              string   // "bool", "int", "string"
	Default           string   // Default value as string
	Usage             string   // Help text
	MutuallyExclusive []string // Flags that cannot be used together
}
