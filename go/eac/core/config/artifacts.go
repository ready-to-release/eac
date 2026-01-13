package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ArtifactResolver resolves artifact patterns to concrete file paths
type ArtifactResolver struct {
	Moniker  string            // Module moniker
	BuildDir string            // Build output directory (e.g., out/build/<module>)
	OS       string            // Target OS (linux, windows, darwin)
	Arch     string            // Target architecture (amd64, arm64)
	Metadata map[string]string // Module metadata for custom artifact names
}

// NewArtifactResolver creates a new resolver for a module
func NewArtifactResolver(moniker, buildDir string) *ArtifactResolver {
	return &ArtifactResolver{
		Moniker:  moniker,
		BuildDir: buildDir,
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
	}
}

// NewArtifactResolverWithMetadata creates a resolver with module metadata
func NewArtifactResolverWithMetadata(moniker, buildDir string, metadata map[string]string) *ArtifactResolver {
	return &ArtifactResolver{
		Moniker:  moniker,
		BuildDir: buildDir,
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Metadata: metadata,
	}
}

// NewArtifactResolverWithPlatform creates a resolver for a specific platform
func NewArtifactResolverWithPlatform(moniker, buildDir, os, arch string) *ArtifactResolver {
	return &ArtifactResolver{
		Moniker:  moniker,
		BuildDir: buildDir,
		OS:       os,
		Arch:     arch,
	}
}

// NewArtifactResolverFull creates a resolver with all options
func NewArtifactResolverFull(moniker, buildDir, os, arch string, metadata map[string]string) *ArtifactResolver {
	return &ArtifactResolver{
		Moniker:  moniker,
		BuildDir: buildDir,
		OS:       os,
		Arch:     arch,
		Metadata: metadata,
	}
}

// ResolvePattern resolves an artifact pattern to a concrete path
func (r *ArtifactResolver) ResolvePattern(pattern string) string {
	result := pattern
	result = strings.ReplaceAll(result, "{moniker}", r.Moniker)
	result = strings.ReplaceAll(result, "{os}", r.OS)
	result = strings.ReplaceAll(result, "{arch}", r.Arch)
	result = strings.ReplaceAll(result, "{ext}", r.getExtension())
	return result
}

// ResolvePatternWithMetadata resolves a pattern, checking metadata for custom names.
// Supports all artifact types via {type}-{variant} metadata keys (e.g., executable-linux-amd64, directory-site, file-pdf).
func (r *ArtifactResolver) ResolvePatternWithMetadata(pattern string, artifact Artifact) string {
	if r.Metadata == nil {
		return r.ResolvePattern(pattern)
	}

	// Determine variant ID for metadata key
	variantID := artifact.ID
	if variantID == "" {
		// Derive variant ID from pattern/context
		variantID = r.deriveVariantID(pattern, artifact)
	}

	// Check for metadata override: {type}-{variant}
	// Uses FULL type name (e.g., executable-linux-amd64, not exe-linux-amd64)
	metadataKey := fmt.Sprintf("%s-%s", artifact.Type, variantID)
	if customValue, ok := r.Metadata[metadataKey]; ok && customValue != "" {
		return customValue
	}

	// No override - use pattern
	return r.ResolvePattern(pattern)
}

// deriveVariantID derives a variant ID from the pattern and artifact.
// For executables: {os}-{arch} (e.g., linux-amd64), with optional compression suffix
// For files/directories: base filename/dirname from pattern
func (r *ArtifactResolver) deriveVariantID(pattern string, artifact Artifact) string {
	switch artifact.Type {
	case ArtifactTypeExecutable:
		// Executables use {os}-{arch} format, with optional compression suffix
		id := fmt.Sprintf("%s-%s", r.OS, r.Arch)
		// Add compression suffix for compressed artifacts
		if artifact.GetCompression() == CompressionUPX {
			id += "-upx"
		}
		return id

	case ArtifactTypeFile, ArtifactTypeDirectory, ArtifactTypeImage:
		// For other types, resolve pattern and use base name
		resolved := r.ResolvePattern(pattern)
		base := filepath.Base(resolved)
		// Remove extension for files
		if artifact.Type == ArtifactTypeFile {
			ext := filepath.Ext(base)
			if ext != "" {
				base = base[:len(base)-len(ext)]
			}
		}
		return base

	default:
		// Markers, globs, etc. - use pattern as-is
		return pattern
	}
}

