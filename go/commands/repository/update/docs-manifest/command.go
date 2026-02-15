package docsmanifest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

type updateDocsManifestCommand struct{}

var _ core.SimpleCommandPort = (*updateDocsManifestCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&updateDocsManifestCommand{},
	}
}

func (c *updateDocsManifestCommand) Name() string { return "update docs-manifest" }

func (c *updateDocsManifestCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "update-docs-manifest",
		Short:         "Update the documentation assets manifest",
		Long:          "Scans docs/assets/ for documentation assets (drawio diagrams, images)\nand updates the asset tracking files.\n\nThis command manages two files:\n  - descriptions.yml: Human-maintained descriptions and active status (git-tracked)\n  - .manifest-cache.json: Auto-generated usage and metadata (git-ignored)\n\nThe descriptions file tracks:\n  - Asset descriptions (human/LLM-authored, preserved on update)\n  - Active status flag (mark assets as deprecated with active: false)\n\nThe cache file tracks:\n  - Usage references (auto-detected from markdown files)\n  - File metadata (size, hash, last modified)\n  - Statistics (total, used, unused by category)\n\nExpected Output:\n  - Updates .manifest-cache.json with current usage\n  - Updates descriptions.yml only when new assets are discovered\n  - Reports added/removed/changed assets\n  - Lists new assets needing descriptions\n\nExample:\n  update docs-manifest              # Update manifest files\n  update docs-manifest --check      # Validate manifest is up-to-date (CI)\n  update docs-manifest --dry-run    # Show what would change",
		Flags: []core.FlagSpec{
			{Name: "check", Type: "bool", DefaultValue: "false", Usage: "Validate manifest is up-to-date (exits non-zero if stale)"},
			{Name: "dry-run", Type: "bool", DefaultValue: "false", Usage: "Show what would change without writing"},
			{Name: "verbose", Shorthand: "v", Type: "bool", DefaultValue: "false", Usage: "Show detailed progress"},
		},
	}
}

func (c *updateDocsManifestCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return UpdateDocsManifest()
}

var log = logging.C()
// UpdateDocsManifest updates the documentation assets manifest.
func UpdateDocsManifest() int {
	// Validate flags
	args := os.Args[3:] // Skip program name, "update", and "docs-manifest"
	if err := flags.ValidateFlagsFromRegistry(args); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse flags
	checkMode := false
	dryRun := false
	verbose := false

	for _, arg := range args {
		switch arg {
		case "--check":
			checkMode = true
		case "--dry-run":
			dryRun = true
		case "-v", "--verbose":
			verbose = true
		}
	}

	// Get repo root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Define paths
	docsDir := paths.DocsSourcePath(repoRoot)
	assetsDir := filepath.Join(docsDir, "assets")
	descriptionsPath := filepath.Join(assetsDir, DescriptionsFileName)
	cachePath := filepath.Join(assetsDir, CacheFileName)

	// Check if assets directory exists
	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		log.Errorf("Error: assets directory not found: %s", assetsDir)
		return 1
	}

	if verbose {
		fmt.Printf("Docs directory: %s\n", docsDir)
		fmt.Printf("Assets directory: %s\n", assetsDir)
		fmt.Printf("Descriptions path: %s\n", descriptionsPath)
		fmt.Printf("Cache path: %s\n", cachePath)
		fmt.Println()
	}

	// Normal operation
	return runUpdate(descriptionsPath, cachePath, docsDir, assetsDir, checkMode, dryRun, verbose)
}

// runUpdate handles normal update operation
func runUpdate(descriptionsPath, cachePath, docsDir, assetsDir string, checkMode, dryRun, verbose bool) int {
	// Phase 1: Load existing files
	descriptions, existingCache, ok := loadExistingFiles(descriptionsPath, cachePath)
	if !ok {
		return 1
	}

	// Phase 2: Scan for assets
	discovered, ok := scanAndReportAssets(assetsDir, verbose)
	if !ok {
		return 1
	}

	// Phase 3: Scan for usage
	usage, ok := scanAndReportUsage(docsDir, assetsDir)
	if !ok {
		return 1
	}

	// Phase 4: Merge and report
	mergeResult := mergeAndReport(descriptions, existingCache, discovered, usage, descriptionsPath, cachePath, verbose)

	// Handle check mode
	if checkMode {
		return handleCheckMode(mergeResult.Result)
	}

	// Handle dry-run mode
	if dryRun {
		return handleDryRunMode(mergeResult.Result, descriptionsPath, cachePath)
	}

	// Write files
	return writeManifestFiles(mergeResult)
}

