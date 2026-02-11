package release

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/changelog"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/git"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
)

type releaseTagPendingCommand struct{}

var _ core.SimpleCommandPort = (*releaseTagPendingCommand)(nil)

func (c *releaseTagPendingCommand) Name() string { return "release tag-pending" }

func (c *releaseTagPendingCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "release-tag-pending",
		Short:         "Check for changelog versions without corresponding git tags",
		Long:          "Scans changelog files for version entries and checks if the corresponding\ngit tag exists. Returns versions that need tagging.\n\nThis is used by CI to detect merged releases that need tags created.\n\nExpected Output:\n  - JSON list of versions needing tags, including module, version, tag, and needs_tag fields\n\nExamples:\n  release tag-pending clie        # Check single module\n  release tag-pending --all          # Check all modules with changelogs",
		Args:          "modules",
		Flags: []core.FlagSpec{
			{Name: "all", Type: "bool", Usage: "Check all modules with changelogs"},
		},
	}
}

func (c *releaseTagPendingCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ReleaseTagPending()
}

var tagPendingLog = logging.C("release")

// TagPendingResult contains info about a version needing a tag.
type TagPendingResult struct {
	Module   string `json:"module"`
	Version  string `json:"version"`
	Tag      string `json:"tag"`
	NeedsTag bool   `json:"needs_tag"`
}

// TagPendingReport contains results for multiple modules.
type TagPendingReport struct {
	Results    []TagPendingResult `json:"results"`
	HasPending bool               `json:"has_pending"`
}

func ReleaseTagPending() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		tagPendingLog.Errorf("%v", err)
		return 1
	}

	// Parse flags
	module := ""
	checkAll := false

	args := os.Args[3:] // Skip binary, "release", "tag-pending"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--all":
			checkAll = true
		default:
			if !strings.HasPrefix(arg, "--") && module == "" {
				module = arg
			}
		}
	}

	if !checkAll && module == "" {
		tagPendingLog.Error("module moniker required (or use --all)")
		tagPendingLog.Info("Usage: release tag-pending <module>")
		tagPendingLog.Info("       release tag-pending --all")
		return 1
	}

	// Load workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		tagPendingLog.Errorf("failed to get workspace root: %v", err)
		return 1
	}

	// Open git repository
	// Open git repository
	gitMgr := git.NewManager(logging.C().Zap())
	repo, err := gitMgr.Open("")
	if err != nil {
		tagPendingLog.Errorf("failed to open git repository: %v", err)
		return 1
	}

	// Load module contracts
	moduleRegistry, err := modules.LoadFromWorkspace("")
	if err != nil {
		tagPendingLog.Errorf("failed to load modules: %v", err)
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
			tagPendingLog.Warnf("module '%s' not found", mod)
			continue
		}
		result, err := checkTagPending(mod, moduleContract, repo, workspaceRoot)
		if err != nil {
			tagPendingLog.Warnf("failed to check module '%s': %v", mod, err)
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
		tagPendingLog.Errorf("failed to marshal JSON: %v", err)
		return 1
	}
	tagPendingLog.Info(string(jsonBytes))

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
		return result, fmt.Errorf("failed to parse changelog: %w", err)
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
		return result, fmt.Errorf("failed to check tag: %w", err)
	}

	result.NeedsTag = !exists
	return result, nil
}
