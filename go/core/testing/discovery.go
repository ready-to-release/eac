package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/logging"
)

// Package-level logger for discovery debugging
var log = logging.C()

// DiscoveryConfig holds configuration for test discovery.
// All values come from repository.yml - no defaults, no fallbacks.
type DiscoveryConfig struct {
	// SpecsRoot is the root directory for specifications (from repository.yml paths.specs_root)
	SpecsRoot string
}

// NewDiscoveryConfig creates a DiscoveryConfig from global configuration.
// Returns an error if configuration is not loaded.
func NewDiscoveryConfig() (*DiscoveryConfig, error) {
	cfg := config.Global()
	if cfg == nil || cfg.Repository == nil {
		return nil, fmt.Errorf("discovery: repository configuration not loaded - ensure config.LoadGlobal() is called first")
	}

	return &DiscoveryConfig{
		SpecsRoot: cfg.Repository.Paths.SpecsRoot,
	}, nil
}

// IsRunnerFile checks if the given filename is a registered test runner file.
// Queries the adapter registry for runner file conventions.
func (dc *DiscoveryConfig) IsRunnerFile(filename string) bool {
	conventions := getRunnerFileConventions()
	return conventions[filename]
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
	dc, err := NewDiscoveryConfig()
	if err != nil {
		return nil, err
	}
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
					if !slices.Contains(goRefs[i].Tags, depmTag) {
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
				if !slices.Contains(featureRefs[i].Tags, depmTag) {
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
		// Gherkin feature file - already handled by DiscoverFeatureTags
		// Skip to avoid duplicates
		log.Debugf("discoverTestsInFile: %s -> skipped (handled by P3)", name)
		return nil, nil
	case ext == ".go" && strings.HasSuffix(name, "_test.go"):
		// Go test file - parse it directly
		// Skip godog runners as they're not unit tests
		if dc.IsRunnerFile(name) {
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
		if !slices.Contains(refs[i].Tags, depmTag) {
			refs[i].Tags = append(refs[i].Tags, depmTag)
		}
		if !slices.Contains(refs[i].ModuleDependencies, moniker) {
			refs[i].ModuleDependencies = append(refs[i].ModuleDependencies, moniker)
		}
	}

	return refs, nil
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
// Queries the adapter registry to determine which BDD test type owns features
// for this module based on its component types.
func getFeatureTestTypeForModule(module *modules.ModuleContract) string {
	hasTS := module.HasComponent("typescript")
	hasGo := module.HasComponent("go")
	hasPython := module.HasComponent("python")
	hasDotnet := module.HasComponent("dotnet")
	log.Debugf("getFeatureTestTypeForModule: %s hasTypescript=%v hasGo=%v hasPython=%v hasDotnet=%v", module.Moniker, hasTS, hasGo, hasPython, hasDotnet)
	return resolveFeatureTestType(hasTS, hasGo, hasPython, hasDotnet)
}
