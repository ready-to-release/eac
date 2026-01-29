package flags

import (
	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
)

// DryRunFlags holds dry run control flags.
type DryRunFlags struct {
	DryRun bool // --dry-run: Simulate execution
}

// DryRunFlagSet implements FlagSet for dry run control flags.
type DryRunFlagSet struct {
	flags *DryRunFlags
}

// NewDryRunFlagSet creates a new dry run flag set.
func NewDryRunFlagSet() *DryRunFlagSet {
	return &DryRunFlagSet{
		flags: &DryRunFlags{},
	}
}

// Name returns the unique identifier for this flag set.
func (s *DryRunFlagSet) Name() string {
	return "dryrun"
}

// Description returns a human-readable description for documentation.
func (s *DryRunFlagSet) Description() string {
	return "Simulation mode"
}

// Flags returns the flag definitions for this set.
func (s *DryRunFlagSet) Flags() []FlagDef {
	return []FlagDef{
		{
			Name:    "dry-run",
			Type:    "bool",
			Default: "false",
			Usage:   "Simulate execution without making changes",
		},
	}
}

// Parse processes arguments and extracts flags belonging to this set.
func (s *DryRunFlagSet) Parse(args []string, env *environment.Env) ([]string, error) {
	var remaining []string

	for _, arg := range args {
		switch arg {
		case "--dry-run":
			s.flags.DryRun = true
		default:
			remaining = append(remaining, arg)
		}
	}

	return remaining, nil
}

// Validate checks for invalid flag combinations within this set.
func (s *DryRunFlagSet) Validate() error {
	return nil
}

// ApplyDefaults sets environment-aware default values.
func (s *DryRunFlagSet) ApplyDefaults(env *environment.Env) {}

// Values returns the parsed flag values.
func (s *DryRunFlagSet) Values() *DryRunFlags {
	return s.flags
}
