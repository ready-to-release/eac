package release

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/tool"
)

type releasePruneCommand struct{}

var _ core.SimpleCommandPort = (*releasePruneCommand)(nil)

func (c *releasePruneCommand) Name() string { return "release prune" }

func (c *releasePruneCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "release-prune",
		Short:         "Remove old pre-releases and their tags, keeping only the newest N",
		Long:          "Prunes old pre-releases for a module, keeping only the specified number of newest releases.\n\nThis command:\n  - Lists all releases matching the module prefix (e.g., docs/*)\n  - Keeps the newest N releases (default: 3)\n  - Deletes older releases and their associated git tags\n\nUse this after successful releases to maintain a clean release history.\n\nExpected Output:\n  - Count of releases found\n  - List of releases deleted (if any)\n  - Summary of cleanup results\n\nExample:\n  release prune docs              # Keep 3 newest docs/* releases\n  release prune books --keep 5    # Keep 5 newest books/* releases\n  release prune --all --keep 3    # Prune all modules",
		Flags: []core.FlagSpec{
			{Name: "keep", Type: "int", DefaultValue: "1", Usage: "Number of releases to keep per module"},
			{Name: "all", Type: "bool", Usage: "Prune all modules with releases"},
			{Name: "dry-run", Type: "bool", Usage: "Show what would be deleted without deleting"},
		},
	}
}

func (c *releasePruneCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ReleasePrune()
}

func ReleasePrune() int {
	s, exitCode := newReleaseScaffold()
	if s == nil {
		return exitCode
	}

	// Parse flags
	keepCount := 1
	dryRun := false
	pruneAll := false
	module := ""

	args := os.Args[3:] // Skip: program, "release", "prune"
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--keep":
			if i+1 < len(args) {
				_, _ = fmt.Sscanf(args[i+1], "%d", &keepCount) //nolint:errcheck // default value on parse error
				i++
			}
		case "--dry-run":
			dryRun = true
		case "--all":
			pruneAll = true
		default:
			if !strings.HasPrefix(arg, "-") && module == "" {
				module = arg
			}
		}
	}

	if module == "" && !pruneAll {
		log.Errorf("Error: module name required (or use --all)")
		return 1
	}

	if keepCount < 1 {
		log.Errorf("Error: --keep must be at least 1")
		return 1
	}

	// Get workspace root for module discovery if --all
	var modules []string
	if pruneAll {
		var err error
		modules, err = getModulesWithReleases(s.WorkspaceRoot)
		if err != nil {
			log.Errorf("Failed to get modules: %v", err)
			return 1
		}
	} else {
		modules = []string{module}
	}

	totalDeleted := 0
	for _, mod := range modules {
		deleted, err := pruneModule(mod, keepCount, dryRun)
		if err != nil {
			log.Errorf("Error pruning %s: %v", mod, err)
			continue
		}
		totalDeleted += deleted
	}

	if dryRun {
		log.Infof("")
		log.Infof("Dry run complete. Would delete %d releases.", totalDeleted)
	} else {
		log.Infof("")
		log.Infof("Cleanup complete. Deleted %d old releases.", totalDeleted)
	}

	return 0
}

func pruneModule(module string, keepCount int, dryRun bool) (int, error) {
	log.Infof("Pruning old releases for: %s (keeping %d newest)", module, keepCount)

	// List releases for this module
	releases, err := listModuleReleases(module)
	if err != nil {
		return 0, fmt.Errorf("failed to list releases: %w", err)
	}

	if len(releases) == 0 {
		log.Infof("  No releases found for %s", module)
		return 0, nil
	}

	log.Infof("  Found %d releases", len(releases))

	if len(releases) <= keepCount {
		log.Infof("  Nothing to prune (only %d releases, keeping %d)", len(releases), keepCount)
		return 0, nil
	}

	// Releases are sorted newest first, so skip first keepCount
	toDelete := releases[keepCount:]
	log.Infof("  Will delete %d old releases", len(toDelete))

	deleted := 0
	for _, tag := range toDelete {
		if dryRun {
			log.Infof("  [dry-run] Would delete: %s", tag)
			deleted++
		} else {
			log.Infof("  Deleting: %s", tag)
			if err := deleteReleaseAndTag(tag); err != nil {
				log.Errorf("    Warning: %v", err)
			} else {
				deleted++
			}
		}
	}

	return deleted, nil
}

func listModuleReleases(module string) ([]string, error) {
	// Use gh release list with JSON output to correctly parse tag names
	output, err := tool.GlobalToolSystem().RunTool(context.Background(), "gh", ".", "release", "list", "--limit", "100", "--json", "tagName", "-q", ".[].tagName")
	if err != nil {
		return nil, err
	}

	var releases []string
	prefix := module + "/"
	for _, line := range strings.Split(string(output), "\n") {
		tag := strings.TrimSpace(line)
		if tag == "" {
			continue
		}
		if strings.HasPrefix(tag, prefix) {
			releases = append(releases, tag)
		}
	}

	return releases, nil
}

func deleteReleaseAndTag(tagName string) error {
	// Delete the release first
	output, _, err := tool.GlobalToolSystem().RunToolCombined(context.Background(), "gh", ".", "release", "delete", tagName, "--yes")
	if err != nil {
		// Release might not exist, continue to delete tag
		log.Debugf("Release delete output: %s", string(output))
	}

	// Delete the tag via API (URL-encode the slash)
	repoOutput, err := tool.GlobalToolSystem().RunTool(context.Background(), "gh", ".", "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner")
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}
	repo := strings.TrimSpace(string(repoOutput))

	// URL-encode the tag name (/ becomes %2F)
	encodedTag := url.PathEscape(tagName)
	apiPath := fmt.Sprintf("repos/%s/git/refs/tags/%s", repo, encodedTag)

	output, exitCode, err := tool.GlobalToolSystem().RunToolCombined(context.Background(), "gh", ".", "api", "--method", "DELETE", apiPath)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("failed to delete tag: %s", strings.TrimSpace(string(output)))
	}

	return nil
}

func getModulesWithReleases(workspaceRoot string) ([]string, error) {
	// Get all releases and extract unique module prefixes using JSON output
	output, err := tool.GlobalToolSystem().RunTool(context.Background(), "gh", ".", "release", "list", "--limit", "100", "--json", "tagName", "-q", ".[].tagName")
	if err != nil {
		return nil, err
	}

	moduleSet := make(map[string]bool)
	for _, line := range strings.Split(string(output), "\n") {
		tag := strings.TrimSpace(line)
		if tag == "" {
			continue
		}
		if idx := strings.Index(tag, "/"); idx > 0 {
			moduleSet[tag[:idx]] = true
		}
	}

	var modules []string
	for mod := range moduleSet {
		modules = append(modules, mod)
	}
	return modules, nil
}
