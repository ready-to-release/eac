package github

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafetyChecker_PreservePatterns(t *testing.T) {
	checker := NewPackageSafetyChecker(
		[]string{"v*", "latest"},
		[]string{"sha-*"},
		7,
		nil,
	)
	// Set time for consistent testing
	checker.SetNow(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name      string
		tags      []string
		protected bool
		reason    VersionProtectionReason
	}{
		{
			name:      "release tag v1.0.0 is protected",
			tags:      []string{"v1.0.0"},
			protected: true,
			reason:    ReasonTagMatchesPreserve,
		},
		{
			name:      "latest tag is protected",
			tags:      []string{"latest"},
			protected: true,
			reason:    ReasonTagMatchesPreserve,
		},
		{
			name:      "sha tag is prunable",
			tags:      []string{"sha-abc1234"},
			protected: false,
		},
		{
			name:      "mixed tags with v* is protected",
			tags:      []string{"sha-abc", "v1.0.0"},
			protected: true,
			reason:    ReasonTagMatchesPreserve,
		},
		{
			name:      "v prefix only matches at start",
			tags:      []string{"sha-v1.0.0"},
			protected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := PackageVersion{
				Tags:      tt.tags,
				CreatedAt: time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC), // 45 days ago
			}
			result := checker.Assess(v)
			assert.Equal(t, tt.protected, result.Protected)
			if tt.protected {
				assert.Equal(t, tt.reason, result.Reason)
			}
		})
	}
}

func TestSafetyChecker_ReleaseCorrelation(t *testing.T) {
	releases := []Release{{TagName: "ext-eac/1.0.0"}}
	checker := NewPackageSafetyChecker(
		[]string{}, // No preserve patterns - testing release protection only
		[]string{"sha-*"},
		0, // No age check
		releases,
	)

	t.Run("version with release tag is protected", func(t *testing.T) {
		v := PackageVersion{
			Tags:      []string{"sha-abc1234", "ext-eac/1.0.0"},
			CreatedAt: time.Now().AddDate(0, 0, -30),
		}
		result := checker.Assess(v)
		assert.True(t, result.Protected)
		assert.Equal(t, ReasonHasRelease, result.Reason)
		assert.Equal(t, "ext-eac/1.0.0", result.MatchedTag)
	})

	t.Run("version without release tag is prunable", func(t *testing.T) {
		v := PackageVersion{
			Tags:      []string{"sha-abc1234"},
			CreatedAt: time.Now().AddDate(0, 0, -30),
		}
		result := checker.Assess(v)
		assert.False(t, result.Protected)
	})
}

// Note: TestSafetyChecker_ReleaseCorrelationDisabled was removed.
// Release protection is now ALWAYS enabled - this is a non-configurable safety feature.
// Released packages must never be auto-deleted.

func TestSafetyChecker_DigestMatching(t *testing.T) {
	checker := NewPackageSafetyChecker(
		[]string{"v*"},
		[]string{"sha-*"},
		0, // No age check
		nil,
	)

	// Simulate a released version's digest
	releasedDigest := "sha256:abc123def456"
	checker.AddReleasedDigest(releasedDigest)

	t.Run("CI version with same digest as release is protected", func(t *testing.T) {
		v := PackageVersion{
			Tags:      []string{"sha-xyz789"},
			Digest:    releasedDigest,
			CreatedAt: time.Now().AddDate(0, 0, -30),
		}
		result := checker.Assess(v)
		assert.True(t, result.Protected)
		assert.Equal(t, ReasonDigestMatchesRelease, result.Reason)
	})

	t.Run("version with different digest is prunable", func(t *testing.T) {
		v := PackageVersion{
			Tags:      []string{"sha-xyz789"},
			Digest:    "sha256:differentdigest",
			CreatedAt: time.Now().AddDate(0, 0, -30),
		}
		result := checker.Assess(v)
		assert.False(t, result.Protected)
	})
}

