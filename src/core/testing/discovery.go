package testing

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// DiscoverGoTestTags discovers Go test functions and their build tags
func DiscoverGoTestTags(pkgPath string) ([]TestReference, error) {
	refs := []TestReference{}

	// Walk the directory
	err := filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip testdata directories (Go convention for test fixtures)
		if info.IsDir() && info.Name() == "testdata" {
			return filepath.SkipDir
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process *_test.go files
		if !strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}

		// Parse the file
		fileRefs, err := parseGoTestFile(path)
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

// parseGoTestFile parses a single Go test file
func parseGoTestFile(filePath string) ([]TestReference, error) {
	fset := token.NewFileSet()

	// Parse with comments to get build tags
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Extract build tags
	tags := extractBuildTags(file)

	// Find all Test* functions
	refs := []TestReference{}
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// Check if function name starts with Test
		if !strings.HasPrefix(funcDecl.Name.Name, "Test") {
			continue
		}

		// Check if it has testing.T parameter
		if !hasTestingParam(funcDecl) {
			continue
		}

		refs = append(refs, TestReference{
			FilePath: filePath,
			Type:     "gotest",
			TestName: funcDecl.Name.Name,
			Tags:     copyTags(tags),
		})
	}

	return refs, nil
}

// extractBuildTags extracts build constraint tags from file comments
func extractBuildTags(file *ast.File) []string {
	tags := []string{}

	// Check all comment groups
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			text := comment.Text

			// Check for //go:build directive
			if strings.HasPrefix(text, "//go:build ") {
				buildExpr := strings.TrimPrefix(text, "//go:build ")
				buildExpr = strings.TrimSpace(buildExpr)

				// Simple parsing: look for L0, L1, L2 tags
				// TODO: Handle complex expressions if needed
				if strings.Contains(buildExpr, "L0") {
					tags = append(tags, "@L0")
				} else if strings.Contains(buildExpr, "L1") {
					tags = append(tags, "@L1")
				} else if strings.Contains(buildExpr, "L2") {
					tags = append(tags, "@L2")
				}
			}

			// Also check old-style // +build
			if strings.HasPrefix(text, "// +build ") {
				buildExpr := strings.TrimPrefix(text, "// +build ")
				buildExpr = strings.TrimSpace(buildExpr)

				if strings.Contains(buildExpr, "L0") && !contains(tags, "@L0") {
					tags = append(tags, "@L0")
				} else if strings.Contains(buildExpr, "L1") && !contains(tags, "@L1") {
					tags = append(tags, "@L1")
				} else if strings.Contains(buildExpr, "L2") && !contains(tags, "@L2") {
					tags = append(tags, "@L2")
				}
			}
		}
	}

	return tags
}

// hasTestingParam checks if function has *testing.T parameter
func hasTestingParam(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) == 0 {
		return false
	}

	// Check first parameter
	param := funcDecl.Type.Params.List[0]

	// Check if it's *testing.T or *testing.B
	starExpr, ok := param.Type.(*ast.StarExpr)
	if !ok {
		return false
	}

	selExpr, ok := starExpr.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selExpr.X.(*ast.Ident)
	if !ok {
		return false
	}

	// Check if it's testing.T or testing.B
	return ident.Name == "testing" && (selExpr.Sel.Name == "T" || selExpr.Sel.Name == "B")
}

// copyTags creates a copy of tags slice
func copyTags(tags []string) []string {
	if len(tags) == 0 {
		return []string{}
	}
	copied := make([]string, len(tags))
	copy(copied, tags)
	return copied
}

// contains checks if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// DiscoverAllTests discovers all tests from a root path
func DiscoverAllTests(rootPath string) ([]TestReference, error) {
	refs := []TestReference{}

	// Discover Go tests from src/
	srcPath := filepath.Join(rootPath, "src")
	if _, err := os.Stat(srcPath); err == nil {
		goRefs, err := DiscoverGoTestTags(srcPath)
		if err != nil {
			return nil, fmt.Errorf("failed to discover Go tests: %w", err)
		}
		refs = append(refs, goRefs...)
	}

	// Discover Godog features from specs/
	specsPath := filepath.Join(rootPath, "specs")
	if _, err := os.Stat(specsPath); err == nil {
		godogRefs, err := DiscoverGodogFeatureTags(specsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to discover Godog features: %w", err)
		}
		refs = append(refs, godogRefs...)
	}

	return refs, nil
}

