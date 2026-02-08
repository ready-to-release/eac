// Command: get token-size
// Short: Estimate token counts for source files
// Long: Estimate token counts for source files using low-cost heuristics.
// Long:
// Long: This command helps identify files that may exceed Claude's token limits
// Long: (typically 25,000 tokens). It uses a characters/4 heuristic which provides
// Long: a reasonable approximation for code files.
// Long:
// Long: Supports glob patterns for processing multiple files at once.
// Long:
// Long: Examples:
// Long:   eac get token-size main.go
// Long:   eac get token-size "go/**/*.go" --threshold 20000
// Long:   eac get token-size main.go --as-json
// Flag.threshold: type=int, usage=Only show files exceeding this token limit
// Flag.as-json: type=bool, usage=Output as JSON
// Args: files
package get

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/tokensize"
)

type tokenSizeOptions struct {
	Threshold    int
	HasThreshold bool
	AsJSON       bool
	Files        []string
}

func GetTokenSize() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	opts, err := parseTokenSizeFlags(os.Args[3:]) // Skip "eac get token-size"
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if len(opts.Files) == 0 {
		fmt.Fprintln(os.Stderr, "Error: at least one file or pattern required")
		printTokenSizeUsage()
		return 1
	}

	// Get workspace root for glob expansion
	workspaceRoot, err := registry.GetWorkspaceRoot()
	if err != nil {
		workspaceRoot, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot determine working directory: %v\n", err)
			return 1
		}
	}

	// Expand glob patterns
	files, err := tokensize.ExpandGlobPatterns(workspaceRoot, opts.Files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error expanding patterns: %v\n", err)
		return 1
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "No files matched the specified patterns")
		return 1
	}

	// Process files
	var results []*tokensize.Estimate

	for _, file := range files {
		estimate, err := tokensize.EstimateFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", file, err)
			continue
		}

		// Filter by threshold if specified
		if opts.HasThreshold && estimate.Tokens <= opts.Threshold {
			continue
		}

		// Make path relative to workspace root for cleaner output
		relPath, err := filepath.Rel(workspaceRoot, file)
		if err == nil {
			estimate.FilePath = filepath.ToSlash(relPath)
		}

		results = append(results, estimate)
	}

	// Output results
	if opts.AsJSON {
		output, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			return 1
		}
		fmt.Println(string(output))
	} else {
		for _, r := range results {
			fmt.Printf("%s: %d tokens\n", r.FilePath, r.Tokens)
		}
	}

	// Return exit code 1 if threshold was set and files exceeded it
	if opts.HasThreshold && len(results) > 0 {
		return 1
	}
	return 0
}

func parseTokenSizeFlags(args []string) (*tokenSizeOptions, error) {
	opts := &tokenSizeOptions{}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "--threshold":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--threshold requires a value")
			}
			val, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("--threshold must be an integer")
			}
			opts.Threshold = val
			opts.HasThreshold = true
			i++
		case "--as-json":
			opts.AsJSON = true
		case "--help", "-h":
			printTokenSizeUsage()
			os.Exit(0)
		default:
			if len(arg) > 0 && arg[0] != '-' {
				opts.Files = append(opts.Files, arg)
			}
		}
	}

	return opts, nil
}

func printTokenSizeUsage() {
	fmt.Println("Usage: eac get token-size <file|pattern>... [flags]")
	fmt.Println("")
	fmt.Println("Estimate token counts for source files using characters/4 heuristic.")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  --threshold <n>  Only show files exceeding this token limit")
	fmt.Println("  --as-json        Output as JSON")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  eac get token-size main.go")
	fmt.Println("  eac get token-size \"go/**/*.go\" --threshold 25000")
	fmt.Println("  eac get token-size main.go --as-json")
}
