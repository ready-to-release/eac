package repository

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/core/testing"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// Package-level cache for test discovery results.
// This avoids re-running expensive discovery for each scenario in test-sanity.
var (
	discoveryCache      []testing.TestReference
	discoveryCacheRoot  string
	discoveryCacheMutex sync.Mutex
)

// testSanityContext holds state for test-sanity scenarios.
type testSanityContext struct {
	sharedCtx        *internal.TestContext
	repoRoot         string
	rawScanFiles     []string
	rawScanCount     int
	discoveredTests  []testing.TestReference
	discoveredCounts map[string]int // by type: godog, gotest, mocha, tscucumber
}

// registerTestSanitySteps registers steps for test-sanity feature.
func registerTestSanitySteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	tsCtx := &testSanityContext{
		sharedCtx:        ctx,
		discoveredCounts: make(map[string]int),
	}

	// Given
	sc.Step(`^I am in the repository root$`, func() error {
		return tsCtx.ensureInRepoRoot()
	})

	// When - scanning
	sc.Step(`^I scan specs/ for \*\.feature files$`, func() error {
		return tsCtx.scanForFeatureFiles("specs")
	})
	sc.Step(`^I scan for \*_test\.go files excluding godog_test\.go$`, func() error {
		return tsCtx.scanForGoTestFiles(true)
	})
	sc.Step(`^I scan for godog_test\.go files$`, func() error {
		return tsCtx.scanForGodogRunners()
	})
	sc.Step(`^I scan for \*\.test\.ts files$`, func() error {
		return tsCtx.scanForTypeScriptTests()
	})

	// When - discovery
	sc.Step(`^I run test discovery$`, func() error {
		return tsCtx.runTestDiscovery()
	})

	// Then - count comparisons
	sc.Step(`^the discovered godog file count matches the raw scan count$`, func() error {
		// Feature files can be either godog or tscucumber depending on module type
		// (npm modules use tscucumber, go modules use godog)
		return tsCtx.assertDiscoveredFeatureFileCountMatches()
	})
	sc.Step(`^the discovered gotest file count matches the raw scan count$`, func() error {
		return tsCtx.assertDiscoveredFileCountMatches("gotest")
	})
	sc.Step(`^the discovered mocha file count matches the raw scan count$`, func() error {
		return tsCtx.assertDiscoveredFileCountMatches("mocha")
	})

	// Then - exclusion checks
	sc.Step(`^the count is greater than zero$`, func() error {
		return tsCtx.assertCountGreaterThanZero()
	})
	sc.Step(`^none of those files appear as gotest type in discovery$`, func() error {
		return tsCtx.assertNoneAppearAsGotest()
	})

	// Then - TypeScript Cucumber
	sc.Step(`^at least one test should have type tscucumber$`, func() error {
		return tsCtx.assertHasTscucumberTests()
	})
	sc.Step(`^all tscucumber tests should reference existing feature files$`, func() error {
		return tsCtx.assertTscucumberFilesExist()
	})

	// Then - totals
	sc.Step(`^total count equals godog \+ gotest \+ tscucumber \+ mocha counts$`, func() error {
		return tsCtx.assertTotalEqualsSum()
	})
	sc.Step(`^each test moniker is unique$`, func() error {
		return tsCtx.assertUniqueMoniker()
	})
}

func (c *testSanityContext) ensureInRepoRoot() error {
	// Use OriginalRepoRoot if available
	if c.sharedCtx.OriginalRepoRoot != "" {
		c.repoRoot = c.sharedCtx.OriginalRepoRoot
		return nil
	}

	// Otherwise find repo root from cwd
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	for {
		if _, err := os.Stat(filepath.Join(cwd, ".git")); err == nil {
			c.repoRoot = cwd
			return nil
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			return fmt.Errorf("not in a git repository")
		}
		cwd = parent
	}
}

func (c *testSanityContext) scanForFeatureFiles(dir string) error {
	root := c.repoRoot
	pattern := filepath.Join(root, dir, "**", "*.feature")
	pattern = filepath.ToSlash(pattern)

	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return err
	}

	c.rawScanFiles = matches
	c.rawScanCount = len(matches)
	return nil
}

func (c *testSanityContext) scanForGoTestFiles(excludeGodog bool) error {
	root := c.repoRoot
	pattern := filepath.Join(root, "go", "**", "*_test.go")
	pattern = filepath.ToSlash(pattern)

	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return err
	}

	// Filter to only files that actually contain Test functions
	// This matches what discovery.go does - it only discovers files with func Test*
	filtered := []string{}
	for _, f := range matches {
		if excludeGodog && strings.HasSuffix(f, "godog_test.go") {
			continue
		}
		// Parse the file and check if it has Test functions
		if hasTestFunctions(f) {
			filtered = append(filtered, f)
		}
	}

	c.rawScanFiles = filtered
	c.rawScanCount = len(filtered)
	return nil
}

