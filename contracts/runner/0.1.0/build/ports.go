// Package build defines the port interface for build handler adapters.
// Adapters implement BuilderPort; the execution engine (go/core/tool.BuildBridge)
// stores and dispatches them directly.
//
// # Convention for new adapters
//
// Place build handlers alongside test runners in the adapter package:
//
//	go/adapters/<name>/runner.go   — implements test.TestRunnerPort
//	go/adapters/<name>/builder.go  — implements build.BuilderPort
//
// Existing handlers in go/commands/build/builders/ implement BuilderPort directly.
// Register with the global bridge from init():
//
//	func init() {
//	    tool.GlobalBuildBridge().RegisterNativeHandler(&MyHandler{})
//	}
//
// # Options parameter
//
// The opts parameter in Build is typed as any to avoid importing
// implementation-specific option types from go/core/tool or go/core/cache.
// The execution engine always passes a valid tool.BuildOptions, so the
// blank-identifier idiom is safe in production. Use the two-return form
// in tests to catch missing opts early:
//
//	func (h *MyHandler) Build(m core.ModuleContractPort, root, out string, w io.Writer, opts any) int {
//	    o, ok := opts.(tool.BuildOptions)
//	    if !ok {
//	        return 1 // wrong type passed — programming error
//	    }
//	    // use o.Component, o.Version, etc.
//	}
package build

import (
	"io"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// BuilderPort is the interface that build handler adapters implement.
// Each adapter handles one or more component types.
type BuilderPort interface {
	// Name returns the handler identifier (e.g., "go", "dotnet", "pac").
	// This name is used for registration and lookup in the build bridge.
	Name() string

	// Build executes the build for a module and returns an exit code.
	// Returns 0 for success, non-zero for failure.
	//
	// workspaceRoot is the absolute path to the repository root.
	// outputDir is the pre-created output directory for this build.
	// logWriter receives all build output (stdout + stderr).
	// opts is tool.BuildOptions from go/core/tool; type-assert to access fields.
	Build(module core.ModuleContractPort, workspaceRoot, outputDir string,
		logWriter io.Writer, opts any) int

	// ListArtifacts returns artifact paths produced by a successful Build.
	// Paths are relative to the module output directory.
	ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string

	// Requirements returns names of system binaries required by this handler.
	// Used for pre-flight validation (e.g., ["dotnet"], ["go", "docker"]).
	Requirements() []string

	// ValidateModule checks whether a module's configuration is valid for
	// a specific component. Returns nil if valid, or a descriptive error.
	ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error

	// IsContainer returns true if this handler runs builds in a Docker container.
	IsContainer() bool

	// IsHostInstalled returns true if this handler uses host-installed tooling.
	IsHostInstalled() bool
}
