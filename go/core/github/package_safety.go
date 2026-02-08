package github

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// VersionProtectionReason explains why a version is protected.
type VersionProtectionReason string

const (
	// ReasonTagMatchesPreserve indicates a tag matches a preserve pattern.
	ReasonTagMatchesPreserve VersionProtectionReason = "tag matches preserve pattern"

	// ReasonHasRelease indicates the version is associated with a GitHub release.
	ReasonHasRelease VersionProtectionReason = "associated with GitHub release"

	// ReasonInReleaseBundle indicates the version is referenced by a release bundle.
	ReasonInReleaseBundle VersionProtectionReason = "referenced by release bundle"

	// ReasonDigestMatchesRelease indicates the digest matches a released version.
	ReasonDigestMatchesRelease VersionProtectionReason = "digest matches released version"

	// ReasonTooNew indicates the version was created too recently.
	ReasonTooNew VersionProtectionReason = "created less than min_age_days ago"

	// ReasonNotMatchPrune indicates no tags match the prune patterns.
	ReasonNotMatchPrune VersionProtectionReason = "no tags match prune patterns"
)

// VersionAssessment contains the safety assessment for a package version.
type VersionAssessment struct {
	Version    PackageVersion
	Protected  bool
	Reason     VersionProtectionReason
	MatchedTag string // The tag that triggered protection (if any)
}

// PackageSafetyChecker assesses which package versions are safe to delete.
// Released packages are ALWAYS protected - this is a non-configurable safety feature.
type PackageSafetyChecker struct {
	preservePatterns []string        // Image tag patterns to preserve (OCI/Docker tags)
	prunePatterns    []string        // Image tag patterns eligible for pruning
	minAgeDays       int             // Minimum age before pruning
	releasedTags     map[string]bool // GitHub Release tags -> exists (for API correlation)
	releasedDigests  map[string]bool // digest -> exists (cross-reference protection)
	bundleModules    map[string]bool // Module versions referenced by release bundles
	now              time.Time       // for testing
}

// NewPackageSafetyChecker creates a new safety checker.
// Released packages are ALWAYS protected - this is a non-configurable safety feature.
//
// Parameters:
//   - preservePatterns: Image tag patterns (glob) that are never deleted
//   - prunePatterns: Image tag patterns (glob) eligible for cleanup
//   - minAgeDays: Minimum age in days before a version can be pruned
//   - releases: GitHub Releases for tag correlation (from GitHub API)
func NewPackageSafetyChecker(
	preservePatterns, prunePatterns []string,
	minAgeDays int,
	releases []Release,
) *PackageSafetyChecker {
	releasedTags := make(map[string]bool)

	// All releases are protected - this is non-configurable for safety
	for _, r := range releases {
		releasedTags[r.TagName] = true
	}

	return &PackageSafetyChecker{
		preservePatterns: preservePatterns,
		prunePatterns:    prunePatterns,
		minAgeDays:       minAgeDays,
		releasedTags:     releasedTags,
		releasedDigests:  make(map[string]bool),
		bundleModules:    make(map[string]bool),
		now:              time.Now(),
	}
}

// SetNow sets the current time for testing purposes.
func (c *PackageSafetyChecker) SetNow(t time.Time) {
	c.now = t
}

// AddReleasedDigest records a digest as belonging to a release.
func (c *PackageSafetyChecker) AddReleasedDigest(digest string) {
	c.releasedDigests[digest] = true
}

// AddBundleModuleVersion records a module version as being referenced by a release bundle.
// Format: "module/version" (e.g., "eac-ext/1.0.0").
func (c *PackageSafetyChecker) AddBundleModuleVersion(moduleVersion string) {
	c.bundleModules[moduleVersion] = true
}

