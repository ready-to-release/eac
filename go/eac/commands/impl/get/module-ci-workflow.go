// Command: get module-ci-workflow
// Short: Get CI workflow filename for a module
// Long: Returns the CI workflow filename for the specified module.
// Long:
// Long: This replaces the jq pattern:
// Long:   echo "$MODULES_JSON" | jq -r '.[] | select(.moniker == "X") | .files.workflows.ci'
// Long:
// Long: Exit codes:
// Long:   0 - Workflow found, outputs filename (e.g., ci-r2r-cli.yaml)
// Long:   1 - Module not found or no CI workflow configured
// Long:
// Long: Example:
// Long:   get module-ci-workflow r2r-cli       # Outputs: ci-r2r-cli.yaml
// Long:   get module-ci-workflow eac-core      # Outputs: ci-eac-core.yaml
// Flag.basename: type=bool, usage=Output only the basename (default true)
package get

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetModuleCIWorkflow)
}

func GetModuleCIWorkflow() int {
	// Parse arguments
	moniker := ""
	basename := true // Default: output basename only

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--basename":
			basename = true
		case arg == "--full-path":
			basename = false
		case !strings.HasPrefix(arg, "--") && moniker == "":
			moniker = arg
		}
	}

	if moniker == "" {
		fmt.Fprintln(os.Stderr, "Error: module moniker required")
		fmt.Fprintln(os.Stderr, "Usage: get module-ci-workflow <moniker>")
		return 1
	}

	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return 1
	}

	// Find the module
	var module *config.Module
	for i := range cfg.Repository.Modules {
		if cfg.Repository.Modules[i].Moniker == moniker {
			module = &cfg.Repository.Modules[i]
			break
		}
	}

	if module == nil {
		fmt.Fprintf(os.Stderr, "Error: module '%s' not found\n", moniker)
		return 1
	}

	// Get CI workflow path
	ciWorkflow := ""
	if module.Files.Workflows.CI != "" {
		ciWorkflow = module.Files.Workflows.CI
	}

	if ciWorkflow == "" {
		// No output, exit 1 (no CI workflow configured)
		return 1
	}

	// Output the workflow
	if basename {
		fmt.Println(filepath.Base(ciWorkflow))
	} else {
		fmt.Println(ciWorkflow)
	}

	return 0
}