func TestSafetyChecker_MinAgeDays(t *testing.T) {
	checker := NewPackageSafetyChecker(
		[]string{},
		[]string{"sha-*"},
		7, // 7 days minimum age
		nil,
	)
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	checker.SetNow(now)

	t.Run("version created 3 days ago is protected", func(t *testing.T) {
		v := PackageVersion{
			Tags:      []string{"sha-abc123"},
			CreatedAt: now.AddDate(0, 0, -3),
		}
		result := checker.Assess(v)
		assert.True(t, result.Protected)
		assert.Equal(t, ReasonTooNew, result.Reason)
	})

	t.Run("version created 7 days ago is protected (boundary)", func(t *testing.T) {
		v := PackageVersion{
			Tags:      []string{"sha-abc123"},
			CreatedAt: now.AddDate(0, 0, -7).Add(time.Hour), // Just under 7 days
		}
		result := checker.Assess(v)
		assert.True(t, result.Protected)
		assert.Equal(t, ReasonTooNew, result.Reason)
	})

	t.Run("version created 8 days ago is prunable", func(t *testing.T) {
		v := PackageVersion{
			Tags:      []string{"sha-abc123"},
			CreatedAt: now.AddDate(0, 0, -8),
		}
		result := checker.Assess(v)
		assert.False(t, result.Protected)
	})
}

func TestSafetyChecker_PrunePatterns(t *testing.T) {
	checker := NewPackageSafetyChecker(
		[]string{},
		[]string{"sha-*", "dev-*", "pr-*"},
		0,
		nil,
	)

	tests := []struct {
		name      string
		tags      []string
		protected bool
		reason    VersionProtectionReason
	}{
		{
			name:      "sha-* matches prune pattern",
			tags:      []string{"sha-abc123"},
			protected: false,
		},
		{
			name:      "dev-* matches prune pattern",
			tags:      []string{"dev-latest"},
			protected: false,
		},
		{
			name:      "pr-* matches prune pattern",
			tags:      []string{"pr-42"},
			protected: false,
		},
		{
			name:      "feature-branch does not match prune pattern",
			tags:      []string{"feature-branch"},
			protected: true,
			reason:    ReasonNotMatchPrune,
		},
		{
			name:      "main does not match prune pattern",
			tags:      []string{"main"},
			protected: true,
			reason:    ReasonNotMatchPrune,
		},
		{
			name:      "untagged version is prunable",
			tags:      []string{},
			protected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := PackageVersion{
				Tags:      tt.tags,
				CreatedAt: time.Now().AddDate(0, 0, -30),
			}
			result := checker.Assess(v)
			assert.Equal(t, tt.protected, result.Protected)
			if tt.protected {
				assert.Equal(t, tt.reason, result.Reason)
			}
		})
	}
}

func TestSafetyChecker_AssessAll(t *testing.T) {
	checker := NewPackageSafetyChecker(
		[]string{"v*"},
		[]string{"sha-*"},
		0,
		nil,
	)

	versions := []PackageVersion{
		{ID: 1, Tags: []string{"v1.0.0"}, CreatedAt: time.Now().AddDate(0, 0, -30)},
		{ID: 2, Tags: []string{"sha-abc"}, CreatedAt: time.Now().AddDate(0, 0, -30)},
		{ID: 3, Tags: []string{"sha-def"}, CreatedAt: time.Now().AddDate(0, 0, -20)},
		{ID: 4, Tags: []string{"main"}, CreatedAt: time.Now().AddDate(0, 0, -10)},
	}

	protected, prunable := checker.AssessAll(versions)

	require.Len(t, protected, 2)
	require.Len(t, prunable, 2)

	// Check protected versions
	protectedIDs := make([]int, len(protected))
	for i, a := range protected {
		protectedIDs[i] = a.Version.ID
	}
	assert.Contains(t, protectedIDs, 1) // v1.0.0
	assert.Contains(t, protectedIDs, 4) // main (no prune match)

	// Check prunable versions
	prunableIDs := make([]int, len(prunable))
	for i, a := range prunable {
		prunableIDs[i] = a.Version.ID
	}
	assert.Contains(t, prunableIDs, 2) // sha-abc
	assert.Contains(t, prunableIDs, 3) // sha-def
}

