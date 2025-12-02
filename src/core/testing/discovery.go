package testing

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/ready-to-release/eac/src/core/config"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
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

		// Skip godog_test.go files - these are godog runners, not Go unit tests
		// The actual tests are in .feature files and discovered separately
		if info.Name() == "godog_test.go" {
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

// DiscoverAllTests discovers all tests from a root path using module contracts.
// All test discovery is driven by module definitions in modules.yml and module-types.yml:
// - files.tests: patterns for test files within module (e.g., "test/**/*.ts", "**/*_test.go")
// - repo.specs: patterns for specification files (e.g., "specs/{moniker}/**")
// - repo.test_impl: path to test implementation (e.g., "{root}/tests")
func DiscoverAllTests(rootPath string) ([]TestReference, error) {
	refs := []TestReference{}

	// Load module registry
	registry, err := modules.LoadFromWorkspaceLatest(rootPath)
	if err != nil {
		// If no registry found, return empty (not an error)
		return refs, nil
	}

	// Discover tests for each module based on its contracts
	for _, module := range registry.All() {
		moduleRefs, err := discoverModuleAllTests(rootPath, module)
		if err != nil {
			continue // Skip modules that fail to parse
		}
		refs = append(refs, moduleRefs...)
	}

	return refs, nil
}

// discoverModuleAllTests discovers all tests for a single module.
// It uses the module's contract to find:
// - Go tests from files.tests patterns (e.g., "**/*_test.go")
// - Godog specs from repo.specs patterns (e.g., "specs/{moniker}/**")
// - Other test files from files.tests patterns (e.g., "test/**/*.ts")
func discoverModuleAllTests(rootPath string, module *modules.ModuleContract) ([]TestReference, error) {
	refs := []TestReference{}
	moduleRoot := filepath.Join(rootPath, module.Files.Root)

	// 1. Discover tests from files.tests patterns
	for _, pattern := range module.Files.Tests {
		fullPattern := filepath.Join(moduleRoot, pattern)
		fullPattern = filepath.ToSlash(fullPattern)

		matches, err := doublestar.FilepathGlob(fullPattern)
		if err != nil {
			continue
		}

		for _, testFile := range matches {
			fileRefs, err := discoverTestsInFile(testFile, module.Moniker, module.Type)
			if err != nil {
				continue
			}
			refs = append(refs, fileRefs...)
		}
	}

	// 2. Discover Go tests from test_impl path if specified
	if module.Files.Repo.TestImpl != "" {
		testImplPath := expandModuleVars(module.Files.Repo.TestImpl, module, rootPath)
		if _, err := os.Stat(testImplPath); err == nil {
			goRefs, err := DiscoverGoTestTags(testImplPath)
			if err == nil {
				// Tag tests with module dependency
				for i := range goRefs {
					depmTag := "@depm:" + module.Moniker
					if !contains(goRefs[i].Tags, depmTag) {
						goRefs[i].Tags = append(goRefs[i].Tags, depmTag)
					}
					goRefs[i].ModuleDependencies = append(goRefs[i].ModuleDependencies, module.Moniker)
				}
				refs = append(refs, goRefs...)
			}
		}
	}

	// 3. Discover Gherkin specs from repo.specs patterns
	// Determine the test type based on module's test framework:
	// - Module metadata test_framework: cucumber → Type: "cucumber"
	// - Module type test_framework: cucumber → Type: "cucumber"
	// - Default → Type: "godog"
	featureTestType := getFeatureTestTypeForModule(module)

	for _, specPattern := range module.Files.Repo.Specs {
		expandedPattern := expandModuleVars(specPattern, module, rootPath)
		fullPattern := filepath.Join(rootPath, expandedPattern)
		fullPattern = filepath.ToSlash(fullPattern)

		matches, err := doublestar.FilepathGlob(fullPattern)
		if err != nil {
			continue
		}

		for _, specFile := range matches {
			if strings.HasSuffix(specFile, ".feature") {
				featureRefs, err := parseFeatureFile(specFile)
				if err != nil {
					continue
				}
				// Tag specs with module dependency and set correct test type
				for i := range featureRefs {
					featureRefs[i].Type = featureTestType
					depmTag := "@depm:" + module.Moniker
					if !contains(featureRefs[i].Tags, depmTag) {
						featureRefs[i].Tags = append(featureRefs[i].Tags, depmTag)
					}
					featureRefs[i].ModuleDependencies = append(featureRefs[i].ModuleDependencies, module.Moniker)
				}
				refs = append(refs, featureRefs...)
			}
		}
	}

	return refs, nil
}

// expandModuleVars expands variables in a path pattern.
// Supported: {moniker}, {root}, {type}
func expandModuleVars(pattern string, module *modules.ModuleContract, rootPath string) string {
	result := pattern
	result = strings.ReplaceAll(result, "{moniker}", module.Moniker)
	result = strings.ReplaceAll(result, "{root}", module.Files.Root)
	result = strings.ReplaceAll(result, "{type}", module.Type)
	return result
}

// discoverTestsInFile discovers tests in a single file based on its type.
// Returns test references with the module moniker attached.
// moduleType is used to determine the test framework from module-types.yml.
func discoverTestsInFile(filePath string, moniker string, moduleType string) ([]TestReference, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	name := filepath.Base(filePath)

	var refs []TestReference
	var err error

	// Get test framework from module type configuration
	testFramework := getTestFrameworkForType(moduleType)

	switch {
	case ext == ".ts" && strings.HasSuffix(name, ".test.ts"):
		// TypeScript test file
		refs, err = parseNodeTestFile(filePath, testFramework)
	case ext == ".js" && strings.HasSuffix(name, ".test.js"):
		// JavaScript test file
		refs, err = parseNodeTestFile(filePath, testFramework)
	case ext == ".feature":
		// Gherkin feature file - already handled by DiscoverGodogFeatureTags
		// Skip to avoid duplicates
		return nil, nil
	case ext == ".go" && strings.HasSuffix(name, "_test.go"):
		// Go test file - already handled by DiscoverGoTestTags
		// Skip to avoid duplicates
		return nil, nil
	default:
		// Unknown test file type
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	// Attach module moniker to all discovered tests
	for i := range refs {
		// Add @depm:<moniker> tag if not already present
		depmTag := "@depm:" + moniker
		if !contains(refs[i].Tags, depmTag) {
			refs[i].Tags = append(refs[i].Tags, depmTag)
		}
		if !contains(refs[i].ModuleDependencies, moniker) {
			refs[i].ModuleDependencies = append(refs[i].ModuleDependencies, moniker)
		}
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
			test.IsCriticalAspect = contains(test.Tags, "@gmp-critical-aspect")

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

// getTestFrameworkForType returns the test framework for a module type.
// Looks up test_framework from module-types.yml configuration.
// Returns "mocha" as default for npm-based modules if not specified.
func getTestFrameworkForType(moduleType string) string {
	cfg := config.Global()
	if cfg != nil && cfg.ModuleTypes != nil {
		framework := cfg.ModuleTypes.GetTestFramework(moduleType)
		if framework != "" {
			return framework
		}
	}

	// Default: return "mocha" for backwards compatibility with npm-based test files
	return "mocha"
}

// getFeatureTestTypeForModule returns the test type for Gherkin feature files.
// Determined by primary build dependency:
// - npm → "tscucumber" (TypeScript cucumber-js)
// - go → "godog" (Go BDD framework)
func getFeatureTestTypeForModule(module *modules.ModuleContract) string {
	cfg := config.Global()
	if cfg != nil && cfg.ModuleTypes != nil {
		primaryDep := cfg.ModuleTypes.GetPrimaryBuildDep(module.Type)
		if primaryDep == "npm" {
			return "tscucumber"
		}
	}
	return "godog"
}

// parseNodeTestFile extracts describe blocks with tags from TypeScript/JavaScript test file.
// Pattern: describe('@L0 ComponentName', ...) or describe('@L0 @deps:foo ComponentName', ...)
// Tags must be at the start of the describe name, space-separated before the actual name.
// testFramework determines the Type field (e.g., "mocha", "jest").
// Note: Module dependency is attached by the caller (discoverTestsInFile) based on module contract.
func parseNodeTestFile(filePath string, testFramework string) ([]TestReference, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	refs := []TestReference{}
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for describe( patterns
		if !strings.Contains(trimmed, "describe(") {
			continue
		}

		// Extract the string inside describe('...' or describe("..."
		descName := extractDescribeName(trimmed)
		if descName == "" {
			continue
		}

		// Parse tags from the describe name
		// Format: "@L0 @deps:foo ComponentName" -> tags=[@L0, @deps:foo], name=ComponentName
		tags, testName := parseDescribeTags(descName)

		if len(tags) == 0 {
			// No tags in this describe, skip it (we only track tagged tests)
			continue
		}

		ref := TestReference{
			FilePath: filePath,
			Type:     testFramework,
			TestName: testName,
			Tags:     tags,
		}

		// Set execution control fields
		ref.IsIgnored, ref.SkipReason = extractSkipReason(ref.Tags)
		ref.SystemDependencies = extractSystemDependencies(ref.Tags)
		ref.ModuleDependencies = extractModuleDependencies(ref.Tags)

		refs = append(refs, ref)
	}

	return refs, nil
}

// extractDescribeName extracts the string content from describe('...' or describe("..."
func extractDescribeName(line string) string {
	// Find describe( and then the opening quote
	idx := strings.Index(line, "describe(")
	if idx == -1 {
		return ""
	}

	rest := line[idx+len("describe("):]
	rest = strings.TrimSpace(rest)

	// Find the quote type (single or double)
	if len(rest) == 0 {
		return ""
	}

	quote := rest[0]
	if quote != '\'' && quote != '"' && quote != '`' {
		return ""
	}

	// Find the closing quote
	rest = rest[1:]
	endIdx := strings.IndexByte(rest, quote)
	if endIdx == -1 {
		return ""
	}

	return rest[:endIdx]
}

// parseDescribeTags parses tags from the start of a describe name.
// Input: "@L0 @deps:foo ComponentName"
// Output: tags=["@L0", "@deps:foo"], name="ComponentName"
func parseDescribeTags(descName string) ([]string, string) {
	parts := strings.Fields(descName)
	if len(parts) == 0 {
		return nil, ""
	}

	var tags []string
	var nameStart int

	for i, part := range parts {
		if strings.HasPrefix(part, "@") {
			tags = append(tags, part)
		} else {
			// First non-tag part starts the name
			nameStart = i
			break
		}
	}

	// If all parts were tags, no name found
	if nameStart == 0 && len(tags) == len(parts) {
		return tags, ""
	}

	// Reconstruct the name from remaining parts
	name := strings.Join(parts[nameStart:], " ")
	return tags, name
}