// ResolvePath resolves an artifact pattern to a full file path
func (r *ArtifactResolver) ResolvePath(pattern string) string {
	resolved := r.ResolvePattern(pattern)
	return filepath.Join(r.BuildDir, resolved)
}

// getExtension returns the executable extension for the current OS
func (r *ArtifactResolver) getExtension() string {
	if r.OS == "windows" {
		return ".exe"
	}
	return ""
}

// ArtifactVerificationResult contains the result of verifying an artifact
type ArtifactVerificationResult struct {
	Artifact    Artifact
	Pattern     string // Resolved pattern
	Path        string // Full path
	Exists      bool
	IsDirectory bool
	Error       error
}

// VerifyArtifact checks if an artifact exists
func (r *ArtifactResolver) VerifyArtifact(artifact Artifact) ArtifactVerificationResult {
	result := ArtifactVerificationResult{
		Artifact: artifact,
		Pattern:  r.ResolvePatternWithMetadata(artifact.Pattern, artifact),
	}
	result.Path = filepath.Join(r.BuildDir, result.Pattern)

	info, err := os.Stat(result.Path)
	if err != nil {
		if os.IsNotExist(err) {
			result.Exists = false
		} else {
			result.Error = err
		}
		return result
	}

	result.Exists = true
	result.IsDirectory = info.IsDir()

	// Verify type matches expectation
	if artifact.Type == ArtifactTypeDirectory && !result.IsDirectory {
		result.Error = fmt.Errorf("expected directory but found file")
	} else if artifact.Type != ArtifactTypeDirectory && result.IsDirectory {
		result.Error = fmt.Errorf("expected file but found directory")
	}

	return result
}

// VerifyArtifacts verifies all artifacts for a module type
func (r *ArtifactResolver) VerifyArtifacts(artifacts []Artifact) []ArtifactVerificationResult {
	var results []ArtifactVerificationResult

	for _, artifact := range artifacts {
		switch artifact.Type {
		case ArtifactTypeExecutable:
			// Handle platform-specific executables
			results = append(results, r.verifyExecutableArtifact(artifact)...)
		case ArtifactTypeGlob:
			// Handle glob patterns
			results = append(results, r.verifyGlobArtifact(artifact))
		case ArtifactTypeImage:
			// Docker images require actual verification - check if image exists locally
			results = append(results, r.verifyImageArtifact(artifact))
		default:
			// Standard artifact verification
			results = append(results, r.VerifyArtifact(artifact))
		}
	}

	return results
}

// verifyExecutableArtifact handles executable artifacts with platform considerations
func (r *ArtifactResolver) verifyExecutableArtifact(artifact Artifact) []ArtifactVerificationResult {
	var results []ArtifactVerificationResult
	verifyMode := artifact.GetVerifyMode()

	// Determine which platforms to check
	platforms := artifact.Platforms
	if len(platforms) == 0 {
		platforms = []string{"linux", "windows", "darwin"}
	}

	switch verifyMode {
	case VerifyCurrentPlatform:
		// Only verify for current platform if it's in the list
		// Also check architecture - patterns may contain hardcoded arch (e.g., "-arm64")
		patternArch := extractArchFromPattern(artifact.Pattern)
		if patternArch != "" && patternArch != r.Arch {
			// Pattern specifies a different architecture, skip verification
			break
		}
		for _, p := range platforms {
			if p == r.OS {
				results = append(results, r.VerifyArtifact(artifact))
				break
			}
		}
	case VerifyAll:
		// Verify all specified platforms exist
		for _, p := range platforms {
			resolver := NewArtifactResolverFull(r.Moniker, r.BuildDir, p, r.Arch, r.Metadata)
			results = append(results, resolver.VerifyArtifact(artifact))
		}
	case VerifyAny:
		// At least one platform must exist
		found := false
		for _, p := range platforms {
			resolver := NewArtifactResolverFull(r.Moniker, r.BuildDir, p, r.Arch, r.Metadata)
			result := resolver.VerifyArtifact(artifact)
			if result.Exists {
				found = true
				results = append(results, result)
				break
			}
		}
		if !found {
			// Return last check result showing not found
			resolver := NewArtifactResolverFull(r.Moniker, r.BuildDir, platforms[0], r.Arch, r.Metadata)
			result := resolver.VerifyArtifact(artifact)
			result.Error = fmt.Errorf("no executable found for any platform: %v", platforms)
			results = append(results, result)
		}
	}

	return results
}

