// Command: test list-suites
// Short: List all available test suites
// Long: List all available test suites defined in the repository.
// Long:
// Long: Test suites are logical groupings of tests defined in domain.
// Long: This command displays all configured suites, making it easy to discover
// Long: what test suites are available for execution.
// Long:
// Long: Expected Output:
// Long:   - List of configured test suite names (unit, integration, acceptance, etc.)
// Long:   - Suite name, display name, and description for each suite
// Long:
// Long: Example:
// Long:   test list-suites
package test

import (
	"os"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/testing"
)

// ListSuites lists all available test suites.
func ListSuites() int {
	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	log.Info("Available test suites:")
	log.Info("")

	suites := testing.ListSuites()
	for _, moniker := range suites {
		suite, err := testing.GetSuite(moniker)
		if err != nil {
			log.Errorf("%v", err)
			continue
		}

		log.Infof("  %s", suite.Moniker)
		log.Infof("    Name: %s", suite.Name)
		log.Infof("    Description: %s", suite.Description)
		log.Info("")
	}

	return 0
}
