// mkdocs_adapter.go - Adapts mkdocs package handlers to the builders.Handler interface
package builders

import (
	"io"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	build "github.com/ready-to-release/eac/contracts/runner/0.1.0/build"
	"github.com/ready-to-release/eac/go/commands/build/builders/mkdocs"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	tool.GlobalBuildBridge().RegisterNativeHandler(&mkdocsPreprocessAdapter{})
	tool.GlobalBuildBridge().RegisterNativeHandler(&mkdocsSiteAdapter{})
	tool.GlobalBuildBridge().RegisterNativeHandler(&mkdocsPDFAdapter{})
}

// mkdocsPreprocessAdapter adapts mkdocs.PreprocessHandler to builders.Handler.
// It lazily creates the actual handler when Build is called.
type mkdocsPreprocessAdapter struct {
	handler *mkdocs.PreprocessHandler
}

func (a *mkdocsPreprocessAdapter) Name() string { return "mkdocs-preprocess" }

func (a *mkdocsPreprocessAdapter) Requirements() []string { return nil }

func (a *mkdocsPreprocessAdapter) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	a.ensureHandler(workspaceRoot)
	return a.handler.ValidateModule(module, workspaceRoot, component)
}

func (a *mkdocsPreprocessAdapter) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	a.ensureHandler(workspaceRoot)
	return a.handler.ListArtifacts(module, workspaceRoot)
}

func (a *mkdocsPreprocessAdapter) IsContainer() bool { return false }

func (a *mkdocsPreprocessAdapter) IsHostInstalled() bool { return true }

func (a *mkdocsPreprocessAdapter) Build(module core.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, rawOpts any) int {
	opts, _ := rawOpts.(BuildOptions)
	a.ensureHandler(workspaceRoot)

	// Convert BuildOptions to mkdocs.BuildOptions
	mkdocsOpts := mkdocs.BuildOptions{
		Component:     opts.Component,
		Force:         opts.ForceRebuild || isForceRebuildCLI(),
		Reproducible:  opts.Reproducible,
		PDFMode:       false, // Determined from book config
		Tidy:          opts.TidyFirst,
		Weight:        opts.Weight,         // Pass weight for container resource scaling
		ArtifactsMode: opts.ArtifactsMode,  // Pass artifacts mode for PDF page limits
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
// mkdocs-render-oci adapter
// ============================================================================

// mkdocsSiteAdapter adapts mkdocs.SiteRenderHandler to builders.Handler.
type mkdocsSiteAdapter struct {
	handler *mkdocs.SiteRenderHandler
}

func (a *mkdocsSiteAdapter) Name() string { return "mkdocs-render-oci" }

func (a *mkdocsSiteAdapter) Requirements() []string { return []string{"docker"} }

func (a *mkdocsSiteAdapter) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	a.ensureHandler(workspaceRoot)
	return a.handler.ValidateModule(module, workspaceRoot, component)
}

func (a *mkdocsSiteAdapter) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	a.ensureHandler(workspaceRoot)
	return a.handler.ListArtifacts(module, workspaceRoot)
}

func (a *mkdocsSiteAdapter) IsContainer() bool { return true }

func (a *mkdocsSiteAdapter) IsHostInstalled() bool { return false }

func (a *mkdocsSiteAdapter) Build(module core.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, rawOpts any) int {
	opts, _ := rawOpts.(BuildOptions)
	a.ensureHandler(workspaceRoot)

	mkdocsOpts := mkdocs.BuildOptions{
		Component:     opts.Component,
		Force:         opts.ForceRebuild || isForceRebuildCLI(),
		Reproducible:  opts.Reproducible,
		Weight:        opts.Weight,         // Pass weight for container resource scaling
		ArtifactsMode: opts.ArtifactsMode,  // Pass artifacts mode for PDF page limits
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
// pdf-oci adapter
// ============================================================================

// mkdocsPDFAdapter adapts mkdocs.PDFRenderHandler to builders.Handler.
type mkdocsPDFAdapter struct {
	handler *mkdocs.PDFRenderHandler
}

func (a *mkdocsPDFAdapter) Name() string { return "pdf-oci" }

func (a *mkdocsPDFAdapter) Requirements() []string { return []string{"docker"} }

func (a *mkdocsPDFAdapter) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	a.ensureHandler(workspaceRoot)
	return a.handler.ValidateModule(module, workspaceRoot, component)
}

func (a *mkdocsPDFAdapter) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	a.ensureHandler(workspaceRoot)
	return a.handler.ListArtifacts(module, workspaceRoot)
}

func (a *mkdocsPDFAdapter) IsContainer() bool { return true }

func (a *mkdocsPDFAdapter) IsHostInstalled() bool { return false }

func (a *mkdocsPDFAdapter) Build(module core.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, rawOpts any) int {
	opts, _ := rawOpts.(BuildOptions)
	a.ensureHandler(workspaceRoot)

	// Extract theme from component name if present (e.g., "tutorials-dark" -> theme="dark")
	theme := "dark" // default
	component := opts.Component
	if idx := len(component) - 5; idx > 0 && (component[idx:] == "-dark" || component[idx:] == "-light") {
		theme = component[idx+1:]
		component = component[:idx]
	}

	mkdocsOpts := mkdocs.BuildOptions{
		Component:     component,
		Force:         opts.ForceRebuild || isForceRebuildCLI(),
		Reproducible:  opts.Reproducible,
		Weight:        opts.Weight,         // Pass weight for container resource scaling
		ArtifactsMode: opts.ArtifactsMode,  // Pass artifacts mode for PDF page limits
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

var _ build.BuilderPort = (*mkdocsPreprocessAdapter)(nil)
var _ build.BuilderPort = (*mkdocsSiteAdapter)(nil)
var _ build.BuilderPort = (*mkdocsPDFAdapter)(nil)
