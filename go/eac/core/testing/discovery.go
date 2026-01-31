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
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/domain/modules"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

// Package-level logger for discovery debugging
var log = logging.C()

// DiscoveryConfig holds configuration for test discovery.
// All values come from repository.yml - no defaults, no fallbacks.
type DiscoveryConfig struct {
	// SpecsRoot is the root directory for specifications (from repository.yml paths.specs_root)
	SpecsRoot string
	// GodogTestFile is the conventional filename for godog test files (from repository.yml conventions.godog_test)
	GodogTestFile string
}

// NewDiscoveryConfig creates a DiscoveryConfig from global configuration.
// Panics if configuration is not loaded - configuration is required, not optional.
func NewDiscoveryConfig() *DiscoveryConfig {
	cfg := config.Global()
	if cfg == nil || cfg.Repository == nil {
		panic("discovery: repository configuration not loaded - ensure config.LoadGlobal() is called first")
	}

	return &DiscoveryConfig{
		SpecsRoot:     cfg.Repository.Paths.SpecsRoot,
		GodogTestFile: cfg.Repository.Conventions.GodogTest,
	}
}

// IsGodogTestFile checks if the given filename matches the configured godog test file name.
func (dc *DiscoveryConfig) IsGodogTestFile(filename string) bool {
	return filename == dc.GodogTestFile
}

// DiscoverGoTestTags discovers Go test functions and their build tags.
// Public API - creates config internally.
func DiscoverGoTestTags(pkgPath string) ([]TestReference, error) {
	return discoverGoTestTagsInPath(pkgPath, NewDiscoveryConfig())
}

// discoverGoTestTagsInPath discovers Go test functions using provided config.
func discoverGoTestTagsInPath(pkgPath string, dc *DiscoveryConfig) ([]TestReference, error) {
	refs := []TestReference{}

	err := filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == "testdata" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}

		if dc.IsGodogTestFile(info.Name()) {
			return nil
		}

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

// parseGoTestFile parses a single Go test file.
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

// extractBuildTags extracts build constraint tags from file comments.
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

// hasTestingParam checks if function has *testing.T parameter.
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

// copyTags creates a copy of tags slice.
func copyTags(tags []string) []string {
	if len(tags) == 0 {
		return []string{}
	}
	copied := make([]string, len(tags))
	copy(copied, tags)
	return copied
}

// contains checks if slice contains string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// DiscoveryOptions configures post-discovery processing for DiscoverAndEnrich.
type DiscoveryOptions struct {
	// Inferences to apply after discovery (nil = skip inference)
	Inferences []Inference

	// ModuleRegistry for module-dependency inference (nil = skip)
	ModuleRegistry *modules.Registry

	// Environments for environment-dependency inference (nil = skip)
	Environments *config.EnvironmentsConfig
}

// DiscoverAndEnrich is the unified discovery entry point.
// It discovers all tests and applies configured enrichments.
// This consolidates: DiscoverAllTests + ApplyInferences + InferSystemDepsFromModuleDeps + InferSystemDepsFromEnv.
func DiscoverAndEnrich(repoRoot string, opts DiscoveryOptions) ([]TestReference, error) {
	// Phase 1: Discovery
	tests, err := discoverAllTests(repoRoot)
	if err != nil {
		return nil, err
	}

	// Phase 2: Apply inferences (if configured)
	if opts.Inferences != nil {
		tests = ApplyInferences(tests, opts.Inferences)
	}

	// Phase 3: Module-based inference (if registry provided)
	if opts.ModuleRegistry != nil {
		tests = InferSystemDepsFromModuleDeps(tests, opts.ModuleRegistry)
	}

	// Phase 4: Environment-based inference (if environments provided)
	if opts.Environments != nil {
		tests = InferSystemDepsFromEnv(tests, opts.Environments)
	}

	return tests, nil
}

// discoverAllTests is the internal discovery implementation.
// Use DiscoverAndEnrich as the public API.
func discoverAllTests(rootPath string) ([]TestReference, error) {
	dc := NewDiscoveryConfig()
	refs := []TestReference{}

	// Load module registry
	registry, err := modules.LoadFromWorkspace(rootPath)
	if err != nil {
		// If no registry found, return empty (not an error)
		return refs, nil
	}

	// Discover tests for each module based on its contracts
	for _, module := range registry.All() {
		moduleRefs, err := discoverModuleAllTests(rootPath, module, dc)
		if err != nil {
			continue // Skip modules that fail to parse
		}
		refs = append(refs, moduleRefs...)
	}

	return refs, nil
}

