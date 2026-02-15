package get

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

type getChangedModulesCommand struct{}

var _ core.SimpleCommandPort = (*getChangedModulesCommand)(nil)

func (c *getChangedModulesCommand) Name() string { return "get changed-modules" }

func (c *getChangedModulesCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-changed-modules",
		Short:         "Get modules affected by changed files",
		Long: "Expected Output:\n" +
			"YAML list of module monikers that have changes based on git diff against the specified base ref,\n" +
			"or based on file paths read from stdin when --from-stdin is used.\n" +
			"Only includes modules directly containing changed files.\n" +
			"\n" +
			"Examples:\n" +
			"  get changed-modules                           # Use git diff HEAD\n" +
			"  get changed-modules --base main               # Use git diff main\n" +
			"  echo \"path/to/file.go\" | get changed-modules --from-stdin  # Read from stdin",
		Flags: []core.FlagSpec{
			{Name: "as-yaml", Type: "bool", Usage: "Output as YAML (default)"},
			{Name: "as-json", Type: "bool", Usage: "Output as JSON"},
			{Name: "as-toml", Type: "bool", Usage: "Output as TOML"},
			{Name: "base", Type: "string", Usage: "Base ref to compare against (default: HEAD)"},
			{Name: "from-stdin", Type: "bool", Usage: "Read file paths from stdin (one per line) instead of git diff"},
		},
	}
}

func (c *getChangedModulesCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetChangedModules()
}

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
		ts := tool.GlobalToolSystem()
		output, err := ts.RunTool(context.Background(), "git", workspaceRoot, "diff", "--name-only", baseRef)
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
