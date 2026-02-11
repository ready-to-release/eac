package get

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/environments"
)

type getModuleTriggerReasonCommand struct{}

var _ core.SimpleCommandPort = (*getModuleTriggerReasonCommand)(nil)

func (c *getModuleTriggerReasonCommand) Name() string { return "get module-trigger-reason" }

func (c *getModuleTriggerReasonCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-module-trigger-reason",
		Short:         "Get the reason a module was triggered for CI",
		Long: "Extracts the trigger reason for a specific module from MODULE_STATUS JSON.\n" +
			"\n" +
			"This is used in CI summaries to explain why each module needs CI.\n" +
			"\n" +
			"Example:\n" +
			"  get module-trigger-reason docs --json \"$MODULE_STATUS\"\n" +
			"  # Output: \"files_changed_since_abc1234\" or \"dependency eac needs CI\"",
		Flags: []core.FlagSpec{
			{Name: "json", Type: "string", Usage: "MODULE_STATUS JSON (defaults to env var)"},
		},
	}
}

func (c *getModuleTriggerReasonCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetModuleTriggerReason()
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
		// "dependency eac needs CI" -> "eac changed"
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
