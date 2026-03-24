package release

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/changelog"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/git"
)

type releaseCheckPendingCommand struct{}

var _ core.SimpleCommandPort = (*releaseCheckPendingCommand)(nil)

func (c *releaseCheckPendingCommand) Name() string { return "release check-pending" }

func (c *releaseCheckPendingCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "release-check-pending",
		Short:         "Check for all pending releases (semver and calver)",
		Long: "Comprehensive release detection that combines semver (from changelog) and calver (from CI dispatch).\n\nThis command:\n  1. Checks semver modules for changelog versions without git tags\n  2. Checks calver modules with CI workflows (release when CI was dispatched)\n  3. Checks calver bundle modules (release when any dependency was dispatched)\n  4. Returns enriched layers ready for release execute-layers",
		Notes: "Expected Output:\n  - JSON object with has_pending, modules_json, layers_json, layer_count\n  - layers_json contains enriched module info [{module, version, tag, type}, ...]",
		Examples: []string{
			"eac release check-pending --dispatched \"docs books\"  # Check with dispatched modules",
			"eac release check-pending                            # Check semver only",
		},
		Flags: []core.FlagSpec{
			{Name: "dispatched", Type: "string", Usage: "Space-separated list of modules that had CI dispatched"},
		},
	}
}

func (c *releaseCheckPendingCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ReleaseCheckPending()
}

// PendingModule represents a module needing release.
type PendingModule struct {
	Module   string `json:"module"`
	Version  string `json:"version"`
	Tag      string `json:"tag"`
	Type     string `json:"type"` // "semver" or "calver"
	NeedsTag bool   `json:"needs_tag"`
}

// CheckPendingResult is the comprehensive output.
type CheckPendingResult struct {
	HasPending  bool              `json:"has_pending"`
	ModulesJSON []PendingModule   `json:"modules_json"`
	LayersJSON  [][]PendingModule `json:"layers_json"`
	LayerCount  int               `json:"layer_count"`
}

func ReleaseCheckPending() int {
	// Parse flags
	dispatched := ""
	format := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--dispatched" && i+1 < len(os.Args):
			dispatched = os.Args[i+1]
			i++
		case arg == "--format" && i+1 < len(os.Args):
			format = os.Args[i+1]
			i++
		}
	}

	dispatchedModules := parseDispatchedModules(dispatched)

	s, exitCode := newReleaseScaffoldNoFlags(withModules(), withGit())
	if s == nil {
		return exitCode
	}

	workspaceRoot := s.WorkspaceRoot
	moduleRegistry := s.ModuleRegistry
	repo := s.Repo

	// Collect all pending releases
	var allPending []PendingModule

	// 1. Check semver modules (changelog versions without tags)
	semverPending, err := checkSemverPending(workspaceRoot, moduleRegistry, repo)
	if err != nil {
		log.Warnf("failed to check semver pending: %v", err)
	} else {
		allPending = append(allPending, semverPending...)
	}

	// 2. Check calver modules with CI (release when CI was dispatched)
	calverPending := checkCalverWithCI(moduleRegistry, dispatchedModules)
	allPending = append(allPending, calverPending...)

	// 3. Check calver bundle modules (release when any dependency was dispatched)
	bundlePending := checkCalverBundles(moduleRegistry, dispatchedModules)
	allPending = append(allPending, bundlePending...)

	// Calculate execution order for pending modules
	if len(allPending) == 0 {
		result := CheckPendingResult{
			HasPending:  false,
			ModulesJSON: []PendingModule{},
			LayersJSON:  [][]PendingModule{},
			LayerCount:  0,
		}
		outputCheckPendingResult(result, format)
		return 0
	}

	// Build release layers using bundle-aware strategy:
	// - Layer 0: All non-bundle releases (can run in parallel)
	// - Layer 1: All bundle releases (must wait for dependencies)
	//
	// This is different from build layering which uses depends_on.
	// For releases, only bundles need to wait because they aggregate
	// release notes from their dependencies.
	var layer0 []PendingModule // non-bundles
	var layer1 []PendingModule // bundles

	for _, p := range allPending {
		mod, exists := moduleRegistry.Get(p.Module)
		if !exists {
			// Module not found, treat as non-bundle
			layer0 = append(layer0, p)
			continue
		}

		if mod.GetReleaseBundle() != nil {
			// Bundle release - must wait for dependencies
			layer1 = append(layer1, p)
		} else {
			// Regular release - can run in parallel
			layer0 = append(layer0, p)
		}
	}

	// Build enriched layers (only include non-empty layers)
	var enrichedLayers [][]PendingModule
	if len(layer0) > 0 {
		enrichedLayers = append(enrichedLayers, layer0)
	}
	if len(layer1) > 0 {
		enrichedLayers = append(enrichedLayers, layer1)
	}

	result := CheckPendingResult{
		HasPending:  true,
		ModulesJSON: allPending,
		LayersJSON:  enrichedLayers,
		LayerCount:  len(enrichedLayers),
	}

	outputCheckPendingResult(result, format)
	return 0
}

