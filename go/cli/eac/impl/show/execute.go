package show

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/clibase/flags"
)

// ExecuteShowCommand is a helper that wraps the common pattern for show commands:
// 1. Validate flags against registry metadata
// 2. Execute the command function
// 3. Return the exit code.
func ExecuteShowCommand(fn func() int) int {
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return fn()
}
