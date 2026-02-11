package pipeline

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/gitexec"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/github"
	"github.com/ready-to-release/eac/go/core/repository"

	"github.com/ready-to-release/eac/go/clibase/ghexec"
)

type pipelineCheckRecentRunCommand struct{}

var _ core.SimpleCommandPort = (*pipelineCheckRecentRunCommand)(nil)

func (c *pipelineCheckRecentRunCommand) Name() string { return "pipeline check-recent-run" }

func (c *pipelineCheckRecentRunCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "pipeline-check-recent-run",
		Short:         "Check if a recent successful workflow run exists",
		Long:          "Checks if a successful workflow run exists for the given SHA within\nthe specified time window. Used to skip redundant CI runs.\n\nOutput:\n  Default: \"true\" or \"false\"\n  --format shell: HAS_RECENT=\"true/false\"\n\nExample:\n  pipeline check-recent-run --workflow ci-trigger.yaml --since 2h",
	}
}

func (c *pipelineCheckRecentRunCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return PipelineCheckRecentRun()
}

func PipelineCheckRecentRun() int {
	// Parse flags
	workflow := ""
	sha := ""
	since := config.CIRecentRunWindow()
	format := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--workflow" && i+1 < len(os.Args):
			workflow = os.Args[i+1]
			i++
		case arg == "--sha" && i+1 < len(os.Args):
			sha = os.Args[i+1]
			i++
		case arg == "--since" && i+1 < len(os.Args):
			d, err := time.ParseDuration(os.Args[i+1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid duration: %v\n", err)
				return 1
			}
			since = d
			i++
		case arg == "--format" && i+1 < len(os.Args):
			format = os.Args[i+1]
			i++
		}
	}

	if workflow == "" {
		fmt.Fprintf(os.Stderr, "Error: --workflow is required\n")
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Default to HEAD
	if sha == "" {
		output, err := gitexec.Run(workspaceRoot, "rev-parse", "HEAD")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get HEAD SHA: %v\n", err)
			return 1
		}
		sha = strings.TrimSpace(string(output))
	}

	// Use GitHub API
	api := github.NewGHClient(ghexec.New(workspaceRoot), workspaceRoot)

	hasRecent, err := api.HasRecentSuccess(workflow, sha, since)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if format == "shell" {
		fmt.Printf("HAS_RECENT=\"%t\"\n", hasRecent)
	} else {
		fmt.Println(hasRecent)
	}

	return 0
}
