// Command: update docs-manifest
// Short: Update the documentation assets manifest
// Long: Scans docs/assets/ for documentation assets (drawio diagrams, images)
// Long: and updates the asset tracking files.
// Long:
// Long: This command manages two files:
// Long:   - descriptions.yml: Human-maintained descriptions and active status (git-tracked)
// Long:   - .manifest-cache.json: Auto-generated usage and metadata (git-ignored)
// Long:
// Long: The descriptions file tracks:
// Long:   - Asset descriptions (human/LLM-authored, preserved on update)
// Long:   - Active status flag (mark assets as deprecated with active: false)
// Long:
// Long: The cache file tracks:
// Long:   - Usage references (auto-detected from markdown files)
// Long:   - File metadata (size, hash, last modified)
// Long:   - Statistics (total, used, unused by category)
// Long:
// Long: Expected Output:
// Long:   - Updates .manifest-cache.json with current usage
// Long:   - Updates descriptions.yml only when new assets are discovered
// Long:   - Reports added/removed/changed assets
// Long:   - Lists new assets needing descriptions
// Long:
// Long: Example:
// Long:   update docs-manifest              # Update manifest files
// Long:   update docs-manifest --check      # Validate manifest is up-to-date (CI)
// Long:   update docs-manifest --dry-run    # Show what would change
// Long:   update docs-manifest --migrate    # Migrate from legacy manifest.json
// Flag.check: type=bool, default=false, usage=Validate manifest is up-to-date (exits non-zero if stale)
// Flag.dry-run: type=bool, default=false, usage=Show what would change without writing
// Flag.migrate: type=bool, default=false, usage=Migrate from legacy manifest.json format
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show detailed progress
package docsmanifest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

var log = logging.C()

