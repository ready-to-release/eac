// Command: get ci-workflows
// Short: Get list of CI workflow modules
// Long: Discovers all CI workflows (ci-*.yaml) and returns module names.
// Long:
// Long: This replaces the bash pattern:
// Long:   for workflow in .github/workflows/ci-*.yaml; do
// Long:     module=$(basename "$workflow" .yaml | sed 's/^ci-//')
// Long:   done
// Long:
// Long: Output formats:
// Long:   --format space: "mod1 mod2 mod3" (default)
// Long:   --format list: One module per line
// Long:   --format json: ["mod1", "mod2", "mod3"]
// Long:
// Long: Example:
// Long:   get ci-workflows                    # Space-separated
// Long:   get ci-workflows --format list      # One per line
// Long:   get ci-workflows --format json      # JSON array
// Flag.format: type=string, usage=Output format (space, list, json)
package get

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetCIWorkflows)
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

	// Output based on format
	switch format {
	case "list":
		for _, m := range modules {
			fmt.Println(m)
		}
	case "json":
		output, _ := json.Marshal(modules)
		fmt.Println(string(output))
	default: // "space"
		fmt.Println(strings.Join(modules, " "))
	}

	return 0
}
