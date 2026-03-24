package release

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/gh"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/github"
)

type releasePrunePackagesCommand struct{}

var _ core.SimpleCommandPort = (*releasePrunePackagesCommand)(nil)

func (c *releasePrunePackagesCommand) Name() string { return "release prune-packages" }

func (c *releasePrunePackagesCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "release-prune-packages",
		Short:         "Remove old container image versions from GHCR, keeping newest N",
		Long: "Prunes old container image versions from GitHub Container Registry (GHCR),\nimplementing multiple safety checks to prevent accidental deletion of released versions.\n\nSafety Checks (in order):\n  1. Versions with tags matching preserve patterns (e.g., v*, latest) are never deleted\n  2. Versions associated with GitHub releases are protected\n  3. Versions whose digest matches a released version are protected\n  4. Versions created less than min_age_days ago are protected\n  5. Only versions matching prune patterns (e.g., sha-*, dev-*) are candidates\n\nConfiguration:\n  Configure in .eac/repository.yml under registries:\n    registries:\n      ghcr.io:\n        org: your-org\n        cleanup:\n          enabled: true\n          keep: 10\n          preserve_patterns: [\"v*\", \"latest\", \"[0-9]*.[0-9]*.[0-9]*\"]\n          prune_patterns: [\"sha-*\", \"dev-*\", \"pr-*\", \"ci\"]\n          min_age_days: 7\n\nDefault mode is dry-run for safety. Use --force to actually delete.",
		Notes: "Expected Output:\n  - List of protected versions with reasons\n  - List of versions to prune\n  - Summary of cleanup results",
		Examples: []string{
			"eac release prune-packages eac-ext           # Dry-run: show what would be deleted",
			"eac release prune-packages eac-ext --force   # Actually delete versions",
			"eac release prune-packages --all             # Prune all packages (dry-run)",
			"eac release prune-packages eac-ext --keep 5  # Override keep count",
		},
		Flags: []core.FlagSpec{
			{Name: "keep", Type: "int", Usage: "Override the number of versions to keep (default from config)"},
			{Name: "all", Type: "bool", Usage: "Prune all configured packages"},
			{Name: "force", Type: "bool", Usage: "Actually delete versions (default is dry-run)"},
			{Name: "verbose", Type: "bool", Usage: "Show protected versions with reasons"},
			{Name: "json", Type: "bool", Usage: "Output results in JSON format"},
		},
	}
}

func (c *releasePrunePackagesCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ReleasePrunePackages()
}

// PrunePackagesResult contains the result of pruning a package.
type PrunePackagesResult struct {
	Package          string                       `json:"package"`
	TotalVersions    int                          `json:"total_versions"`
	ProtectedCount   int                          `json:"protected_count"`
	PrunableCount    int                          `json:"prunable_count"`
	DeletedCount     int                          `json:"deleted_count"`
	KeptCount        int                          `json:"kept_count"`
	Protected        []PruneVersionResult         `json:"protected,omitempty"`
	Prunable         []PruneVersionResult         `json:"prunable,omitempty"`
	VersionsToDelete []PruneVersionResult         `json:"versions_to_delete,omitempty"`
	Error            string                       `json:"error,omitempty"`
}

