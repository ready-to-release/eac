// Command: get changed modules
// Description: Get modules affected by changed files
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
//   --base <ref>: Base ref to compare against (default: HEAD)
package get

import (
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/src/commands/impl/get/internal"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(GetChangedModules)
}

func GetChangedModules() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Get base ref from flags
	baseRef := "HEAD"
	for i, arg := range os.Args {
		if arg == "--base" && i+1 < len(os.Args) {
			baseRef = os.Args[i+1]
			break
		}
	}

	// Get list of changed files from git
	cmd := exec.Command("git", "diff", "--name-only", baseRef)
	cmd.Dir = workspaceRoot
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting changed files: %v\n", err)
		return 1
	}

	changedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(changedFiles) == 1 && changedFiles[0] == "" {
		changedFiles = []string{}
	}

	// Use the shared get command helper
	return get.ExecuteGetCommand(func() (interface{}, error) {
		modules, err := repository.GetChangedModules(changedFiles, workspaceRoot, "0.1.0")
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
