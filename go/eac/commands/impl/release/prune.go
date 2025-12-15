// Command: release prune
// Short: Remove old pre-releases and their tags, keeping only the newest N
// Long: Prunes old pre-releases for a module, keeping only the specified number of newest releases.
// Long:
// Long: This command:
// Long:   - Lists all releases matching the module prefix (e.g., docs/*)
// Long:   - Keeps the newest N releases (default: 3)
// Long:   - Deletes older releases and their associated git tags
// Long:
// Long: Use this after successful releases to maintain a clean release history.
// Long:
// Long: Expected Output:
// Long:   - Count of releases found
// Long:   - List of releases deleted (if any)
// Long:   - Summary of cleanup results
// Long:
// Long: Example:
// Long:   release prune docs              # Keep 3 newest docs/* releases
// Long:   release prune books --keep 5    # Keep 5 newest books/* releases
// Long:   release prune --all --keep 3    # Prune all modules
// Flag.keep: type=int, default=3, usage=Number of releases to keep per module
// Flag.all: type=bool, usage=Prune all modules with releases
// Flag.dry-run: type=bool, usage=Show what would be deleted without deleting
package release

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ReleasePrune)
}

func ReleasePrune() int {
	// Parse flags
	keepCount := 3
	dryRun := false
	pruneAll := false
	module := ""

	args := os.Args[3:] // Skip: program, "release", "prune"
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--keep":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &keepCount)
				i++
			}
		case "--dry-run":
			dryRun = true
		case "--all":
			pruneAll = true
		case "--help", "-h":
			printPruneHelp()
			return 0
		default:
			if !strings.HasPrefix(arg, "-") && module == "" {
				module = arg
			}
		}
	}

	if module == "" && !pruneAll {
		log.Errorf("Error: module name required (or use --all)")
		printPruneHelp()
		return 1
	}

	if keepCount < 1 {
		log.Errorf("Error: --keep must be at least 1")
		return 1
	}

	// Get workspace root for module discovery if --all
	var modules []string
	if pruneAll {
		workspaceRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			log.Errorf("Failed to find repository root: %v", err)
			return 1
		}
		modules, err = getModulesWithReleases(workspaceRoot)
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
	// Use gh release list and filter by module prefix
	cmd := exec.Command("gh", "release", "list", "--limit", "100")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var releases []string
	prefix := module + "/"
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// First column is the tag name
		parts := strings.Fields(line)
		if len(parts) > 0 && strings.HasPrefix(parts[0], prefix) {
			releases = append(releases, parts[0])
		}
	}

	return releases, nil
}

func deleteReleaseAndTag(tagName string) error {
	// Delete the release first
	cmd := exec.Command("gh", "release", "delete", tagName, "--yes")
	if output, err := cmd.CombinedOutput(); err != nil {
		// Release might not exist, continue to delete tag
		log.Debugf("Release delete output: %s", string(output))
	}

	// Delete the tag via API (URL-encode the slash)
	repoCmd := exec.Command("gh", "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner")
	repoOutput, err := repoCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}
	repo := strings.TrimSpace(string(repoOutput))

	// URL-encode the tag name (/ becomes %2F)
	encodedTag := url.PathEscape(tagName)
	apiPath := fmt.Sprintf("repos/%s/git/refs/tags/%s", repo, encodedTag)

	cmd = exec.Command("gh", "api", "--method", "DELETE", apiPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete tag: %s", strings.TrimSpace(string(output)))
	}

	return nil
}

func getModulesWithReleases(workspaceRoot string) ([]string, error) {
	// Get all releases and extract unique module prefixes
	cmd := exec.Command("gh", "release", "list", "--limit", "100")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	moduleSet := make(map[string]bool)
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) > 0 {
			tag := parts[0]
			if idx := strings.Index(tag, "/"); idx > 0 {
				moduleSet[tag[:idx]] = true
			}
		}
	}

	var modules []string
	for mod := range moduleSet {
		modules = append(modules, mod)
	}
	return modules, nil
}

func printPruneHelp() {
	log.Info("Remove old pre-releases and their tags, keeping only the newest N")
	log.Info("")
	log.Info("Usage: release prune <module> [flags]")
	log.Info("       release prune --all [flags]")
	log.Info("")
	log.Info("Arguments:")
	log.Info("  <module>     Module name (e.g., docs, books, r2r-cli)")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --keep N     Number of releases to keep (default: 3)")
	log.Info("  --all        Prune all modules with releases")
	log.Info("  --dry-run    Show what would be deleted without deleting")
	log.Info("")
	log.Info("Examples:")
	log.Info("  release prune docs              # Keep 3 newest docs/* releases")
	log.Info("  release prune books --keep 5    # Keep 5 newest books/* releases")
	log.Info("  release prune --all --keep 3    # Prune all modules")
	log.Info("  release prune docs --dry-run    # Preview what would be deleted")
}