// verifyGlobArtifact handles glob pattern artifacts
func (r *ArtifactResolver) verifyGlobArtifact(artifact Artifact) ArtifactVerificationResult {
	result := ArtifactVerificationResult{
		Artifact: artifact,
		Pattern:  r.ResolvePattern(artifact.Pattern),
	}
	result.Path = filepath.Join(r.BuildDir, result.Pattern)

	matches, err := filepath.Glob(result.Path)
	if err != nil {
		result.Error = err
		return result
	}

	result.Exists = len(matches) > 0
	if !result.Exists {
		result.Error = fmt.Errorf("no files match pattern: %s", result.Path)
	}

	return result
}

// verifyImageArtifact handles Docker image artifacts by checking if the image exists locally
func (r *ArtifactResolver) verifyImageArtifact(artifact Artifact) ArtifactVerificationResult {
	imageRef := r.ResolvePattern(artifact.Pattern)

	result := ArtifactVerificationResult{
		Artifact: artifact,
		Pattern:  imageRef,
		Path:     imageRef, // For images, Path is the image reference
	}

	// Check if Docker is available
	if !isDockerAvailable() {
		result.Error = fmt.Errorf("docker not available for image verification")
		result.Exists = false
		return result
	}

	// Check if image exists locally using `docker images -q <ref>`
	cmd := exec.Command("docker", "images", "-q", imageRef)
	output, err := cmd.Output()
	if err != nil {
		result.Error = fmt.Errorf("failed to check docker image: %w", err)
		result.Exists = false
		return result
	}

	// If output is non-empty, image exists locally
	if strings.TrimSpace(string(output)) != "" {
		result.Exists = true
		return result
	}

	// Image not found locally - for registry images, try manifest inspect
	if strings.Contains(imageRef, "/") {
		// Looks like a registry image (e.g., ghcr.io/...)
		cmd = exec.Command("docker", "manifest", "inspect", imageRef)
		if err := cmd.Run(); err == nil {
			result.Exists = true
			return result
		}
	}

	result.Exists = false
	result.Error = fmt.Errorf("docker image not found: %s", imageRef)
	return result
}

// isDockerAvailable checks if Docker CLI is available
func isDockerAvailable() bool {
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	return cmd.Run() == nil
}

// AllSuccessful returns true if all verification results are successful
func AllSuccessful(results []ArtifactVerificationResult) bool {
	for _, r := range results {
		if !r.Exists || r.Error != nil {
			return false
		}
	}
	return true
}

// GetFailures returns only the failed verification results
func GetFailures(results []ArtifactVerificationResult) []ArtifactVerificationResult {
	var failures []ArtifactVerificationResult
	for _, r := range results {
		if !r.Exists || r.Error != nil {
			failures = append(failures, r)
		}
	}
	return failures
}

// FormatVerificationResults formats verification results for display
func FormatVerificationResults(results []ArtifactVerificationResult) string {
	var sb strings.Builder
	for _, r := range results {
		if r.Exists && r.Error == nil {
			sb.WriteString(fmt.Sprintf("  ✅ %s\n", r.Pattern))
		} else if r.Error != nil {
			sb.WriteString(fmt.Sprintf("  ❌ %s: %v\n", r.Pattern, r.Error))
		} else {
			sb.WriteString(fmt.Sprintf("  ❌ %s: not found\n", r.Pattern))
		}
	}
	return sb.String()
}

// extractArchFromPattern extracts hardcoded architecture from an artifact pattern.
// Returns the architecture if found (e.g., "amd64", "arm64"), or empty string if
// the pattern uses a placeholder like {arch}.
func extractArchFromPattern(pattern string) string {
	// Common architecture suffixes in patterns
	// Check for hardcoded architectures (not placeholders)
	archs := []string{"amd64", "arm64", "386", "arm"}

	for _, arch := range archs {
		// Look for arch in pattern, but not as a placeholder
		if strings.Contains(pattern, "{arch}") {
			// Pattern uses placeholder, architecture is variable
			return ""
		}
		// Check for arch as a component (e.g., "-amd64", "-arm64")
		if strings.Contains(pattern, "-"+arch) || strings.Contains(pattern, "_"+arch) {
			return arch
		}
	}
	return ""
}
