package build

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/config"
)

// ProcessArtifactDerivations handles deriving and compressing artifacts after build
func ProcessArtifactDerivations(
	moniker string,
	moduleTypeDef *config.ModuleTypeDef,
	buildDir string,
	requestedArtifacts []string,
	metadata map[string]string,
	logWriter io.Writer,
) error {
	if moduleTypeDef.Build == nil {
		return nil
	}

	artifacts := moduleTypeDef.Build.Artifacts

	// Collect all derived artifact variants to process
	type derivedVariant struct {
		artifact config.Artifact
		os       string
		arch     string
	}
	var derivedVariants []derivedVariant

	for _, art := range artifacts {
		if !art.IsDerived() {
			continue
		}

		// For executable artifacts, check all platform/arch combinations
		if art.IsExecutable() && len(art.Platforms) > 0 {
			for _, targetOS := range art.Platforms {
				// Determine supported architectures for this OS
				// Pattern analysis: {moniker}-{os}-amd64-upx{ext} suggests only amd64 is supported for UPX
				var archs []string
				if art.Pattern != "" && (art.Pattern == "{moniker}-{os}-amd64-upx{ext}" ||
					art.Pattern == "{moniker}-{os}-amd64{ext}") {
					archs = []string{"amd64"}
				} else if targetOS == "windows" {
					archs = []string{"amd64"}
				} else {
					archs = []string{"amd64", "arm64"}
				}

				for _, arch := range archs {
					// Derive the artifact ID for this platform
					artifactID := deriveArtifactIDForPlatform(art, targetOS, arch)

					// Check if this variant was requested
					isRequested := false
					for _, reqID := range requestedArtifacts {
						if artifactID == reqID {
							isRequested = true
							break
						}
					}

					if isRequested {
						derivedVariants = append(derivedVariants, derivedVariant{
							artifact: art,
							os:       targetOS,
							arch:     arch,
						})
					}
				}
			}
		} else {
			// Non-executable derived artifacts (rare, but handle generically)
			// Use current platform
			derivedVariants = append(derivedVariants, derivedVariant{
				artifact: art,
				os:       "",
				arch:     "",
			})
		}
	}

	if len(derivedVariants) == 0 {
		return nil
	}

	fmt.Fprintf(logWriter, "\n📦 Processing %d derived artifact(s)\n", len(derivedVariants))

	// Process each derived artifact variant
	for _, variant := range derivedVariants {
		if err := processDerivedArtifact(moniker, variant.artifact, variant.os, variant.arch, buildDir, metadata, logWriter); err != nil {
			return fmt.Errorf("failed to process derived artifact %s: %w", variant.artifact.Pattern, err)
		}
	}

	fmt.Fprintf(logWriter, "✅ Derived artifacts processed successfully\n")
	return nil
}

// deriveArtifactIDForPlatform derives an artifact ID based on the pattern and compression
// For UPX compressed artifacts with pattern {moniker}-{os}-amd64-upx{ext}, the ID is {os}-amd64-upx
func deriveArtifactIDForPlatform(artifact config.Artifact, targetOS, targetArch string) string {
	if artifact.ID != "" {
		return artifact.ID
	}

	// For executables with "-upx" in pattern, append -upx to the ID
	if artifact.IsExecutable() && artifact.GetCompression() == config.CompressionUPX {
		return fmt.Sprintf("%s-%s-upx", targetOS, targetArch)
	}

	// Default executable ID
	if artifact.IsExecutable() {
		return fmt.Sprintf("%s-%s", targetOS, targetArch)
	}

	// For other types, use base of pattern
	return filepath.Base(artifact.Pattern)
}

// processDerivedArtifact derives a single artifact from its source
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
		Type:     art.Type,
		Pattern:  art.DeriveFrom,
		ID:       "",
		Platforms: art.Platforms,
	}
	sourceName := resolver.ResolvePatternWithMetadata(art.DeriveFrom, sourceArtifact)
	sourcePath := filepath.Join(buildDir, sourceName)

	// Resolve target path
	targetName := resolver.ResolvePatternWithMetadata(art.Pattern, art)
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
		if err := os.Chmod(targetPath, 0755); err != nil {
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

// compressArtifact applies compression to an artifact
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

// stripArtifact strips debug symbols from a binary
func stripArtifact(targetPath string, logWriter io.Writer) error {
	stripPath, err := exec.LookPath("strip")
	if err != nil {
		return fmt.Errorf("strip command not found")
	}

	fmt.Fprintf(logWriter, "  Stripping debug symbols: %s\n", filepath.Base(targetPath))

	cmd := exec.Command(stripPath, targetPath)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("strip command failed: %w", err)
	}

	return nil
}

// upxCompressArtifact compresses a binary with UPX
func upxCompressArtifact(targetPath string, logWriter io.Writer) error {
	upxPath, err := exec.LookPath("upx")
	if err != nil {
		return fmt.Errorf("upx command not found")
	}

	fmt.Fprintf(logWriter, "  Compressing with UPX: %s\n", filepath.Base(targetPath))

	cmd := exec.Command(upxPath, "--best", "--lzma", targetPath)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("upx command failed: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst
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
