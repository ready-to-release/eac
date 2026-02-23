// Package dependabot provides shared logic for discovering, parsing, and comparing
// Dependabot configuration against actual dependency sources in the repository.
package dependabot

import "sort"

// DependabotConfig represents the full .github/dependabot.yml file.
type DependabotConfig struct {
	Version int           `yaml:"version"`
	Updates []UpdateEntry `yaml:"updates"`
}

// UpdateEntry represents a single ecosystem update configuration in dependabot.yml.
type UpdateEntry struct {
	PackageEcosystem string         `yaml:"package-ecosystem"`
	Directory        string         `yaml:"directory"`
	Schedule         Schedule       `yaml:"schedule"`
	CommitMessage    *CommitMessage `yaml:"commit-message,omitempty"`
	Labels           []string       `yaml:"labels,omitempty"`
	Reviewers        []string       `yaml:"reviewers,omitempty"`
	Assignees        []string       `yaml:"assignees,omitempty"`
	OpenPullRequests *int           `yaml:"open-pull-requests-limit,omitempty"`
}

// Schedule defines how often Dependabot checks for updates.
type Schedule struct {
	Interval string `yaml:"interval"`
	Day      string `yaml:"day,omitempty"`
	Time     string `yaml:"time,omitempty"`
	Timezone string `yaml:"timezone,omitempty"`
}

// CommitMessage configures the commit message prefix for Dependabot PRs.
type CommitMessage struct {
	Prefix string `yaml:"prefix,omitempty"`
}

// EcosystemEntry represents a discovered dependency source in the repository.
type EcosystemEntry struct {
	Ecosystem string // Dependabot ecosystem identifier (gomod, npm, pip, docker, github-actions)
	Directory string // Path relative to repo root with leading slash (e.g., "/go/core")
}

// Key returns a unique identifier for matching entries.
func (e EcosystemEntry) Key() string {
	return e.Ecosystem + ":" + e.Directory
}

// Key returns a unique identifier for matching entries.
func (u UpdateEntry) Key() string {
	return u.PackageEcosystem + ":" + u.Directory
}

// ComparisonReport contains the results of comparing declared vs discovered entries.
type ComparisonReport struct {
	Declared   []UpdateEntry
	Discovered []EcosystemEntry
	Missing    []EcosystemEntry // In filesystem but not in dependabot.yml
	Extra      []UpdateEntry    // In dependabot.yml but no filesystem source
	Matched    []EcosystemEntry // Present in both
}

// HasIssues returns true if there are missing or extra entries.
func (r *ComparisonReport) HasIssues() bool {
	return len(r.Missing) > 0 || len(r.Extra) > 0
}

// ecosystemOrder defines the sort order for ecosystems in the YAML output.
var ecosystemOrder = map[string]int{
	"github-actions": 0,
	"gomod":          1,
	"npm":            2,
	"docker":         3,
	"pip":            4,
}

// SortUpdates sorts update entries by ecosystem order, then by directory.
func SortUpdates(updates []UpdateEntry) {
	sort.Slice(updates, func(i, j int) bool {
		oi := ecosystemOrder[updates[i].PackageEcosystem]
		oj := ecosystemOrder[updates[j].PackageEcosystem]
		if oi != oj {
			return oi < oj
		}
		return updates[i].Directory < updates[j].Directory
	})
}
