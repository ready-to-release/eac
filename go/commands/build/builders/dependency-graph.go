// dependency-graph.go - Build handler for rendering repository dependency graph as PlantUML SVG.
package builders

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/build/docprep/diagrams"
	"github.com/ready-to-release/eac/go/clibase/caching/itemcache"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	tool.GlobalBuildBridge().RegisterNativeHandler(&DependencyGraphHandler{})
}

// DependencyGraphHandler generates and renders the repository dependency graph
// as a PlantUML diagram. Uses the existing GetPlantUMLDiagram() function to
// generate the .puml source and renders it via the plantuml container.
// Uses .cache/eac/plantuml/ for incremental builds (shared with plantuml-render).
type DependencyGraphHandler struct{}

func (h *DependencyGraphHandler) Name() string         { return "dependency-graph" }
func (h *DependencyGraphHandler) Requirements() []string { return []string{"docker"} }
func (h *DependencyGraphHandler) IsContainer() bool     { return true }
func (h *DependencyGraphHandler) IsHostInstalled() bool  { return false }

func (h *DependencyGraphHandler) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	if !IsDockerAvailable() {
		if IsDockerInDocker() {
			return fmt.Errorf("Docker socket not mounted")
		}
		return fmt.Errorf("Docker is not available (required for dependency graph rendering)")
	}
	return nil
}

func (h *DependencyGraphHandler) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	return []string{"dependency-graph/"}
}

// DependencyGraphIndex maps the generated dependency graph to its rendered SVG.
type DependencyGraphIndex struct {
	Entries []DependencyGraphIndexEntry `json:"entries"`
}

// DependencyGraphIndexEntry maps a single diagram to its SVG.
type DependencyGraphIndexEntry struct {
	DiagramType string `json:"diagram_type"` // "dependency-graph"
	ContentHash string `json:"content_hash"` // First 8 chars of SHA256 of PlantUML source
	SVGFilename string `json:"svg_filename"` // Filename in the output directory
}

func (h *DependencyGraphHandler) Build(
	module core.ModuleContractPort, workspaceRoot, outputDir string,
	logWriter io.Writer, opts BuildOptions,
) int {
	Logln(logWriter, "=== Generating repository dependency graph ===")

	cacheDir := filepath.Join(paths.PlantUMLAccelCachePath(workspaceRoot), module.GetMoniker(), filepath.Base(outputDir))
	outputGraphDir := filepath.Join(outputDir, "dependency-graph")

	// 1. Load dependency graph and generate PlantUML source
	graph, err := repository.GetModuleDependencyGraph(workspaceRoot)
	if err != nil {
		Logln(logWriter, "Error loading dependency graph: %v", err)
		return 1
	}

	pumlSource := repository.GetPlantUMLDiagram(graph)
	if pumlSource == "" {
		Logln(logWriter, "No dependency graph to render")
		_ = os.MkdirAll(outputGraphDir, 0o755)
		return 0
	}

	// 2. Hash the PlantUML source for cache key
	contentHash := diagrams.HashPlantUMLContent(pumlSource)
	cacheFilename := fmt.Sprintf("dependency-graph_%s.svg", contentHash)

	Logln(logWriter, "Dependency graph: %d modules, %d edges (hash: %s)",
		len(graph.Modules), len(graph.Edges), contentHash)

	// 3. Create single-item cache
	items := []itemcache.Item{
		{
			Key:           "dependency-graph",
			ContentHash:   contentHash,
			CacheFilename: cacheFilename,
			OutputRelPath: "dependency-graph.svg",
		},
	}

	// 4. Build function renders the diagram via plantuml container
	buildFunc := func(misses []itemcache.Item, dir string) (int, error) {
		if len(misses) == 0 {
			return 0, nil
		}
		return h.renderGraph(pumlSource, misses[0].CacheFilename, workspaceRoot, dir, logWriter)
	}

	// 5. Execute with cache
	cache := itemcache.New(cacheDir)
	result, err := cache.Execute(items, outputGraphDir, buildFunc, opts.ForceRebuild)
	if err != nil {
		Logln(logWriter, "Error: %v", err)
		return 1
	}

	// 6. Write index manifest
	manifest := DependencyGraphIndex{
		Entries: []DependencyGraphIndexEntry{
			{
				DiagramType: "dependency-graph",
				ContentHash: contentHash,
				SVGFilename: "dependency-graph.svg",
			},
		},
	}
	manifestPath := filepath.Join(outputGraphDir, "dependency-graph-index.json")
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	_ = os.WriteFile(manifestPath, manifestData, 0o644)

	if result.CacheMisses == 0 {
		Logln(logWriter, "Dependency graph: cached (hash: %s)", contentHash)
	} else {
		Logln(logWriter, "Dependency graph: rendered (hash: %s)", contentHash)
	}
	Logln(logWriter, "Output: %s (1 SVG, 1 index)", outputGraphDir)

	return 0
}

// renderGraph renders the PlantUML source to SVG via the plantuml container.
func (h *DependencyGraphHandler) renderGraph(
	pumlSource, cacheFilename, workspaceRoot, cacheDir string, logWriter io.Writer,
) (int, error) {
	// Create work directory
	workDir := filepath.Join(cacheDir, ".depgraph-work")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return 0, fmt.Errorf("creating work directory: %w", err)
	}
	defer os.RemoveAll(workDir)

	// Write .puml file
	pumlName := strings.TrimSuffix(cacheFilename, ".svg") + ".puml"
	pumlPath := filepath.Join(workDir, pumlName)
	if err := os.WriteFile(pumlPath, []byte(pumlSource), 0o644); err != nil {
		return 0, fmt.Errorf("writing .puml file: %w", err)
	}

	// Execute plantuml container
	bridge := tool.GlobalHandlerToolBridge()
	tc := &tool.ToolContext{
		WorkspaceRoot: workspaceRoot,
		OutputDir:     workDir,
		LogWriter:     logWriter,
	}

	exitCode, execErr := bridge.ExecuteTool(context.Background(), "plantuml", tc)
	if execErr != nil {
		return 0, fmt.Errorf("plantuml tool failed: %w", execErr)
	}
	if exitCode != 0 {
		return 0, fmt.Errorf("plantuml tool exited with code %d", exitCode)
	}

	// Copy rendered SVG to cache
	svgPath := filepath.Join(workDir, strings.TrimSuffix(cacheFilename, ".svg")+".svg")
	cachePath := filepath.Join(cacheDir, cacheFilename)

	if _, statErr := os.Stat(svgPath); statErr != nil {
		return 0, fmt.Errorf("rendered SVG not found at %s", svgPath)
	}

	if cpErr := CopyFile(svgPath, cachePath); cpErr != nil {
		return 0, fmt.Errorf("copying SVG to cache: %w", cpErr)
	}

	return 1, nil
}