// loadExistingFiles loads the descriptions and cache files, reporting progress.
// Returns the loaded data and true on success, or zero values and false on error.
func loadExistingFiles(descriptionsPath, cachePath string) (map[string]Description, *CacheManifest, bool) {
	fmt.Println("Loading existing files...")
	descriptions, err := LoadDescriptions(descriptionsPath)
	if err != nil {
		log.Errorf("Error loading descriptions: %v", err)
		return nil, nil, false
	}
	if len(descriptions) > 0 {
		fmt.Printf("  Loaded %d descriptions\n", len(descriptions))
	} else {
		fmt.Println("  No existing descriptions (will create new)")
	}

	existingCache, err := LoadCache(cachePath)
	if err != nil {
		log.Errorf("Error loading cache: %v", err)
		return nil, nil, false
	}
	if existingCache != nil {
		fmt.Printf("  Loaded cache with %d assets\n", len(existingCache.Assets))
	} else {
		fmt.Println("  No existing cache (will create new)")
	}

	return descriptions, existingCache, true
}

// scanAndReportAssets scans the assets directory and reports what was found.
// Returns the discovered assets and true on success, or nil and false on error.
func scanAndReportAssets(assetsDir string, verbose bool) (map[string]DiscoveredAsset, bool) {
	fmt.Println("\nScanning assets...")
	discovered, err := ScanAssets(assetsDir)
	if err != nil {
		log.Errorf("Error scanning assets: %v", err)
		return nil, false
	}

	// Count by category and type
	categories := make(map[string]int)
	types := make(map[string]int)
	for _, asset := range discovered {
		if asset.Category != "" {
			categories[asset.Category]++
		}
		types[asset.Type]++
	}
	fmt.Printf("  Found %d assets in %d categories\n", len(discovered), len(categories))

	if verbose {
		fmt.Println("  By category:")
		for cat, count := range categories {
			fmt.Printf("    %s: %d\n", cat, count)
		}
		fmt.Println("  By type:")
		for typ, count := range types {
			fmt.Printf("    %s: %d\n", typ, count)
		}
	}

	return discovered, true
}

// scanAndReportUsage scans markdown files for asset references and reports results.
// Returns the usage map and true on success, or nil and false on error.
func scanAndReportUsage(docsDir, assetsDir string) (map[string][]string, bool) {
	fmt.Println("\nScanning markdown files...")
	usage, err := ScanUsage(docsDir, assetsDir)
	if err != nil {
		log.Errorf("Error scanning usage: %v", err)
		return nil, false
	}

	usedCount := 0
	for _, refs := range usage {
		if len(refs) > 0 {
			usedCount++
		}
	}
	fmt.Printf("  Found %d assets referenced in documentation\n", usedCount)

	return usage, true
}

// mergeAndReport merges discovered data with existing state and prints a change summary.
func mergeAndReport(
	descriptions map[string]Description,
	existingCache *CacheManifest,
	discovered map[string]DiscoveredAsset,
	usage map[string][]string,
	descriptionsPath, cachePath string,
	verbose bool,
) *MergeResult {
	fmt.Println("\nUpdating manifest...")
	mergeResult := MergeManifest(descriptions, existingCache, discovered, usage)
	result := mergeResult.Result
	result.DescriptionsPath = descriptionsPath
	result.CachePath = cachePath

	reportChanges(result)
	reportSummaryStats(result)
	reportMissingDescriptions(mergeResult, verbose)

	return mergeResult
}

// reportChanges prints the added/removed/updated asset counts.
func reportChanges(result *UpdateResult) {
	if len(result.Added) > 0 {
		fmt.Printf("  + %d new asset(s)\n", len(result.Added))
	}
	if len(result.Removed) > 0 {
		fmt.Printf("  - %d removed asset(s)\n", len(result.Removed))
	}
	if len(result.UpdatedUsedIn) > 0 {
		fmt.Printf("  ~ %d asset(s) with updated references\n", len(result.UpdatedUsedIn))
	}
	if !result.NeedsDescriptionsWrite && !result.NeedsCacheWrite {
		fmt.Println("  No changes detected")
	}
}