// Assess evaluates whether a package version is safe to delete.
func (c *PackageSafetyChecker) Assess(version PackageVersion) VersionAssessment {
	assessment := VersionAssessment{
		Version:   version,
		Protected: false,
	}

	// Check 1: Does any tag match preserve patterns?
	for _, tag := range version.Tags {
		for _, pattern := range c.preservePatterns {
			matched, err := filepath.Match(pattern, tag)
			if err != nil {
				continue // Invalid pattern, skip it
			}
			if !matched && strings.Contains(tag, "/") && pattern == "*" {
				matched = true
			}
			if matched {
				assessment.Protected = true
				assessment.Reason = ReasonTagMatchesPreserve
				assessment.MatchedTag = tag
				return assessment
			}
		}
	}

	// Check 2: Does any tag have a corresponding GitHub Release?
	// Released packages are ALWAYS protected - this is non-configurable for safety.
	for _, tag := range version.Tags {
		if c.releasedTags[tag] {
			assessment.Protected = true
			assessment.Reason = ReasonHasRelease
			assessment.MatchedTag = tag
			return assessment
		}
	}

	// Check 2b: Is this version referenced by a release bundle?
	// Release bundles aggregate multiple module releases - protect all referenced versions.
	for _, tag := range version.Tags {
		if c.bundleModules[tag] {
			assessment.Protected = true
			assessment.Reason = ReasonInReleaseBundle
			assessment.MatchedTag = tag
			return assessment
		}
	}

	// Check 3: Does the digest match any released version's digest?
	if c.releasedDigests[version.Digest] {
		assessment.Protected = true
		assessment.Reason = ReasonDigestMatchesRelease
		return assessment
	}

	// Check 4: Is it too new?
	cutoff := c.now.AddDate(0, 0, -c.minAgeDays)
	if version.CreatedAt.After(cutoff) {
		assessment.Protected = true
		assessment.Reason = ReasonTooNew
		return assessment
	}

	// Check 5: Does at least one tag match prune patterns?
	hasPruneMatch := false
	for _, tag := range version.Tags {
		for _, pattern := range c.prunePatterns {
			matched, err := filepath.Match(pattern, tag)
			if err != nil {
				continue // Invalid pattern, skip it
			}
			if !matched && strings.Contains(tag, "/") && pattern == "*" {
				matched = true
			}
			if matched {
				hasPruneMatch = true
				break
			}
		}
		if hasPruneMatch {
			break
		}
	}

	// Untagged versions (empty tags) are eligible for pruning
	if len(version.Tags) == 0 {
		hasPruneMatch = true
	}

	if !hasPruneMatch {
		assessment.Protected = true
		assessment.Reason = ReasonNotMatchPrune
		return assessment
	}

	// Version is safe to prune
	return assessment
}

// AssessAll evaluates all versions and returns protected and prunable ones.
func (c *PackageSafetyChecker) AssessAll(versions []PackageVersion) (protected, prunable []VersionAssessment) {
	for _, v := range versions {
		assessment := c.Assess(v)
		if assessment.Protected {
			protected = append(protected, assessment)
		} else {
			prunable = append(prunable, assessment)
		}
	}
	return protected, prunable
}

// bundleTagPattern matches release bundle tags like "clie-eac-bundle/2025.01.15".
var bundleTagPattern = regexp.MustCompile(`^[a-z][a-z0-9-]*-bundle/`)

// moduleVersionPattern matches module version references in release notes.
// Matches patterns like:
//   - "eac-ext/1.0.0" (module/version format)
//   - "- eac-ext: v1.0.0" (markdown list format)
//   - "- eac-ext: 1.0.0" (markdown list format without v prefix)
var moduleVersionPattern = regexp.MustCompile(`(?:^|\s)([a-z][a-z0-9-]*)[/:]\s*v?(\d+\.\d+\.\d+(?:-[a-zA-Z0-9.]+)?)`)

// ExtractBundleModuleVersions extracts module versions referenced by release bundles.
// It identifies bundle releases (tags matching *-bundle/*) and parses their bodies
// to extract module/version references.
//
// Returns a list of module/version strings in the format "module/version".
func ExtractBundleModuleVersions(releases []Release) []string {
	var versions []string
	seen := make(map[string]bool)

	for _, r := range releases {
		// Check if this is a bundle release
		if !bundleTagPattern.MatchString(r.TagName) {
			continue
		}

		// Parse the release body for module versions
		matches := moduleVersionPattern.FindAllStringSubmatch(r.Body, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				module := strings.TrimSpace(match[1])
				version := strings.TrimSpace(match[2])
				moduleVersion := module + "/" + version

				// Skip duplicates
				if seen[moduleVersion] {
					continue
				}
				seen[moduleVersion] = true
				versions = append(versions, moduleVersion)
			}
		}
	}

	return versions
}

// AddBundleModuleVersionsFromReleases extracts and registers module versions
// from release bundles. This ensures that packages referenced by any release
// bundle are protected from deletion.
func (c *PackageSafetyChecker) AddBundleModuleVersionsFromReleases(releases []Release) {
	versions := ExtractBundleModuleVersions(releases)
	for _, v := range versions {
		c.AddBundleModuleVersion(v)
	}
}