// discoverModuleAllTests discovers all tests for a single module.
// It uses the module's contract to find:
// - Go tests from go package test patterns (e.g., "**/*_test.go")
// - Godog specs from specs package (e.g., "specs/{moniker}/**")
// - Other test files from package test patterns.
func discoverModuleAllTests(rootPath string, module *modules.ModuleContract, dc *DiscoveryConfig) ([]TestReference, error) {
	refs := []TestReference{}

	// Log module discovery start with component count
	enabledComps := module.GetEnabledComponents()
	log.Debugf("Discovery: module=%s components=%d (%v)", module.Moniker, len(module.Components), enabledComps)

	// 1. Discover tests from all components with test patterns
	for pkgType, pkg := range module.Components {
		if pkg == nil {
			log.Debugf("Discovery P1: %s:%s skipped (nil entry)", module.Moniker, pkgType)
			continue
		}
		if pkg.Patterns == nil {
			log.Debugf("Discovery P1: %s:%s skipped (nil patterns)", module.Moniker, pkgType)
			continue
		}
		if len(pkg.Patterns.Tests) == 0 {
			log.Debugf("Discovery P1: %s:%s skipped (no test patterns)", module.Moniker, pkgType)
			continue
		}
		log.Debugf("Discovery P1: %s:%s root=%s patterns=%v", module.Moniker, pkgType, pkg.Root, pkg.Patterns.Tests)
		pkgRoot := filepath.Join(rootPath, pkg.Root)

		for _, pattern := range pkg.Patterns.Tests {
			fullPattern := filepath.Join(pkgRoot, pattern)
			fullPattern = filepath.ToSlash(fullPattern)

			matches, err := doublestar.FilepathGlob(fullPattern)
			if err != nil {
				continue
			}

			for _, testFile := range matches {
				fileRefs, err := discoverTestsInFile(testFile, module.Moniker, pkgType, dc)
				if err != nil {
					continue
				}
				refs = append(refs, fileRefs...)
			}
		}
	}

	// 2. Discover Go tests from test-impl package if specified
	testImplPath := module.GetTestImplementationPath()
	if testImplPath != "" {
		log.Debugf("Discovery P2: %s test-impl=%s", module.Moniker, testImplPath)
		fullTestImplPath := filepath.Join(rootPath, testImplPath)
		if _, err := os.Stat(fullTestImplPath); err == nil {
			goRefs, err := discoverGoTestTagsInPath(fullTestImplPath, dc)
			if err == nil {
				log.Debugf("Discovery P2: %s found %d gotest from test-impl", module.Moniker, len(goRefs))
				// Tag tests with module dependency
				for i := range goRefs {
					depmTag := "@depm:" + module.Moniker
					if !contains(goRefs[i].Tags, depmTag) {
						goRefs[i].Tags = append(goRefs[i].Tags, depmTag)
					}
					goRefs[i].ModuleDependencies = append(goRefs[i].ModuleDependencies, module.Moniker)
				}
				refs = append(refs, goRefs...)
			} else {
				log.Debugf("Discovery P2: %s test-impl error: %v", module.Moniker, err)
			}
		} else {
			log.Debugf("Discovery P2: %s test-impl path not found: %s", module.Moniker, fullTestImplPath)
		}
	} else {
		log.Debugf("Discovery P2: %s skipped (no test-impl)", module.Moniker)
	}

	// 3. Discover Gherkin specs from specs package
	featureTestType := getFeatureTestTypeForModule(module)
	specsRoot := module.GetSpecsRoot()

	log.Debugf("Discovery P3: %s specsRoot=%s type=%s", module.Moniker, specsRoot, featureTestType)

	specPattern := filepath.Join(rootPath, specsRoot, "**/*.feature")
	specPattern = filepath.ToSlash(specPattern)

	matches, err := doublestar.FilepathGlob(specPattern)
	if err == nil {
		log.Debugf("Discovery P3: %s found %d .feature files", module.Moniker, len(matches))
		for _, specFile := range matches {
			featureRefs, err := parseFeatureFile(specFile, dc)
			if err != nil {
				log.Debugf("Discovery P3: %s parse error in %s: %v", module.Moniker, specFile, err)
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
	} else {
		log.Debugf("Discovery P3: %s glob error: %v", module.Moniker, err)
	}

	log.Debugf("Discovery: %s total=%d tests discovered", module.Moniker, len(refs))
	return refs, nil
}

// expandModuleVars expands variables in a path pattern.
// Supported: {moniker}, {specs_root}.
// Note: {type} is no longer supported - use package-specific patterns instead.
// Note: {test_impl_root} removed - use explicit test-impl component root instead.
func expandModuleVars(pattern string, module *modules.ModuleContract, rootPath string, dc *DiscoveryConfig) string {
	result := pattern
	result = strings.ReplaceAll(result, "{moniker}", module.Moniker)
	result = strings.ReplaceAll(result, "{specs_root}", dc.SpecsRoot)
	return result
}

// discoverTestsInFile discovers tests in a single file based on its type.
// Returns test references with the module moniker attached.
// packageType is used to determine the test framework (e.g., go, typescript).
func discoverTestsInFile(filePath, moniker, packageType string, dc *DiscoveryConfig) ([]TestReference, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	name := filepath.Base(filePath)

	var refs []TestReference
	var err error

	// Get test framework from package type configuration
	testFramework := getTestFrameworkForPackage(packageType)

	log.Debugf("discoverTestsInFile: %s ext=%s pkg=%s framework=%s", filePath, ext, packageType, testFramework)

	switch {
	case ext == ".ts" && strings.HasSuffix(name, ".test.ts"):
		// TypeScript test file
		log.Debugf("discoverTestsInFile: %s -> mocha (typescript test)", name)
		refs, err = parseNodeTestFile(filePath, testFramework)
	case ext == ".js" && strings.HasSuffix(name, ".test.js"):
		// JavaScript test file
		log.Debugf("discoverTestsInFile: %s -> mocha (javascript test)", name)
		refs, err = parseNodeTestFile(filePath, testFramework)
	case ext == ".feature":
		// Gherkin feature file - already handled by DiscoverGodogFeatureTags
		// Skip to avoid duplicates
		log.Debugf("discoverTestsInFile: %s -> skipped (handled by P3)", name)
		return nil, nil
	case ext == ".go" && strings.HasSuffix(name, "_test.go"):
		// Go test file - parse it directly
		// Skip godog runners as they're not unit tests
		if dc.IsGodogTestFile(name) {
			log.Debugf("discoverTestsInFile: %s -> skipped (godog runner)", name)
			return nil, nil
		}
		log.Debugf("discoverTestsInFile: %s -> gotest", name)
		refs, err = parseGoTestFile(filePath)
	default:
		// Unknown test file type
		log.Debugf("discoverTestsInFile: %s -> skipped (unknown type)", name)
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

// DiscoverGodogFeatureTags discovers Godog feature files and their tags.
func DiscoverGodogFeatureTags(specsPath string) ([]TestReference, error) {
	dc := NewDiscoveryConfig()
	refs := []TestReference{}

	err := filepath.Walk(specsPath, func(path string, info os.FileInfo, err error) error {
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
			// Example: specs/r2r-cli/verify-configuration/... → @deps:r2r-cli
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
			test.IsManual = contains(test.Tags, "@Manual")
			test.IsSequential = contains(test.Tags, "@sequential")

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
// Example: specs/r2r-cli/verify-configuration/specification.feature → @depm:r2r-cli.
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
// Example: @depm:r2r-cli → "r2r-cli".
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

// getTestFrameworkForPackage returns the test framework for a package type.
// Used when inferring test framework from module packages:
// - go → "go" (go test)
// - typescript → "mocha"
// - Returns empty string if not configured - caller must handle.
func getTestFrameworkForPackage(packageType string) string {
	switch packageType {
	case "go":
		return "go"
	case "typescript":
		return "mocha"
	default:
		return ""
	}
}

// getFeatureTestTypeForModule returns the test type for Gherkin feature files.
// Infers BDD framework from module's package types:
// - typescript package → "tscucumber" (TypeScript cucumber-js)
// - go package (or anything else) → "godog" (Go BDD framework).
func getFeatureTestTypeForModule(module *modules.ModuleContract) string {
	// Check if module has typescript package
	hasTS := module.HasComponent("typescript")
	log.Debugf("getFeatureTestTypeForModule: %s hasTypescript=%v", module.Moniker, hasTS)
	if hasTS {
		return "tscucumber"
	}
	// Default to godog for Go and other modules
	return "godog"
}

// parseNodeTestFile extracts describe blocks with tags from TypeScript/JavaScript test file.
// Pattern: describe('@L0 ComponentName', ...) or describe('@L0 @deps:foo ComponentName', ...)
// Tags must be at the start of the describe name, space-separated before the actual name.
// testFramework determines the Type field (e.g., "mocha", "jest").
// Note: Module dependency is attached by the caller (discoverTestsInFile) based on module contract.
func parseNodeTestFile(filePath, testFramework string) ([]TestReference, error) {
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
		ref.IsManual = contains(ref.Tags, "@Manual")
		ref.IsSequential = contains(ref.Tags, "@sequential")
		ref.SystemDependencies = extractSystemDependencies(ref.Tags)
		ref.ModuleDependencies = extractModuleDependencies(ref.Tags)

		refs = append(refs, ref)
	}

	return refs, nil
}

// extractDescribeName extracts the string content from describe('...' or describe("...".
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
// Output: tags=["@L0", "@deps:foo"], name="ComponentName".
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
