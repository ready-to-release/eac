// structurizr.go - Build handler for rendering Structurizr architecture diagrams to SVG.
// Renders a single module's workspace.dsl via structurizr-cli + plantuml containers,
// and writes SVGs + index manifest to the module's output directory.
package builders

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/caching/itemcache"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	tool.GlobalBuildBridge().RegisterNativeHandler(&StructurizrRenderHandler{})
}

// StructurizrRenderHandler renders Structurizr architecture diagrams to SVG.
// It renders the current module's workspace.dsl, exports views via
// structurizr-cli + plantuml containers, and writes SVGs + index manifest.
// Uses .cache/eac/structurizr/ for incremental builds.
type StructurizrRenderHandler struct{}

func (h *StructurizrRenderHandler) Name() string           { return "structurizr-render" }
func (h *StructurizrRenderHandler) Requirements() []string { return []string{"docker"} }
func (h *StructurizrRenderHandler) IsContainer() bool      { return true }
func (h *StructurizrRenderHandler) IsHostInstalled() bool  { return false }

func (h *StructurizrRenderHandler) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	if !IsDockerAvailable() {
		if IsDockerInDocker() {
			return fmt.Errorf("Docker socket not mounted")
		}
		return fmt.Errorf("Docker is not available (required for structurizr rendering)")
	}
	return nil
}

func (h *StructurizrRenderHandler) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	return []string{"structurizr/"}
}

// StructurizrIndex maps source workspace.dsl modules to their rendered SVG filenames.
type StructurizrIndex struct {
	Entries []StructurizrIndexEntry `json:"entries"`
}

// StructurizrIndexEntry maps a single Structurizr view to its SVG file.
type StructurizrIndexEntry struct {
	Module      string `json:"module"`       // Module name (e.g., "eac")
	ViewKey     string `json:"view_key"`     // View key from DSL (e.g., "SystemContext")
	DSLHash     string `json:"dsl_hash"`     // First 8 chars of SHA256 of workspace.dsl
	SVGFilename string `json:"svg_filename"` // Filename in the output directory
}

// structurizrModuleWork tracks the state of a single module during build.
type structurizrModuleWork struct {
	ModuleName    string
	WorkspacePath string // Absolute path to workspace.dsl
	DSLHash       string
}

