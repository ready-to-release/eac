package build

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/tool"
)

// ProcessArtifactDerivations handles deriving and compressing artifacts after build.
// artifacts should be the merged artifacts from cfg.GetBuildArtifacts() - this function
// does NOT access component type definitions directly to ensure module-level artifacts take priority.
func ProcessArtifactDerivations(
	moniker string,
	artifacts []config.Artifact,
	buildDir string,
	requestedArtifacts []string,
	metadata map[string]string,
	logWriter io.Writer,
) error {
	if len(artifacts) == 0 {
		return nil
	}

	// Collect all derived artifact variants to process
	type derivedVariant struct {
		artifact config.Artifact
		os       string
		arch     string
	}
	var derivedVariants []derivedVariant

	// Check if all artifacts are requested (wildcard "*")
	allRequested := false
	for _, reqID := range requestedArtifacts {
		if reqID == "*" {
			allRequested = true
			break
		}
	}

	for _, art := range artifacts {
		if !art.IsDerived() {
			continue
		}

		// Check if this artifact was requested (or all are requested via "*")
		isRequested := allRequested
		if !isRequested {
			for _, reqID := range requestedArtifacts {
				if art.ID == reqID {
					isRequested = true
					break
				}
			}
		}
		if !isRequested {
			continue
		}

		// Artifacts from cfg.GetBuildArtifacts are already expanded per-platform
		// Extract OS and arch from artifact ID (format: {os}-{arch} or {os}-{arch}-upx)
		targetOS, targetArch := extractPlatformFromID(art.ID)

		derivedVariants = append(derivedVariants, derivedVariant{
			artifact: art,
			os:       targetOS,
			arch:     targetArch,
		})
	}

	if len(derivedVariants) == 0 {
		return nil
	}

	fmt.Fprintf(logWriter, "\n📦 Processing %d derived artifact(s)\n", len(derivedVariants))

	// Process each derived artifact variant
	for _, variant := range derivedVariants {
		if err := processDerivedArtifact(moniker, variant.artifact, variant.os, variant.arch, buildDir, metadata, logWriter); err != nil {
			return fmt.Errorf("failed to process derived artifact %s (os=%s arch=%s): %w", variant.artifact.Pattern, variant.os, variant.arch, err)
		}
	}

	fmt.Fprintf(logWriter, "✅ Derived artifacts processed successfully\n")
	return nil
}

// extractPlatformFromID extracts OS and architecture from an artifact ID.
// Expected formats: "{os}-{arch}" or "{os}-{arch}-upx"
// Examples: "linux-amd64" -> ("linux", "amd64"), "linux-amd64-upx" -> ("linux", "amd64").
func extractPlatformFromID(id string) (osName, arch string) {
	// Known OS values
	knownOS := []string{"linux", "darwin", "windows"}

	for _, osVal := range knownOS {
		if len(id) > len(osVal) && id[:len(osVal)] == osVal && id[len(osVal)] == '-' {
			remainder := id[len(osVal)+1:]
			// Extract arch (everything before next dash or end)
			archEnd := len(remainder)
			for i, c := range remainder {
				if c == '-' {
					archEnd = i
					break
				}
			}
			return osVal, remainder[:archEnd]
		}
	}
	return "", ""
}

