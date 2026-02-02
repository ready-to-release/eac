package flags

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/environment"
)

// ExecutionFlags holds parallelism and execution control flags.
type ExecutionFlags struct {
	Turbo          bool // --turbo: Enable turbo mode (+25% parallelism)
	Roof           int  // --roof N: Max parallel capacity (0=auto, 1=sequential)
	UnlayeredBuild bool // --unlayered-build, --unlayered: Disable layered execution

	// Declarative tracking fields
	ParallelExplicit bool // True if --parallel or --sequential was used
	LayeredExplicit  bool // True if --layered or --unlayered was used
}

// ExecutionFlagSet implements FlagSet for execution control flags.
type ExecutionFlagSet struct {
	flags *ExecutionFlags
}

// NewExecutionFlagSet creates a new execution flag set.
func NewExecutionFlagSet() *ExecutionFlagSet {
	return &ExecutionFlagSet{
		flags: &ExecutionFlags{},
	}
}

// Name returns the unique identifier for this flag set.
func (s *ExecutionFlagSet) Name() string {
	return "execution"
}

// Description returns a human-readable description for documentation.
func (s *ExecutionFlagSet) Description() string {
	return "Control parallel execution behavior"
}

// Flags returns the flag definitions for this set.
func (s *ExecutionFlagSet) Flags() []FlagDef {
	return []FlagDef{
		{
			Name:    "turbo",
			Type:    "bool",
			Default: "false",
			Usage:   "Enable turbo mode for faster execution (increases parallelism)",
		},
		{
			Name:    "roof",
			Type:    "int",
			Default: "0",
			Usage:   "Limit max parallel capacity (0=auto, 1=sequential)",
		},
		{
			Name:    "parallel",
			Type:    "bool",
			Default: "true",
			Usage:   "Enable parallel execution (default, for self-documenting CI)",
		},
		{
			Name:              "sequential",
			Type:              "bool",
			Default:           "false",
			Usage:             "Disable parallel execution (equivalent to --roof 1)",
			MutuallyExclusive: []string{"parallel"},
		},
		{
			Name:    "layered",
			Type:    "bool",
			Default: "true",
			Usage:   "Enable layered/dependency-ordered execution (default, for self-documenting CI)",
		},
		{
			Name:              "unlayered",
			Type:              "bool",
			Default:           "false",
			Usage:             "Disable layered execution (run all modules in parallel)",
			MutuallyExclusive: []string{"layered"},
		},
		{
			Name:    "unlayered-build",
			Type:    "bool",
			Default: "false",
			Usage:   "Disable layered execution (legacy alias for --unlayered)",
		},
	}
}

// Parse processes arguments and extracts flags belonging to this set.
func (s *ExecutionFlagSet) Parse(args []string, env *environment.Env) ([]string, error) {
	var remaining []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		consumed, advance, err := s.parseFlag(arg, args, i)
		if err != nil {
			return nil, err
		}
		if consumed {
			i += advance
		} else {
			remaining = append(remaining, arg)
		}
	}

	return remaining, nil
}

func (s *ExecutionFlagSet) parseFlag(arg string, args []string, i int) (consumed bool, advance int, err error) {
	switch arg {
	case "--turbo":
		s.flags.Turbo = true
		return true, 0, nil

	// Declarative parallel flags
	case "--parallel":
		// Parallel is default, just mark as explicit
		s.flags.ParallelExplicit = true
		return true, 0, nil
	case "--sequential":
		s.flags.Roof = 1 // Sequential = roof 1
		s.flags.ParallelExplicit = true
		return true, 0, nil

	// Declarative layered flags
	case "--layered":
		s.flags.UnlayeredBuild = false
		s.flags.LayeredExplicit = true
		return true, 0, nil
	case "--unlayered":
		s.flags.UnlayeredBuild = true
		s.flags.LayeredExplicit = true
		return true, 0, nil

	// Legacy layered flag (backward compat)
	case "--unlayered-build":
		s.flags.UnlayeredBuild = true
		s.flags.LayeredExplicit = true
		return true, 0, nil

	case "--roof":
		if i+1 >= len(args) {
			return true, 0, fmt.Errorf("--roof requires a value")
		}
		roof, err := strconv.Atoi(args[i+1])
		if err != nil || roof < 0 {
			return true, 0, fmt.Errorf("--roof must be a non-negative integer")
		}
		s.flags.Roof = roof
		return true, 1, nil
	}

	if strings.HasPrefix(arg, "--roof=") {
		roofStr := strings.TrimPrefix(arg, "--roof=")
		roof, err := strconv.Atoi(roofStr)
		if err != nil || roof < 0 {
			return true, 0, fmt.Errorf("--roof must be a non-negative integer")
		}
		s.flags.Roof = roof
		return true, 0, nil
	}

	return false, 0, nil
}

// Validate checks for invalid flag combinations within this set.
func (s *ExecutionFlagSet) Validate() error {
	return nil
}

// ApplyDefaults sets environment-aware default values.
func (s *ExecutionFlagSet) ApplyDefaults(env *environment.Env) {}

// Values returns the parsed flag values.
func (s *ExecutionFlagSet) Values() *ExecutionFlags {
	return s.flags
}

// DeclarativeFlags returns the declarative flag definitions for this set.
func (s *ExecutionFlagSet) DeclarativeFlags() []DeclarativeFlagDef {
	return []DeclarativeFlagDef{
		{
			Behavior:    "parallel",
			EnableFlag:  "--parallel",
			DisableFlag: "--sequential",
			LegacyFlags: []LegacyFlagMapping{
				{LegacyFlag: "--roof", MapsTo: "disable"}, // --roof 1 implies sequential
			},
			DefaultOn:   true,
			EnvAware:    false,
			Description: "Parallel execution of modules within layers",
		},
		{
			Behavior:    "layered",
			EnableFlag:  "--layered",
			DisableFlag: "--unlayered",
			LegacyFlags: []LegacyFlagMapping{
				{LegacyFlag: "--unlayered-build", MapsTo: "disable"},
			},
			DefaultOn:   true,
			EnvAware:    false,
			Description: "Layered/dependency-ordered execution of modules",
		},
	}
}

// GetDeclarativeState returns the state of a specific behavior.
// Returns nil if the behavior is not tracked by this set.
func (s *ExecutionFlagSet) GetDeclarativeState(behavior string) *DeclarativeState {
	switch behavior {
	case "parallel":
		// Parallel is OFF when Roof == 1 (sequential mode)
		isSequential := s.flags.Roof == 1
		return &DeclarativeState{
			Behavior:      "parallel",
			ExplicitlyOn:  s.flags.ParallelExplicit && !isSequential,
			ExplicitlyOff: s.flags.ParallelExplicit && isSequential,
		}
	case "layered":
		return &DeclarativeState{
			Behavior:      "layered",
			ExplicitlyOn:  s.flags.LayeredExplicit && !s.flags.UnlayeredBuild,
			ExplicitlyOff: s.flags.LayeredExplicit && s.flags.UnlayeredBuild,
		}
	default:
		return nil
	}
}

// AllDeclarativeStates returns states for all behaviors tracked by this set.
func (s *ExecutionFlagSet) AllDeclarativeStates() []*DeclarativeState {
	return []*DeclarativeState{
		s.GetDeclarativeState("parallel"),
		s.GetDeclarativeState("layered"),
	}
}
