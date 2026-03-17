package dependabot

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBytes(t *testing.T) {
	t.Run("complete dependabot.yml with all supported fields", func(t *testing.T) {
		input := `version: 2

updates:
  # GitHub Actions
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
      time: "09:00"
      timezone: "America/New_York"
    commit-message:
      prefix: "chore(deps)"
    labels:
      - "dependencies"
      - "github-actions"

  # Go modules - go/core
  - package-ecosystem: "gomod"
    directory: "/go/core"
    schedule:
      interval: "daily"
    labels:
      - "dependencies"
      - "go"
`
		config, err := ParseBytes([]byte(input))
		require.NoError(t, err)

		assert.Equal(t, 2, config.Version)
		require.Len(t, config.Updates, 2)

		// First entry: github-actions
		u0 := config.Updates[0]
		assert.Equal(t, "github-actions", u0.PackageEcosystem)
		assert.Equal(t, "/", u0.Directory)
		assert.Equal(t, "weekly", u0.Schedule.Interval)
		assert.Equal(t, "monday", u0.Schedule.Day)
		assert.Equal(t, "09:00", u0.Schedule.Time)
		assert.Equal(t, "America/New_York", u0.Schedule.Timezone)
		require.NotNil(t, u0.CommitMessage)
		assert.Equal(t, "chore(deps)", u0.CommitMessage.Prefix)
		assert.Equal(t, []string{"dependencies", "github-actions"}, u0.Labels)

		// Second entry: gomod
		u1 := config.Updates[1]
		assert.Equal(t, "gomod", u1.PackageEcosystem)
		assert.Equal(t, "/go/core", u1.Directory)
		assert.Equal(t, "daily", u1.Schedule.Interval)
		assert.Equal(t, []string{"dependencies", "go"}, u1.Labels)
	})

	t.Run("empty input returns empty config with version 2", func(t *testing.T) {
		config, err := ParseBytes([]byte(""))
		require.NoError(t, err)

		assert.Equal(t, 2, config.Version)
		assert.Empty(t, config.Updates)
	})

	t.Run("version only returns empty updates", func(t *testing.T) {
		input := `version: 2
updates:
`
		config, err := ParseBytes([]byte(input))
		require.NoError(t, err)

		assert.Equal(t, 2, config.Version)
		assert.Empty(t, config.Updates)
	})

	t.Run("single entry without optional fields", func(t *testing.T) {
		input := `version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/web/app"
    schedule:
      interval: "weekly"
`
		config, err := ParseBytes([]byte(input))
		require.NoError(t, err)

		require.Len(t, config.Updates, 1)
		u := config.Updates[0]
		assert.Equal(t, "npm", u.PackageEcosystem)
		assert.Equal(t, "/web/app", u.Directory)
		assert.Equal(t, "weekly", u.Schedule.Interval)
		assert.Nil(t, u.CommitMessage)
		assert.Empty(t, u.Labels)
	})

	t.Run("entries with comments between them", func(t *testing.T) {
		input := `version: 2
updates:
  # First entry
  - package-ecosystem: "pip"
    directory: "/scripts"
    schedule:
      interval: "monthly"
  # Second entry
  - package-ecosystem: "docker"
    directory: "/containers/app"
    schedule:
      interval: "weekly"
`
		config, err := ParseBytes([]byte(input))
		require.NoError(t, err)

		require.Len(t, config.Updates, 2)
		assert.Equal(t, "pip", config.Updates[0].PackageEcosystem)
		assert.Equal(t, "/scripts", config.Updates[0].Directory)
		assert.Equal(t, "docker", config.Updates[1].PackageEcosystem)
		assert.Equal(t, "/containers/app", config.Updates[1].Directory)
	})

	t.Run("consolidated entry with directories and groups", func(t *testing.T) {
		input := `version: 2

updates:
  - package-ecosystem: "gomod"
    directories:
      - "/go/core"
      - "/go/cli/eac"
      - "/go/clibase"
    schedule:
      interval: "weekly"
      day: "monday"
    commit-message:
      prefix: "chore(deps)"
    labels:
      - "dependencies"
      - "go"
    groups:
      golang-x:
        patterns:
          - "golang.org/x/*"
      all-go-deps:
        group-by: "dependency-name"
`
		config, err := ParseBytes([]byte(input))
		require.NoError(t, err)

		require.Len(t, config.Updates, 1)
		u := config.Updates[0]
		assert.Equal(t, "gomod", u.PackageEcosystem)
		assert.Empty(t, u.Directory)
		assert.Equal(t, []string{"/go/core", "/go/cli/eac", "/go/clibase"}, u.Directories)
		assert.Equal(t, "weekly", u.Schedule.Interval)
		assert.Equal(t, "monday", u.Schedule.Day)
		assert.Equal(t, []string{"dependencies", "go"}, u.Labels)
		assert.True(t, u.IsConsolidated())

		require.Len(t, u.Groups, 2)
		assert.Equal(t, []string{"golang.org/x/*"}, u.Groups["golang-x"].Patterns)
		assert.Equal(t, "dependency-name", u.Groups["all-go-deps"].GroupBy)
	})

	t.Run("mixed singular and consolidated entries", func(t *testing.T) {
		input := `version: 2

updates:
  - package-ecosystem: "gomod"
    directories:
      - "/go/core"
    schedule:
      interval: "weekly"
    labels:
      - "dependencies"
  - package-ecosystem: "npm"
    directory: "/web"
    schedule:
      interval: "weekly"
    labels:
      - "dependencies"
`
		config, err := ParseBytes([]byte(input))
		require.NoError(t, err)

		require.Len(t, config.Updates, 2)
		assert.True(t, config.Updates[0].IsConsolidated())
		assert.False(t, config.Updates[1].IsConsolidated())
	})

	t.Run("values without quotes", func(t *testing.T) {
		input := `version: 2
updates:
  - package-ecosystem: gomod
    directory: /go/core
    schedule:
      interval: weekly
`
		config, err := ParseBytes([]byte(input))
		require.NoError(t, err)

		require.Len(t, config.Updates, 1)
		assert.Equal(t, "gomod", config.Updates[0].PackageEcosystem)
		assert.Equal(t, "/go/core", config.Updates[0].Directory)
		assert.Equal(t, "weekly", config.Updates[0].Schedule.Interval)
	})
}

