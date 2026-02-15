package lint

import (
	"fmt"
	"strings"
)

// LintSpecificFlags holds flags that are specific to the lint command.
// These are not shared with other commands (build, test, scan).
type LintSpecificFlags struct {
	Fix        bool   // --fix: Auto-fix issues where possible
	ConfigPath string // --config: Override lint config file path
}

// ParseLintSpecificFlags parses lint-specific flags from remaining args.
// Returns the flags and any unknown/unprocessed args.
func ParseLintSpecificFlags(args []string) (*LintSpecificFlags, []string, error) {
	flags := &LintSpecificFlags{}
	var remaining []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		consumed, advance, err := parseLintFlag(arg, args, i, flags)
		if err != nil {
			return nil, nil, err
		}
		if consumed {
			i += advance
		} else {
			remaining = append(remaining, arg)
		}
	}

	return flags, remaining, nil
}

func parseLintFlag(arg string, args []string, i int, flags *LintSpecificFlags) (consumed bool, advance int, err error) {
	switch arg {
	case "--fix":
		flags.Fix = true
		return true, 0, nil
	case "--config":
		if i+1 >= len(args) {
			return true, 0, fmt.Errorf("--config requires a value")
		}
		flags.ConfigPath = args[i+1]
		return true, 1, nil
	}

	// Handle --config=value syntax
	if strings.HasPrefix(arg, "--config=") {
		flags.ConfigPath = strings.TrimPrefix(arg, "--config=")
		return true, 0, nil
	}

	return false, 0, nil
}
