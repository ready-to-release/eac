// types.go - Type aliases for build handler types.
// The canonical definitions live in go/core/tool/handler_adapter.go.
// These aliases allow handlers in this package to reference the types
// without qualifying every usage with the tool package prefix.
package builders

import (
	"github.com/ready-to-release/eac/go/core/tool"
)

// Handler is the interface for build handlers.
// Canonical definition: tool.BuildHandler in go/core/tool/handler_adapter.go.
type Handler = tool.BuildHandler

// BuildOptions contains flags for controlling the build process.
// Canonical definition: tool.BuildOptions in go/core/tool/handler_adapter.go.
type BuildOptions = tool.BuildOptions
