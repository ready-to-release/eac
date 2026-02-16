// site.go - Unified Site build handler
// Combines preprocessing and container rendering in a single handler.
package builders

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/build/builders/mkdocs"
	"github.com/ready-to-release/eac/go/commands/build/docprep"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	tool.GlobalBuildBridge().RegisterNativeHandler(&SiteHandler{})
}

// SiteHandler is a unified handler for docs-site component type.
// It combines preprocessing (native Go) and HTML site rendering (container)
// in a single handler, eliminating the need for tool_chain orchestration.
type SiteHandler struct{}

func (h *SiteHandler) Name() string { return "site" }

func (h *SiteHandler) Capabilities() []string { return []string{"documentation", "html", "container"} }

func (h *SiteHandler) Requirements() []string { return []string{"docker"} }

// IsContainer returns true as site generation requires Docker for the rendering step.
func (h *SiteHandler) IsContainer() bool { return true }

// IsHostInstalled returns false as site generation uses Docker.
func (h *SiteHandler) IsHostInstalled() bool { return false }

// ValidateModule checks if a module has valid book configuration for site generation.
func (h *SiteHandler) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	// Check Docker availability
	if !IsDockerAvailable() {
		if IsDockerInDocker() {
			return fmt.Errorf("Docker socket not mounted")
		}
		return fmt.Errorf("Docker is not available but required for site builds")
	}

	// Load config to check for book configuration
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot, LazyLoad: true})
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := cfg.LoadBooks(false); err != nil {
		return fmt.Errorf("loading books config: %w", err)
	}

	// Resolve book name from component config or naming convention
	bookName := resolveBookNameForSite(module, component)

	// Look up book by name
	book := cfg.GetBookByName(bookName)
	if book != nil {
		return nil
	}

	// Fallback: try module's book list
	moduleBooks := cfg.GetBooksByModule(module.GetMoniker())
	for _, b := range moduleBooks {
		if b.Name == bookName {
			return nil
		}
	}

	return fmt.Errorf("book '%s' not found for site generation (component: %s, module: %s)", bookName, component, module.GetMoniker())
}

// ListArtifacts returns artifact paths that would be produced.
func (h *SiteHandler) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	return []string{"site/"}
}

