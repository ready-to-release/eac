package get

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/repository/reports"
)

type getFilesCommand struct{}

var _ core.SimpleCommandPort = (*getFilesCommand)(nil)

func (c *getFilesCommand) Name() string { return "get files" }

func (c *getFilesCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-files",
		Short:         "Get repository files with their module ownership",
		Long:          "Get repository files with their module ownership in structured format.\n\nThis command analyzes the repository and maps each file to its owning module\nbased on module domain. The output can be formatted as YAML, JSON, or TOML\nfor programmatic processing.\n\nUse filters to scope the output to specific files or modules:\n  --changed-only: Only modified/unstaged files\n  --staged-only: Only staged files\n  --module: Files owned by a specific module\n  --pattern: Files matching a glob pattern\n\nExample:\n  get files --as-json\n  get files --module core\n  get files --changed-only --as-yaml\n\nExpected Output:\nList of files with owning module information. Format depends on --as-* flag:\n  - YAML (default): Structured list with file name and modules array\n  - JSON: Same structure as YAML in JSON format\n  - TOML: Same structure as YAML in TOML format\nEach entry contains the file name and list of owning module monikers.",
		Flags: []core.FlagSpec{
			{Name: "as-yaml", Type: "bool", DefaultValue: "", Usage: "Output as YAML (default format)"},
			{Name: "as-json", Type: "bool", DefaultValue: "", Usage: "Output as JSON"},
			{Name: "as-toml", Type: "bool", DefaultValue: "", Usage: "Output as TOML"},
			{Name: "changed-only", Type: "bool", DefaultValue: "", Usage: "Only show modified/unstaged files"},
			{Name: "staged-only", Type: "bool", DefaultValue: "", Usage: "Only show staged files"},
			{Name: "module", Type: "string", DefaultValue: "", Usage: "Filter to files owned by specific module moniker"},
			{Name: "pattern", Type: "string", DefaultValue: "", Usage: "Filter files matching glob pattern"},
		},
	}
}

func (c *getFilesCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetFiles()
}

// filesFlags defines valid flags for the get files command

func GetFiles() int {
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

	// Parse filter flags
	filterOpts, err := parseFilterFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Use the shared get command helper
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		// Determine if we need staged-only files
		stagedOnly := filterOpts.StagedOnly

		// Generate report for files (tracked only, no ignored, staged-only if requested)
		report, err := reports.GetFilesModulesReport(true, false, stagedOnly, workspaceRoot)
		if err != nil {
			return nil, err
		}

		// Apply filters
		filteredFiles, err := applyFilters(report.AllFiles, filterOpts, workspaceRoot)
		if err != nil {
			return nil, err
		}

		// Sort by last module in the list (if multiple modules)
		sort.Slice(filteredFiles, func(i, j int) bool {
			// Get last module for each file (or empty string if no modules)
			lastModuleI := ""
			if len(filteredFiles[i].Modules) > 0 {
				lastModuleI = filteredFiles[i].Modules[len(filteredFiles[i].Modules)-1]
			}

			lastModuleJ := ""
			if len(filteredFiles[j].Modules) > 0 {
				lastModuleJ = filteredFiles[j].Modules[len(filteredFiles[j].Modules)-1]
			}

			// Sort by last module, then by file name if modules are equal
			if lastModuleI != lastModuleJ {
				return lastModuleI < lastModuleJ
			}
			return filteredFiles[i].Name < filteredFiles[j].Name
		})

		return filteredFiles, nil
	})
}

// filterOptions holds parsed filter flags.
type filterOptions struct {
	ChangedOnly bool
	StagedOnly  bool
	Module      string
	Pattern     string
}

// parseFilterFlags extracts filter flags from command arguments.
func parseFilterFlags(args []string) (*filterOptions, error) {
	opts := &filterOptions{}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "--changed-only":
			opts.ChangedOnly = true
		case "--staged-only":
			opts.StagedOnly = true
		case "--module":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--module requires a value")
			}
			opts.Module = args[i+1]
			i++ // Skip next arg
		case "--pattern":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--pattern requires a value")
			}
			opts.Pattern = args[i+1]
			i++ // Skip next arg
		}
	}

	// Validate: --changed-only and --staged-only are mutually exclusive
	if opts.ChangedOnly && opts.StagedOnly {
		return nil, fmt.Errorf("--changed-only and --staged-only are mutually exclusive")
	}

	return opts, nil
}

// applyFilters applies all active filters to the file list.
func applyFilters(files []repository.RepositoryFileWithModule, opts *filterOptions, workspaceRoot string) ([]repository.RepositoryFileWithModule, error) {
	result := files

	// Filter by changed files
	if opts.ChangedOnly {
		changedFiles, err := getChangedFiles(workspaceRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to get changed files: %w", err)
		}

		changedMap := make(map[string]bool)
		for _, f := range changedFiles {
			changedMap[f] = true
		}

		filtered := make([]repository.RepositoryFileWithModule, 0)
		for _, file := range result {
			if changedMap[file.Name] {
				filtered = append(filtered, file)
			}
		}
		result = filtered
	}

	// Filter by module
	if opts.Module != "" {
		filtered := make([]repository.RepositoryFileWithModule, 0)
		for _, file := range result {
			for _, module := range file.Modules {
				if module == opts.Module {
					filtered = append(filtered, file)
					break
				}
			}
		}
		result = filtered
	}

	// Filter by pattern
	if opts.Pattern != "" {
		// Normalize pattern to use forward slashes
		pattern := strings.ReplaceAll(opts.Pattern, "\\", "/")

		filtered := make([]repository.RepositoryFileWithModule, 0)
		for _, file := range result {
			// Normalize file path to forward slashes
			normalizedPath := strings.ReplaceAll(file.Name, "\\", "/")
			matched, err := doublestar.Match(pattern, normalizedPath)
			if err != nil {
				return nil, fmt.Errorf("invalid pattern %q: %w", opts.Pattern, err)
			}
			if matched {
				filtered = append(filtered, file)
			}
		}
		result = filtered
	}

	return result, nil
}

// getChangedFiles returns list of modified/unstaged files from git.
func getChangedFiles(workspaceRoot string) ([]string, error) {
	ts := tool.GlobalToolSystem()
	output, err := ts.RunTool(context.Background(), "git", workspaceRoot, "diff", "--name-only", "HEAD")
	if err != nil {
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}

	return files, nil
}
