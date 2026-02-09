// Command: pipeline get-tree-files
// Short: Get repository file list from GitHub Trees API
//
//	--sha <sha>: Tree SHA to fetch (default: HEAD)
//
// Long:
// Long: Fetches the list of all files in the repository at a given SHA using
// Long: the GitHub Trees API. This is much faster than git ls-files for large repos.
// Long:
// Long: Output: One file path per line
// Long:
// Long: Example:
// Long:   pipeline get-tree-files --sha abc123
// Long:   pipeline get-tree-files  # uses HEAD
package pipeline

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/ghexec"
	"github.com/ready-to-release/eac/go/clibase/gitexec"
	"github.com/ready-to-release/eac/go/core/github"
	"github.com/ready-to-release/eac/go/core/repository"
)

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
		output, err := gitexec.Run(workspaceRoot, "rev-parse", "HEAD")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get HEAD SHA: %v\n", err)
			return 1
		}
		sha = strings.TrimSpace(string(output))
	}

	// Use GitHub API
	api := github.NewGHClient(ghexec.New(workspaceRoot), workspaceRoot)

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
