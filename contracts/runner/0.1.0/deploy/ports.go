// Package deploy defines the port interface for deploy handler adapters.
// Adapters implement DeployerPort; the execution engine (go/core/tool.DeployBridge)
// stores and dispatches them directly.
//
// # Convention for new adapters
//
// Place deploy handlers in the adapter package:
//
//	go/adapters/<name>/deployer.go — implements deploy.DeployerPort
//
// Register with the global bridge from init():
//
//	func init() {
//	    tool.GlobalDeployBridge().RegisterNativeHandler(&MyDeployer{})
//	}
//
// # Options parameter
//
// The opts parameter in Deploy/DryRun is typed as any to avoid importing
// implementation-specific option types from go/core/tool.
// The execution engine always passes a valid tool.DeployOptions, so the
// blank-identifier idiom is safe in production.
package deploy

import (
	"context"
	"io"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// DeployerPort is the interface that deploy handler adapters implement.
// Each adapter handles one or more component types (e.g., bicep, helm, terraform).
type DeployerPort interface {
	// Name returns the handler identifier (e.g., "az-bicep", "helm", "terraform").
	Name() string

	// Deploy executes the deployment for a module to a target environment.
	// Returns 0 for success, non-zero for failure.
	//
	// workspaceRoot is the absolute path to the repository root.
	// outputDir is the pre-created output directory for deploy evidence.
	// logWriter receives all deploy output (stdout + stderr).
	// opts is tool.DeployOptions; type-assert to access fields.
	Deploy(ctx context.Context, module core.ModuleContractPort, workspaceRoot, outputDir string,
		logWriter io.Writer, opts any) int

	// DryRun executes a what-if deployment (preview changes without applying).
	// Returns 0 for success, non-zero for failure.
	DryRun(ctx context.Context, module core.ModuleContractPort, workspaceRoot, outputDir string,
		logWriter io.Writer, opts any) int

	// Requirements returns names of system binaries required by this handler.
	// Used for pre-flight validation (e.g., ["docker"], ["az"]).
	Requirements() []string

	// ValidateModule checks whether a module is valid for deployment to the
	// specified environment. Returns nil if valid, or a descriptive error.
	ValidateModule(module core.ModuleContractPort, workspaceRoot, environment string) error

	// IsContainer returns true if this handler runs in a Docker container.
	IsContainer() bool

	// IsHostInstalled returns true if this handler uses host-installed tooling.
	IsHostInstalled() bool
}
