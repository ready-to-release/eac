package test

import (
	"fmt"
	"strings"
)

// TestSpecificFlags holds flags that are specific to the test command.
// These are not shared with other commands (build, lint, scan).
type TestSpecificFlags struct {
	SuiteName string // --suite: Filter tests by suite
	Coverage  bool   // --coverage: Enable coverage reporting
	ListOnly  bool   // --list-only: List tests without running them
}

// ParseTestSpecificFlags parses test-specific flags from remaining args.
// Returns the flags and any unknown/unprocessed args.
func ParseTestSpecificFlags(args []string) (*TestSpecificFlags, []string, error) {
	flags := &TestSpecificFlags{}
	var remaining []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		consumed, advance, err := parseTestFlag(arg, args, i, flags)
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

func parseTestFlag(arg string, args []string, i int, flags *TestSpecificFlags) (consumed bool, advance int, err error) {
	switch arg {
	case "--suite":
		if i+1 >= len(args) {
			return true, 0, fmt.Errorf("--suite requires a suite name")
		}
		flags.SuiteName = args[i+1]
		return true, 1, nil
	case "--coverage":
		flags.Coverage = true
		return true, 0, nil
	case "--list-only":
		flags.ListOnly = true
		return true, 0, nil
	}

	// Handle --suite=value syntax
	if strings.HasPrefix(arg, "--suite=") {
		flags.SuiteName = strings.TrimPrefix(arg, "--suite=")
		return true, 0, nil
	}

	return false, 0, nil
}
