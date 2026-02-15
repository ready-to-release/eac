package get

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/environments"
)

type getFilesByModuleCommand struct{}

var _ core.SimpleCommandPort = (*getFilesByModuleCommand)(nil)

func (c *getFilesByModuleCommand) Name() string { return "get files-by-module" }

func (c *getFilesByModuleCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-files-by-module",
		Short:         "Get files owned by a module from cached analysis",
		Long:          "Returns files that belong to a specific module from the FILES_BY_MODULE JSON.\n\nThis command parses the FILES_BY_MODULE JSON (from get changed-modules-ci)\nand returns files for a specific module in various formats.\n\nInput: FILES_BY_MODULE environment variable or --json flag\n\nOutput formats:\n  --count: Just the count of files\n  --list: One file per line (default)\n  --format shell: COUNT=\"N\" FILES=\"f1 f2 f3\"\n\nExample:\n  get files-by-module docs                    # List files\n  get files-by-module docs --count            # Just count\n  get files-by-module docs --json \"$JSON\"     # From explicit JSON\n  eval $(get files-by-module docs --format shell)",
		Flags: []core.FlagSpec{
			{Name: "count", Type: "bool", DefaultValue: "", Usage: "Output only the file count"},
			{Name: "json", Type: "string", DefaultValue: "", Usage: "FILES_BY_MODULE JSON (defaults to env var)"},
			{Name: "format", Type: "string", DefaultValue: "", Usage: "Output format (list, shell)"},
		},
	}
}

func (c *getFilesByModuleCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetFilesByModule()
}

func GetFilesByModule() int {
	// Parse arguments
	module := ""
	countOnly := false
	jsonInput := ""
	format := "list"

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--count":
			countOnly = true
		case arg == "--json" && i+1 < len(os.Args):
			jsonInput = os.Args[i+1]
			i++
		case arg == "--format" && i+1 < len(os.Args):
			format = os.Args[i+1]
			i++
		case !strings.HasPrefix(arg, "--") && module == "":
			module = arg
		}
	}

	if module == "" {
		fmt.Fprintln(os.Stderr, "Error: module name required")
		fmt.Fprintln(os.Stderr, "Usage: get files-by-module <module> [--count] [--json <json>] [--format shell]")
		return 1
	}

	// Get JSON from flag or environment
	if jsonInput == "" {
		jsonInput = os.Getenv(environments.EnvFilesByModule)
	}

	if jsonInput == "" {
		// No files data - return empty
		if countOnly || format == "shell" {
			if format == "shell" {
				fmt.Printf("COUNT=\"0\"\n")
				fmt.Printf("FILES=\"\"\n")
			} else {
				fmt.Println("0")
			}
		}
		return 0
	}

	// Parse JSON: map[module][]files
	var filesByModule map[string][]string
	if err := json.Unmarshal([]byte(jsonInput), &filesByModule); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse JSON: %v\n", err)
		return 1
	}

	files := filesByModule[module]
	if files == nil {
		files = []string{}
	}

	// Output based on format
	switch {
	case format == "shell":
		fmt.Printf("COUNT=\"%d\"\n", len(files))
		fmt.Printf("FILES=\"%s\"\n", strings.Join(files, " "))
	case countOnly:
		fmt.Println(len(files))
	default:
		for _, f := range files {
			fmt.Println(f)
		}
	}

	return 0
}
