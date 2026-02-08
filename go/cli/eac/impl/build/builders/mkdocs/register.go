package mkdocs

import (
	"io"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// Handler defines the interface that mkdocs handlers implement.
// This mirrors the builders.Handler interface for registration.
type Handler interface {
	Name() string
	Build(module core.ModuleContractPort, workspaceRoot, outputDir string,
		logWriter io.Writer, opts BuildOptions) BuildResult
	ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string
	Requirements() []string
	ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error
	IsContainer() bool
	IsHostInstalled() bool
}

// HandlerRegistry is a function type for registering handlers.
// This allows the mkdocs package to register handlers without importing
// the builders package (avoiding circular imports).
type HandlerRegistry func(name string, h Handler)

// Handlers returns all mkdocs handlers for registration.
// Call this from the builders package to register all mkdocs handlers.
func Handlers(workspaceRoot string) map[string]Handler {
	return map[string]Handler{
		"mkdocs-preprocess": NewPreprocessHandler(workspaceRoot),
		"mkdocs-render-oci":   NewSiteRenderHandler(workspaceRoot),
		"pdf-oci":           NewPDFRenderHandler(workspaceRoot),
	}
}

// HandlerNames returns the names of all mkdocs handlers.
func HandlerNames() []string {
	return []string{
		"mkdocs-preprocess",
		"mkdocs-render-oci",
		"pdf-oci",
	}
}
