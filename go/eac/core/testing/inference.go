package testing

import (
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

// ApplyInferences applies inference rules to enrich test tags
// L-level tags (@L1) are inferred for gotest since Go files cannot have explicit tags.
// Godog scenarios MUST have explicit L-tags (no inference for godog).
// Verification tags (@ov) ARE inferred: if no verification tag is present, @ov is added.
// Tracks which tags were inferred in TestReference.InferredTags
// Captures SourceTags (pre-inference tags) if not already set.
func ApplyInferences(tests []TestReference, inferences []Inference) []TestReference {
	enriched := make([]TestReference, len(tests))

	for i, test := range tests {
		enriched[i] = test
		enriched[i].Tags = copyTags(test.Tags)
		enriched[i].InferredTags = copyTags(test.InferredTags) // Preserve existing

		// Capture source tags before inference (if not already set)
		if len(enriched[i].SourceTags) == 0 {
			enriched[i].SourceTags = copyTags(test.Tags)
		}

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
				// Add inferred tags and track them
				for _, tag := range inference.ThenAddTags {
					if !contains(enriched[i].Tags, tag) {
						enriched[i].Tags = append(enriched[i].Tags, tag)
						enriched[i].InferredTags = append(enriched[i].InferredTags, tag)
					}
				}
			}
		}

		// Derive @ov if no other verification tag is present
		hadOV := contains(enriched[i].Tags, "@ov")
		enriched[i].Tags = DeriveOperationalVerification(enriched[i].Tags)
		// Track @ov as inferred if it was added
		if !hadOV && contains(enriched[i].Tags, "@ov") {
			enriched[i].InferredTags = append(enriched[i].InferredTags, "@ov")
		}
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
// Returns true if config unavailable (fail-closed: assume level tags may be present).
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
// Returns true if config unavailable (fail-closed: skip inference if we can't verify).
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

// matchesConditions checks if tags match inference conditions.
func matchesConditions(tags, conditions, thenAddTags []string) bool {
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

// isDependencyInference checks if inference adds dependency tags.
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
// L-level tags are inferred for gotest (Go unit tests) since Go source files
// cannot have explicit tag annotations like feature files can.
// Godog scenarios MUST have explicit L-tags (no inference for godog).
// Verification tags (@ov) are inferred if no verification tag is present.
func GetGlobalInferences() []Inference {
	return []Inference{
		// Type-based: Go tests require Go toolchain
		{
			TestTypes:   []string{"gotest"},
			IfTags:      []string{},
			ThenAddTags: []string{"@deps:go"},
			Description: "Go tests require Go toolchain",
		},
		// Go unit tests default to @L1 (unit test level)
		// This is necessary because Go source files cannot have explicit tag annotations
		// Unlike .feature files where you can add @L0/@L1/@L2 tags directly
		{
			TestTypes:   []string{"gotest"},
			IfTags:      []string{},
			ThenAddTags: []string{"@L1"},
			Description: "Go unit tests default to L1 (unit test level)",
		},
	}
}

// InferSystemDepsFromModuleDeps infers system dependencies based on module dependencies
// For example, if a test has @depm:eac-commands and eac-commands is a go-* module,
// then @deps:go should be inferred
// Tracks inferred deps in TestReference.InferredDeps.
func InferSystemDepsFromModuleDeps(tests []TestReference, registry *modules.Registry) []TestReference {
	if registry == nil {
		return tests // No registry available, return unchanged
	}

	enriched := make([]TestReference, len(tests))

	for i, test := range tests {
		enriched[i] = test
		enriched[i].Tags = copyTags(test.Tags)
		enriched[i].InferredDeps = copyTags(test.InferredDeps) // Preserve existing

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
				depTag := "@deps:go"
				if !contains(enriched[i].Tags, depTag) {
					enriched[i].Tags = append(enriched[i].Tags, depTag)
					enriched[i].InferredDeps = append(enriched[i].InferredDeps, depTag)
				}
			}

			// npm-* modules -> @deps:npm
			if strings.HasPrefix(moduleType, "npm-") {
				depTag := "@deps:npm"
				if !contains(enriched[i].Tags, depTag) {
					enriched[i].Tags = append(enriched[i].Tags, depTag)
					enriched[i].InferredDeps = append(enriched[i].InferredDeps, depTag)
				}
			}
		}
	}

	return enriched
}

// GetOSPlatformTags returns OS platform monikers from system dependencies config.
// Falls back to built-in defaults if config cannot be loaded.
func GetOSPlatformTags() []string {
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil || cfg.SystemDependencies == nil {
		// Fallback to built-in defaults
		return []string{"linux", "macos", "windows", "darwin"}
	}

	var platforms []string
	for _, dep := range cfg.SystemDependencies.Dependencies {
		if dep.Verify.IsOSPlatformBased() {
			platforms = append(platforms, dep.Moniker)
		}
	}
	return platforms
}

// GetOSPlatformTagsFull returns OS platform tags with @deps: prefix as a map.
func GetOSPlatformTagsFull() map[string]bool {
	platforms := GetOSPlatformTags()
	result := make(map[string]bool, len(platforms))
	for _, p := range platforms {
		result["@deps:"+p] = true
	}
	return result
}

// IsOSPlatformDep checks if a dependency name is an OS platform.
func IsOSPlatformDep(dep string) bool {
	platforms := GetOSPlatformTags()
	for _, p := range platforms {
		if dep == p {
			return true
		}
	}
	return false
}

// FilterByCurrentOS filters tests to only those compatible with the current OS.
// Tests with OS-specific deps (e.g., @deps:linux) only run on that OS.
// Tests without any OS-specific deps are OS-agnostic and run everywhere.
// Returns (compatible tests, filtered count).
func FilterByCurrentOS(tests []TestReference, currentOS string) ([]TestReference, int) {
	compatible := []TestReference{}

	for _, test := range tests {
		hasOSDep := false
		matchesCurrentOS := false

		for _, dep := range test.SystemDependencies {
			if IsOSPlatformDep(dep) {
				hasOSDep = true
				if dep == currentOS {
					matchesCurrentOS = true
				}
			}
		}

		// Include if no OS deps (agnostic) or matches current OS
		if !hasOSDep || matchesCurrentOS {
			compatible = append(compatible, test)
		}
	}

	filteredCount := len(tests) - len(compatible)
	return compatible, filteredCount
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
