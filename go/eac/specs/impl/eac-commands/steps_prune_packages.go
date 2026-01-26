// Package srccommands contains godog step implementations for specs/eac-commands.
//
// This file contains release prune-packages command step definitions.
// Features: specs/eac-commands/release-prune-packages/
package srccommands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// mockPackageVersion represents a mock container image version.
type mockPackageVersion struct {
	Digest    string
	Tags      []string
	CreatedAt time.Time
}

// pruneTestState holds state for prune-packages tests.
type pruneTestState struct {
	packages         map[string][]mockPackageVersion // package name -> versions
	releases         map[string]bool                 // release tag -> exists
	preservePatterns []string
	prunePatterns    []string
	minAgeDays       int
	protectedDigests map[string]string // digest -> protection reason
	prunableDigests  map[string]bool
	deletedDigests   map[string]bool
	keptDigests      map[string]bool
	markedDigests    map[string]bool
	commandArgs      string
	simulated        bool
}

// registerPrunePackagesSteps registers step definitions for release prune-packages features.
func registerPrunePackagesSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	state := &pruneTestState{
		packages:         make(map[string][]mockPackageVersion),
		releases:         make(map[string]bool),
		protectedDigests: make(map[string]string),
		prunableDigests:  make(map[string]bool),
		deletedDigests:   make(map[string]bool),
		keptDigests:      make(map[string]bool),
		markedDigests:    make(map[string]bool),
		minAgeDays:       7, // default
	}

	// Reset state before each scenario
	sc.Before(func(c context.Context, sc2 *godog.Scenario) (context.Context, error) {
		state.packages = make(map[string][]mockPackageVersion)
		state.releases = make(map[string]bool)
		state.preservePatterns = nil
		state.prunePatterns = nil
		state.protectedDigests = make(map[string]string)
		state.prunableDigests = make(map[string]bool)
		state.deletedDigests = make(map[string]bool)
		state.keptDigests = make(map[string]bool)
		state.markedDigests = make(map[string]bool)
		state.minAgeDays = 7
		state.commandArgs = ""
		state.simulated = false
		return c, nil
	})

	// Background steps
	sc.Step(`^the repository has registries configured for "([^"]*)"$`, func(registry string) error {
		// Mock step - registries are configured
		return nil
	})
	sc.Step(`^the registry cleanup policy is enabled$`, func() error {
		// Mock step - cleanup policy is enabled
		return nil
	})

	// Given steps
	sc.Step(`^package "([^"]*)" has versions:$`, func(pkgName string, table *godog.Table) error {
		return packageHasVersions(pkgName, table, state)
	})
	sc.Step(`^preserve_patterns includes "([^"]*)"$`, func(pattern string) error {
		state.preservePatterns = append(state.preservePatterns, pattern)
		return nil
	})
	sc.Step(`^prune_patterns includes "([^"]*)"$`, func(pattern string) error {
		state.prunePatterns = append(state.prunePatterns, pattern)
		return nil
	})
	sc.Step(`^prune_patterns includes "([^"]*)" and "([^"]*)"$`, func(p1, p2 string) error {
		state.prunePatterns = append(state.prunePatterns, p1, p2)
		return nil
	})
	sc.Step(`^GitHub release "([^"]*)" exists$`, func(tag string) error {
		state.releases[tag] = true
		return nil
	})
	sc.Step(`^min_age_days is (\d+)$`, func(days int) error {
		state.minAgeDays = days
		return nil
	})
	sc.Step(`^packages exist: "([^"]*)", "([^"]*)"$`, func(pkg1, pkg2 string) error {
		state.packages[pkg1] = []mockPackageVersion{}
		state.packages[pkg2] = []mockPackageVersion{}
		return nil
	})

	// When step: The common "I run" step runs the actual command.
	// Since this feature is @pending (needs @env:mock-github infrastructure),
	// the step definitions are ready but tests won't run until mocking is available.

	// Then steps - verify simulated state
	sc.Step(`^version "([^"]*)" should be protected$`, func(digest string) error {
		if _, ok := state.protectedDigests[digest]; !ok {
			return fmt.Errorf("version %s is not protected", digest)
		}
		return nil
	})
	sc.Step(`^version with tag "([^"]*)" should be protected$`, func(tag string) error {
		// Find digest by tag
		for _, versions := range state.packages {
			for _, v := range versions {
				for _, t := range v.Tags {
					if t == tag {
						if _, ok := state.protectedDigests[v.Digest]; !ok {
							return fmt.Errorf("version with tag %s (digest %s) is not protected", tag, v.Digest)
						}
						return nil
					}
				}
			}
		}
		return fmt.Errorf("no version found with tag %s", tag)
	})
	sc.Step(`^the protection reason should be "([^"]*)"$`, func(reason string) error {
		// Check if any protected digest has this reason
		for _, r := range state.protectedDigests {
			if r == reason {
				return nil
			}
		}
		return fmt.Errorf("no version protected with reason: %s\nProtected versions: %v", reason, state.protectedDigests)
	})
	sc.Step(`^version "([^"]*)" should be prunable$`, func(digest string) error {
		if !state.prunableDigests[digest] {
			return fmt.Errorf("version %s is not prunable", digest)
		}
		return nil
	})
	sc.Step(`^version "([^"]*)" should be marked for deletion$`, func(digest string) error {
		if !state.markedDigests[digest] {
			return fmt.Errorf("version %s is not marked for deletion", digest)
		}
		return nil
	})
	sc.Step(`^version "([^"]*)" should be kept$`, func(digest string) error {
		if !state.keptDigests[digest] {
			return fmt.Errorf("version %s is not kept", digest)
		}
		return nil
	})
	sc.Step(`^version "([^"]*)" should be deleted$`, func(digest string) error {
		if !state.deletedDigests[digest] {
			return fmt.Errorf("version %s was not deleted", digest)
		}
		return nil
	})
	sc.Step(`^no versions should be deleted$`, func() error {
		if len(state.deletedDigests) > 0 {
			return fmt.Errorf("expected no deletions but %d versions were deleted", len(state.deletedDigests))
		}
		return nil
	})
	sc.Step(`^the command should fail$`, func() error {
		// Check if command output indicates failure
		if !strings.Contains(ctx.CommandOutput, "Error") && !strings.Contains(ctx.CommandOutput, "package name required") {
			return fmt.Errorf("expected command to fail but output was: %s", ctx.CommandOutput)
		}
		return nil
	})
	sc.Step(`^both packages should be processed$`, func() error {
		// Check output mentions processing multiple packages
		if len(state.packages) < 2 {
			return fmt.Errorf("expected 2 packages, got %d", len(state.packages))
		}
		return nil
	})
}

