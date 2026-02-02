package changelog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseString_BasicChangelog(t *testing.T) {
	content := `# Changelog

All notable changes to **test-module** will be documented in this file.

## [Unreleased]

## [1.2.0] - 2025-12-01

### Added

- feat(core): add new feature
- feat: another feature without scope

### Fixed

- fix(api): resolve bug

## [1.1.0] - 2025-11-15

### Added

- Initial release

[Unreleased]: https://github.com/org/repo/compare/test-module/1.2.0...HEAD
[1.2.0]: https://github.com/org/repo/compare/test-module/1.1.0...test-module/1.2.0
[1.1.0]: https://github.com/org/repo/releases/tag/test-module/1.1.0
`

	changelog, err := ParseString(content)
	require.NoError(t, err)

	assert.Equal(t, "Changelog", changelog.Title)
	assert.Equal(t, Semver, changelog.VersionType)
	assert.NotNil(t, changelog.Unreleased)
	assert.Len(t, changelog.Versions, 2)

	// Check first version
	v := changelog.Versions[0]
	assert.Equal(t, "1.2.0", v.Number)
	assert.Equal(t, 2025, v.Date.Year())
	assert.Equal(t, 12, int(v.Date.Month()))
	assert.Equal(t, 1, v.Date.Day())
	assert.Len(t, v.Added, 2)
	assert.Len(t, v.Fixed, 1)

	// Check entry parsing
	assert.Equal(t, "feat", v.Added[0].CommitType)
	assert.Equal(t, "core", v.Added[0].Scope)
	assert.Equal(t, "add new feature", v.Added[0].Description)

	// Check second entry (no scope)
	assert.Equal(t, "feat", v.Added[1].CommitType)
	assert.Equal(t, "", v.Added[1].Scope)

	// Check repo URL extraction
	assert.Equal(t, "https://github.com/org/repo", changelog.RepoURL)
}

func TestParseString_CalverChangelog(t *testing.T) {
	content := `# Changelog

## [2025.12.01] - 2025-12-01

### Changed

- Documentation updates

## [2025.11.27.1] - 2025-11-27

### Fixed

- Build fix

## [2025.11.27] - 2025-11-27

### Added

- Initial release
`

	changelog, err := ParseString(content)
	require.NoError(t, err)

	assert.Equal(t, Calver, changelog.VersionType)
	assert.Len(t, changelog.Versions, 3)
	assert.Equal(t, "2025.12.01", changelog.Versions[0].Number)
	assert.Equal(t, "2025.11.27.1", changelog.Versions[1].Number)
}

func TestParseString_YankedVersion(t *testing.T) {
	content := `# Changelog

## [1.0.1] - 2025-12-01

### Fixed

- Security fix

## [1.0.0] - 2025-11-01 [YANKED]

### Added

- Initial release (yanked due to critical bug)
`

	changelog, err := ParseString(content)
	require.NoError(t, err)

	assert.Len(t, changelog.Versions, 2)
	assert.False(t, changelog.Versions[0].Yanked)
	assert.True(t, changelog.Versions[1].Yanked)
}

func TestParseString_BreakingChange(t *testing.T) {
	content := `# Changelog

## [2.0.0] - 2025-12-01

### Changed

- feat(api)!: breaking API change
- refactor!: another breaking change
`

	changelog, err := ParseString(content)
	require.NoError(t, err)

	v := changelog.Versions[0]
	assert.Len(t, v.Changed, 2)
	assert.True(t, v.Changed[0].Breaking)
	assert.True(t, v.Changed[1].Breaking)
	assert.True(t, v.HasBreakingChanges())
}

func TestChangelog_LatestVersion(t *testing.T) {
	changelog := &Changelog{
		Versions: []Version{
			{Number: "1.2.0"},
			{Number: "1.1.0"},
			{Number: "1.0.0"},
		},
	}

	latest := changelog.LatestVersion()
	require.NotNil(t, latest)
	assert.Equal(t, "1.2.0", latest.Number)
}

func TestChangelog_LatestVersion_Empty(t *testing.T) {
	changelog := &Changelog{}

	latest := changelog.LatestVersion()
	assert.Nil(t, latest)
	assert.Equal(t, "", changelog.LatestVersionNumber())
}

func TestChangelog_GetVersion(t *testing.T) {
	changelog := &Changelog{
		Versions: []Version{
			{Number: "1.2.0"},
			{Number: "1.1.0"},
			{Number: "1.0.0"},
		},
	}

	v := changelog.GetVersion("1.1.0")
	require.NotNil(t, v)
	assert.Equal(t, "1.1.0", v.Number)

	v = changelog.GetVersion("9.9.9")
	assert.Nil(t, v)
}

func TestVersion_HasEntries(t *testing.T) {
	tests := []struct {
		name     string
		version  Version
		expected bool
	}{
		{
			name:     "empty version",
			version:  Version{},
			expected: false,
		},
		{
			name: "with added entries",
			version: Version{
				Added: []Entry{{Description: "test"}},
			},
			expected: true,
		},
		{
			name: "with fixed entries",
			version: Version{
				Fixed: []Entry{{Description: "test"}},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.version.HasEntries())
		})
	}
}
