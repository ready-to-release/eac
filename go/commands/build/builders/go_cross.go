// go_cross.go - Cross-compilation for Go modules.
//
// Builds binaries for multiple GOOS/GOARCH targets from per-module artifact
// definitions. Handles target filtering by artifacts mode, ldflags injection,
// and derived artifact skipping (e.g. UPX).
package builders

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/environments"
)

// buildTarget describes a single cross-compilation target parsed from artifact definitions.
type buildTarget struct {
	goos        string
	goarch      string
	pattern     string
	compression string
	deriveFrom  string
}

// buildCrossCompiledFromArtifacts builds binaries for multiple platforms from per-module artifact definitions.
func buildCrossCompiledFromArtifacts(module *modules.ModuleContract, moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions, artifacts []config.ModuleArtifact) int {
	targets := parseBuildTargets(artifacts)

	if len(targets) == 0 {
		Logln(logWriter, "❌ No valid platform targets found in artifacts")
		return 1
	}

	// Filter targets based on artifacts mode (reduced vs all)
	targets = filterTargetsByMode(targets, opts.ArtifactsMode, logWriter)
	if len(targets) == 0 {
		return 0 // filtered to nothing, not an error
	}

	// Build ldflags for version injection
	ldflags := buildLdflags(module, moduleRoot, workspaceRoot, opts.Version, logWriter)

	successCount := 0
	for _, target := range targets {
		// Skip derived artifacts (UPX) - they're processed after base builds
		if target.deriveFrom != "" {
			continue
		}

		if exitCode := buildOneTarget(target, module, moduleRoot, workspaceRoot, outputDir, logWriter, ldflags); exitCode == 0 {
			successCount++
		}
	}

	if successCount == 0 {
		Logln(logWriter, "❌ All builds failed")
		return 1
	}

	Logln(logWriter, "✅ Built %d/%d targets successfully", successCount, len(targets))

	// Generate checksums
	generateChecksums(outputDir, module.Moniker, logWriter)

	return 0
}

// parseBuildTargets extracts platform targets from artifact IDs.
func parseBuildTargets(artifacts []config.ModuleArtifact) []buildTarget {
	var targets []buildTarget
	for _, artifact := range artifacts {
		// Parse platform from artifact ID (e.g., "linux-amd64", "windows-amd64-upx")
		parts := strings.Split(artifact.ID, "-")
		if len(parts) < 2 {
			continue
		}
		goos := parts[0]
		goarch := parts[1]

		targets = append(targets, buildTarget{
			goos:        goos,
			goarch:      goarch,
			pattern:     artifact.Pattern,
			compression: artifact.Compression,
			deriveFrom:  artifact.DeriveFrom,
		})
	}
	return targets
}

// filterTargetsByMode filters targets based on the artifacts mode (reduced vs all).
// Returns the filtered list. Logs when targets are reduced. Returns nil if all
// targets were filtered out (caller should treat as skip, not error).
func filterTargetsByMode(targets []buildTarget, mode environments.ArtifactsMode, logWriter io.Writer) []buildTarget {
	if !mode.IsValid() {
		return targets
	}

	var filtered []buildTarget
	for _, target := range targets {
		if mode.ShouldBuildTarget(target.goos, target.goarch, target.compression) {
			filtered = append(filtered, target)
		}
	}
	if len(filtered) == 0 {
		Logln(logWriter, "⚠️  No targets match artifacts mode '%s' - skipping build", mode)
		return nil
	}
	if len(filtered) < len(targets) {
		Logln(logWriter, "📦 Artifacts mode '%s': building %d/%d targets", mode, len(filtered), len(targets))
	}
	return filtered
}

// buildOneTarget compiles a single GOOS/GOARCH target binary.
func buildOneTarget(target buildTarget, module *modules.ModuleContract, moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, ldflags string) int {
	resolver := config.NewArtifactResolverWithPlatform(module.Moniker, "", target.goos, target.goarch)
	outputName := resolver.ResolvePattern(target.pattern)
	outputPath := filepath.Join(outputDir, outputName)

	Logln(logWriter, "Building: %s/%s → %s", target.goos, target.goarch, outputName)

	args := []string{"build", "-o", outputPath}
	if ldflags != "" {
		args = append(args, "-ldflags", ldflags)
	}

	// Use tool system with env overrides for cross-compilation
	exitCode := executeGoToolWithEnv(moduleRoot, workspaceRoot, logWriter, args, target.goos, target.goarch)
	if exitCode != 0 {
		Logln(logWriter, "❌ Failed to build %s/%s", target.goos, target.goarch)
	}
	return exitCode
}
