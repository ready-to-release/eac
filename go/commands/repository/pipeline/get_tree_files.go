package pipeline

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/gh"
	"github.com/ready-to-release/eac/go/core/github"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

type pipelineGetTreeFilesCommand struct{}

var _ core.SimpleCommandPort = (*pipelineGetTreeFilesCommand)(nil)

func (c *pipelineGetTreeFilesCommand) Name() string { return "pipeline get-tree-files" }

func (c *pipelineGetTreeFilesCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "pipeline-get-tree-files",
		Short:         "Get repository file list from GitHub Trees API",
		Long:          "Fetches the list of all files in the repository at a given SHA using\nthe GitHub Trees API. This is much faster than git ls-files for large repos.\n\nOutput: One file path per line\n\nExample:\n  pipeline get-tree-files --sha abc123\n  pipeline get-tree-files  # uses HEAD",
	}
}

func (c *pipelineGetTreeFilesCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return PipelineGetTreeFiles()
}

func PipelineGetTreeFiles() int {
	// Parse flags
	sha := ""
	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--sha" && i+1 < len(os.Args) {
			sha = os.Args[i+1]
			i++
		}
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Default to HEAD
	if sha == "" {
		output, err := tool.GlobalToolSystem().RunTool(context.Background(), "git", workspaceRoot, "rev-parse", "HEAD")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get HEAD SHA: %v\n", err)
			return 1
		}
		sha = strings.TrimSpace(string(output))
	}

	// Use GitHub API
	api := github.NewGHClient(gh.New(tool.GlobalToolSystem(), workspaceRoot), workspaceRoot)

	files, err := api.GetTreeFiles(sha)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get tree files: %v\n", err)
		return 1
	}

	// Output one file per line
	for _, file := range files {
		fmt.Println(file)
	}

	return 0
}