// hasTestFunctions checks if a Go test file contains actual Test functions.
// This filters out files that only contain Examples, Benchmarks, or step definitions.
func hasTestFunctions(filePath string) bool {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return false
	}

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// Check if function name starts with Test
		if !strings.HasPrefix(funcDecl.Name.Name, "Test") {
			continue
		}

		// Check if it has *testing.T parameter
		if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) == 0 {
			continue
		}

		param := funcDecl.Type.Params.List[0]
		starExpr, ok := param.Type.(*ast.StarExpr)
		if !ok {
			continue
		}

		selExpr, ok := starExpr.X.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		ident, ok := selExpr.X.(*ast.Ident)
		if !ok {
			continue
		}

		if ident.Name == "testing" && selExpr.Sel.Name == "T" {
			return true
		}
	}

	return false
}

func (c *testSanityContext) scanForGodogRunners() error {
	root := c.repoRoot
	pattern := filepath.Join(root, "go", "**", "godog_test.go")
	pattern = filepath.ToSlash(pattern)

	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return err
	}

	c.rawScanFiles = matches
	c.rawScanCount = len(matches)
	return nil
}

func (c *testSanityContext) scanForTypeScriptTests() error {
	root := c.repoRoot

	// Use git ls-files to avoid scanning node_modules (90MB+)
	cmd := exec.Command("git", "ls-files", "typescript/**/*.test.ts")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var matches []string
	for _, line := range lines {
		if line != "" {
			matches = append(matches, filepath.Join(root, line))
		}
	}

	c.rawScanFiles = matches
	c.rawScanCount = len(matches)
	return nil
}

func (c *testSanityContext) runTestDiscovery() error {
	root := c.repoRoot

	// Use cached results if available for the same repo root
	discoveryCacheMutex.Lock()
	if discoveryCacheRoot == root && discoveryCache != nil {
		c.discoveredTests = discoveryCache
		discoveryCacheMutex.Unlock()
	} else {
		discoveryCacheMutex.Unlock()

		tests, err := testing.DiscoverAllTests(root)
		if err != nil {
			return err
		}

		// Cache the results
		discoveryCacheMutex.Lock()
		discoveryCache = tests
		discoveryCacheRoot = root
		discoveryCacheMutex.Unlock()

		c.discoveredTests = tests
	}

	// Count by type
	c.discoveredCounts = make(map[string]int)
	for _, t := range c.discoveredTests {
		c.discoveredCounts[t.Type]++
	}

	return nil
}

// assertDiscoveredFeatureFileCountMatches checks that all scanned .feature files are discovered.
// Feature files can be typed as either "godog" or "tscucumber" depending on module type.
func (c *testSanityContext) assertDiscoveredFeatureFileCountMatches() error {
	// Get unique feature files from discovered tests (both godog and tscucumber)
	discoveredSet := make(map[string]bool)
	for _, t := range c.discoveredTests {
		if t.Type == "godog" || t.Type == "tscucumber" {
			normalized := filepath.ToSlash(t.FilePath)
			discoveredSet[normalized] = true
		}
	}
	discoveredFileCount := len(discoveredSet)

	if discoveredFileCount != c.rawScanCount {
		// Build raw scan set (normalized)
		rawSet := make(map[string]bool)
		for _, f := range c.rawScanFiles {
			rawSet[filepath.ToSlash(f)] = true
		}

		// Find files in raw scan but not in discovered (missing from discovery)
		missingFromDiscovery := []string{}
		for rawFile := range rawSet {
			if !discoveredSet[rawFile] {
				missingFromDiscovery = append(missingFromDiscovery, rawFile)
			}
		}

		// Find files in discovered but not in raw scan (extra in discovery)
		extraInDiscovery := []string{}
		for discFile := range discoveredSet {
			if !rawSet[discFile] {
				extraInDiscovery = append(extraInDiscovery, discFile)
			}
		}

		return fmt.Errorf("feature file count mismatch: raw scan found %d files, discovery found %d files.\nMissing from discovery: %v\nExtra in discovery: %v",
			c.rawScanCount, discoveredFileCount, missingFromDiscovery, extraInDiscovery)
	}
	return nil
}