// processDerivedArtifact derives a single artifact from its source.
func processDerivedArtifact(moniker string, art config.Artifact, targetOS, targetArch, buildDir string, metadata map[string]string, logWriter io.Writer) error {
	// Resolve source and target paths with platform-specific resolver including metadata
	var resolver *config.ArtifactResolver
	if targetOS != "" && targetArch != "" {
		resolver = config.NewArtifactResolverFull(moniker, buildDir, targetOS, targetArch, metadata)
	} else {
		resolver = config.NewArtifactResolverWithMetadata(moniker, buildDir, metadata)
	}

	// Resolve source path - check metadata for custom name
	// Create a temporary artifact definition for the source to check metadata overrides
	sourceArtifact := config.Artifact{
		Type:      art.Type,
		Pattern:   art.DeriveFrom,
		ID:        "",
		Platforms: art.Platforms,
	}
	sourceName := resolver.ResolvePatternWithMetadata(art.DeriveFrom, &sourceArtifact)
	sourcePath := filepath.Join(buildDir, sourceName)

	// Resolve target path
	targetName := resolver.ResolvePatternWithMetadata(art.Pattern, &art)
	targetPath := filepath.Join(buildDir, targetName)

	// Verify source exists
	if _, err := os.Stat(sourcePath); err != nil {
		return fmt.Errorf("source artifact not found: %s", sourcePath)
	}

	fmt.Fprintf(logWriter, "  Deriving %s from %s\n", filepath.Base(targetPath), filepath.Base(sourcePath))

	// Copy source to target
	if err := copyFile(sourcePath, targetPath); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	// Make executable if needed
	if art.IsExecutable() {
		if err := os.Chmod(targetPath, 0o755); err != nil {
			return fmt.Errorf("failed to set executable permissions: %w", err)
		}
	}

	// Apply compression if specified
	if art.RequiresCompression() {
		if err := compressArtifact(art, targetPath, logWriter); err != nil {
			return fmt.Errorf("compression failed: %w", err)
		}
	}

	return nil
}

// compressArtifact applies compression to an artifact.
func compressArtifact(art config.Artifact, targetPath string, logWriter io.Writer) error {
	compression := art.GetCompression()

	switch compression {
	case config.CompressionStrip:
		return stripArtifact(targetPath, logWriter)

	case config.CompressionUPX:
		// UPX implies stripping first
		if err := stripArtifact(targetPath, logWriter); err != nil {
			fmt.Fprintf(logWriter, "  ⚠️  Strip before UPX failed: %v\n", err)
		}
		return upxCompressArtifact(targetPath, logWriter)

	default:
		return fmt.Errorf("unsupported compression type: %s", compression)
	}
}

// stripArtifact strips debug symbols from a binary.
func stripArtifact(targetPath string, logWriter io.Writer) error {
	if _, err := exec.LookPath("strip"); err != nil {
		return fmt.Errorf("strip command not found")
	}

	fmt.Fprintf(logWriter, "  Stripping debug symbols: %s\n", filepath.Base(targetPath))

	toolDef := tool.GlobalRegistry().GetOrAdhoc("strip")
	execCtx := &tool.ExecutionContext{
		LogWriter:     logWriter,
		StdoutWriter:  logWriter,
		StderrWriter:  logWriter,
		ArgsOverrides: []string{targetPath},
	}

	result, err := tool.GlobalExecutor().Execute(context.Background(), toolDef, execCtx)
	if err != nil {
		return fmt.Errorf("strip command failed: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("strip command failed with exit code %d", result.ExitCode)
	}

	return nil
}

// upxCompressArtifact compresses a binary with UPX.
func upxCompressArtifact(targetPath string, logWriter io.Writer) error {
	if _, err := exec.LookPath("upx"); err != nil {
		return fmt.Errorf("upx command not found")
	}

	fmt.Fprintf(logWriter, "  Compressing with UPX: %s\n", filepath.Base(targetPath))

	toolDef := tool.GlobalRegistry().GetOrAdhoc("upx")
	execCtx := &tool.ExecutionContext{
		LogWriter:     logWriter,
		StdoutWriter:  logWriter,
		StderrWriter:  logWriter,
		ArgsOverrides: []string{"--best", "--lzma", targetPath},
	}

	result, err := tool.GlobalExecutor().Execute(context.Background(), toolDef, execCtx)
	if err != nil {
		return fmt.Errorf("upx command failed: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("upx command failed with exit code %d", result.ExitCode)
	}

	return nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return dstFile.Sync()
}
