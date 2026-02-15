package lint

import (
	"encoding/json"
	"os"

	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
)

// buildLintInitSummary creates the init summary for lint commands.
func buildLintInitSummary(ctx *cmdframework.ExecutionContext) {
	summary := initsummary.New("lint").
		SetRequest(ctx.Config.Monikers, ctx.GetExecutionMonikers()).
		SetExecutionContext(string(logging.GetExecutionContext())).
		SetFlags(initsummary.Flags{
			DebugMode: ctx.Config.DebugMode,
			UseTUI:    ctx.Config.UseTUI,
		}).
		SetOutputDir(paths.OutLintRelPath)

	ctx.InitSummary = summary
}

// countLintIssues counts the number of issues from a golangci-lint JSON output.
func countLintIssues(jsonPath string) (int, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return 0, err
	}

	if len(data) == 0 {
		return 0, nil
	}

	var jsonOutput struct {
		Issues []interface{} `json:"Issues"`
	}
	if err := json.Unmarshal(data, &jsonOutput); err != nil {
		return 0, err
	}

	return len(jsonOutput.Issues), nil
}
