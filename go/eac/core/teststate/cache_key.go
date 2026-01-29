// Package teststate provides granular test caching with isolation-aware invalidation.
// This file contains the cache key type for granular test result caching.
package teststate

import (
	"fmt"
	"strings"
)

// TestCacheKey uniquely identifies a cached test result.
// The key format is: module:component:type:testset
//
// Examples:
//   - eac-commands:go:gotest:unit
//   - eac-commands:gherkin:godog:integration
//   - r2r-cli:go:gotest:unit
type TestCacheKey struct {
	// Module moniker (e.g., "eac-commands")
	Module string

	// Component type (e.g., "go", "gherkin", "typescript")
	Component string

	// Test type (e.g., "gotest", "godog", "mocha", "tscucumber")
	TestType string

	// Test isolation level (unit, integration)
	TestSet TestSet
}

// String returns the cache key as a string for storage.
func (k TestCacheKey) String() string {
	return fmt.Sprintf("%s:%s:%s:%s", k.Module, k.Component, k.TestType, k.TestSet)
}

// ParseTestCacheKey parses a cache key string.
// Returns an error if the format is invalid.
func ParseTestCacheKey(s string) (TestCacheKey, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 4 {
		return TestCacheKey{}, fmt.Errorf("invalid cache key format: %s (expected module:component:type:testset)", s)
	}

	testSet := TestSet(parts[3])
	if testSet != TestSetUnit && testSet != TestSetIntegration {
		return TestCacheKey{}, fmt.Errorf("invalid test set in cache key: %s (expected 'unit' or 'integration')", parts[3])
	}

	return TestCacheKey{
		Module:    parts[0],
		Component: parts[1],
		TestType:  parts[2],
		TestSet:   testSet,
	}, nil
}

// IsValid returns true if the cache key has all required fields.
func (k TestCacheKey) IsValid() bool {
	return k.Module != "" && k.Component != "" && k.TestType != "" &&
		(k.TestSet == TestSetUnit || k.TestSet == TestSetIntegration)
}
