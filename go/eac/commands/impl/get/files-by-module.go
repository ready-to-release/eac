// Command: get files-by-module
// Short: Get files owned by a module from cached analysis
// Long: Returns files that belong to a specific module from the FILES_BY_MODULE JSON.
// Long:
// Long: This command parses the FILES_BY_MODULE JSON (from get changed-modules-ci)
// Long: and returns files for a specific module in various formats.
// Long:
// Long: Input: FILES_BY_MODULE environment variable or --json flag
// Long:
// Long: Output formats:
// Long:   --count: Just the count of files
// Long:   --list: One file per line (default)
// Long:   --format shell: COUNT="N" FILES="f1 f2 f3"
// Long:
// Long: Example:
// Long:   get files-by-module docs                    # List files
// Long:   get files-by-module docs --count            # Just count
// Long:   get files-by-module docs --json "$JSON"     # From explicit JSON
// Long:   eval $(get files-by-module docs --format shell)
// Flag.count: type=bool, usage=Output only the file count
// Flag.json: type=string, usage=FILES_BY_MODULE JSON (defaults to env var)
// Flag.format: type=string, usage=Output format (list, shell)
package get

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(GetFilesByModule)
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
		jsonInput = os.Getenv("FILES_BY_MODULE")
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
