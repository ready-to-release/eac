// Command: test list-suites
// Description: List all available test suites
// Usage: test list-suites
// HasSideEffects: false
package test

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/testing"
)

func init() {
	registry.Register(ListSuites)
}

// ListSuites lists all available test suites
func ListSuites() int {
	fmt.Println("Available test suites:")
	fmt.Println()

	suites := testing.ListSuites()
	for _, moniker := range suites {
		suite, err := testing.GetSuite(moniker)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		fmt.Printf("  %s\n", suite.Moniker)
		fmt.Printf("    Name: %s\n", suite.Name)
		fmt.Printf("    Description: %s\n", suite.Description)
		fmt.Println()
	}

	return 0
}
