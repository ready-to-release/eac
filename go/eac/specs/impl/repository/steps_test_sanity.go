package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/core/testing"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
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
	sc.Step(`^I scan templates/ for \*\.feature files$`, func() error {
		return tsCtx.scanForFeatureFiles("templates")
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
		return tsCtx.assertDiscoveredFileCountMatches("godog")
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
	sc.Step(`^none of those files appear in test discovery$`, func() error {
		return tsCtx.assertNoneAppearInDiscovery()
	})
	sc.Step(`^none of those files appear as gotest type in discovery$`, func() error {
		return tsCtx.assertNoneAppearAsGotest()
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

	if excludeGodog {
		filtered := []string{}
		for _, f := range matches {
			if !strings.HasSuffix(f, "godog_test.go") {
				filtered = append(filtered, f)
			}
		}
		matches = filtered
	}

	c.rawScanFiles = matches
	c.rawScanCount = len(matches)
	return nil
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
	pattern := filepath.Join(root, "typescript", "**", "*.test.ts")
	pattern = filepath.ToSlash(pattern)

	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return err
	}

	c.rawScanFiles = matches
	c.rawScanCount = len(matches)
	return nil
}

func (c *testSanityContext) runTestDiscovery() error {
	root := c.repoRoot

	tests, err := testing.DiscoverAllTests(root)
	if err != nil {
		return err
	}

	c.discoveredTests = tests

	// Count by type
	c.discoveredCounts = make(map[string]int)
	for _, t := range tests {
		c.discoveredCounts[t.Type]++
	}

	return nil
}

func (c *testSanityContext) assertDiscoveredFileCountMatches(testType string) error {
	// Get unique files from discovered tests of this type (normalize paths)
	fileSet := make(map[string]bool)
	for _, t := range c.discoveredTests {
		if t.Type == testType {
			normalized := filepath.ToSlash(t.FilePath)
			fileSet[normalized] = true
		}
	}
	discoveredFileCount := len(fileSet)

	if discoveredFileCount != c.rawScanCount {
		// Find missing files for better error message
		rawSet := make(map[string]bool)
		for _, f := range c.rawScanFiles {
			rawSet[filepath.ToSlash(f)] = true
		}

		missing := []string{}
		for f := range rawSet {
			found := false
			for df := range fileSet {
				if strings.HasSuffix(df, f) || strings.HasSuffix(f, filepath.Base(df)) {
					found = true
					break
				}
			}
			if !found {
				missing = append(missing, f)
			}
		}

		return fmt.Errorf("%s file count mismatch: raw scan found %d files, discovery found %d files. Missing: %v",
			testType, c.rawScanCount, discoveredFileCount, missing)
	}
	return nil
}

func (c *testSanityContext) assertCountGreaterThanZero() error {
	if c.rawScanCount == 0 {
		return fmt.Errorf("expected count > 0, got 0")
	}
	return nil
}

func (c *testSanityContext) assertNoneAppearInDiscovery() error {
	// Build set of discovered file paths
	discoveredFiles := make(map[string]bool)
	for _, t := range c.discoveredTests {
		discoveredFiles[filepath.ToSlash(t.FilePath)] = true
	}

	// Check none of raw scan files appear
	for _, f := range c.rawScanFiles {
		normalized := filepath.ToSlash(f)
		if discoveredFiles[normalized] {
			return fmt.Errorf("file %s should NOT be discovered but was found in discovery", f)
		}
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

func (c *testSanityContext) assertUniqueMoniker() error {
	seen := make(map[string]bool)
	for _, t := range c.discoveredTests {
		// Use file + test name as unique key (same test name in different files is OK)
		normalized := filepath.ToSlash(t.FilePath)
		moniker := fmt.Sprintf("%s:%s:%s", t.Type, normalized, t.TestName)
		if seen[moniker] {
			return fmt.Errorf("duplicate test entry: %s in %s", t.TestName, t.FilePath)
		}
		seen[moniker] = true
	}
	return nil
}