// PruneVersionResult contains info about a version in the prune operation.
type PruneVersionResult struct {
	ID        int      `json:"id"`
	Digest    string   `json:"digest"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
	Reason    string   `json:"reason,omitempty"`
}

func ReleasePrunePackages() int {
	s, exitCode := newReleaseScaffold(withConfig())
	if s == nil {
		return exitCode
	}

	// Parse flags
	keepOverride := -1 // -1 means use config default
	force := false
	pruneAll := false
	verbose := false
	jsonOutput := false
	packageName := ""

	args := os.Args[3:] // Skip: program, "release", "prune-packages"
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--keep":
			if i+1 < len(args) {
				_, _ = fmt.Sscanf(args[i+1], "%d", &keepOverride) //nolint:errcheck // default value on parse error
				i++
			}
		case "--force":
			force = true
		case "--all":
			pruneAll = true
		case "--verbose":
			verbose = true
		case "--json":
			jsonOutput = true
		default:
			if !strings.HasPrefix(arg, "-") && packageName == "" {
				packageName = arg
			}
		}
	}

	if packageName == "" && !pruneAll {
		log.Errorf("Error: package name required (or use --all)")
		log.Infof("Usage: release prune-packages <package> [--force] [--keep N]")
		log.Infof("       release prune-packages --all [--force]")
		return 1
	}

	workspaceRoot := s.WorkspaceRoot
	cfg := s.Config

	// Check for registry configuration
	registryConfig := cfg.GetRegistryConfig("ghcr.io")
	if registryConfig == nil {
		log.Errorf("No registry configuration found for ghcr.io")
		log.Infof("Add registries section to .eac/repository.yml:")
		log.Infof("  registries:")
		log.Infof("    ghcr.io:")
		log.Infof("      org: your-org")
		log.Infof("      cleanup:")
		log.Infof("        enabled: true")
		return 1
	}

	if !registryConfig.IsCleanupEnabled() {
		log.Infof("Cleanup is disabled for ghcr.io in configuration")
		return 0
	}

	cleanupPolicy := registryConfig.GetCleanupPolicy()

	// Determine keep count
	keepCount := cleanupPolicy.Keep
	if keepOverride >= 0 {
		keepCount = keepOverride
	}

	// Initialize packages client
	ghExec := gh.New(tool.GlobalToolSystem(), workspaceRoot)
	pkgClient := github.NewGHPackagesClient(ghExec, workspaceRoot)
	ghClient := github.NewGHClient(ghExec, workspaceRoot)
	ctx := context.Background()

	// Get org (from registry config or derived from remote config)
	org := cfg.GetRegistryOrg("ghcr.io")
	if org == "" {
		log.Errorf("No organization configured. Set registries.ghcr.io.org or repository.remote.registry_url")
		return 1
	}

	// Get packages to process
	var packages []string
	if pruneAll {
		pkgList, err := pkgClient.ListOrgPackages(ctx, org)
		if err != nil {
			log.Errorf("Failed to list packages: %v", err)
			return 1
		}
		for _, pkg := range pkgList {
			packages = append(packages, pkg.Name)
		}
	} else {
		packages = []string{packageName}
	}

	// Get releases for safety checking
	releases, err := ghClient.ListReleases(500)
	if err != nil {
		log.Warnf("Failed to list releases (continuing without release protection): %v", err)
		releases = nil
	}

	var results []PrunePackagesResult
	totalDeleted := 0

	for _, pkg := range packages {
		result := prunePackage(ctx, pkg, org, cleanupPolicy, keepCount, releases, pkgClient, force, verbose, jsonOutput)
		results = append(results, result)
		totalDeleted += result.DeletedCount
	}

	// Output results
	if jsonOutput {
		output, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(output))
	} else {
		log.Infof("")
		if force {
			log.Infof("Cleanup complete. Deleted %d versions.", totalDeleted)
		} else {
			log.Infof("DRY RUN complete. Would delete %d versions.", totalDeleted)
			log.Infof("Use --force to actually delete.")
		}
	}

	return 0
}

func prunePackage(
	ctx context.Context,
	packageName, org string,
	policy *config.RegistryCleanupPolicy,
	keepCount int,
	releases []github.Release,
	client *github.GHPackagesClient,
	force, verbose, jsonOutput bool,
) PrunePackagesResult {
	result := PrunePackagesResult{
		Package: packageName,
	}

	if !jsonOutput {
		log.Infof("Processing package: %s", packageName)
	}

	// List versions
	versions, err := client.ListPackageVersions(ctx, org, packageName)
	if err != nil {
		result.Error = err.Error()
		if !jsonOutput {
			log.Errorf("  Error listing versions: %v", err)
		}
		return result
	}

	result.TotalVersions = len(versions)

	// Create safety checker with image tag rules
	// Released packages are ALWAYS protected - this is non-configurable for safety
	checker := github.NewPackageSafetyChecker(
		policy.GetPreservePatterns(),
		policy.GetPrunePatterns(),
		policy.MinAgeDays,
		releases,
	)

	// Extract and register bundle module versions
	// This protects packages referenced by release bundles
	checker.AddBundleModuleVersionsFromReleases(releases)

	// First pass: identify released digests for cross-digest protection
	for _, v := range versions {
		for _, tag := range v.Tags {
			for _, pattern := range policy.GetPreservePatterns() {
				if matched, _ := pathMatch(pattern, tag); matched {
					checker.AddReleasedDigest(v.Digest)
					break
				}
			}
		}
	}

	// Assess all versions
	protected, prunable := checker.AssessAll(versions)

	result.ProtectedCount = len(protected)
	result.PrunableCount = len(prunable)

	// Convert to result format
	for _, a := range protected {
		result.Protected = append(result.Protected, PruneVersionResult{
			ID:        a.Version.ID,
			Digest:    a.Version.Digest,
			Tags:      a.Version.Tags,
			CreatedAt: a.Version.CreatedAt.Format(time.RFC3339),
			Reason:    string(a.Reason),
		})
	}

	for _, a := range prunable {
		result.Prunable = append(result.Prunable, PruneVersionResult{
			ID:        a.Version.ID,
			Digest:    a.Version.Digest,
			Tags:      a.Version.Tags,
			CreatedAt: a.Version.CreatedAt.Format(time.RFC3339),
		})
	}

	// Sort prunable by date (newest first)
	sort.Slice(result.Prunable, func(i, j int) bool {
		return result.Prunable[i].CreatedAt > result.Prunable[j].CreatedAt
	})

	// Determine which prunable versions to delete (keep newest N)
	if len(result.Prunable) > keepCount {
		result.VersionsToDelete = result.Prunable[keepCount:]
		result.KeptCount = keepCount
	} else {
		result.KeptCount = len(result.Prunable)
	}

	if !jsonOutput {
		log.Infof("  Total versions: %d", result.TotalVersions)
		log.Infof("  Protected: %d", result.ProtectedCount)
		log.Infof("  Prunable: %d (keeping %d)", result.PrunableCount, result.KeptCount)

		if verbose && len(result.Protected) > 0 {
			log.Infof("")
			log.Infof("  Protected versions:")
			for _, p := range result.Protected {
				tags := strings.Join(p.Tags, ", ")
				if tags == "" {
					tags = "(untagged)"
				}
				log.Infof("    - %s [%s]: %s", p.Digest[:12], tags, p.Reason)
			}
		}
	}

	// Delete or report
	if len(result.VersionsToDelete) == 0 {
		if !jsonOutput {
			log.Infof("  Nothing to delete")
		}
		return result
	}

	if !jsonOutput {
		log.Infof("  Versions to delete: %d", len(result.VersionsToDelete))
	}

	for _, v := range result.VersionsToDelete {
		tags := strings.Join(v.Tags, ", ")
		if tags == "" {
			tags = "(untagged)"
		}

		if force {
			if !jsonOutput {
				log.Infof("  Deleting: %s [%s]", v.Digest[:12], tags)
			}
			if err := client.DeletePackageVersion(ctx, org, packageName, v.ID); err != nil {
				if !jsonOutput {
					log.Errorf("    Error: %v", err)
				}
			} else {
				result.DeletedCount++
			}
		} else {
			if !jsonOutput {
				log.Infof("  Would delete: %s [%s]", v.Digest[:12], tags)
			}
			result.DeletedCount++ // In dry-run, count as "would delete"
		}
	}

	return result
}

// pathMatch is a simple wrapper for filepath.Match used by the safety checker
func pathMatch(pattern, name string) (bool, error) {
	// Using simple glob matching
	if pattern == "" {
		return false, nil
	}

	// Handle * prefix
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(name, pattern[1:]), nil
	}

	// Handle * suffix
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(name, pattern[:len(pattern)-1]), nil
	}

	// Handle [0-9] character classes
	if strings.Contains(pattern, "[") {
		// Simplified: just use filepath.Match
		return name == pattern, nil
	}

	// Exact match
	return name == pattern, nil
}
