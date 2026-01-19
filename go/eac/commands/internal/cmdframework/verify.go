package cmdframework

import (
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/initsummary"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

// detectExecutionContext returns a human-readable execution context string.
// Combines container detection (r2r vs implicit) with environment (CI vs devbox).
// Returns: "implicit-cli (devbox)", "implicit-cli (CI)", "r2r-cli (devbox)", "r2r-cli (CI)".
func detectExecutionContext() string {
	// CLI mode: r2r-cli (container) or implicit-cli (local)
	cliMode := string(logging.GetExecutionContext())

	// Environment: CI or devbox
	env := "devbox"
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		env = "CI"
	}

	return cliMode + " (" + env + ")"
}

// phaseVerify handles the verification phase:
// - Build init summary (or enhance existing one from hooks)
// - System dependency verification (via hook)
// - Module dependency validation (via hook).
func phaseVerify(ctx *ExecutionContext) error {
	// Use existing summary if set by a hook (e.g., test framework), otherwise create new one
	summary := ctx.InitSummary
	if summary == nil {
		summary = initsummary.New(string(ctx.Config.Type))
		summary.RequestedModules = ctx.Config.Monikers
		summary.CalculatedModules = ctx.GetExecutionMonikers()
		summary.AddedDepm = ctx.GetAddedDependencies()
	}

	// Always set/update execution context
	summary.ExecutionContext = detectExecutionContext()

	// Execution plan info
	if ctx.ExecutionPlan != nil {
		summary.ExecutionLayers = ctx.ExecutionPlan.Layers
		summary.LayerCount = len(ctx.ExecutionPlan.Layers)
	}
	summary.FlatExecution = !ctx.Config.Layered

	// Set flags (merge with any existing flags from hooks)
	summary.Flags.DryRun = ctx.Config.DryRun
	summary.Flags.SkipDeps = ctx.Config.SkipDeps
	summary.Flags.SkipDepm = ctx.Config.SkipDepm
	summary.Flags.ForceRebuild = ctx.Config.ForceRebuild

	// System dependency verification (if hook provided)
	if !ctx.Config.SkipDeps && depsVerifier != nil {
		depsStatus := depsVerifier(ctx)
		if depsStatus != nil {
			summary.DepsStatus = *depsStatus
		}
	} else if ctx.Config.SkipDeps {
		summary.DepsStatus.Skipped = true
	}

	// Module dependency (build artifact) validation (test/scan only)
	// Build command creates artifacts - it doesn't need them to exist beforehand
	// Test/scan commands consume artifacts - they need to verify builds exist
	if ctx.Config.Type != CommandTypeBuild && !ctx.Config.SkipDepm && artifactValidator != nil {
		artifactInfo := artifactValidator(ctx)
		if artifactInfo != nil {
			summary.ArtifactValidation = artifactInfo
		}
	}

	ctx.InitSummary = summary
	return nil
}

// DepsVerifier is a function that verifies system dependencies.
// Commands provide their own implementation.
type DepsVerifier func(ctx *ExecutionContext) *initsummary.DepsStatus

// ArtifactValidator is a function that validates build artifacts.
// Commands provide their own implementation since artifacts package
// is internal to impl/.
type ArtifactValidator func(ctx *ExecutionContext) *initsummary.ArtifactValidationInfo

var (
	depsVerifier      DepsVerifier
	artifactValidator ArtifactValidator
)

// SetDepsVerifier sets the global system dependency verifier function.
func SetDepsVerifier(v DepsVerifier) {
	depsVerifier = v
}

// SetArtifactValidator sets the global artifact validator function.
func SetArtifactValidator(v ArtifactValidator) {
	artifactValidator = v
}

// displayInitSummary outputs the initialization summary.
func displayInitSummary(ctx *ExecutionContext) {
	if ctx.InitSummary == nil {
		return
	}

	var formatted string
	if ctx.Config.UseTUI {
		formatted = initsummary.FormatCompact(ctx.InitSummary)
	} else {
		formatted = initsummary.FormatDetailed(ctx.InitSummary)
	}

	for _, line := range strings.Split(strings.TrimSpace(formatted), "\n") {
		if line != "" {
			ctx.WriteInit("%s", line)
		}
	}
}
