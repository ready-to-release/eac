package workunit

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"
)

// UnitState represents the cached state of a unit of work.
// It tracks what was executed and when for cache invalidation decisions.
type UnitState struct {
	ID             UnitID    `json:"id"`
	SourceHash     string    `json:"source_hash"`
	BuildID        string    `json:"build_id,omitempty"`        // For test/scan - the build this was tested against
	DependencyHash string    `json:"dependency_hash,omitempty"` // For integration tests - hash of dependency build IDs
	Passed         bool      `json:"passed"`
	ExecutedAt     time.Time `json:"executed_at"`
}

// InvalidationRule defines when a cached unit should be re-executed.
type InvalidationRule struct {
	OnSourceChange     bool // Source files changed
	OnBuildChange      bool // Build artifact changed
	OnDependencyChange bool // Dependency module changed
	OnFailure          bool // Previous run failed
}

// DefaultRules maps each context to its default invalidation rule.
var DefaultRules = map[Context]InvalidationRule{
	ContextBuild: {OnSourceChange: true, OnFailure: true},
	ContextTest:  {OnSourceChange: true, OnBuildChange: true, OnFailure: true},
	ContextLint:  {OnSourceChange: true, OnFailure: true},
	ContextScan:  {OnSourceChange: true, OnBuildChange: true, OnFailure: true},
}

// IntegrationTestRule is the invalidation rule for integration tests.
// It includes OnDependencyChange because integration tests exercise
// cross-module interactions and must re-run when dependencies change.
var IntegrationTestRule = InvalidationRule{
	OnSourceChange:     true,
	OnBuildChange:      true,
	OnDependencyChange: true,
	OnFailure:          true,
}

// TestSet represents a classification of tests by isolation level.
// This determines invalidation behavior for test caching.
type TestSet string

const (
	// TestSetUnit represents fully isolated unit tests (L0, L1).
	// These tests do NOT need to be invalidated when transitive dependencies change.
	TestSetUnit TestSet = "unit"

	// TestSetIntegration represents integration tests that may have external dependencies (L2+).
	// These tests SHOULD be invalidated when transitive dependencies change.
	TestSetIntegration TestSet = "integration"
)

// GetTestSetForLTag returns the test set for a given L-tag.
// L0 and L1 are unit tests (fully isolated).
// L2, L3, and L4 are integration tests (may have external dependencies).
// Unrecognized tags default to unit for safety.
func GetTestSetForLTag(ltag string) TestSet {
	switch ltag {
	case "@L0", "@L1":
		return TestSetUnit
	case "@L2", "@L3", "@L4":
		return TestSetIntegration
	default:
		return TestSetUnit // Default to unit for untagged tests
	}
}

// ClassifyTestByTags determines the test set for a test based on its tags.
// If any tag indicates integration (L2+), the test is classified as integration.
// Otherwise, it defaults to unit.
func ClassifyTestByTags(tags []string) TestSet {
	for _, tag := range tags {
		if ts := GetTestSetForLTag(tag); ts == TestSetIntegration {
			return TestSetIntegration
		}
	}
	return TestSetUnit
}

// GetRuleForTestSet returns the invalidation rule for a test set.
func GetRuleForTestSet(ts TestSet) InvalidationRule {
	if ts == TestSetIntegration {
		return IntegrationTestRule
	}
	return DefaultRules[ContextTest]
}

// ComputeDependencyBuildHash computes a hash of all transitive dependency BuildIDs.
// This is used for integration test invalidation - if any dependency's BuildID changes,
// the hash will change and integration tests need to be re-run.
// Returns empty string if no dependencies.
func ComputeDependencyBuildHash(dependencyBuildIDs map[string]string) string {
	if len(dependencyBuildIDs) == 0 {
		return ""
	}

	// Sort keys for deterministic hashing
	keys := make([]string, 0, len(dependencyBuildIDs))
	for k := range dependencyBuildIDs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build hash input
	var builder strings.Builder
	for _, k := range keys {
		builder.WriteString(k)
		builder.WriteString(":")
		builder.WriteString(dependencyBuildIDs[k])
		builder.WriteString(";")
	}

	hash := sha256.Sum256([]byte(builder.String()))
	return hex.EncodeToString(hash[:8]) // First 8 bytes for readability
}
