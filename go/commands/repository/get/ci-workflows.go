package get

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/get/internal"
	"github.com/ready-to-release/eac/go/core/repository"
)

type getCIWorkflowsCommand struct{}

var _ core.SimpleCommandPort = (*getCIWorkflowsCommand)(nil)

func (c *getCIWorkflowsCommand) Name() string { return "get ci-workflows" }

func (c *getCIWorkflowsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-ci-workflows",
		Short:         "Get list of CI workflow modules",
		Long: "Discovers all CI workflows (ci-*.yaml) and returns module names.\n" +
			"\n" +
			"This replaces the bash pattern:\n" +
			"  for workflow in .github/workflows/ci-*.yaml; do\n" +
			"    module=$(basename \"$workflow\" .yaml | sed 's/^ci-//')\n" +
			"  done\n" +
			"\n" +
			"Output formats:\n" +
			"  --format space: \"mod1 mod2 mod3\" (default)\n" +
			"  --format list: One module per line\n" +
			"  --format json: [\"mod1\", \"mod2\", \"mod3\"]\n" +
			"\n" +
			"Example:\n" +
			"  get ci-workflows                    # Space-separated\n" +
			"  get ci-workflows --format list      # One per line\n" +
			"  get ci-workflows --format json      # JSON array",
		Flags: []core.FlagSpec{
			{Name: "format", Type: "string", Usage: "Output format (space, list, json)"},
		},
	}
}

func (c *getCIWorkflowsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetCIWorkflows()
}

func GetCIWorkflows() int {
	// Parse flags
	format := "space"

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--format" && i+1 < len(os.Args):
			format = os.Args[i+1]
			i++
		}
	}

	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Find all ci-*.yaml workflows
	workflowsDir := filepath.Join(workspaceRoot, ".github", "workflows")
	pattern := filepath.Join(workflowsDir, "ci-*.yaml")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Extract module names
	modules := make([]string, 0, len(matches))
	for _, match := range matches {
		base := filepath.Base(match)
		// ci-foo.yaml -> foo
		module := strings.TrimSuffix(strings.TrimPrefix(base, "ci-"), ".yaml")
		modules = append(modules, module)
	}

	sort.Strings(modules)

	// Legacy formats have special output requirements (shell scripting)
	switch format {
	case "list":
		for _, m := range modules {
			fmt.Println(m)
		}
		return 0
	case "space":
		fmt.Println(strings.Join(modules, " "))
		return 0
	}

	// Default: use shared helper for consistent YAML/JSON/TOML output
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return modules, nil
	})
}
