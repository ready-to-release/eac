package testing

import (
	"strings"

	"github.com/ready-to-release/eac/src/core/config"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// ApplyInferences applies inference rules to enrich test tags
// NOTE: L-level tags (@L0-@L4) are NEVER inferred - tests MUST have explicit L-tags.
// Verification tags (@ov) ARE inferred: if no verification tag is present, @ov is added.
func ApplyInferences(tests []TestReference, inferences []Inference) []TestReference {
	enriched := make([]TestReference, len(tests))

	for i, test := range tests {
		enriched[i] = test
		enriched[i].Tags = copyTags(test.Tags)

		// Apply each inference rule
		for _, inference := range inferences {
			// Skip if inference doesn't apply to this test type
			if len(inference.TestTypes) > 0 && !contains(inference.TestTypes, test.Type) {
				continue
			}

			// Skip level inferences if test already has explicit level
			if isLevelInference(inference) && hasAnyLevelTag(test.Tags) {
				continue
			}

			// Check if conditions match
			if matchesConditions(test.Tags, inference.IfTags, inference.ThenAddTags) {
				// Add inferred tags
				for _, tag := range inference.ThenAddTags {
					if !contains(enriched[i].Tags, tag) {
						enriched[i].Tags = append(enriched[i].Tags, tag)
					}
				}
			}
		}

		// Derive @ov if no other verification tag is present
		enriched[i].Tags = DeriveOperationalVerification(enriched[i].Tags)
	}

	return enriched
}

// DeriveOperationalVerification adds @ov tag if no other verification tags present
// This allows Go tests and other tests that can't explicitly declare verification tags
// to be automatically classified as operational verification tests.
func DeriveOperationalVerification(tags []string) []string {
	verificationTags := GetVerificationTags()
	if verificationTags == nil {
		// Config unavailable - can't determine verification tags
		// Default to adding @ov as a safe fallback
		if !contains(tags, "@ov") {
			return append(tags, "@ov")
		}
		return tags
	}

	// Check if any verification tag is already present
	hasVerificationTag := false
	for _, tag := range tags {
		if contains(verificationTags, tag) {
			hasVerificationTag = true
			break
		}
	}

	// @ov = no other verification tag present
	if !hasVerificationTag {
		return append(tags, "@ov")
	}

	return tags
}

// hasAnyLevelTag checks if tags contain any level tag (@L0-@L4)
// Returns true if config unavailable (fail-closed: assume level tags may be present)
func hasAnyLevelTag(tags []string) bool {
	levelTags := GetLevelTags()
	if levelTags == nil {
		// Fail closed: without config, assume level tags may be present
		// This prevents inferences from running when we can't verify
		return true
	}
	for _, tag := range tags {
		if contains(levelTags, tag) {
			return true
		}
	}
	return false
}

// isLevelInference checks if inference adds level tags
// Returns true if config unavailable (fail-closed: skip inference if we can't verify)
func isLevelInference(inference Inference) bool {
	levelTags := GetLevelTags()
	if levelTags == nil {
		// Fail closed: without config, assume inference might add level tags
		// This causes the inference to be skipped (safe default)
		return true
	}
	for _, tag := range inference.ThenAddTags {
		if contains(levelTags, tag) {
			return true
		}
	}
	return false
}


// matchesConditions checks if tags match inference conditions
func matchesConditions(tags []string, conditions []string, thenAddTags []string) bool {
	// Special case: dependency inferences always apply (regardless of level tags)
	if len(conditions) == 0 && isDependencyInference(thenAddTags) {
		return true
	}

	// Empty conditions = "no level tags present"
	if len(conditions) == 0 {
		return !hasAnyLevelTag(tags)
	}

	// All conditions must be met
	for _, cond := range conditions {
		if !contains(tags, cond) {
			return false
		}
	}
	return true
}

// isDependencyInference checks if inference adds dependency tags
func isDependencyInference(tags []string) bool {
	for _, tag := range tags {
		if strings.HasPrefix(tag, "@deps:") || strings.HasPrefix(tag, "@depm:") {
			return true
		}
	}
	return false
}

// GetGlobalInferences returns the standard inference rules
//
// NOTE: L-level tags are NOT inferred. All tests must have explicit L-tags.
// This ensures godog can filter scenarios correctly without needing inference logic.
// Validation enforces that all tests have proper L-tags and verification tags.
func GetGlobalInferences() []Inference {
	return []Inference{
		// Type-based: Go tests require Go toolchain
		{
			TestTypes:   []string{"gotest"},
			IfTags:      []string{},
			ThenAddTags: []string{"@deps:go"},
			Description: "Go tests require Go toolchain",
		},
	}
}

// unique removes duplicate strings from slice
func unique(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// remove removes all occurrences of item from slice
func remove(slice []string, item string) []string {
	result := []string{}
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// filterTags returns tags matching a pattern
func filterTags(tags []string, pattern string) []string {
	result := []string{}
	for _, tag := range tags {
		if strings.Contains(tag, pattern) {
			result = append(result, tag)
		}
	}
	return result
}

// InferSystemDepsFromModuleDeps infers system dependencies based on module dependencies
// For example, if a test has @depm:src-commands and src-commands is a go-* module,
// then @deps:go should be inferred
func InferSystemDepsFromModuleDeps(tests []TestReference, registry *modules.Registry) []TestReference {
	if registry == nil {
		return tests // No registry available, return unchanged
	}

	enriched := make([]TestReference, len(tests))

	for i, test := range tests {
		enriched[i] = test
		enriched[i].Tags = copyTags(test.Tags)

		// Extract module dependencies from tags
		for _, tag := range test.Tags {
			if !strings.HasPrefix(tag, "@depm:") {
				continue
			}

			// Extract module moniker from @depm:<moniker>
			moniker := strings.TrimPrefix(tag, "@depm:")

			// Look up module in registry
			module, exists := registry.Get(moniker)
			if !exists {
				continue // Module not found, skip
			}

			// Check module type and infer system dependencies
			moduleType := module.Type

			// If module type starts with "go-", infer @deps:go
			if strings.HasPrefix(moduleType, "go-") {
				if !contains(enriched[i].Tags, "@deps:go") {
					enriched[i].Tags = append(enriched[i].Tags, "@deps:go")
				}
			}

			// TODO: Add more module type -> system dependency mappings as needed
			// For example:
			// - python-* modules -> @deps:python
			// - docker-* modules -> @deps:docker
		}
	}

	return enriched
}

// osPlatforms is the hardcoded list of OS platform names.
// These are intrinsic to the test system, not configurable.
var osPlatforms = []string{"linux", "macos", "windows"}

// GetOSPlatformTags returns the hardcoded OS platform names.
func GetOSPlatformTags() []string {
	return osPlatforms
}

// GetOSPlatformTagsFull returns OS platform tags with @deps: prefix as a map.
func GetOSPlatformTagsFull() map[string]bool {
	result := make(map[string]bool, len(osPlatforms))
	for _, p := range osPlatforms {
		result["@deps:"+p] = true
	}
	return result
}

// IsOSPlatformDep checks if a dependency name is an OS platform.
func IsOSPlatformDep(dep string) bool {
	for _, p := range osPlatforms {
		if dep == p {
			return true
		}
	}
	return false
}

// InferSystemDepsFromEnv infers system dependencies based on environment tags
// For example, if a test has @env:local01, look up the local01 environment contract
// and add all its system dependencies (@deps:docker, etc.)
func InferSystemDepsFromEnv(tests []TestReference, envConfig *config.EnvironmentsConfig) []TestReference {
	if envConfig == nil {
		return tests // No environment config available, return unchanged
	}

	enriched := make([]TestReference, len(tests))

	for i, test := range tests {
		enriched[i] = test
		enriched[i].Tags = copyTags(test.Tags)

		// Extract environment tags from test tags
		for _, tag := range test.Tags {
			if !strings.HasPrefix(tag, "@env:") {
				continue
			}

			// Extract environment moniker from @env:<moniker>
			moniker := strings.TrimPrefix(tag, "@env:")

			// Look up environment in config
			env, found := envConfig.GetEnvironment(moniker)
			if !found {
				continue // Environment not found, skip
			}

			// Add all system dependencies from the environment
			for _, sysDep := range env.SystemDeps {
				if !contains(enriched[i].Tags, sysDep) {
					enriched[i].Tags = append(enriched[i].Tags, sysDep)
				}
			}
		}
	}

	return enriched
}
