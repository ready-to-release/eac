// Command: get module-trigger-reason
// Short: Get the reason a module was triggered for CI
// Long: Extracts the trigger reason for a specific module from MODULE_STATUS JSON.
// Long:
// Long: This is used in CI summaries to explain why each module needs CI.
// Long:
// Long: Example:
// Long:   get module-trigger-reason docs --json "$MODULE_STATUS"
// Long:   # Output: "files_changed_since_abc1234" or "dependency eac-cli needs CI"
// Flag.json: type=string, usage=MODULE_STATUS JSON (defaults to env var)
package get

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/environments"
)

func init() {
	registry.Register(GetModuleTriggerReason)
}

func GetModuleTriggerReason() int {
	// Parse arguments
	module := ""
	jsonInput := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--json" && i+1 < len(os.Args):
			jsonInput = os.Args[i+1]
			i++
		case !strings.HasPrefix(arg, "--") && module == "":
			module = arg
		}
	}

	if module == "" {
		fmt.Fprintln(os.Stderr, "Error: module name required")
		fmt.Fprintln(os.Stderr, "Usage: get module-trigger-reason <module> [--json <json>]")
		return 1
	}

	// Get JSON from flag or environment
	if jsonInput == "" {
		jsonInput = os.Getenv(environments.EnvModuleStatus)
	}

	if jsonInput == "" {
		fmt.Println("unknown")
		return 0
	}

	// Parse JSON: map[module]ModuleCIStatus
	var statusByModule map[string]ModuleCIStatus
	if err := json.Unmarshal([]byte(jsonInput), &statusByModule); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse JSON: %v\n", err)
		return 1
	}

	status, ok := statusByModule[module]
	if !ok {
		fmt.Println("unknown")
		return 0
	}

	// Format the reason for human readability
	reason := formatTriggerReason(status.Reason)
	fmt.Println(reason)

	return 0
}

// formatTriggerReason converts internal reason codes to human-readable text.
func formatTriggerReason(reason string) string {
	switch {
	case strings.HasPrefix(reason, "dependency "):
		// "dependency eac-cli needs CI" -> "eac-cli changed"
		parts := strings.SplitN(reason, " ", 3)
		if len(parts) >= 2 {
			return parts[1] + " changed"
		}
		return reason
	case strings.HasPrefix(reason, "files_changed_since_"):
		// "files_changed_since_abc1234" -> "files changed"
		return "files changed"
	case reason == "no_ci_history":
		return "no previous CI"
	case reason == "query_failed":
		return "CI query failed"
	default:
		return reason
	}
}
