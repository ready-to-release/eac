// mkdocs_adapter.go - Adapts mkdocs package handlers to the builders.Handler interface
package builders

import (
	"io"
	"os"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/builders/mkdocs"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	// Register all mkdocs handlers in BOTH registries:
	// 1. builders.handlers - for legacy code paths (GetHandlersForModule)
	// 2. tool.BuildBridge.nativeHandlers - for component resolver (ResolveForBuild)
	preprocess := &mkdocsPreprocessAdapter{}
	site := &mkdocsSiteAdapter{}
	pdf := &mkdocsPDFAdapter{}

	// Register in builders registry (legacy)
	RegisterHandler(preprocess)
	RegisterHandler(site)
	RegisterHandler(pdf)

	// Register in tool bridge (for component resolver)
	tool.GlobalBuildBridge().RegisterNativeHandler(preprocess)
	tool.GlobalBuildBridge().RegisterNativeHandler(site)
	tool.GlobalBuildBridge().RegisterNativeHandler(pdf)
}

// mkdocsPreprocessAdapter adapts mkdocs.PreprocessHandler to builders.Handler.
// It lazily creates the actual handler when Build is called.
type mkdocsPreprocessAdapter struct {
	handler *mkdocs.PreprocessHandler
}

func (a *mkdocsPreprocessAdapter) Name() string { return "mkdocs-preprocess" }

func (a *mkdocsPreprocessAdapter) Requirements() []string { return nil }

func (a *mkdocsPreprocessAdapter) ValidateModule(module interfaces.ModuleContractPort, workspaceRoot, component string) error {
	a.ensureHandler(workspaceRoot)
	return a.handler.ValidateModule(module, workspaceRoot, component)
}

func (a *mkdocsPreprocessAdapter) ListArtifacts(module interfaces.ModuleContractPort, workspaceRoot string) []string {
	a.ensureHandler(workspaceRoot)
	return a.handler.ListArtifacts(module, workspaceRoot)
}

func (a *mkdocsPreprocessAdapter) IsContainer() bool { return false }

func (a *mkdocsPreprocessAdapter) IsHostInstalled() bool { return true }

func (a *mkdocsPreprocessAdapter) Build(module interfaces.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	a.ensureHandler(workspaceRoot)

	// Convert BuildOptions to mkdocs.BuildOptions
	mkdocsOpts := mkdocs.BuildOptions{
		Component:    opts.Component,
		Force:        opts.ForceRebuild || isForceRebuildCLI(),
		Reproducible: opts.Reproducible,
		PDFMode:      false, // Determined from book config
		Tidy:         opts.TidyFirst,
		Weight:       opts.Weight, // Pass weight for container resource scaling
	}

	result := a.handler.Build(module, workspaceRoot, outputDir, logWriter, mkdocsOpts)
	return result.ExitCode
}

func (a *mkdocsPreprocessAdapter) ensureHandler(workspaceRoot string) {
	if a.handler == nil {
		a.handler = mkdocs.NewPreprocessHandler(workspaceRoot)
	}
}

// isForceRebuildCLI checks if force rebuild is requested via CLI flags.
func isForceRebuildCLI() bool {
	for _, arg := range os.Args {
		if arg == "--force" || arg == "-f" || arg == "--skip-cache" {
			return true
		}
	}
	return false
}

// ============================================================================
// site-render-tool adapter
// ============================================================================

// mkdocsSiteAdapter adapts mkdocs.SiteRenderHandler to builders.Handler.
type mkdocsSiteAdapter struct {
	handler *mkdocs.SiteRenderHandler
}

func (a *mkdocsSiteAdapter) Name() string { return "site-render-tool" }

func (a *mkdocsSiteAdapter) Requirements() []string { return []string{"docker"} }

func (a *mkdocsSiteAdapter) ValidateModule(module interfaces.ModuleContractPort, workspaceRoot, component string) error {
	a.ensureHandler(workspaceRoot)
	return a.handler.ValidateModule(module, workspaceRoot, component)
}

func (a *mkdocsSiteAdapter) ListArtifacts(module interfaces.ModuleContractPort, workspaceRoot string) []string {
	a.ensureHandler(workspaceRoot)
	return a.handler.ListArtifacts(module, workspaceRoot)
}

func (a *mkdocsSiteAdapter) IsContainer() bool { return true }

func (a *mkdocsSiteAdapter) IsHostInstalled() bool { return false }

func (a *mkdocsSiteAdapter) Build(module interfaces.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	a.ensureHandler(workspaceRoot)

	mkdocsOpts := mkdocs.BuildOptions{
		Component:    opts.Component,
		Force:        opts.ForceRebuild || isForceRebuildCLI(),
		Reproducible: opts.Reproducible,
		Weight:       opts.Weight, // Pass weight for container resource scaling
	}

	result := a.handler.Build(module, workspaceRoot, outputDir, logWriter, mkdocsOpts)
	return result.ExitCode
}

func (a *mkdocsSiteAdapter) ensureHandler(workspaceRoot string) {
	if a.handler == nil {
		a.handler = mkdocs.NewSiteRenderHandler(workspaceRoot)
	}
}

// ============================================================================
// pdf-render-tool adapter
// ============================================================================

// mkdocsPDFAdapter adapts mkdocs.PDFRenderHandler to builders.Handler.
type mkdocsPDFAdapter struct {
	handler *mkdocs.PDFRenderHandler
}

func (a *mkdocsPDFAdapter) Name() string { return "pdf-render-tool" }

func (a *mkdocsPDFAdapter) Requirements() []string { return []string{"docker"} }

func (a *mkdocsPDFAdapter) ValidateModule(module interfaces.ModuleContractPort, workspaceRoot, component string) error {
	a.ensureHandler(workspaceRoot)
	return a.handler.ValidateModule(module, workspaceRoot, component)
}

func (a *mkdocsPDFAdapter) ListArtifacts(module interfaces.ModuleContractPort, workspaceRoot string) []string {
	a.ensureHandler(workspaceRoot)
	return a.handler.ListArtifacts(module, workspaceRoot)
}

func (a *mkdocsPDFAdapter) IsContainer() bool { return true }

func (a *mkdocsPDFAdapter) IsHostInstalled() bool { return false }

func (a *mkdocsPDFAdapter) Build(module interfaces.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	a.ensureHandler(workspaceRoot)

	// Extract theme from component name if present (e.g., "tutorials-dark" -> theme="dark")
	theme := "dark" // default
	component := opts.Component
	if idx := len(component) - 5; idx > 0 && (component[idx:] == "-dark" || component[idx:] == "-light") {
		theme = component[idx+1:]
		component = component[:idx]
	}

	mkdocsOpts := mkdocs.BuildOptions{
		Component:    component,
		Force:        opts.ForceRebuild || isForceRebuildCLI(),
		Reproducible: opts.Reproducible,
		Weight:       opts.Weight, // Pass weight for container resource scaling
		Metadata: map[string]string{
			"theme": theme,
		},
	}

	result := a.handler.Build(module, workspaceRoot, outputDir, logWriter, mkdocsOpts)
	return result.ExitCode
}

func (a *mkdocsPDFAdapter) ensureHandler(workspaceRoot string) {
	if a.handler == nil {
		a.handler = mkdocs.NewPDFRenderHandler(workspaceRoot)
	}
}