func TestSafetyChecker_ProtectionPriority(t *testing.T) {
	// Test that protection checks are applied in order:
	// 1. Preserve patterns (image tags)
	// 2. Release tags (GitHub API)
	// 3. Digest matching
	// 4. Age check
	// 5. Prune patterns (image tags)

	t.Run("preserve pattern takes priority over prune pattern", func(t *testing.T) {
		checker := NewPackageSafetyChecker(
			[]string{"v*"},
			[]string{"v*"}, // Both preserve and prune match v*
			0,
			nil,
		)
		v := PackageVersion{
			Tags:      []string{"v1.0.0"},
			CreatedAt: time.Now().AddDate(0, 0, -30),
		}
		result := checker.Assess(v)
		assert.True(t, result.Protected)
		assert.Equal(t, ReasonTagMatchesPreserve, result.Reason)
	})

	t.Run("release tag takes priority over age", func(t *testing.T) {
		releases := []Release{{TagName: "ext-eac/1.0.0"}}
		checker := NewPackageSafetyChecker(
			[]string{},
			[]string{"*"}, // Everything matches prune
			30,            // 30 day age requirement
			releases,
		)
		v := PackageVersion{
			Tags:      []string{"ext-eac/1.0.0"},
			CreatedAt: time.Now().AddDate(0, 0, -1), // Very recent
		}
		result := checker.Assess(v)
		assert.True(t, result.Protected)
		assert.Equal(t, ReasonHasRelease, result.Reason) // Not ReasonTooNew
	})
}

func TestSafetyChecker_SemverPatterns(t *testing.T) {
	// Test the default semver pattern [0-9]*.[0-9]*.[0-9]*
	checker := NewPackageSafetyChecker(
		[]string{"[0-9]*.[0-9]*.[0-9]*"},
		[]string{"sha-*"},
		0,
		nil,
	)

	tests := []struct {
		name      string
		tags      []string
		protected bool
	}{
		{
			name:      "1.0.0 matches semver pattern",
			tags:      []string{"1.0.0"},
			protected: true,
		},
		{
			name:      "10.20.30 matches semver pattern",
			tags:      []string{"10.20.30"},
			protected: true,
		},
		{
			name:      "sha-1.0.0 does not match (has prefix)",
			tags:      []string{"sha-1.0.0"},
			protected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := PackageVersion{
				Tags:      tt.tags,
				CreatedAt: time.Now().AddDate(0, 0, -30),
			}
			result := checker.Assess(v)
			assert.Equal(t, tt.protected, result.Protected)
		})
	}
}

func TestExtractBundleModuleVersions(t *testing.T) {
	tests := []struct {
		name     string
		releases []Release
		expected []string
	}{
		{
			name: "extracts module versions from bundle release body",
			releases: []Release{
				{
					TagName: "r2r-eac-bundle/2025.01.15",
					Body: `## Core Tools
- ext-eac: v1.0.0
- r2r-cli: v2.3.4

## Documentation
- docs: 2025.01.15`,
				},
			},
			expected: []string{"ext-eac/1.0.0", "r2r-cli/2.3.4", "docs/2025.01.15"},
		},
		{
			name: "extracts module/version format",
			releases: []Release{
				{
					TagName: "test-bundle/1.0.0",
					Body:    "Includes ext-eac/1.2.3 and r2r-cli/4.5.6",
				},
			},
			expected: []string{"ext-eac/1.2.3", "r2r-cli/4.5.6"},
		},
		{
			name: "ignores non-bundle releases",
			releases: []Release{
				{
					TagName: "ext-eac/1.0.0",
					Body:    "This is a regular release",
				},
			},
			expected: nil,
		},
		{
			name: "handles empty body",
			releases: []Release{
				{
					TagName: "my-bundle/1.0.0",
					Body:    "",
				},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractBundleModuleVersions(tt.releases)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSafetyChecker_BundleProtection(t *testing.T) {
	releases := []Release{
		{
			TagName: "r2r-eac-bundle/2025.01.15",
			Body:    "Includes ext-eac/1.0.0 and r2r-cli/2.0.0",
		},
	}
	checker := NewPackageSafetyChecker(
		[]string{},
		[]string{"*"}, // Everything matches prune
		0,
		nil, // No individual releases
	)
	checker.AddBundleModuleVersionsFromReleases(releases)

	t.Run("version referenced by bundle is protected", func(t *testing.T) {
		v := PackageVersion{
			Tags:      []string{"ext-eac/1.0.0"},
			CreatedAt: time.Now().AddDate(0, 0, -30),
		}
		result := checker.Assess(v)
		assert.True(t, result.Protected)
		assert.Equal(t, ReasonInReleaseBundle, result.Reason)
	})

	t.Run("version not referenced by bundle is prunable", func(t *testing.T) {
		v := PackageVersion{
			Tags:      []string{"ext-eac/0.9.0"},
			CreatedAt: time.Now().AddDate(0, 0, -30),
		}
		result := checker.Assess(v)
		assert.False(t, result.Protected)
	})
}