// reportSummaryStats prints total, used, and unused asset statistics.
func reportSummaryStats(result *UpdateResult) {
	fmt.Println("\nSummary:")
	fmt.Printf("  Total: %d assets\n", result.TotalAssets)
	if result.TotalAssets > 0 {
		usedPct := float64(result.UsedAssets) / float64(result.TotalAssets) * 100
		fmt.Printf("  Used: %d (%.0f%%)\n", result.UsedAssets, usedPct)
		fmt.Printf("  Unused: %d (%.0f%%)\n", result.UnusedAssets, 100-usedPct)
	}
}

// reportMissingDescriptions lists assets that still need descriptions.
func reportMissingDescriptions(mergeResult *MergeResult, verbose bool) {
	result := mergeResult.Result

	needsDescription := []string{}
	for path, desc := range mergeResult.Descriptions {
		if desc.Description == "" {
			needsDescription = append(needsDescription, path)
		}
	}

	if len(needsDescription) > 0 {
		fmt.Printf("\nAssets needing descriptions: %d\n", len(needsDescription))
		if verbose || len(result.Added) > 0 {
			// Show new assets or all if verbose
			shown := 0
			for _, path := range needsDescription {
				// In non-verbose mode, only show newly added assets
				if !verbose && !slices.Contains(result.Added, path) {
					continue
				}
				fmt.Printf("  - %s\n", path)
				shown++
				if !verbose && shown >= 10 {
					remaining := len(needsDescription) - shown
					if remaining > 0 {
						fmt.Printf("  ... and %d more (use --verbose to see all)\n", remaining)
					}
					break
				}
			}
		}
	}
}

// handleCheckMode validates whether the manifest is up-to-date and returns the appropriate exit code.
func handleCheckMode(result *UpdateResult) int {
	if result.NeedsCacheWrite {
		fmt.Println("\nCache is out of date")
		if len(result.Added) > 0 {
			fmt.Println("   New assets not in manifest:")
			for _, p := range result.Added {
				fmt.Printf("     + %s\n", p)
			}
		}
		if len(result.Removed) > 0 {
			fmt.Println("   Deleted assets still in manifest:")
			for _, p := range result.Removed {
				fmt.Printf("     - %s\n", p)
			}
		}
		fmt.Println("\n   Run 'eac update docs-manifest' to update")
		return 1
	}
	fmt.Println("\n✓ Manifest is up to date")
	return 0
}

// handleDryRunMode reports what would change without writing and returns exit code 0.
func handleDryRunMode(result *UpdateResult, descriptionsPath, cachePath string) int {
	if result.NeedsDescriptionsWrite {
		fmt.Printf("\n[DRY RUN] Would update descriptions: %s\n", descriptionsPath)
	}
	if result.NeedsCacheWrite {
		fmt.Printf("[DRY RUN] Would write cache: %s\n", cachePath)
	}
	if !result.NeedsDescriptionsWrite && !result.NeedsCacheWrite {
		fmt.Println("\n[DRY RUN] No changes to write")
	}
	return 0
}

// writeManifestFiles persists the descriptions and cache files to disk.
func writeManifestFiles(mergeResult *MergeResult) int {
	result := mergeResult.Result

	if result.NeedsDescriptionsWrite {
		if err := SaveDescriptions(result.DescriptionsPath, mergeResult.Descriptions); err != nil {
			log.Errorf("Error writing descriptions: %v", err)
			return 1
		}
		fmt.Printf("\n✓ Written: %s\n", result.DescriptionsPath)
	}

	if result.NeedsCacheWrite {
		if err := SaveCache(mergeResult.Cache, result.CachePath); err != nil {
			log.Errorf("Error writing cache: %v", err)
			return 1
		}
		fmt.Printf("✓ Written: %s\n", result.CachePath)
	}

	if !result.NeedsDescriptionsWrite && !result.NeedsCacheWrite {
		fmt.Println("\n✓ Manifest is already up to date")
	}

	return 0
}
