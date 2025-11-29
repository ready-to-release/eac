// dispatch.go - Build function dispatch and type definitions
package build

import (
	"github.com/ready-to-release/eac/src/commands/impl/build/builders"
	"github.com/ready-to-release/eac/src/core/config"
)

// BuildOptions is an alias to the builders package type
type BuildOptions = builders.BuildOptions

// BuildFunc is an alias to the builders package type
type BuildFunc = builders.BuildFunc

// buildFunctions maps module types to their build functions.
// Types not in this map fall back to build system handlers.
var buildFunctions = map[string]BuildFunc{
	// Go family
	"go-cli":           builders.BuildGoCLI,
	"go-commands":      builders.BuildGoCommands,
	"go-mcp":           builders.BuildGoMCP,
	"go-library":       builders.BuildGoLibrary,
	"go-tests":         builders.BuildGoTests,
	"go-r2r-extension": builders.BuildR2RExtension,

	// Documentation
	"mkdocs-site": builders.BuildMkDocsSite,

	// VS Code
	"vscode-ext":    builders.BuildVSCodeExtension,
	"vscode-config": builders.BuildNoop,

	// Scripts
	"scripts-package": builders.BuildScriptsPackage,

	// Configuration
	"configuration": builders.BuildConfig,
	"claude-config": builders.BuildNoop,

	// Repository structure
	"repository-root": builders.BuildRepositoryRoot,
	"catch-all":       builders.BuildNoop,
	"templates":       builders.BuildTemplates,
}

// buildSystemHandlers maps build systems to default build functions.
// Used when no type-specific handler exists in buildFunctions.
var buildSystemHandlers = map[string]BuildFunc{
	"go":     builders.BuildGoDefault,
	"mkdocs": builders.BuildMkDocsDefault,
	"docker": builders.BuildDockerDefault,
	"vscode": builders.BuildVSCodeDefault,
	"none":   builders.BuildNoop,
}

// GetBuildFunc returns the appropriate build function for a module type.
// It first checks for a type-specific handler, then falls back to build system handlers.
func GetBuildFunc(moduleType string) BuildFunc {
	// First, check for type-specific handler
	if fn, exists := buildFunctions[moduleType]; exists {
		return fn
	}

	// Fall back to build system handler from type registry
	cfg := config.Global()
	if cfg != nil && cfg.ModuleTypes != nil {
		buildSystem := cfg.ModuleTypes.GetBuildSystem(moduleType)
		if fn, exists := buildSystemHandlers[buildSystem]; exists {
			return fn
		}
	}

	// Ultimate fallback: no-op build
	return builders.BuildNoop
}

// IsGoModuleType returns true if the module type uses Go tooling (has go_module capability)
func IsGoModuleType(moduleType string) bool {
	cfg := config.Global()
	if cfg != nil && cfg.ModuleTypes != nil {
		return cfg.ModuleTypes.HasCapability(moduleType, "go_module")
	}
	return false
}
