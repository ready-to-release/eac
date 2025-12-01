// Command: test list-suites
// Short: List all available test suites
// Long: List all available test suites defined in the repository.
// Long:
// Long: Test suites are logical groupings of tests defined in contracts.
// Long: This command displays all configured suites, making it easy to discover
// Long: what test suites are available for execution.
// Long:
// Long: Example:
// Long:   test list-suites
package test

import (
	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/testing"
)

func init() {
	registry.Register(ListSuites)
}

// ListSuites lists all available test suites
func ListSuites() int {
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
