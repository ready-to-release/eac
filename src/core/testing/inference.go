package testing

import (
	"strings"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/environments"
)

// ApplyInferences applies inference rules to enrich test tags
// NOTE: L-level tags (@L0-@L4) and verification tags (@ov/@iv/@pv/@piv/@ppv) are NEVER inferred.
// Tests MUST have these tags explicitly present for the executor to find them.
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

			// NEVER infer L-level tags or verification tags - they must be explicit
			if isLevelInference(inference) || isVerificationInference(inference) {
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

		// NOTE: Verification tags (@ov etc.) are NO LONGER derived/inferred
		// They MUST be explicitly present in the test
	}

	return enriched
}

// DeriveOperationalVerification is DEPRECATED and no longer used
// Verification tags MUST be explicitly present in tests
// This function is kept for backwards compatibility but does nothing
func DeriveOperationalVerification(tags []string) []string {
	// NO-OP: Verification tags are never derived anymore
	return tags
}

// hasAnyLevelTag checks if tags contain any level tag (@L0-@L4)
func hasAnyLevelTag(tags []string) bool {
	levelTags := []string{"@L0", "@L1", "@L2", "@L3", "@L4"}
	for _, tag := range tags {
		if contains(levelTags, tag) {
			return true
		}
	}
	return false
}

// isLevelInference checks if inference adds level tags
func isLevelInference(inference Inference) bool {
	levelTags := []string{"@L0", "@L1", "@L2", "@L3", "@L4"}
	for _, tag := range inference.ThenAddTags {
		if contains(levelTags, tag) {
			return true
		}
	}
	return false
}

// isVerificationInference checks if inference adds verification tags
func isVerificationInference(inference Inference) bool {
	verificationTags := []string{"@ov", "@iv", "@pv", "@piv", "@ppv"}
	for _, tag := range inference.ThenAddTags {
		if contains(verificationTags, tag) {
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

// OsPlatformTags lists OS-specific dependency tag values (without @deps: prefix)
var OsPlatformTags = []string{"linux", "macos", "windows"}

// OsPlatformTagsFull lists OS-specific dependency tags (with @deps: prefix) for filtering
var OsPlatformTagsFull = map[string]bool{
	"@deps:linux":       true,
	"@deps:macos":       true,
	"@deps:windows":     true,
	"@deps:os-agnostic": true,
}

// InferOSPlatform adds @deps:os-agnostic to tests that don't have any OS-specific deps
// This runs after other inference phases to ensure all explicit OS deps are already set
func InferOSPlatform(tests []TestReference) []TestReference {
	enriched := make([]TestReference, len(tests))

	for i, test := range tests {
		enriched[i] = test
		enriched[i].Tags = copyTags(test.Tags)

		// Check if test has any OS-specific dependency
		hasOSDep := false
		for _, dep := range test.SystemDependencies {
			for _, osDep := range OsPlatformTags {
				if dep == osDep {
					hasOSDep = true
					break
				}
			}
			if hasOSDep {
				break
			}
		}

		// If no OS-specific dep, add os-agnostic
		if !hasOSDep {
			if !contains(enriched[i].Tags, "@deps:os-agnostic") {
				enriched[i].Tags = append(enriched[i].Tags, "@deps:os-agnostic")
			}
			// Also add to SystemDependencies for consistency
			if !contains(enriched[i].SystemDependencies, "os-agnostic") {
				enriched[i].SystemDependencies = append(enriched[i].SystemDependencies, "os-agnostic")
			}
		}
	}

	return enriched
}

// InferSystemDepsFromEnv infers system dependencies based on environment tags
// For example, if a test has @env:local01, look up the local01 environment contract
// and add all its system dependencies (@deps:docker, etc.)
func InferSystemDepsFromEnv(tests []TestReference, envContract *environments.EnvironmentContract) []TestReference {
	if envContract == nil {
		return tests // No environment contract available, return unchanged
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

			// Look up environment in contract
			env, err := envContract.GetEnvironment(moniker)
			if err != nil {
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