func TestFormatConfig(t *testing.T) {
	t.Run("multiple ecosystems formatted correctly", func(t *testing.T) {
		config := &DependabotConfig{
			Version: 2,
			Updates: []UpdateEntry{
				{
					PackageEcosystem: "github-actions",
					Directory:        "/",
					Schedule:         Schedule{Interval: "weekly", Day: "monday"},
					CommitMessage:    &CommitMessage{Prefix: "chore(deps)"},
					Labels:           []string{"dependencies", "github-actions"},
				},
				{
					PackageEcosystem: "gomod",
					Directory:        "/go/core",
					Schedule:         Schedule{Interval: "weekly"},
					Labels:           []string{"dependencies", "go"},
				},
			},
		}

		output := FormatConfig(config)

		// Must start with version header
		assert.True(t, strings.HasPrefix(output, "version: 2\n\nupdates:\n"),
			"output should start with version header")

		// Verify ecosystem comments
		assert.Contains(t, output, "# GitHub Actions")
		assert.Contains(t, output, "# Go modules - go/core")

		// Verify quoted values
		assert.Contains(t, output, `package-ecosystem: "github-actions"`)
		assert.Contains(t, output, `directory: "/"`)
		assert.Contains(t, output, `package-ecosystem: "gomod"`)
		assert.Contains(t, output, `directory: "/go/core"`)
		assert.Contains(t, output, `interval: "weekly"`)

		// Verify labels are quoted
		assert.Contains(t, output, `- "dependencies"`)
		assert.Contains(t, output, `- "github-actions"`)
		assert.Contains(t, output, `- "go"`)
	})

	t.Run("optional schedule fields included when set", func(t *testing.T) {
		config := &DependabotConfig{
			Version: 2,
			Updates: []UpdateEntry{
				{
					PackageEcosystem: "npm",
					Directory:        "/web",
					Schedule: Schedule{
						Interval: "weekly",
						Day:      "monday",
						Time:     "09:00",
						Timezone: "America/New_York",
					},
				},
			},
		}

		output := FormatConfig(config)

		assert.Contains(t, output, `day: "monday"`)
		assert.Contains(t, output, `time: "09:00"`)
		assert.Contains(t, output, `timezone: "America/New_York"`)
	})

	t.Run("optional schedule fields omitted when empty", func(t *testing.T) {
		config := &DependabotConfig{
			Version: 2,
			Updates: []UpdateEntry{
				{
					PackageEcosystem: "npm",
					Directory:        "/web",
					Schedule:         Schedule{Interval: "daily"},
				},
			},
		}

		output := FormatConfig(config)

		assert.NotContains(t, output, "day:")
		assert.NotContains(t, output, "time:")
		assert.NotContains(t, output, "timezone:")
	})

	t.Run("commit-message section included when set", func(t *testing.T) {
		config := &DependabotConfig{
			Version: 2,
			Updates: []UpdateEntry{
				{
					PackageEcosystem: "gomod",
					Directory:        "/go/core",
					Schedule:         Schedule{Interval: "weekly"},
					CommitMessage:    &CommitMessage{Prefix: "chore(deps)"},
				},
			},
		}

		output := FormatConfig(config)

		assert.Contains(t, output, "commit-message:")
		assert.Contains(t, output, `prefix: "chore(deps)"`)
	})

	t.Run("commit-message section omitted when nil", func(t *testing.T) {
		config := &DependabotConfig{
			Version: 2,
			Updates: []UpdateEntry{
				{
					PackageEcosystem: "gomod",
					Directory:        "/go/core",
					Schedule:         Schedule{Interval: "weekly"},
				},
			},
		}

		output := FormatConfig(config)

		assert.NotContains(t, output, "commit-message:")
		assert.NotContains(t, output, "prefix:")
	})

	t.Run("section separator between different ecosystems", func(t *testing.T) {
		config := &DependabotConfig{
			Version: 2,
			Updates: []UpdateEntry{
				{
					PackageEcosystem: "github-actions",
					Directory:        "/",
					Schedule:         Schedule{Interval: "weekly"},
				},
				{
					PackageEcosystem: "gomod",
					Directory:        "/go/core",
					Schedule:         Schedule{Interval: "weekly"},
				},
			},
		}

		output := FormatConfig(config)

		// There should be a blank line between the two entries (different ecosystems)
		lines := strings.Split(output, "\n")
		foundBlankBetween := false
		inFirstEntry := false
		for _, line := range lines {
			if strings.Contains(line, "github-actions") {
				inFirstEntry = true
			}
			if inFirstEntry && line == "" {
				foundBlankBetween = true
			}
			if foundBlankBetween && strings.Contains(line, "gomod") {
				break
			}
		}
		assert.True(t, foundBlankBetween, "should have blank line between different ecosystems")
	})

	t.Run("open-pull-requests-limit formatted when set", func(t *testing.T) {
		limit := 5
		config := &DependabotConfig{
			Version: 2,
			Updates: []UpdateEntry{
				{
					PackageEcosystem: "npm",
					Directory:        "/web",
					Schedule:         Schedule{Interval: "weekly"},
					OpenPullRequests: &limit,
				},
			},
		}

		output := FormatConfig(config)
		assert.Contains(t, output, "open-pull-requests-limit: 5")
	})

	t.Run("reviewers and assignees formatted when set", func(t *testing.T) {
		config := &DependabotConfig{
			Version: 2,
			Updates: []UpdateEntry{
				{
					PackageEcosystem: "gomod",
					Directory:        "/go/core",
					Schedule:         Schedule{Interval: "weekly"},
					Reviewers:        []string{"alice", "bob"},
					Assignees:        []string{"charlie"},
				},
			},
		}

		output := FormatConfig(config)
		assert.Contains(t, output, "reviewers:")
		assert.Contains(t, output, `- "alice"`)
		assert.Contains(t, output, `- "bob"`)
		assert.Contains(t, output, "assignees:")
		assert.Contains(t, output, `- "charlie"`)
	})

	t.Run("consolidated entry formats directories and groups", func(t *testing.T) {
		config := &DependabotConfig{
			Version: 2,
			Updates: []UpdateEntry{
				{
					PackageEcosystem: "gomod",
					Directories:      []string{"/go/core", "/go/cli/eac"},
					Schedule:         Schedule{Interval: "weekly", Day: "monday"},
					CommitMessage:    &CommitMessage{Prefix: "chore(deps)"},
					Labels:           []string{"dependencies", "go"},
					Groups: map[string]Group{
						"golang-x":    {Patterns: []string{"golang.org/x/*"}},
						"all-go-deps": {GroupBy: "dependency-name"},
					},
				},
			},
		}

		output := FormatConfig(config)

		assert.Contains(t, output, "# Go modules - all workspace modules (consolidated)")
		assert.Contains(t, output, "directories:")
		assert.Contains(t, output, `- "/go/core"`)
		assert.Contains(t, output, `- "/go/cli/eac"`)
		assert.NotContains(t, output, "directory:")
		assert.Contains(t, output, "groups:")
		assert.Contains(t, output, "golang-x:")
		assert.Contains(t, output, `- "golang.org/x/*"`)
		assert.Contains(t, output, "all-go-deps:")
		assert.Contains(t, output, `group-by: "dependency-name"`)
	})

	t.Run("consolidated entry roundtrips through format and parse", func(t *testing.T) {
		config := &DependabotConfig{
			Version: 2,
			Updates: []UpdateEntry{
				{
					PackageEcosystem: "gomod",
					Directories:      []string{"/go/core", "/go/cli/eac"},
					Schedule:         Schedule{Interval: "weekly", Day: "monday"},
					CommitMessage:    &CommitMessage{Prefix: "chore(deps)"},
					Labels:           []string{"dependencies", "go"},
					Groups: map[string]Group{
						"golang-x":    {Patterns: []string{"golang.org/x/*"}},
						"all-go-deps": {GroupBy: "dependency-name"},
					},
				},
			},
		}

		formatted := FormatConfig(config)
		parsed, err := ParseBytes([]byte(formatted))
		require.NoError(t, err)

		require.Len(t, parsed.Updates, 1)
		u := parsed.Updates[0]
		assert.Equal(t, []string{"/go/core", "/go/cli/eac"}, u.Directories)
		assert.Empty(t, u.Directory)
		assert.True(t, u.IsConsolidated())
		assert.Equal(t, []string{"golang.org/x/*"}, u.Groups["golang-x"].Patterns)
		assert.Equal(t, "dependency-name", u.Groups["all-go-deps"].GroupBy)
	})

	t.Run("empty updates produces header only", func(t *testing.T) {
		config := &DependabotConfig{
			Version: 2,
			Updates: []UpdateEntry{},
		}

		output := FormatConfig(config)
		assert.Equal(t, "version: 2\n\nupdates:\n", output)
	})

	t.Run("entry comment for each ecosystem type", func(t *testing.T) {
		ecosystems := []struct {
			ecosystem string
			directory string
			wantComment string
		}{
			{"github-actions", "/", "# GitHub Actions"},
			{"gomod", "/go/core", "# Go modules - go/core"},
			{"npm", "/web/app", "# npm - web/app"},
			{"docker", "/containers/app", "# Docker - containers/app"},
			{"pip", "/scripts", "# Python - scripts"},
		}

		for _, tc := range ecosystems {
			t.Run(tc.ecosystem, func(t *testing.T) {
				config := &DependabotConfig{
					Version: 2,
					Updates: []UpdateEntry{
						{
							PackageEcosystem: tc.ecosystem,
							Directory:        tc.directory,
							Schedule:         Schedule{Interval: "weekly"},
						},
					},
				}

				output := FormatConfig(config)
				assert.Contains(t, output, tc.wantComment)
			})
		}
	})
}

