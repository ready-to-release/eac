package flags

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
	"github.com/ready-to-release/eac/go/eac/adapters/tui"
)

// OutputFlags holds TUI and output control flags.
type OutputFlags struct {
	UseTUI           bool // --tui: Enable TUI console
	NoTUI            bool // --no-tui: Disable TUI console
	TUIHeight        int  // --tui-height N: TUI height (3-20)
	TUIASCIIMode     bool // --ascii: ASCII-only TUI characters
	SkipTUIDelay     bool // --skip-tui-delay: Exit immediately when done (skip user interaction)
	Debug            bool // --debug, -d: Enable debug logging
	ShowTimings      bool // --timings: Show timing breakdown
	TUIExplicitlySet bool // Internal: whether TUI was explicitly set
}

// OutputFlagSet implements FlagSet for output control flags.
type OutputFlagSet struct {
	flags *OutputFlags
}

// NewOutputFlagSet creates a new output flag set.
func NewOutputFlagSet() *OutputFlagSet {
	return &OutputFlagSet{
		flags: &OutputFlags{
			TUIHeight: tui.DefaultHeight,
		},
	}
}

// Name returns the unique identifier for this flag set.
func (s *OutputFlagSet) Name() string {
	return "output"
}

// Description returns a human-readable description for documentation.
func (s *OutputFlagSet) Description() string {
	return "Control terminal user interface and output display"
}

// Flags returns the flag definitions for this set.
func (s *OutputFlagSet) Flags() []FlagDef {
	return []FlagDef{
		{
			Name:              "tui",
			Type:              "bool",
			Default:           "auto",
			Usage:             "Enable TUI console",
			MutuallyExclusive: []string{"no-tui"},
		},
		{
			Name:              "no-tui",
			Type:              "bool",
			Default:           "false",
			Usage:             "Disable TUI console (use plain output)",
			MutuallyExclusive: []string{"tui"},
		},
		{
			Name:    "tui-height",
			Type:    "int",
			Default: fmt.Sprintf("%d", tui.DefaultHeight),
			Usage:   fmt.Sprintf("Set TUI console height (3-20, default: %d)", tui.DefaultHeight),
		},
		{
			Name:    "ascii",
			Type:    "bool",
			Default: "false",
			Usage:   "Use ASCII-only characters in TUI",
		},
		{
			Name:    "skip-tui-delay",
			Type:    "bool",
			Default: "false",
			Usage:   "Skip TUI exit delay (exit immediately when done)",
		},
		{
			Name:      "debug",
			Shorthand: "d",
			Type:      "bool",
			Default:   "false",
			Usage:     "Enable debug logs to console",
		},
		{
			Name:    "timings",
			Type:    "bool",
			Default: "false",
			Usage:   "Show detailed timing breakdown",
		},
	}
}

// Parse processes arguments and extracts flags belonging to this set.
func (s *OutputFlagSet) Parse(args []string, env *environment.Env) ([]string, error) {
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

func (s *OutputFlagSet) parseFlag(arg string, args []string, i int) (consumed bool, advance int, err error) {
	switch arg {
	case "--tui":
		s.flags.UseTUI = true
		s.flags.TUIExplicitlySet = true
		return true, 0, nil
	case "--no-tui":
		s.flags.NoTUI = true
		s.flags.UseTUI = false
		s.flags.TUIExplicitlySet = true
		return true, 0, nil
	case "--ascii":
		s.flags.TUIASCIIMode = true
		return true, 0, nil
	case "--skip-tui-delay":
		s.flags.SkipTUIDelay = true
		return true, 0, nil
	case "--debug", "-d":
		s.flags.Debug = true
		return true, 0, nil
	case "--timings":
		s.flags.ShowTimings = true
		return true, 0, nil
	case "--tui-height":
		if i+1 >= len(args) {
			return true, 0, fmt.Errorf("--tui-height requires a value")
		}
		height, err := strconv.Atoi(args[i+1])
		if err != nil || height < 3 || height > 20 {
			return true, 0, fmt.Errorf("--tui-height must be between 3 and 20")
		}
		s.flags.TUIHeight = height
		return true, 1, nil
	}

	if strings.HasPrefix(arg, "--tui-height=") {
		heightStr := strings.TrimPrefix(arg, "--tui-height=")
		height, err := strconv.Atoi(heightStr)
		if err != nil || height < 3 || height > 20 {
			return true, 0, fmt.Errorf("--tui-height must be between 3 and 20")
		}
		s.flags.TUIHeight = height
		return true, 0, nil
	}

	return false, 0, nil
}

// Validate checks for invalid flag combinations within this set.
func (s *OutputFlagSet) Validate() error {
	return nil
}

// ApplyDefaults sets environment-aware default values.
func (s *OutputFlagSet) ApplyDefaults(env *environment.Env) {
	if !s.flags.TUIExplicitlySet {
		s.flags.UseTUI = env.ShouldUseTUI()
	}
}

// Values returns the parsed flag values.
func (s *OutputFlagSet) Values() *OutputFlags {
	return s.flags
}

// ValidateTUI validates TUI usage against environment constraints.
func (s *OutputFlagSet) ValidateTUI(env *environment.Env) error {
	return env.ValidateTUI(s.flags.TUIExplicitlySet, s.flags.UseTUI)
}