// Build executes the unified site build: preprocessing + container rendering.
func (h *SiteHandler) Build(module core.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	startTime := time.Now()

	concrete := adapters.UnwrapModule(module)
	if concrete == nil {
		Logln(logWriter, "Error: invalid module type")
		return 1
	}

	// Load book and repository configuration
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot, LazyLoad: true})
	if err != nil {
		Logln(logWriter, "❌ Failed to load config: %v", err)
		return 1
	}
	if err := cfg.LoadRepository(false); err != nil {
		Logln(logWriter, "❌ Failed to load repository config: %v", err)
		return 1
	}
	if err := cfg.LoadBooks(false); err != nil {
		Logln(logWriter, "❌ Failed to load books config: %v", err)
		return 1
	}

	// Resolve book name from component config or naming convention
	bookName := resolveBookNameForSite(module, opts.Component)

	// Look up book
	book := cfg.GetBookByName(bookName)
	if book == nil {
		moduleBooks := cfg.GetBooksByModule(concrete.Moniker)
		for _, b := range moduleBooks {
			if b.Name == bookName {
				book = b
				break
			}
		}
	}
	if book == nil {
		Logln(logWriter, "❌ Book '%s' not found (component: %s, module: %s)", bookName, opts.Component, concrete.Moniker)
		return 1
	}

	Logln(logWriter, "\n=== Building Site: %s/%s ===", concrete.Moniker, bookName)

	// ━━━ Step 1: Preprocessing (native Go) ━━━
	Logln(logWriter, "📝 Running preprocessing...")

	stagingDir := paths.BookStagingCachePath(workspaceRoot, concrete.Moniker, bookName)

	// Clean staging on force rebuild
	if opts.ForceRebuild {
		Logln(logWriter, "   🔄 Force rebuild: clearing staging directory")
		if err := os.RemoveAll(stagingDir); err != nil && !os.IsNotExist(err) {
			Logln(logWriter, "❌ Failed to clean staging directory: %v", err)
			return 1
		}
	}

	if err := os.MkdirAll(stagingDir, 0o755); err != nil {
		Logln(logWriter, "❌ Failed to create staging directory: %v", err)
		return 1
	}

	// Run preprocessing pipeline (site mode)
	pctx := docprep.NewPreprocessContext(context.Background(), book, workspaceRoot, stagingDir, logWriter, docprep.SiteMode{})
	pctx.Moniker = concrete.Moniker
	pctx.DesignComponents = cfg.Repository.DesignComponents()
	pctx.DiagramComponents = cfg.Repository.DiagramComponents(concrete.Moniker)
	if err := docprep.DefaultPipeline().Execute(pctx); err != nil {
		Logln(logWriter, "❌ Preprocessing failed: %v", err)
		return 1
	}

	preprocessDuration := time.Since(startTime)
	Logln(logWriter, "   ✅ Preprocessing complete (%v)", preprocessDuration.Round(time.Millisecond))

	// ━━━ Step 2: Generate mkdocs.yml config ━━━
	configPath := filepath.Join(outputDir, "mkdocs.yml")

	configOpts := mkdocs.ConfigOptions{
		SiteName:     bookName,
		DocsDir:      "/staging", // Container mount path
		OutputFormat: "site",
	}
	if err := mkdocs.WriteMkDocsConfig(workspaceRoot, configPath, configOpts); err != nil {
		Logln(logWriter, "❌ Failed to generate mkdocs.yml: %v", err)
		return 1
	}
	Logln(logWriter, "   Config: %s", configPath)

	// Copy mkdocs macros script for footer generation
	macrosSource := filepath.Join(workspaceRoot, "containers", "mkdocs-render-oci", "mkdocs_macros.py")
	macrosTarget := filepath.Join(outputDir, "main.py")
	if macrosData, err := os.ReadFile(macrosSource); err == nil {
		if err := os.WriteFile(macrosTarget, macrosData, 0o644); err != nil {
			Logln(logWriter, "   ⚠️  Failed to copy mkdocs macros script: %v", err)
		} else {
			Logln(logWriter, "   Macros: %s", macrosTarget)
		}
	}

	// Create output directory
	siteDir := filepath.Join(outputDir, "site")
	if err := os.MkdirAll(siteDir, 0o755); err != nil {
		Logln(logWriter, "❌ Failed to create output directory: %v", err)
		return 1
	}

	// ━━━ Step 3: Invoke mkdocs-render-oci container ━━━
	Logln(logWriter, "📦 Invoking site render container...")

	bridge := tool.GlobalHandlerToolBridge()

	// Get weight for resource scaling
	weight := opts.Weight
	if weight <= 0 {
		weight = 1
	}

	tc := &tool.ToolContext{
		WorkspaceRoot: workspaceRoot,
		StagingDir:    stagingDir,
		OutputDir:     outputDir,
		ConfigPath:    configPath,
		LogWriter:     logWriter,
		Weight:        weight,
	}

	exitCode, execErr := bridge.ExecuteTool(context.Background(), "mkdocs-render-oci", tc)
	if execErr != nil {
		Logln(logWriter, "❌ Tool execution failed: %v", execErr)
		return 1
	}
	if exitCode != 0 {
		Logln(logWriter, "❌ MkDocs site build failed with exit code %d", exitCode)
		return exitCode
	}

	totalDuration := time.Since(startTime)
	Logln(logWriter, "✅ Site build complete")
	Logln(logWriter, "   HTML Output: %s", siteDir)
	Logln(logWriter, "   Duration: %v", totalDuration.Round(time.Millisecond))

	return 0
}

// resolveBookNameForSite determines the book name for a site component.
// Resolution order:
// 1. config.book field (explicit override)
// 2. Strip "-site" suffix if present (e.g., "docs-site" -> "docs")
// 3. Use component name as book name
// 4. Default to "site" if component name is empty
func resolveBookNameForSite(module core.ModuleContractPort, componentName string) string {
	if componentName == "" {
		return "site"
	}

	// Try to get book name from component config
	concrete := adapters.UnwrapModule(module)
	if concrete != nil {
		if comp, ok := concrete.Components[componentName]; ok && comp != nil {
			if bookName, ok := comp.Config["book"]; ok && bookName != "" {
				return bookName
			}
		}
	}

	// Convention: strip "-site" suffix to get book name
	// e.g., "docs-site" -> "docs"
	if len(componentName) > 5 && componentName[len(componentName)-5:] == "-site" {
		return componentName[:len(componentName)-5]
	}

	// Fall back to component name as book name
	return componentName
}