func TestDefaultUpdateEntry(t *testing.T) {
	tests := []struct {
		name       string
		ecosystem  string
		directory  string
		wantLabels []string
	}{
		{
			name:       "gomod ecosystem",
			ecosystem:  "gomod",
			directory:  "/go/core",
			wantLabels: []string{"dependencies", "go"},
		},
		{
			name:       "npm ecosystem",
			ecosystem:  "npm",
			directory:  "/web/app",
			wantLabels: []string{"dependencies", "npm"},
		},
		{
			name:       "pip ecosystem",
			ecosystem:  "pip",
			directory:  "/scripts",
			wantLabels: []string{"dependencies", "python"},
		},
		{
			name:       "docker ecosystem",
			ecosystem:  "docker",
			directory:  "/containers/app",
			wantLabels: []string{"dependencies", "docker"},
		},
		{
			name:       "github-actions ecosystem",
			ecosystem:  "github-actions",
			directory:  "/",
			wantLabels: []string{"dependencies", "github-actions"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := EcosystemEntry{
				Ecosystem: tt.ecosystem,
				Directory: tt.directory,
			}

			result := DefaultUpdateEntry(entry)

			assert.Equal(t, tt.ecosystem, result.PackageEcosystem)
			assert.Equal(t, tt.directory, result.Directory)
			assert.Equal(t, "weekly", result.Schedule.Interval)
			assert.Equal(t, "monday", result.Schedule.Day)
			require.NotNil(t, result.CommitMessage)
			assert.Equal(t, "chore(deps)", result.CommitMessage.Prefix)
			assert.Equal(t, tt.wantLabels, result.Labels)
		})
	}
}

