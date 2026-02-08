package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// DiscoverFeatureTags discovers Gherkin feature files and their tags.
// This parses .feature files which is not godog-specific - it's Gherkin-specific
// and works equally for tscucumber.
func DiscoverFeatureTags(specsPath string) ([]TestReference, error) {
	dc, err := NewDiscoveryConfig()
	if err != nil {
		return nil, err
	}
	refs := []TestReference{}

	err = filepath.Walk(specsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == "testdata" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".feature") {
			return nil
		}

		fileRefs, err := parseFeatureFile(path, dc)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		refs = append(refs, fileRefs...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return refs, nil
}

// parseFeatureFile parses a Gherkin feature file and extracts scenarios with tags.
func parseFeatureFile(filePath string, dc *DiscoveryConfig) ([]TestReference, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	refs := []TestReference{}

	var featureTags []string
	var scenarioTags []string
	var inFeature bool
	var scenarioName string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect Feature: keyword first to know when we're done collecting feature tags
		if strings.HasPrefix(trimmed, "Feature:") {
			// We've hit the Feature line, no more feature tags after this
			inFeature = true
			continue
		}

		// Extract feature-level tags (before Feature:)
		// Allow multiple lines of tags before Feature:
		if strings.HasPrefix(trimmed, "@") && !inFeature && len(featureTags) >= 0 {
			tags := extractTagsFromLine(trimmed)
			featureTags = append(featureTags, tags...)
			continue
		}

		// Extract scenario-level tags
		if strings.HasPrefix(trimmed, "@") && !strings.HasPrefix(trimmed, "Feature:") {
			tags := extractTagsFromLine(trimmed)
			scenarioTags = append(scenarioTags, tags...)
		}

		// Detect Scenario: or Scenario Outline:
		if strings.HasPrefix(trimmed, "Scenario:") || strings.HasPrefix(trimmed, "Scenario Outline:") {
			scenarioName = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(trimmed, "Scenario Outline:"), "Scenario:"))

			// Combine tags: scenario level tags OVERRIDE feature level tags
			allTags := mergeFeatureAndScenarioTags(featureTags, scenarioTags)

			// Infer internal module dependency from file path
			// Example: specs/clie-cli/verify-configuration/... → @deps:clie-cli
			inferredDepTag := inferInternalDependencyFromPath(filePath, dc)
			if inferredDepTag != "" {
				allTags = append(allTags, inferredDepTag)
			}

			// Normalize tags to @deps: format
			normalizedTags := normalizeTags(allTags)

			test := TestReference{
				FilePath: filePath,
				Type:     "godog",
				TestName: scenarioName,
				Tags:     normalizedTags,
			}

			// Set execution control fields
			test.IsIgnored, test.SkipReason = extractSkipReason(test.Tags)
			test.IsManual = slices.Contains(test.Tags, "@Manual")
			test.IsSequential = slices.Contains(test.Tags, "@sequential")

			// Extract dependencies
			test.SystemDependencies = extractSystemDependencies(test.Tags)
			test.ModuleDependencies = extractModuleDependencies(test.Tags)

			// Extract risk control references
			test.RiskControls = extractRiskControlTags(test.Tags)

			// Set GxP regulatory fields
			test.IsGxP = slices.Contains(test.Tags, "@gxp")
			test.IsCriticalAspect = slices.Contains(test.Tags, "@gmp-critical-aspect")

			refs = append(refs, test)

			// Reset scenario tags for next scenario
			scenarioTags = []string{}
		}
	}

	return refs, nil
}

// mergeFeatureAndScenarioTags combines feature and scenario tags with proper override semantics
// Rules:
// - Scenario LEVEL tags (@L0-@L4) OVERRIDE feature level tags
// - Scenario VERIFICATION tags (@ov/@iv/@pv/@piv/@ppv) OVERRIDE feature verification tags
// - All other scenario tags are ADDED to feature tags
// - Non-overridden feature tags are INHERITED
//
// This prevents scenarios from inheriting conflicting tags that would violate
// "exactly one" constraints in our validation rules.
func mergeFeatureAndScenarioTags(featureTags, scenarioTags []string) []string {
	result := []string{}

	// Get level tags from both
	featureLevelTags := []string{}
	scenarioLevelTags := []string{}

	for _, tag := range featureTags {
		if isLevelTag(tag) {
			featureLevelTags = append(featureLevelTags, tag)
		}
	}

	for _, tag := range scenarioTags {
		if isLevelTag(tag) {
			scenarioLevelTags = append(scenarioLevelTags, tag)
		}
	}

	// RULE: If scenario has level tag(s), use ONLY scenario level tags (override)
	// Otherwise, inherit feature level tags
	if len(scenarioLevelTags) > 0 {
		result = append(result, scenarioLevelTags...)
	} else {
		result = append(result, featureLevelTags...)
	}

	// Get verification tags from both
	featureVerificationTags := []string{}
	scenarioVerificationTags := []string{}

	for _, tag := range featureTags {
		if isVerificationTag(tag) {
			featureVerificationTags = append(featureVerificationTags, tag)
		}
	}

	for _, tag := range scenarioTags {
		if isVerificationTag(tag) {
			scenarioVerificationTags = append(scenarioVerificationTags, tag)
		}
	}

	// RULE: If scenario has verification tag(s), use ONLY scenario verification tags (override)
	// Otherwise, inherit feature verification tags
	if len(scenarioVerificationTags) > 0 {
		result = append(result, scenarioVerificationTags...)
	} else {
		result = append(result, featureVerificationTags...)
	}

	// Add all OTHER tags from feature (not level, not verification)
	for _, tag := range featureTags {
		if !isLevelTag(tag) && !isVerificationTag(tag) && !slices.Contains(result, tag) {
			result = append(result, tag)
		}
	}

	// Add all OTHER tags from scenario (not level, not verification)
	for _, tag := range scenarioTags {
		if !isLevelTag(tag) && !isVerificationTag(tag) && !slices.Contains(result, tag) {
			result = append(result, tag)
		}
	}

	return result
}