// DiscoverGodogFeatureTags discovers Godog feature files and their tags
func DiscoverGodogFeatureTags(specsPath string) ([]TestReference, error) {
	refs := []TestReference{}

	// Walk the specs directory
	err := filepath.Walk(specsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip testdata directories (Go convention for test fixtures)
		if info.IsDir() && info.Name() == "testdata" {
			return filepath.SkipDir
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .feature files
		if !strings.HasSuffix(info.Name(), ".feature") {
			return nil
		}

		// Parse the feature file
		fileRefs, err := parseFeatureFile(path)
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

// parseFeatureFile parses a Gherkin feature file and extracts scenarios with tags
func parseFeatureFile(filePath string) ([]TestReference, error) {
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
			// Example: specs/src-cli/verify-configuration/... → @deps:src-cli
			inferredDepTag := inferInternalDependencyFromPath(filePath)
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
			test.IsManual = contains(test.Tags, "@Manual")

			// Extract dependencies
			test.SystemDependencies = extractSystemDependencies(test.Tags)
			test.ModuleDependencies = extractModuleDependencies(test.Tags)

			// Extract risk control references
			test.RiskControls = extractRiskControlTags(test.Tags)

			// Set GxP regulatory fields
			test.IsGxP = contains(test.Tags, "@gxp")
			test.IsCriticalAspect = contains(test.Tags, "@critical-aspect")

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
func mergeFeatureAndScenarioTags(featureTags []string, scenarioTags []string) []string {
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
		if !isLevelTag(tag) && !isVerificationTag(tag) && !contains(result, tag) {
			result = append(result, tag)
		}
	}

	// Add all OTHER tags from scenario (not level, not verification)
	for _, tag := range scenarioTags {
		if !isLevelTag(tag) && !isVerificationTag(tag) && !contains(result, tag) {
			result = append(result, tag)
		}
	}

	return result
}

// isLevelTag checks if a tag is a level tag (@L0-@L4)
func isLevelTag(tag string) bool {
	return tag == "@L0" || tag == "@L1" || tag == "@L2" || tag == "@L3" || tag == "@L4"
}

// isVerificationTag checks if a tag is a verification tag (@ov/@iv/@pv/@piv/@ppv)
func isVerificationTag(tag string) bool {
	return tag == "@ov" || tag == "@iv" || tag == "@pv" || tag == "@piv" || tag == "@ppv"
}

// extractTagsFromLine extracts all tags from a line
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

// normalizeTags converts tags to standard format
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

// extractRiskControlTags extracts all @risk-control:* tags
func extractRiskControlTags(tags []string) []string {
	controls := []string{}
	for _, tag := range tags {
		if strings.HasPrefix(tag, "@risk-control:") {
			controls = append(controls, tag)
		}
	}
	return controls
}

// inferInternalDependencyFromPath infers @depm:<module> from feature file path
// Example: specs/src-cli/verify-configuration/specification.feature → @depm:src-cli
// Example: C:\projects\eac\specs\src-commands\docs\specification.feature → @depm:src-commands
func inferInternalDependencyFromPath(filePath string) string {
	// Normalize path separators
	normalized := filepath.ToSlash(filePath)

	// Split by "/"
	parts := strings.Split(normalized, "/")

	// Find "specs" in the path
	specsIndex := -1
	for i, part := range parts {
		if part == "specs" {
			specsIndex = i
			break
		}
	}

	// Expected format: specs/<module>/...
	// Extract module name (the directory right after "specs")
	if specsIndex >= 0 && len(parts) > specsIndex+1 {
		moduleName := parts[specsIndex+1]
		return fmt.Sprintf("@depm:%s", moduleName)
	}

	return ""
}

// extractSkipReason extracts skip reason from @skip:<reason> tags
// Returns (isIgnored, reason) where:
// - isIgnored: true if test has any @skip:<reason> tag
// - reason: the reason code (e.g., "wip", "broken"), empty if not skipped
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
// Example: @deps:docker → "docker"
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
// Example: @depm:src-cli → "src-cli"
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