func TestSortUpdates(t *testing.T) {
	t.Run("sorts by ecosystem order then directory", func(t *testing.T) {
		updates := []UpdateEntry{
			{PackageEcosystem: "pip", Directory: "/scripts"},
			{PackageEcosystem: "gomod", Directory: "/go/core"},
			{PackageEcosystem: "docker", Directory: "/containers/app"},
			{PackageEcosystem: "github-actions", Directory: "/"},
			{PackageEcosystem: "npm", Directory: "/web"},
			{PackageEcosystem: "gomod", Directory: "/go/adapters/ai"},
		}

		SortUpdates(updates)

		assert.Equal(t, "github-actions", updates[0].PackageEcosystem)
		assert.Equal(t, "/", updates[0].Directory)

		assert.Equal(t, "gomod", updates[1].PackageEcosystem)
		assert.Equal(t, "/go/adapters/ai", updates[1].Directory)

		assert.Equal(t, "gomod", updates[2].PackageEcosystem)
		assert.Equal(t, "/go/core", updates[2].Directory)

		assert.Equal(t, "npm", updates[3].PackageEcosystem)
		assert.Equal(t, "/web", updates[3].Directory)

		assert.Equal(t, "docker", updates[4].PackageEcosystem)
		assert.Equal(t, "/containers/app", updates[4].Directory)

		assert.Equal(t, "pip", updates[5].PackageEcosystem)
		assert.Equal(t, "/scripts", updates[5].Directory)
	})

	t.Run("alphabetical within same ecosystem", func(t *testing.T) {
		updates := []UpdateEntry{
			{PackageEcosystem: "gomod", Directory: "/go/core"},
			{PackageEcosystem: "gomod", Directory: "/go/adapters/docker"},
			{PackageEcosystem: "gomod", Directory: "/go/adapters/ai"},
			{PackageEcosystem: "gomod", Directory: "/go/cli/clie"},
		}

		SortUpdates(updates)

		assert.Equal(t, "/go/adapters/ai", updates[0].Directory)
		assert.Equal(t, "/go/adapters/docker", updates[1].Directory)
		assert.Equal(t, "/go/cli/clie", updates[2].Directory)
		assert.Equal(t, "/go/core", updates[3].Directory)
	})

	t.Run("empty slice does not panic", func(t *testing.T) {
		updates := []UpdateEntry{}
		assert.NotPanics(t, func() {
			SortUpdates(updates)
		})
	})

	t.Run("single element remains unchanged", func(t *testing.T) {
		updates := []UpdateEntry{
			{PackageEcosystem: "npm", Directory: "/web"},
		}

		SortUpdates(updates)

		assert.Equal(t, "npm", updates[0].PackageEcosystem)
		assert.Equal(t, "/web", updates[0].Directory)
	})
}
