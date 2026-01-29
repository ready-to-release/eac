// Package teststate provides granular test caching with isolation-aware invalidation.
// This file contains test set classification and invalidation rules.
package teststate

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"
)

// TestSet represents a classification of tests by isolation level.
type TestSet string

const (
	// TestSetUnit represents fully isolated unit tests (L0, L1).
	// These tests do NOT need to be invalidated when transitive dependencies change.
	TestSetUnit TestSet = "unit"

	// TestSetIntegration represents integration tests that may have external dependencies (L2+).
	// These tests SHOULD be invalidated when transitive dependencies change.
	TestSetIntegration TestSet = "integration"
)

// InvalidationRule defines when cached tests should be invalidated.
type InvalidationRule struct {
	// Test set this rule applies to
	TestSet TestSet

	// Invalidate when module's own source files change
	OnModuleChange bool

	// Invalidate when a transitive dependency changes
	OnTransitiveChange bool

	// Invalidate when module's build changes (rebuild detected)
	OnBuildChange bool

	// Invalidate when dependency's build changes
	OnDependencyBuildChange bool
}

// DefaultInvalidationRules returns the standard invalidation rules.
// Unit tests (L0/L1) are isolated and ignore transitive dependency changes.
// Integration tests (L2+) are invalidated when any dependency in the chain changes.
func DefaultInvalidationRules() map[TestSet]InvalidationRule {
	return map[TestSet]InvalidationRule{
		TestSetUnit: {
			TestSet:                 TestSetUnit,
			OnModuleChange:          true,  // YES: module source changed
			OnTransitiveChange:      false, // NO: unit tests are isolated
			OnBuildChange:           true,  // YES: rebuild detected
			OnDependencyBuildChange: false, // NO: unit tests are isolated
		},
		TestSetIntegration: {
			TestSet:                 TestSetIntegration,
			OnModuleChange:          true, // YES: module source changed
			OnTransitiveChange:      true, // YES: integration may depend on deps
			OnBuildChange:           true, // YES: rebuild detected
			OnDependencyBuildChange: true, // YES: deps affect integration tests
		},
	}
}

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

// TestSetState represents cached state for a specific test set within a module.
// This enables granular caching where unit tests can be cached independently from
// integration tests, with different invalidation rules for each.
type TestSetState struct {
	// Hash of source files in the module
	SourceHash string `json:"source_hash"`

	// Hash of test files for this test set
	TestSetHash string `json:"testset_hash"`

	// Build ID at time of test
	BuildID string `json:"build_id,omitempty"`

	// Whether last run passed
	Passed bool `json:"passed"`

	// Last test timestamp
	TestedAt time.Time `json:"tested_at"`

	// L-tags included in this test set
	LevelTags []string `json:"level_tags"`

	// Hash of transitive dependency BuildIDs (only for integration tests)
	DependencyBuildHash string `json:"dependency_build_hash,omitempty"`

	// Test count for reference
	TestCount int `json:"test_count"`

	// Duration of last run
	DurationSeconds float64 `json:"duration_seconds"`
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