func (h *StructurizrRenderHandler) Build(
	module core.ModuleContractPort, workspaceRoot, outputDir string,
	logWriter io.Writer, opts BuildOptions,
) int {
	moduleName := module.GetMoniker()
	Logln(logWriter, "=== Rendering structurizr diagrams for %s via container ===", moduleName)

	specsDir := filepath.Join(workspaceRoot, "specs")
	cacheDir := filepath.Join(paths.StructurizrAccelCachePath(workspaceRoot), moduleName, filepath.Base(outputDir))
	outputStructurizrDir := filepath.Join(outputDir, "structurizr")

	// Resolve workspace.dsl from component root (per-component builds)
	// or fall back to module-level design path
	var workspacePath string
	if opts.Component != "" {
		compRoot := module.GetComponentRoot(opts.Component)
		if compRoot != "" {
			workspacePath = filepath.Join(workspaceRoot, compRoot, "workspace.dsl")
		}
	}
	if workspacePath == "" {
		workspacePath = filepath.Join(specsDir, moduleName, ".design", "workspace.dsl")
	}

	content, err := os.ReadFile(workspacePath)
	if err != nil {
		if os.IsNotExist(err) {
			Logln(logWriter, "No workspace.dsl found for module %s", moduleName)
			_ = os.MkdirAll(outputStructurizrDir, 0o755)
			return 0
		}
		Logln(logWriter, "Error reading workspace.dsl: %v", err)
		return 1
	}

	// Derive index module name from the component name.
	// For per-component builds, strip known prefixes to get the short name
	// (e.g., "structurizr-tui-eac" -> "tui-eac", "design-godog" -> "godog")
	// which is used in index entries and cache filenames.
	indexName := moduleName
	if opts.Component != "" {
		short := opts.Component
		for _, prefix := range []string{"structurizr-", "design-"} {
			if trimmed := strings.TrimPrefix(short, prefix); trimmed != short {
				short = trimmed
				break
			}
		}
		if short != opts.Component {
			indexName = short
		}
	}

	dslHash := hashDSLContent(string(content))
	mod := structurizrModuleWork{
		ModuleName:    indexName,
		WorkspacePath: workspacePath,
		DSLHash:       dslHash,
	}

	// Discover items: either from existing cache (if DSL hash matches) or via rendering.
	// Structurizr-cli discovers views during export, so the item list comes from either:
	//   - Previous cached SVGs (if DSL hash matches AND not force-rebuild)
	//   - Fresh render output (if DSL hash changed or force-rebuild)
	var items []itemcache.Item
	var allEntries []StructurizrIndexEntry
	alreadyRendered := false

	cachedSVGs := findCachedSVGs(cacheDir, mod.ModuleName, mod.DSLHash)
	if len(cachedSVGs) > 0 && !opts.ForceRebuild {
		// Items from cache -- all will be hits in itemcache.Execute
		for _, svgPath := range cachedSVGs {
			filename := filepath.Base(svgPath)
			viewKey := extractViewKeyFromCacheFilename(filename, mod.ModuleName)
			items = append(items, itemcache.Item{
				Key:           indexName + ":" + viewKey,
				ContentHash:   dslHash,
				CacheFilename: filename,
				OutputRelPath: filename,
			})
			allEntries = append(allEntries, StructurizrIndexEntry{
				Module:      indexName,
				ViewKey:     viewKey,
				DSLHash:     dslHash,
				SVGFilename: filename,
			})
		}
	} else {
		// Render to discover views, then build items from the output
		Logln(logWriter, "Rendering %s (hash: %s)...", mod.ModuleName, mod.DSLHash)
		views, renderErr := renderStructurizrModule(mod, workspaceRoot, specsDir, cacheDir, logWriter)
		if renderErr != nil {
			Logln(logWriter, "Error rendering %s: %v", mod.ModuleName, renderErr)
			return 1
		}
		alreadyRendered = true

		for _, view := range views {
			items = append(items, itemcache.Item{
				Key:           indexName + ":" + view.ViewKey,
				ContentHash:   dslHash,
				CacheFilename: view.SVGFilename,
				OutputRelPath: view.SVGFilename,
			})
			allEntries = append(allEntries, view)
		}
	}

	// Execute with cache -- handles copy-to-output, manifest update, and pruning.
	// For structurizr, the buildFunc is a no-op since rendering happened above when needed.
	buildFunc := func(misses []itemcache.Item, dir string) (int, error) {
		if alreadyRendered {
			// Files were already rendered to cacheDir above
			return len(misses), nil
		}
		// Shouldn't reach here (all items came from cache), but render if needed
		views, renderErr := renderStructurizrModule(mod, workspaceRoot, specsDir, dir, logWriter)
		if renderErr != nil {
			return 0, renderErr
		}
		return len(views), nil
	}

	cache := itemcache.New(cacheDir)
	result, err := cache.Execute(items, outputStructurizrDir, buildFunc, false)
	if err != nil {
		Logln(logWriter, "Error: %v", err)
		return 1
	}

	// Write index manifest
	manifest := StructurizrIndex{Entries: allEntries}
	manifestPath := filepath.Join(outputStructurizrDir, "structurizr-index.json")
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	_ = os.WriteFile(manifestPath, manifestData, 0o644)

	if result.CacheMisses == 0 {
		Logln(logWriter, "Structurizr: cached (hash: %s)", mod.DSLHash)
	} else {
		Logln(logWriter, "Structurizr: rendered (hash: %s)", mod.DSLHash)
	}
	Logln(logWriter, "Output: %s (%d SVGs, 1 index)", outputStructurizrDir, len(allEntries))

	return 0
}

// hashDSLContent returns the first 8 characters of SHA256 hash of DSL content.
// Line endings are normalized to LF for consistent hashing across platforms.
func hashDSLContent(content string) string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	h := sha256.Sum256([]byte(normalized))
	return fmt.Sprintf("%x", h)[:8]
}

// findCachedSVGs returns all cached SVGs for a module with the given hash.
func findCachedSVGs(cacheDir, moduleName, dslHash string) []string {
	pattern := filepath.Join(cacheDir, moduleName+"_*_"+dslHash+".svg")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil
	}
	return matches
}

// extractViewKeyFromCacheFilename extracts the view key from a cache filename.
// Filename format: {module}_{viewKey}_{hash}.svg
func extractViewKeyFromCacheFilename(filename, moduleName string) string {
	// Remove module prefix and .svg suffix
	name := strings.TrimPrefix(filename, moduleName+"_")
	name = strings.TrimSuffix(name, ".svg")
	// Remove hash suffix (last _XXXXXXXX)
	if idx := strings.LastIndex(name, "_"); idx > 0 {
		name = name[:idx]
	}
	return name
}