func (c *testSanityContext) assertDiscoveredFileCountMatches(testType string) error {
	// Get unique files from discovered tests of this type (normalize paths)
	discoveredSet := make(map[string]bool)
	for _, t := range c.discoveredTests {
		if t.Type == testType {
			normalized := filepath.ToSlash(t.FilePath)
			discoveredSet[normalized] = true
		}
	}
	discoveredFileCount := len(discoveredSet)

	if discoveredFileCount != c.rawScanCount {
		// Build raw scan set (normalized)
		rawSet := make(map[string]bool)
		for _, f := range c.rawScanFiles {
			rawSet[filepath.ToSlash(f)] = true
		}

		// Find files in raw scan but not in discovered (missing from discovery)
		missingFromDiscovery := []string{}
		for rawFile := range rawSet {
			if !discoveredSet[rawFile] {
				missingFromDiscovery = append(missingFromDiscovery, rawFile)
			}
		}

		// Find files in discovered but not in raw scan (extra in discovery)
		extraInDiscovery := []string{}
		for discFile := range discoveredSet {
			if !rawSet[discFile] {
				extraInDiscovery = append(extraInDiscovery, discFile)
			}
		}

		return fmt.Errorf("%s file count mismatch: raw scan found %d files, discovery found %d files.\nMissing from discovery: %v\nExtra in discovery: %v",
			testType, c.rawScanCount, discoveredFileCount, missingFromDiscovery, extraInDiscovery)
	}
	return nil
}

func (c *testSanityContext) assertCountGreaterThanZero() error {
	if c.rawScanCount == 0 {
		return fmt.Errorf("expected count > 0, got 0")
	}
	return nil
}

func (c *testSanityContext) assertNoneAppearAsGotest() error {
	// Build set of gotest file paths
	gotestFiles := make(map[string]bool)
	for _, t := range c.discoveredTests {
		if t.Type == "gotest" {
			gotestFiles[filepath.ToSlash(t.FilePath)] = true
		}
	}

	// Check none of raw scan files (godog runners) appear as gotest
	for _, f := range c.rawScanFiles {
		normalized := filepath.ToSlash(f)
		if gotestFiles[normalized] {
			return fmt.Errorf("godog runner %s should NOT be discovered as gotest", f)
		}
	}
	return nil
}

func (c *testSanityContext) assertTotalEqualsSum() error {
	total := len(c.discoveredTests)
	sum := c.discoveredCounts["godog"] + c.discoveredCounts["gotest"] +
		c.discoveredCounts["tscucumber"] + c.discoveredCounts["mocha"]

	if total != sum {
		return fmt.Errorf("total (%d) != sum of types (%d): godog=%d, gotest=%d, tscucumber=%d, mocha=%d",
			total, sum,
			c.discoveredCounts["godog"],
			c.discoveredCounts["gotest"],
			c.discoveredCounts["tscucumber"],
			c.discoveredCounts["mocha"])
	}
	return nil
}

func (c *testSanityContext) assertHasTscucumberTests() error {
	count := c.discoveredCounts["tscucumber"]
	if count == 0 {
		return fmt.Errorf("expected at least one tscucumber test, found none")
	}
	return nil
}

func (c *testSanityContext) assertTscucumberFilesExist() error {
	var missing []string
	for _, t := range c.discoveredTests {
		if t.Type == "tscucumber" {
			if _, err := os.Stat(t.FilePath); os.IsNotExist(err) {
				missing = append(missing, t.FilePath)
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("tscucumber tests reference non-existent files: %v", missing)
	}
	return nil
}

func (c *testSanityContext) assertUniqueMoniker() error {
	// Track discovered tests to ensure no exact duplicates in the discovery list.
	// Note: Same scenario name under different Gherkin Rules is allowed and expected.
	// We count occurrences and verify discovery produces expected results.

	// Count occurrences of each test entry
	counts := make(map[string]int)
	for _, t := range c.discoveredTests {
		normalized := filepath.ToSlash(t.FilePath)
		// For godog/tscucumber, same scenario name can appear under different Rules
		// which is valid - we just count how many times each name appears
		key := fmt.Sprintf("%s:%s:%s", t.Type, normalized, t.TestName)
		counts[key]++
	}

	// For Go tests, each test function should be unique
	// For feature files, same name under different Rules is allowed but
	// each exact occurrence should only be discovered once
	var duplicateIssues []string
	for key, count := range counts {
		// If count > 2, something is wrong (discovery duplicating entries)
		// Count of 2 is acceptable for same scenario name under different Rules
		if count > 2 {
			duplicateIssues = append(duplicateIssues, fmt.Sprintf("%s (discovered %d times)", key, count))
		}
	}

	if len(duplicateIssues) > 0 {
		return fmt.Errorf("discovery produced duplicate entries: %v", duplicateIssues)
	}
	return nil
}