func outputCheckPendingResult(result CheckPendingResult, format string) {
	if format == "shell" {
		// Shell format for eval
		modules := make([]string, len(result.ModulesJSON))
		for i, m := range result.ModulesJSON {
			modules[i] = m.Module
		}

		fmt.Printf("HAS_PENDING=\"%t\"\n", result.HasPending)
		fmt.Printf("MODULES=\"%s\"\n", strings.Join(modules, " "))
		fmt.Printf("LAYER_COUNT=\"%d\"\n", result.LayerCount)

		// Output LAYERS_JSON as escaped JSON string for complex parsing
		if layersJSON, err := json.Marshal(result.LayersJSON); err == nil {
			fmt.Printf("LAYERS_JSON='%s'\n", string(layersJSON))
		}
		if modulesJSON, err := json.Marshal(result.ModulesJSON); err == nil {
			fmt.Printf("MODULES_JSON='%s'\n", string(modulesJSON))
		}
		return
	}

	// Default: JSON output
	outputJSON(result)
}

func parseDispatchedModules(dispatched string) map[string]bool {
	result := make(map[string]bool)
	if dispatched == "" {
		return result
	}
	for _, mod := range strings.Fields(dispatched) {
		result[mod] = true
	}
	return result
}

func checkSemverPending(workspaceRoot string, moduleRegistry *modules.Registry, repo git.GitRepository) ([]PendingModule, error) {
	var result []PendingModule

	// Find modules with changelogs
	releaseDir := filepath.Join(workspaceRoot, "release")
	entries, err := os.ReadDir(releaseDir)
	if err != nil {
		return result, nil // No release dir is fine
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		module := entry.Name()

		moduleContract, exists := moduleRegistry.Get(module)
		if !exists {
			continue
		}

		// Parse changelog
		changelogPath := filepath.Join(workspaceRoot, moduleContract.GetChangelogPath())
		cl, err := changelog.Parse(changelogPath)
		if err != nil {
			continue
		}

		if len(cl.Versions) == 0 {
			continue
		}

		latestVersion := cl.Versions[0].Number
		tag := fmt.Sprintf("%s/%s", module, latestVersion)

		// Check if tag exists
		exists, err = repo.TagExists(tag)
		if err != nil {
			continue
		}

		if !exists {
			result = append(result, PendingModule{
				Module:   module,
				Version:  latestVersion,
				Tag:      tag,
				Type:     "semver",
				NeedsTag: true,
			})
		}
	}

	return result, nil
}

func checkCalverWithCI(moduleRegistry *modules.Registry, dispatched map[string]bool) []PendingModule {
	var result []PendingModule

	if len(dispatched) == 0 {
		return result
	}

	for _, mod := range moduleRegistry.All() {
		// Check if CalVer with CI workflow
		if mod.Versioning == nil || !strings.EqualFold(mod.Versioning.Scheme, "calver") {
			continue
		}
		if mod.GetCIWorkflowPath() == "" {
			continue
		}

		// Check if this module was dispatched
		if !dispatched[mod.Moniker] {
			continue
		}

		version := time.Now().UTC().Format("2006.0102.1504")
		result = append(result, PendingModule{
			Module:   mod.Moniker,
			Version:  version,
			Tag:      fmt.Sprintf("%s/%s", mod.Moniker, version),
			Type:     "calver",
			NeedsTag: true,
		})
	}

	return result
}

func checkCalverBundles(moduleRegistry *modules.Registry, dispatched map[string]bool) []PendingModule {
	var result []PendingModule

	if len(dispatched) == 0 {
		return result
	}

	for _, mod := range moduleRegistry.All() {
		// Check if CalVer bundle (has release but no CI)
		if mod.Versioning == nil || !strings.EqualFold(mod.Versioning.Scheme, "calver") {
			continue
		}
		if mod.GetCIWorkflowPath() != "" {
			continue // Has CI, not a bundle
		}
		if mod.GetReleaseWorkflowPath() == "" {
			continue // No release workflow
		}

		// Check if any dependency was dispatched
		shouldRelease := false
		for _, dep := range mod.DependsOn {
			if dispatched[dep] {
				shouldRelease = true
				break
			}
		}

		if !shouldRelease {
			continue
		}

		version := time.Now().UTC().Format("2006.0102.1504")
		result = append(result, PendingModule{
			Module:   mod.Moniker,
			Version:  version,
			Tag:      fmt.Sprintf("%s/%s", mod.Moniker, version),
			Type:     "calver",
			NeedsTag: true,
		})
	}

	return result
}

func outputJSON(v interface{}) {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Errorf("failed to marshal JSON: %v", err)
		return
	}
	fmt.Println(string(jsonBytes))
}
