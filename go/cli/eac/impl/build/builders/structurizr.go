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

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	h := &StructurizrRenderHandler{}
	RegisterHandler(h)
	tool.GlobalBuildBridge().RegisterNativeHandler(h)
}

// StructurizrRenderHandler renders Structurizr architecture diagrams to SVG.
// It renders the current module's workspace.dsl, exports views via
// structurizr-cli + plantuml containers, and writes SVGs + index manifest.
// Uses .cache/eac/structurizr/ for incremental builds.
type StructurizrRenderHandler struct{}

func (h *StructurizrRenderHandler) Name() string         { return "structurizr-render" }
func (h *StructurizrRenderHandler) Requirements() []string { return []string{"docker"} }
func (h *StructurizrRenderHandler) IsContainer() bool     { return true }
func (h *StructurizrRenderHandler) IsHostInstalled() bool  { return false }

func (h *StructurizrRenderHandler) ValidateModule(module interfaces.ModuleContractPort, workspaceRoot, component string) error {
	if !IsDockerAvailable() {
		if IsDockerInDocker() {
			return fmt.Errorf("Docker socket not mounted")
		}
		return fmt.Errorf("Docker is not available (required for structurizr rendering)")
	}
	return nil
}

func (h *StructurizrRenderHandler) ListArtifacts(module interfaces.ModuleContractPort, workspaceRoot string) []string {
	return []string{"structurizr/"}
}

// StructurizrIndex maps source workspace.dsl modules to their rendered SVG filenames.
type StructurizrIndex struct {
	Entries []StructurizrIndexEntry `json:"entries"`
}

// StructurizrIndexEntry maps a single Structurizr view to its SVG file.
type StructurizrIndexEntry struct {
	Module      string `json:"module"`       // Module name (e.g., "eac-cli")
	ViewKey     string `json:"view_key"`     // View key from DSL (e.g., "SystemContext")
	DSLHash     string `json:"dsl_hash"`     // First 8 chars of SHA256 of workspace.dsl
	SVGFilename string `json:"svg_filename"` // Filename in the output directory
}

// structurizrModuleWork tracks the state of a single module during build.
type structurizrModuleWork struct {
	ModuleName    string
	WorkspacePath string // Absolute path to workspace.dsl
	DSLHash       string
	CacheHit      bool
}

func (h *StructurizrRenderHandler) Build(
	module interfaces.ModuleContractPort, workspaceRoot, outputDir string,
	logWriter io.Writer, opts BuildOptions,
) int {
	moduleName := module.GetMoniker()
	Logln(logWriter, "=== Rendering structurizr diagrams for %s via container ===", moduleName)

	specsDir := filepath.Join(workspaceRoot, "specs")
	cacheDir := paths.StructurizrAccelCachePath(workspaceRoot)
	outputStructurizrDir := filepath.Join(outputDir, "structurizr")

	// Locate this module's workspace.dsl
	workspacePath := filepath.Join(specsDir, moduleName, ".design", "workspace.dsl")
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

	dslHash := hashDSLContent(string(content))
	mod := structurizrModuleWork{
		ModuleName:    moduleName,
		WorkspacePath: workspacePath,
		DSLHash:       dslHash,
	}

	// Ensure directories exist
	_ = os.MkdirAll(cacheDir, 0o755)
	_ = os.MkdirAll(outputStructurizrDir, 0o755)

	// Check cache and render
	var allEntries []StructurizrIndexEntry
	cacheHit := false

	cachedSVGs := findCachedSVGs(cacheDir, mod.ModuleName, mod.DSLHash)
	if len(cachedSVGs) > 0 && !opts.ForceRebuild {
		// Cache hit - copy to output
		for _, svgPath := range cachedSVGs {
			filename := filepath.Base(svgPath)
			outputPath := filepath.Join(outputStructurizrDir, filename)
			if cpErr := CopyFile(svgPath, outputPath); cpErr != nil {
				Logln(logWriter, "Error copying cached SVG %s: %v", filename, cpErr)
				return 1
			}
			viewKey := extractViewKeyFromCacheFilename(filename, mod.ModuleName)
			allEntries = append(allEntries, StructurizrIndexEntry{
				Module:      mod.ModuleName,
				ViewKey:     viewKey,
				DSLHash:     mod.DSLHash,
				SVGFilename: filename,
			})
		}
		cacheHit = true
	} else {
		// Cache miss - render via two-step Docker pipeline
		Logln(logWriter, "Rendering %s (hash: %s)...", mod.ModuleName, mod.DSLHash)
		views, renderErr := renderStructurizrModule(mod, workspaceRoot, specsDir, cacheDir, logWriter)
		if renderErr != nil {
			Logln(logWriter, "Error rendering %s: %v", mod.ModuleName, renderErr)
			return 1
		}

		// Copy rendered SVGs to output
		for _, view := range views {
			outputPath := filepath.Join(outputStructurizrDir, view.SVGFilename)
			cachePath := filepath.Join(cacheDir, view.SVGFilename)
			if cpErr := CopyFile(cachePath, outputPath); cpErr != nil {
				Logln(logWriter, "Error copying rendered SVG %s: %v", view.SVGFilename, cpErr)
				return 1
			}
			allEntries = append(allEntries, view)
		}
	}

	// Write index manifest
	manifest := StructurizrIndex{Entries: allEntries}
	manifestPath := filepath.Join(outputStructurizrDir, "structurizr-index.json")
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	_ = os.WriteFile(manifestPath, manifestData, 0o644)

	if cacheHit {
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

	// Remove old cached SVGs for this module (different hash)
	removeOldCachedSVGs(cacheDir, mod.ModuleName, mod.DSLHash)

	return entries, nil
}

// removeOldCachedSVGs removes cached SVGs for a module that don't match the current hash.
func removeOldCachedSVGs(cacheDir, moduleName, currentHash string) {
	pattern := filepath.Join(cacheDir, moduleName+"_*.svg")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	suffix := "_" + currentHash + ".svg"
	for _, match := range matches {
		if !strings.HasSuffix(match, suffix) {
			os.Remove(match)
		}
	}
}
