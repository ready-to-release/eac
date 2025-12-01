// Command: release tag-pending
// Short: Check for changelog versions without corresponding git tags
// Long: Scans changelog files for version entries and checks if the corresponding
// Long: git tag exists. Returns versions that need tagging.
// Long:
// Long: This is used by CI to detect merged releases that need tags created.
// Long:
// Long: Examples:
// Long:   release tag-pending src-cli        # Check single module
// Long:   release tag-pending --all          # Check all modules with changelogs
// Flag.all: type=bool, usage=Check all modules with changelogs
// Args: modules
package release

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/changelog"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/git"
)

func init() {
	registry.Register(ReleaseTagPending)
}

// TagPendingResult contains info about a version needing a tag
type TagPendingResult struct {
	Module  string `json:"module"`
	Version string `json:"version"`
	Tag     string `json:"tag"`
	NeedsTag bool   `json:"needs_tag"`
}

// TagPendingReport contains results for multiple modules
type TagPendingReport struct {
	Results     []TagPendingResult `json:"results"`
	HasPending  bool               `json:"has_pending"`
}

func ReleaseTagPending() int {
	// Parse flags
	module := ""
	checkAll := false

	args := os.Args[3:] // Skip binary, "release", "tag-pending"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--all":
			checkAll = true
		default:
			if !strings.HasPrefix(arg, "--") && module == "" {
				module = arg
			}
		}
	}

	if !checkAll && module == "" {
		fmt.Fprintln(os.Stderr, "Error: module moniker required (or use --all)")
		fmt.Fprintln(os.Stderr, "Usage: release tag-pending <module>")
		fmt.Fprintln(os.Stderr, "       release tag-pending --all")
		return 1
	}

	// Load workspace root
	workspaceRoot, err := registry.GetWorkspaceRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get workspace root: %v\n", err)
		return 1
	}

	// Open git repository
	repo, err := git.Open("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to open git repository: %v\n", err)
		return 1
	}

	// Load module contracts
	moduleRegistry, err := modules.LoadFromWorkspace("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load modules: %v\n", err)
		return 1
	}

	// Determine which modules to check
	var modulesToCheck []string
	if checkAll {
		modulesToCheck = findModulesWithChangelogs(workspaceRoot)
	} else {
		modulesToCheck = []string{module}
	}

	// Check each module
	var allResults []TagPendingResult
	for _, mod := range modulesToCheck {
		moduleContract, exists := moduleRegistry.Get(mod)
		if !exists {
			fmt.Fprintf(os.Stderr, "Warning: module '%s' not found\n", mod)
			continue
		}
		result, err := checkTagPending(mod, moduleContract, repo, workspaceRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to check module '%s': %v\n", mod, err)
			continue
		}
		allResults = append(allResults, result)
	}

	// JSON output - single module returns single result, multiple returns report
	var output interface{}
	if !checkAll && len(allResults) == 1 {
		// Single module query - return just that result
		output = allResults[0]
	} else {
		// Multiple modules - return report with only pending ones
		report := TagPendingReport{
			Results:    make([]TagPendingResult, 0),
			HasPending: false,
		}
		for _, r := range allResults {
			if r.NeedsTag {
				report.Results = append(report.Results, r)
				report.HasPending = true
			}
		}
		output = report
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal JSON: %v\n", err)
		return 1
	}
	fmt.Println(string(jsonBytes))

	return 0
}

func checkTagPending(module string, moduleContract *modules.ModuleContract, repo git.GitRepository, workspaceRoot string) (TagPendingResult, error) {
	result := TagPendingResult{
		Module:   module,
		NeedsTag: false,
	}

	// Parse changelog using path from module contract
	changelogPath := filepath.Join(workspaceRoot, moduleContract.GetChangelogPath())
	cl, err := changelog.Parse(changelogPath)
	if err != nil {
		return result, fmt.Errorf("failed to parse changelog: %v", err)
	}

	// Get latest version from changelog
	if len(cl.Versions) == 0 {
		return result, nil
	}

	latestVersion := cl.Versions[0].Number
	result.Version = latestVersion
	result.Tag = fmt.Sprintf("%s/%s", module, latestVersion)

	// Check if tag exists
	exists, err := repo.TagExists(result.Tag)
	if err != nil {
		return result, fmt.Errorf("failed to check tag: %v", err)
	}

	result.NeedsTag = !exists
	return result, nil
}
