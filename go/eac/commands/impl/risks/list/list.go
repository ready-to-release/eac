// Command: risks list
// Short: List risk controls and their linkages to specifications
// Flag.filter: type=string, default=all, shorthand=f, usage=Filter view: all, orphaned, linked, or missing-links, completion=all,orphaned,linked,missing-links
// Flag.json: type=bool, default=false, shorthand=j, usage=Output as JSON instead of table
// Flag.debug: type=bool, default=false, shorthand=D, usage=Enable debug mode

package list

import (
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(RisksList)
}

// RisksList is the entry point for the risks list command
func RisksList() int {
	return Run()
}

// Run executes the risks list command
func Run() int {
	// Parse configuration
	config, err := parseConfig()
	if err != nil {
		if err.Error() == "help requested" {
			showHelp()
			return 0
		}
		log.Errorf("Error: %v", err)
		return 1
	}

	// Analyze linkages
	report, err := analyzeLinkages(config.WorkspaceRoot)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Format output
	var output string
	if config.JSON {
		output, err = formatJSON(report)
		if err != nil {
			log.Errorf("Error formatting JSON: %v", err)
			return 1
		}
	} else {
		output = formatTable(report, config.Filter)
	}

	// Print output
	log.Info(output)

	return 0
}

// showHelp displays help information
func showHelp() {
	help := `Usage: risks list [flags]

List risk controls and their linkages to specifications

Flags:
  -f, --filter <filter>        Filter view: all, orphaned, linked, or missing-links (default: all)
  -j, --json                   Output as JSON instead of table
  -D, --debug                  Enable debug mode
  -h, --help                   Show this help message

Examples:
  risks list                              # List all controls
  risks list -f orphaned                  # Show only orphaned controls
  risks list -f linked                    # Show only linked controls
  risks list -f missing-links             # Show specs missing risk control links
  risks list -j                           # Output as JSON
`
	log.Info(help)
}