// isLevelTag checks if a tag is a level tag (@L0-@L4).
func isLevelTag(tag string) bool {
	return tag == "@L0" || tag == "@L1" || tag == "@L2" || tag == "@L3" || tag == "@L4"
}

// isVerificationTag checks if a tag is a verification tag (@ov/@iv/@pv/@piv/@ppv).
func isVerificationTag(tag string) bool {
	return tag == "@ov" || tag == "@iv" || tag == "@pv" || tag == "@piv" || tag == "@ppv"
}

// extractTagsFromLine extracts all tags from a line.
func extractTagsFromLine(line string) []string {
	tags := []string{}
	parts := strings.Fields(line)

	for _, part := range parts {
		if strings.HasPrefix(part, "@") {
			tags = append(tags, part)
		}
	}

	return tags
}

// normalizeTags converts tags to standard format.
func normalizeTags(tags []string) []string {
	normalized := []string{}

	for _, tag := range tags {
		// Map @docker -> @deps:docker
		if tag == "@docker" {
			normalized = append(normalized, "@deps:docker")
		} else {
			normalized = append(normalized, tag)
		}
	}

	return normalized
}

// extractRiskControlTags extracts all @risk-control:* tags.
func extractRiskControlTags(tags []string) []string {
	controls := []string{}
	for _, tag := range tags {
		if strings.HasPrefix(tag, "@risk-control:") {
			controls = append(controls, tag)
		}
	}
	return controls
}

// inferInternalDependencyFromPath infers @depm:<module> from feature file path.
// Example: specs/clie-cli/verify-configuration/specification.feature → @depm:clie-cli.
func inferInternalDependencyFromPath(filePath string, dc *DiscoveryConfig) string {
	// Normalize path separators
	normalized := filepath.ToSlash(filePath)

	// Split by "/"
	parts := strings.Split(normalized, "/")

	// Find specs_root directory in the path (e.g., "specs" from config)
	specsRoot := dc.SpecsRoot
	specsIndex := -1
	for i, part := range parts {
		if part == specsRoot {
			specsIndex = i
			break
		}
	}

	// Expected format: <specs_root>/<module>/...
	// Extract module name (the directory right after specs_root)
	if specsIndex >= 0 && len(parts) > specsIndex+1 {
		moduleName := parts[specsIndex+1]
		return fmt.Sprintf("@depm:%s", moduleName)
	}

	return ""
}

// extractSkipReason extracts skip reason from @skip:<reason> tags
// Returns (isIgnored, reason) where:
// - isIgnored: true if test has any @skip:<reason> tag
// - reason: the reason code (e.g., "wip", "broken"), empty if not skipped.
func extractSkipReason(tags []string) (bool, string) {
	for _, tag := range tags {
		if strings.HasPrefix(tag, "@skip:") {
			reason := strings.TrimPrefix(tag, "@skip:")
			return true, reason
		}
	}
	return false, ""
}

// extractSystemDependencies extracts system dependency names from @deps:<name> tags
// Example: @deps:docker → "docker".
func extractSystemDependencies(tags []string) []string {
	deps := []string{}
	for _, tag := range tags {
		if strings.HasPrefix(tag, "@deps:") {
			dep := strings.TrimPrefix(tag, "@deps:")
			deps = append(deps, dep)
		}
	}
	return deps
}

// extractModuleDependencies extracts module dependency names from @depm:<module> tags
// Example: @depm:clie-cli → "clie-cli".
func extractModuleDependencies(tags []string) []string {
	deps := []string{}
	for _, tag := range tags {
		if strings.HasPrefix(tag, "@depm:") {
			dep := strings.TrimPrefix(tag, "@depm:")
			deps = append(deps, dep)
		}
	}
	return deps
}