func init() {
	registry.Register(UpdateDocsManifest)
}

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
	migrate := false
	verbose := false

	for _, arg := range args {
		switch arg {
		case "--check":
			checkMode = true
		case "--dry-run":
			dryRun = true
		case "--migrate":
			migrate = true
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
	legacyManifestPath := filepath.Join(assetsDir, LegacyManifestFileName)

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

	// Handle migration mode
	if migrate {
		return runMigration(legacyManifestPath, descriptionsPath, cachePath, docsDir, assetsDir, dryRun, verbose)
	}

	// Normal operation
	return runUpdate(descriptionsPath, cachePath, docsDir, assetsDir, checkMode, dryRun, verbose)
}

// runMigration handles the one-time migration from legacy manifest.json
func runMigration(legacyPath, descriptionsPath, cachePath, docsDir, assetsDir string, dryRun, verbose bool) int {
	fmt.Println("Migration mode: converting from legacy manifest.json")

	// Check if legacy manifest exists
	legacy, err := LoadLegacyManifest(legacyPath)
	if err != nil {
		log.Errorf("Error loading legacy manifest: %v", err)
		return 1
	}
	if legacy == nil {
		fmt.Println("No legacy manifest.json found. Nothing to migrate.")
		return 0
	}

	fmt.Printf("  Found legacy manifest with %d assets\n", len(legacy.Assets))

	// Extract descriptions from legacy format
	descriptions := ExtractDescriptionsFromLegacy(legacy)
	fmt.Printf("  Extracted %d descriptions\n", len(descriptions))

	// Scan current assets and usage
	fmt.Println("\nScanning assets...")
	discovered, err := ScanAssets(assetsDir)
	if err != nil {
		log.Errorf("Error scanning assets: %v", err)
		return 1
	}
	fmt.Printf("  Found %d assets\n", len(discovered))

	fmt.Println("\nScanning markdown files...")
	usage, err := ScanUsage(docsDir, assetsDir)
	if err != nil {
		log.Errorf("Error scanning usage: %v", err)
		return 1
	}

	// Merge to create cache
	mergeResult := MergeManifest(descriptions, nil, discovered, usage)

	// Handle dry-run mode
	if dryRun {
		fmt.Println("\n[DRY RUN] Would perform the following:")
		fmt.Printf("  - Write descriptions.yml with %d entries\n", len(mergeResult.Descriptions))
		fmt.Printf("  - Write .manifest-cache.json with %d entries\n", len(mergeResult.Cache.Assets))
		fmt.Printf("  - Delete legacy manifest.json\n")
		return 0
	}

	// Write new files
	fmt.Println("\nWriting new files...")

	if err := SaveDescriptions(descriptionsPath, mergeResult.Descriptions); err != nil {
		log.Errorf("Error writing descriptions.yml: %v", err)
		return 1
	}
	fmt.Printf("  Written: %s\n", descriptionsPath)

	if err := SaveCache(mergeResult.Cache, cachePath); err != nil {
		log.Errorf("Error writing cache: %v", err)
		return 1
	}
	fmt.Printf("  Written: %s\n", cachePath)

	// Remove legacy manifest
	if err := os.Remove(legacyPath); err != nil {
		log.Errorf("Error removing legacy manifest: %v", err)
		return 1
	}
	fmt.Printf("  Removed: %s\n", legacyPath)

	fmt.Println("\n✓ Migration complete!")
	fmt.Println("  - descriptions.yml is git-tracked (contains descriptions)")
	fmt.Println("  - .manifest-cache.json is git-ignored (auto-generated)")
	return 0
}

// runUpdate handles normal update operation
func runUpdate(descriptionsPath, cachePath, docsDir, assetsDir string, checkMode, dryRun, verbose bool) int {
	// Phase 1: Load existing files
	fmt.Println("Loading existing files...")
	descriptions, err := LoadDescriptions(descriptionsPath)
	if err != nil {
		log.Errorf("Error loading descriptions: %v", err)
		return 1
	}
	if len(descriptions) > 0 {
		fmt.Printf("  Loaded %d descriptions\n", len(descriptions))
	} else {
		fmt.Println("  No existing descriptions (will create new)")
	}

	existingCache, err := LoadCache(cachePath)
	if err != nil {
		log.Errorf("Error loading cache: %v", err)
		return 1
	}
	if existingCache != nil {
		fmt.Printf("  Loaded cache with %d assets\n", len(existingCache.Assets))
	} else {
		fmt.Println("  No existing cache (will create new)")
	}

	// Phase 2: Scan for assets
	fmt.Println("\nScanning assets...")
	discovered, err := ScanAssets(assetsDir)
	if err != nil {
		log.Errorf("Error scanning assets: %v", err)
		return 1
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

	// Phase 3: Scan for usage
	fmt.Println("\nScanning markdown files...")
	usage, err := ScanUsage(docsDir, assetsDir)
	if err != nil {
		log.Errorf("Error scanning usage: %v", err)
		return 1
	}

	usedCount := 0
	for _, refs := range usage {
		if len(refs) > 0 {
			usedCount++
		}
	}
	fmt.Printf("  Found %d assets referenced in documentation\n", usedCount)

	// Phase 4: Merge
	fmt.Println("\nUpdating manifest...")
	mergeResult := MergeManifest(descriptions, existingCache, discovered, usage)
	result := mergeResult.Result
	result.DescriptionsPath = descriptionsPath
	result.CachePath = cachePath

	// Report changes
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

	// Summary
	fmt.Println("\nSummary:")
	fmt.Printf("  Total: %d assets\n", result.TotalAssets)
	if result.TotalAssets > 0 {
		usedPct := float64(result.UsedAssets) / float64(result.TotalAssets) * 100
		fmt.Printf("  Used: %d (%.0f%%)\n", result.UsedAssets, usedPct)
		fmt.Printf("  Unused: %d (%.0f%%)\n", result.UnusedAssets, 100-usedPct)
	}

	// List assets needing descriptions
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
				if !verbose && !contains(result.Added, path) {
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

	// Handle check mode
	if checkMode {
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

	// Handle dry-run mode
	if dryRun {
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

	// Write files
	if result.NeedsDescriptionsWrite {
		if err := SaveDescriptions(descriptionsPath, mergeResult.Descriptions); err != nil {
			log.Errorf("Error writing descriptions: %v", err)
			return 1
		}
		fmt.Printf("\n✓ Written: %s\n", descriptionsPath)
	}

	if result.NeedsCacheWrite {
		if err := SaveCache(mergeResult.Cache, cachePath); err != nil {
			log.Errorf("Error writing cache: %v", err)
			return 1
		}
		fmt.Printf("✓ Written: %s\n", cachePath)
	}

	if !result.NeedsDescriptionsWrite && !result.NeedsCacheWrite {
		fmt.Println("\n✓ Manifest is already up to date")
	}

	return 0
}
