// ArtifactsMode controls the scope of artifact generation during builds.
package environments

import (
	"fmt"
	"runtime"
)

// ArtifactsMode controls the scope of artifact generation during builds.
type ArtifactsMode string

const (
	// ArtifactsModeAll produces all artifacts (CI default).
	// - Builds all platform targets (cross-compilation)
	// - Includes UPX-compressed variants
	// - Includes derived artifacts
	// - Unlimited PDF pages
	ArtifactsModeAll ArtifactsMode = "all"

	// ArtifactsModeReduced produces reduced artifacts for faster local builds.
	// - Builds only current platform (+ linux-amd64 on Windows for WSL)
	// - Skips UPX-compressed variants
	// - Excludes derived artifacts
	// - Limited PDF pages (10)
	ArtifactsModeReduced ArtifactsMode = "reduced"
)

// DefaultReducedPageLimit is the default page limit for PDF generation in reduced mode.
const DefaultReducedPageLimit = 10

// IsValid returns true if the mode is a recognized value.
func (m ArtifactsMode) IsValid() bool {
	return m == ArtifactsModeAll || m == ArtifactsModeReduced
}

// String returns the string representation of the mode.
func (m ArtifactsMode) String() string {
	return string(m)
}

// PDFPageLimit returns the page limit for PDF generation.
// Returns 0 for unlimited (all pages) in "all" mode.
// Returns DefaultReducedPageLimit in "reduced" mode.
func (m ArtifactsMode) PDFPageLimit() int {
	if m == ArtifactsModeReduced {
		return DefaultReducedPageLimit
	}
	return 0
}

// AllArtifactsRequested returns true if all artifacts should be built.
func (m ArtifactsMode) AllArtifactsRequested() bool {
	return m == ArtifactsModeAll
}

// ShouldIncludeDerivedArtifacts returns true if derived artifacts (UPX variants)
// should be included in the build.
func (m ArtifactsMode) ShouldIncludeDerivedArtifacts() bool {
	return m == ArtifactsModeAll
}

// ShouldBuildTarget returns true if the given OS/arch/compression target should be built.
// Uses runtime.GOOS and runtime.GOARCH for current platform detection.
func (m ArtifactsMode) ShouldBuildTarget(targetOS, targetArch, compression string) bool {
	return m.ShouldBuildTargetWithCurrentPlatform(targetOS, targetArch, compression, runtime.GOOS, runtime.GOARCH)
}

// ShouldBuildTargetWithCurrentPlatform returns true if the given target should be built.
// Accepts explicit current platform for testability.
func (m ArtifactsMode) ShouldBuildTargetWithCurrentPlatform(targetOS, targetArch, compression, currentOS, currentArch string) bool {
	if m == ArtifactsModeAll {
		return true
	}

	// Reduced mode: skip UPX variants (release-only artifacts)
	if compression == "upx" {
		return false
	}

	// Reduced mode: current platform only
	if targetOS == currentOS && targetArch == currentArch {
		return true
	}

	// On Windows, also build linux-amd64 for WSL
	if currentOS == "windows" && targetOS == "linux" && targetArch == "amd64" {
		return true
	}

	return false
}

// DefaultArtifactsMode returns the default artifacts mode for the environment.
// CI environments default to "all", everything else defaults to "reduced".
func DefaultArtifactsMode() ArtifactsMode {
	if IsCI() {
		return ArtifactsModeAll
	}
	return ArtifactsModeReduced
}

// ParseArtifactsMode parses a string into an ArtifactsMode.
// Returns an error if the value is not valid.
func ParseArtifactsMode(s string) (ArtifactsMode, error) {
	mode := ArtifactsMode(s)
	if !mode.IsValid() {
		return "", fmt.Errorf("invalid artifacts mode %q: must be 'all' or 'reduced'", s)
	}
	return mode, nil
}
