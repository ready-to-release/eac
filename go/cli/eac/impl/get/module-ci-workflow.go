package get

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/repository"
)

type getModuleCIWorkflowCommand struct{}

var _ core.SimpleCommandPort = (*getModuleCIWorkflowCommand)(nil)

func (c *getModuleCIWorkflowCommand) Name() string { return "get module-ci-workflow" }

func (c *getModuleCIWorkflowCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-module-ci-workflow",
		Short:         "Get CI workflow filename for a module",
		Long: "Returns the CI workflow filename for the specified module.\n" +
			"\n" +
			"This replaces the jq pattern:\n" +
			"  echo \"$MODULES_JSON\" | jq -r '.[] | select(.moniker == \"X\") | .files.workflows.ci'\n" +
			"\n" +
			"Exit codes:\n" +
			"  0 - Workflow found, outputs filename (e.g., ci-clie.yaml)\n" +
			"  1 - Module not found or no CI workflow configured\n" +
			"\n" +
			"Example:\n" +
			"  get module-ci-workflow clie       # Outputs: ci-clie.yaml\n" +
			"  get module-ci-workflow core      # Outputs: ci-core.yaml",
		Flags: []core.FlagSpec{
			{Name: "basename", Type: "bool", Usage: "Output only the basename (default true)"},
		},
	}
}

func (c *getModuleCIWorkflowCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetModuleCIWorkflow()
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

	// Derive CI workflow path from convention
	// Check if workflows package is configured with explicit patterns
	ciWorkflow := ""
	if pkg, ok := module.Components["workflows"]; ok && pkg != nil {
		root := pkg.Root
		if root == "" {
			root = ".github/workflows"
		}
		// Check if patterns specify CI workflow
		if pkg.Patterns != nil {
			for _, pattern := range pkg.Patterns.Source {
				if strings.Contains(pattern, "ci-") {
					ciWorkflow = filepath.Join(root, strings.TrimPrefix(pattern, "**"))
					break
				}
			}
		}
		// Use convention if no explicit pattern
		if ciWorkflow == "" {
			ciWorkflow = filepath.Join(root, "ci-"+moniker+".yaml")
		}
	} else {
		// Default convention - check if file exists
		ciWorkflow = filepath.Join(".github", "workflows", "ci-"+moniker+".yaml")
	}

	// Check if the workflow file actually exists
	workflowPath := filepath.Join(workspaceRoot, ciWorkflow)
	if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
		// No CI workflow file exists, exit 1
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
