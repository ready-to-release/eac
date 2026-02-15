// go_artifacts.go - Artifact listing for Go modules.
//
// Determines which artifacts a Go module will produce based on per-module
// artifact definitions in repository.yml.
package builders

import (
	"runtime"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

// listGoModuleArtifacts returns the artifacts that would be produced by building this Go module.
func listGoModuleArtifacts(module *modules.ModuleContract, workspaceRoot string) []string {
	hasGoModule := module.HasComponent("go")

	if !hasGoModule {
		return nil
	}

	// Check for per-module artifact definitions
	if module.HasBuildArtifacts() {
		return listModuleArtifacts(module)
	}

	// No artifacts = library (compile-only verification)
	// The UoW manifest itself proves the build succeeded (exit_code, timestamp, hash)
	return nil
}

// listModuleArtifacts returns artifacts based on per-module definitions.
func listModuleArtifacts(module *modules.ModuleContract) []string {
	var artifacts []string

	for _, artifact := range module.GetBuildArtifacts() {
		switch artifact.Type {
		case "executable":
			// Resolve pattern with current platform
			resolver := config.NewArtifactResolverWithPlatform(module.Moniker, "", runtime.GOOS, runtime.GOARCH)
			name := resolver.ResolvePattern(artifact.Pattern)
			artifacts = append(artifacts, name)
		case "test":
			// Test artifact
			artifacts = append(artifacts, artifact.Pattern)
		default:
			// Other artifact types - resolve as-is
			resolver := config.NewArtifactResolver(module.Moniker, "")
			name := resolver.ResolvePattern(artifact.Pattern)
			artifacts = append(artifacts, name)
		}
	}

	// Add checksums for cross-platform builds (multiple executables)
	execCount := 0
	for _, a := range module.GetBuildArtifacts() {
		if a.Type == "executable" {
			execCount++
		}
	}
	if execCount > 1 {
		artifacts = append(artifacts, "checksums.txt")
	}

	return artifacts
}