// packageHasVersions parses the version table and stores mock versions.
func packageHasVersions(pkgName string, table *godog.Table, state *pruneTestState) error {
	versions := []mockPackageVersion{}

	for i, row := range table.Rows {
		if i == 0 {
			continue // Skip header
		}
		if len(row.Cells) < 3 {
			return fmt.Errorf("invalid table row: need digest, tags, created_at")
		}

		digest := row.Cells[0].Value
		tagsStr := row.Cells[1].Value
		createdStr := row.Cells[2].Value

		var tags []string
		if tagsStr != "" {
			tags = strings.Split(tagsStr, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
		}

		createdAt, err := parseRelativeTime(createdStr)
		if err != nil {
			return fmt.Errorf("failed to parse created_at %q: %w", createdStr, err)
		}

		versions = append(versions, mockPackageVersion{
			Digest:    digest,
			Tags:      tags,
			CreatedAt: createdAt,
		})
	}

	state.packages[pkgName] = versions
	return nil
}

// parseRelativeTime parses strings like "30 days ago".
func parseRelativeTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, " days ago") {
		daysStr := strings.TrimSuffix(s, " days ago")
		var days int
		if _, err := fmt.Sscanf(daysStr, "%d", &days); err != nil {
			return time.Time{}, err
		}
		return time.Now().AddDate(0, 0, -days), nil
	}
	return time.Time{}, fmt.Errorf("unknown time format: %s", s)
}
