// Command: get changed-modules
// Short: Get modules affected by changed files
// Flag.as-yaml: type=bool, usage=Output as YAML (default)
// Flag.as-json: type=bool, usage=Output as JSON
// Flag.as-toml: type=bool, usage=Output as TOML
// Flag.base: type=string, usage=Base ref to compare against (default: HEAD)
// Flag.from-stdin: type=bool, usage=Read file paths from stdin (one per line) instead of git diff
// Long:
// Long: Expected Output:
// Long: YAML list of module monikers that have changes based on git diff against the specified base ref,
// Long: or based on file paths read from stdin when --from-stdin is used.
// Long: Only includes modules directly containing changed files.
// Long:
// Long: Examples:
// Long:   get changed-modules                           # Use git diff HEAD
// Long:   get changed-modules --base main               # Use git diff main
// Long:   echo "path/to/file.go" | get changed-modules --from-stdin  # Read from stdin
package get

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/gitexec"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/repository"
)

func init() {
	registry.Register(GetChangedModules)
}

// changedModulesFlags defines valid flags for the get changed-modules command

func GetChangedModules() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Parse flags
	baseRef := "HEAD"
	fromStdin := false
	for i, arg := range os.Args {
		if arg == "--base" && i+1 < len(os.Args) {
			baseRef = os.Args[i+1]
		}
		if arg == "--from-stdin" {
			fromStdin = true
		}
	}

	var changedFiles []string

	if fromStdin {
		// Read file paths from stdin (one per line)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				changedFiles = append(changedFiles, line)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: reading from stdin: %v\n", err)
			return 1
		}
	} else {
		// Get list of changed files from git
		output, err := gitexec.Run(workspaceRoot, "diff", "--name-only", baseRef)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: getting changed files: %v\n", err)
			return 1
		}

		changedFiles = strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(changedFiles) == 1 && changedFiles[0] == "" {
			changedFiles = []string{}
		}
	}

	// Use the shared get command helper
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		modules, err := repository.GetChangedModules(changedFiles, workspaceRoot)
		if err != nil {
			return nil, err
		}

		// Return as struct for proper serialization
		return struct {
			Modules []string `json:"modules" yaml:"modules" toml:"modules"`
		}{
			Modules: modules,
		}, nil
	})
}
