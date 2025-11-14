package testing

import (
	"strings"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// ApplyInferences applies inference rules to enrich test tags
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

		// Derive @ov if applicable
		enriched[i].Tags = DeriveOperationalVerification(enriched[i].Tags)
	}

	return enriched
}

// DeriveOperationalVerification adds @ov tag if no other verification tags present
func DeriveOperationalVerification(tags []string) []string {
	hasIV := contains(tags, "@iv")
	hasPV := contains(tags, "@pv")
	hasPIV := contains(tags, "@piv")
	hasPPV := contains(tags, "@ppv")
	hasOV := contains(tags, "@ov")

	// @ov = no IV/PV/PIV/PPV
	if !hasIV && !hasPV && !hasPIV && !hasPPV && !hasOV {
		return append(tags, "@ov")
	}

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
func GetGlobalInferences() []Inference {
	return []Inference{
		// Type-based: Go tests default to L1
		{
			TestTypes:   []string{"gotest"},
			IfTags:      []string{},
			ThenAddTags: []string{"@L1"},
			Description: "Go tests default to L1",
		},
		// Type-based: Go tests require Go toolchain
		{
			TestTypes:   []string{"gotest"},
			IfTags:      []string{},
			ThenAddTags: []string{"@deps:go"},
			Description: "Go tests require Go toolchain",
		},
		// Type-based: Godog features default to L2
		{
			TestTypes:   []string{"godog"},
			IfTags:      []string{},
			ThenAddTags: []string{"@L2"},
			Description: "Godog features default to L2",
		},
		// Verification-based: IV implies L3
		{
			IfTags:      []string{"@iv"},
			ThenAddTags: []string{"@L3"},
			Description: "Installation verification at L3 (PLTE)",
		},
		// Verification-based: PV implies L3
		{
			IfTags:      []string{"@pv"},
			ThenAddTags: []string{"@L3"},
			Description: "Performance verification at L3 (PLTE)",
		},
		// Production: PIV implies L4
		{
			IfTags:      []string{"@piv"},
			ThenAddTags: []string{"@L4"},
			Description: "Production installation verification at L4",
		},
		// Production: PPV implies L4
		{
			IfTags:      []string{"@ppv"},
			ThenAddTags: []string{"@L4"},
			Description: "Production performance verification at L4",
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
