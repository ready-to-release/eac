package pipeline

import (
	"context"
	"os"
	"strconv"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/get"
	"github.com/ready-to-release/eac/go/core/repository"
)

type pipelineAwaitReleaseCommand struct{}

var _ core.SimpleCommandPort = (*pipelineAwaitReleaseCommand)(nil)

func (c *pipelineAwaitReleaseCommand) Name() string { return "pipeline await-release" }

func (c *pipelineAwaitReleaseCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "pipeline-await-release",
		Short:         "Wait for release workflows to complete for a specific commit",
		Long:          "Wait for release workflows (release-*.yaml) to complete for a specific commit SHA.\n\nThis command polls GitHub Actions for in_progress or queued release workflow\nruns that match the specified SHA and waits until all complete.\n\nSHA Detection (in order of precedence):\n  1. --sha flag (explicit override)\n  2. GITHUB_SHA environment variable (GitHub Actions)\n  3. origin/main HEAD after fetch (local devbox)\n\nExpected Output:\n  - Live progress display showing active workflow count\n  - Exit code 0 when all workflows complete successfully\n  - Exit code 1 on timeout or failure\n\nExample:\n  pipeline await-release                              # Auto-detect SHA\n  pipeline await-release --sha abc123                 # Explicit SHA\n  pipeline await-release --timeout 300                # 5 minute timeout\n  pipeline await-release --exclude clie-eac-bundle     # Exclude bundle workflow",
		Flags: []core.FlagSpec{
			{Name: "sha", Type: "string", Usage: "Commit SHA to filter runs (auto-detected if not provided)"},
			{Name: "timeout", Type: "int", DefaultValue: "600", Usage: "Maximum wait time in seconds (default: 600)"},
			{Name: "interval", Type: "int", DefaultValue: "30", Usage: "Poll interval in seconds (default: 30)"},
			{Name: "pattern", Type: "string", DefaultValue: "release-*.yaml", Usage: "Workflow file pattern to match"},
			{Name: "exclude", Type: "string", Usage: "Workflow name substring to exclude (e.g., clie-eac-bundle)"},
		},
	}
}

func (c *pipelineAwaitReleaseCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return PipelineAwaitRelease()
}

func PipelineAwaitRelease() int {
	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	// Parse flags
	sha := ""
	timeout := 600 // 10 minutes default
	interval := 30 // 30 seconds default
	pattern := "release-*.yaml"
	exclude := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--sha" && i+1 < len(os.Args):
			sha = os.Args[i+1]
			i++
		case arg == "--timeout" && i+1 < len(os.Args):
			if v, err := strconv.Atoi(os.Args[i+1]); err == nil {
				timeout = v
			}
			i++
		case arg == "--interval" && i+1 < len(os.Args):
			if v, err := strconv.Atoi(os.Args[i+1]); err == nil {
				interval = v
			}
			i++
		case arg == "--pattern" && i+1 < len(os.Args):
			pattern = os.Args[i+1]
			i++
		case arg == "--exclude" && i+1 < len(os.Args):
			exclude = os.Args[i+1]
			i++
		}
	}

	// Auto-detect SHA using shared detection logic
	result, err := get.DetectCurrentSHA(workspaceRoot, sha)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	sha = result.SHA

	return awaitWorkflows(workspaceRoot, pattern, exclude, sha, timeout, interval, "release")
}
