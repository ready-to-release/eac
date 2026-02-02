package mkdocs

import (
	"io"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
)

// Handler defines the interface that mkdocs handlers implement.
// This mirrors the builders.Handler interface for registration.
type Handler interface {
	Name() string
	Build(module interfaces.ModuleContractPort, workspaceRoot, outputDir string,
		logWriter io.Writer, opts BuildOptions) BuildResult
	ListArtifacts(module interfaces.ModuleContractPort, workspaceRoot string) []string
	Requirements() []string
	ValidateModule(module interfaces.ModuleContractPort, workspaceRoot, component string) error
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
		"site-render-tool":       NewSiteRenderHandler(workspaceRoot),
		"pdf-render-tool":        NewPDFRenderHandler(workspaceRoot),
	}
}

// HandlerNames returns the names of all mkdocs handlers.
func HandlerNames() []string {
	return []string{
		"mkdocs-preprocess",
		"site-render-tool",
		"pdf-render-tool",
	}
}
