package flags

import (
	"strings"

	"github.com/ready-to-release/eac/go/clibase/environment"
)

// ModuleFlags holds module selection and dependency flags.
type ModuleFlags struct {
	Exclude  string // --exclude: Module exclusion pattern
	SkipDepm bool   // --skip-depm: Skip module dependency handling (build/test only)
}

// ModuleFlagSet implements FlagSet for module control flags.
type ModuleFlagSet struct {
	flags *ModuleFlags
}

// NewModuleFlagSet creates a new module flag set.
func NewModuleFlagSet() *ModuleFlagSet {
	return &ModuleFlagSet{
		flags: &ModuleFlags{},
	}
}

// Name returns the unique identifier for this flag set.
func (s *ModuleFlagSet) Name() string {
	return "module"
}

// Description returns a human-readable description for documentation.
func (s *ModuleFlagSet) Description() string {
	return "Module selection and dependencies"
}

// Flags returns the flag definitions for this set.
func (s *ModuleFlagSet) Flags() []FlagDef {
	return []FlagDef{
		{
			Name:    "exclude",
			Type:    "string",
			Default: "",
			Usage:   "Exclude modules matching pattern",
		},
		{
			Name:    "skip-depm",
			Type:    "bool",
			Default: "false",
			Usage:   "Skip module dependency handling (build: exclude deps, test: skip validation)",
		},
	}
}

// Parse processes arguments and extracts flags belonging to this set.
func (s *ModuleFlagSet) Parse(args []string, env *environment.Env) ([]string, error) {
	var remaining []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--exclude" && i+1 < len(args):
			s.flags.Exclude = args[i+1]
			i++
		case strings.HasPrefix(arg, "--exclude="):
			s.flags.Exclude = strings.TrimPrefix(arg, "--exclude=")
		case arg == "--skip-depm":
			s.flags.SkipDepm = true
		default:
			remaining = append(remaining, arg)
		}
	}

	return remaining, nil
}

// Validate checks for invalid flag combinations within this set.
func (s *ModuleFlagSet) Validate() error {
	return nil
}

// ApplyDefaults sets environment-aware default values.
func (s *ModuleFlagSet) ApplyDefaults(env *environment.Env) {}

// Values returns the parsed flag values.
func (s *ModuleFlagSet) Values() *ModuleFlags {
	return s.flags
}