// viewKeyPattern extracts the view key from Structurizr CLI export filename.
// Structurizr CLI exports files as "structurizr-<key>.svg" or "structurizr-<key>-<number>.svg".
var structurizrViewKeyPattern = regexp.MustCompile(`^structurizr-(.+?)(?:-\d+)?\.svg$`)

func extractViewKeyFromExport(filename string) string {
	matches := structurizrViewKeyPattern.FindStringSubmatch(filename)
	if len(matches) >= 2 {
		return matches[1]
	}
	// Fallback: remove extension and prefix
	name := strings.TrimSuffix(filename, ".svg")
	name = strings.TrimPrefix(name, "structurizr-")
	return name
}

// renderStructurizrModule renders a single module's workspace.dsl to SVGs.
// Uses a two-step pipeline: structurizr-cli (DSL -> PUML) then plantuml (PUML -> SVG).
// Writes rendered SVGs to cacheDir with proper naming.
func renderStructurizrModule(mod structurizrModuleWork, workspaceRoot, specsDir, cacheDir string, logWriter io.Writer) ([]StructurizrIndexEntry, error) {
	// Create temp directory for export intermediates
	tempDir, err := os.MkdirTemp("", "structurizr-export-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Calculate relative path from specs dir to workspace file
	relWorkspacePath, err := filepath.Rel(specsDir, mod.WorkspacePath)
	if err != nil {
		return nil, fmt.Errorf("calculating relative path: %w", err)
	}
	relWorkspacePath = filepath.ToSlash(relWorkspacePath)

	// Step 1: Export to PlantUML using structurizr-export via tool bridge
	// Mount specs/ as /workspace so workspace.dsl relative includes resolve correctly
	bridge := tool.GlobalHandlerToolBridge()
	tc1 := &tool.ToolContext{
		WorkspaceRoot: specsDir,
		OutputDir:     tempDir,
		LogWriter:     logWriter,
		Variables: map[string]string{
			"workspace_path": "/workspace/" + relWorkspacePath,
		},
	}

	exitCode, execErr := bridge.ExecuteTool(context.Background(), "structurizr-export", tc1)
	if execErr != nil {
		return nil, fmt.Errorf("structurizr-export failed: %w", execErr)
	}
	if exitCode != 0 {
		return nil, fmt.Errorf("structurizr-export exited with code %d", exitCode)
	}

	// Check for .puml files
	pumlFiles, _ := filepath.Glob(filepath.Join(tempDir, "*.puml"))
	if len(pumlFiles) == 0 {
		// No views exported - not an error
		return nil, nil
	}

	// Step 2: Render PlantUML files to SVG using plantuml via tool bridge
	tc2 := &tool.ToolContext{
		WorkspaceRoot: workspaceRoot,
		OutputDir:     tempDir,
		LogWriter:     logWriter,
	}

	exitCode, execErr = bridge.ExecuteTool(context.Background(), "plantuml", tc2)
	if execErr != nil {
		return nil, fmt.Errorf("plantuml render failed: %w", execErr)
	}
	if exitCode != 0 {
		return nil, fmt.Errorf("plantuml exited with code %d", exitCode)
	}

	// Clean up .puml files (only need SVGs)
	for _, pumlFile := range pumlFiles {
		os.Remove(pumlFile)
	}

	// Ensure cache directory exists for the deeper per-UoW path
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating cache dir: %w", err)
	}

	// Collect exported SVGs and copy to cache with proper naming
	var entries []StructurizrIndexEntry

	svgFiles, _ := filepath.Glob(filepath.Join(tempDir, "*.svg"))
	for _, svgPath := range svgFiles {
		viewKey := extractViewKeyFromExport(filepath.Base(svgPath))
		if viewKey == "" {
			continue
		}

		// Build cache filename: {module}_{viewKey}_{hash}.svg
		cacheFilename := fmt.Sprintf("%s_%s_%s.svg", mod.ModuleName, viewKey, mod.DSLHash)
		cachePath := filepath.Join(cacheDir, cacheFilename)

		// Copy to cache
		if cpErr := CopyFile(svgPath, cachePath); cpErr != nil {
			return nil, fmt.Errorf("copying SVG to cache: %w", cpErr)
		}

		entries = append(entries, StructurizrIndexEntry{
			Module:      mod.ModuleName,
			ViewKey:     viewKey,
			DSLHash:     mod.DSLHash,
			SVGFilename: cacheFilename,
		})
	}

	return entries, nil
}
