// Package manifests provides test manifest generation utilities
package manifests

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/git"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/testing"
)

var log = logging.C()

// PackageResult holds test results for a single package
type PackageResult struct {
	ModuleMoniker string // Module this package belongs to (for aggregation)
	PackageName   string
	LogFilePath   string
	TestsPassed   int
	TestsFailed   int
	TestsSkipped  int
	TestsTotal    int
	PackageFailed bool
	Duration      time.Duration
}

// CucumberTestResult holds parsed test result data from a cucumber.json file.
// This is passed in from the test command which handles cucumber parsing.
type CucumberTestResult struct {
	ScenarioName string
	FeaturePath  string
	Status       string
	DurationMs   int64
	Tags         []string
}

// GenerateTestManifests creates/updates test.manifest.json for each tested module.
// The manifest aggregates results across suite runs and tracks all test artifacts.
// cucumberResultsByModule maps module moniker to pre-parsed cucumber test results.
func GenerateTestManifests(
	results []PackageResult,
	testsByModulePath map[string][]testing.TestReference,
	cucumberResultsByModule map[string][]CucumberTestResult,
	suiteName string,
	duration time.Duration,
	workspaceRoot string,
	repoCfg *config.RepositoryConfig,
	eacCfg *config.EACConfig,
) {
	// Get current git commit
	gitCommit := git.GetCommitSHA(workspaceRoot)

	// Group results by module moniker
	resultsByModule := make(map[string][]PackageResult)
	for _, result := range results {
		moniker := result.ModuleMoniker
		if moniker == "" {
			// Extract from path if not set
			parts := strings.Split(result.PackageName, "/")
			if len(parts) > 0 {
				moniker = parts[0]
			}
		}
		if moniker != "" {
			resultsByModule[moniker] = append(resultsByModule[moniker], result)
		}
	}

	// Group tests by module for test entries
	testsByModule := make(map[string][]testing.TestReference)
	for modulePath, tests := range testsByModulePath {
		// Extract module moniker from path (first component)
		parts := strings.Split(modulePath, "/")
		if len(parts) > 0 {
			moniker := parts[0]
			testsByModule[moniker] = append(testsByModule[moniker], tests...)
		}
	}

	// Generate manifest for each module
	for moniker, moduleResults := range resultsByModule {
		moduleTestDir := repoCfg.TestModuleDirAbs(workspaceRoot, moniker)

		// Load or create manifest (to aggregate with previous suite runs)
		module := eacCfg.Repository.GetByMoniker(moniker)
		moduleType := ""
		if module != nil {
			moduleType = module.Type
		}

		manifest, err := implinternal.LoadOrCreateTestManifest(moduleTestDir, moniker, moduleType, gitCommit)
		if err != nil {
			log.Warnf("Failed to load test manifest for %s: %v", moniker, err)
			manifest = implinternal.NewTestManifest(moniker, moduleType, gitCommit)
		}

		// For composite suites (e.g., "unit+integration"), record each constituent suite separately
		// based on test L-levels. For single suites, record under the config suite name.
		constituentSuites := GetSuitesIncluded(suiteName)
		if constituentSuites != nil {
			// Composite suite: calculate per-suite summaries based on test L-levels
			suiteSummaries := make(map[string]*implinternal.TestSummary)
			for _, s := range constituentSuites {
				suiteSummaries[s] = &implinternal.TestSummary{}
			}

			// Get tests for this module to determine L-levels
			moduleTests := testsByModule[moniker]

			for _, result := range moduleResults {
				// Determine effective suite from test L-levels
				var pkgTests []testing.TestReference
				for _, t := range moduleTests {
					if strings.Contains(result.PackageName, filepath.Base(filepath.Dir(t.FilePath))) {
						pkgTests = append(pkgTests, t)
						break // Just need one to determine L-level
					}
				}
				// Use first default suite as fallback
				fallbackSuite := ""
				if defaults := eacCfg.TestSuites.ListDefault(); len(defaults) > 0 {
					fallbackSuite = defaults[0]
				}
				effectiveSuite := GetEffectiveSuiteForTests(pkgTests, fallbackSuite)
				if summary, ok := suiteSummaries[effectiveSuite]; ok {
					summary.Total += result.TestsTotal
					summary.Passed += result.TestsPassed
					summary.Failed += result.TestsFailed
					summary.Skipped += result.TestsSkipped
				}
			}

			// Record each constituent suite (even if empty - marks it as "run")
			for _, s := range constituentSuites {
				manifest.AddSuiteResult(s, implinternal.SuiteResult{
					RunTime:         time.Now(),
					DurationSeconds: duration.Seconds() / float64(len(constituentSuites)), // Approximate
					Tests:           *suiteSummaries[s],
				})
				manifest.ClearSuiteTests(s)
			}
		} else {
			// Non-composite suite: record under the config suite name
			suiteSummary := implinternal.TestSummary{}
			for _, result := range moduleResults {
				suiteSummary.Total += result.TestsTotal
				suiteSummary.Passed += result.TestsPassed
				suiteSummary.Failed += result.TestsFailed
				suiteSummary.Skipped += result.TestsSkipped
			}

			manifest.AddSuiteResult(suiteName, implinternal.SuiteResult{
				RunTime:         time.Now(),
				DurationSeconds: duration.Seconds(),
				Tests:           suiteSummary,
			})
			manifest.ClearSuiteTests(suiteName)
		}

		// Get pre-parsed cucumber test results for this module
		cucumberResults := cucumberResultsByModule[moniker]

		// Helper to determine effective suite name for a test
		getEffectiveSuite := func(tags []string) string {
			if constituentSuites != nil {
				// Composite suite: determine from L-level tags using config
				for _, tag := range tags {
					if len(tag) >= 2 && tag[0] == '@' && tag[1] == 'L' {
						if suite := eacCfg.TestSuites.GetSuiteForLTag(tag); suite != "" {
							return suite
						}
					}
				}
				// Default to first default suite for composite suites
				if defaults := eacCfg.TestSuites.ListDefault(); len(defaults) > 0 {
					return defaults[0]
				}
				return ""
			}
			return suiteName // Non-composite: use config suite name
		}

		// Add test entries from cucumber results (godog tests)
		for _, result := range cucumberResults {
			manifest.AddTest(implinternal.TestEntry{
				Name:       result.ScenarioName,
				Package:    result.FeaturePath,
				Type:       "godog",
				Suite:      getEffectiveSuite(result.Tags),
				Status:     result.Status,
				DurationMs: result.DurationMs,
				Tags:       result.Tags,
				FilePath:   result.FeaturePath,
			})
		}

		// Add non-godog tests from test references (go tests without cucumber reports)
		if tests, ok := testsByModule[moniker]; ok {
			// Build a set of scenarios already added from cucumber results
			addedScenarios := make(map[string]bool)
			for _, r := range cucumberResults {
				addedScenarios[r.ScenarioName] = true
			}

			for _, test := range tests {
				// Skip if already added from cucumber results
				if addedScenarios[test.TestName] {
					continue
				}

				// Determine status based on results (simplified - we track at package level)
				status := implinternal.TestStatusPassed
				for _, result := range moduleResults {
					if result.TestsFailed > 0 && strings.Contains(result.PackageName, filepath.Base(filepath.Dir(test.FilePath))) {
						status = implinternal.TestStatusFailed
						break
					}
				}
				if test.IsIgnored {
					status = implinternal.TestStatusSkipped
				}

				manifest.AddTest(implinternal.TestEntry{
					Name:     test.TestName,
					Package:  filepath.Dir(test.FilePath),
					Type:     test.Type,
					Suite:    getEffectiveSuite(test.Tags),
					Status:   status,
					Tags:     test.Tags,
					FilePath: test.FilePath,
				})
			}
		}

		// Collect artifacts (log files, cucumber reports, coverage)
		CollectTestArtifacts(manifest, moduleTestDir)

		// Update timing
		manifest.DurationSeconds = duration.Seconds()

		// Save manifest
		if err := manifest.Save(moduleTestDir); err != nil {
			log.Warnf("Failed to save test manifest for %s: %v", moniker, err)
		} else {
			log.Debugf("Generated test manifest for %s", moniker)
		}
	}
}

// CollectTestArtifacts walks the module test directory and adds artifacts to the manifest
func CollectTestArtifacts(manifest *implinternal.TestManifest, moduleTestDir string) {
	// Walk packages directory to find test artifacts
	packagesDir := filepath.Join(moduleTestDir, "packages")
	if _, err := os.Stat(packagesDir); os.IsNotExist(err) {
		return
	}

	_ = filepath.Walk(packagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(moduleTestDir, path)
		if err != nil {
			return nil
		}
		relPath = filepath.ToSlash(relPath)

		// Determine artifact type and ID
		var artifactType, artifactID string
		switch {
		case strings.HasSuffix(path, "test.log"):
			artifactType = implinternal.TestArtifactTypeLog
			artifactID = "log-" + strings.ReplaceAll(filepath.Dir(relPath), "/", "-")
		case strings.HasSuffix(path, ".cucumber.json") || strings.HasSuffix(path, "cucumber.json"):
			artifactType = implinternal.TestArtifactTypeReport
			artifactID = "report-" + strings.ReplaceAll(filepath.Dir(relPath), "/", "-")
		case strings.HasSuffix(path, "coverage.out"):
			artifactType = implinternal.TestArtifactTypeCoverage
			artifactID = "coverage-" + strings.ReplaceAll(filepath.Dir(relPath), "/", "-")
		default:
			// Skip other files
			return nil
		}

		// Get file size and hash
		size, sha256, _ := implinternal.HashArtifactFile(path)

		manifest.AddArtifact(implinternal.TestArtifactInfo{
			Type:   artifactType,
			ID:     artifactID,
			Name:   info.Name(),
			Path:   relPath,
			Size:   size,
			SHA256: sha256,
		})

		return nil
	})
}

// GetSuitesIncluded parses composite suite syntax.
// Handles composite suite syntax: "unit+integration" -> ["unit", "integration"]
// For single suites, returns nil.
func GetSuitesIncluded(suiteMoniker string) []string {
	// Handle "+" joined composite suites (e.g., "unit+integration")
	if strings.Contains(suiteMoniker, "+") {
		return strings.Split(suiteMoniker, "+")
	}
	return nil
}

// GetEffectiveSuiteForTests determines which suite a set of tests belongs to based on L-level tags.
// Uses config to map L-tags to suite monikers.
// If tests have mixed levels or no L-level tags, returns the provided fallback.
func GetEffectiveSuiteForTests(tests []testing.TestReference, fallback string) string {
	if len(tests) == 0 {
		return fallback
	}

	cfg := config.Global()
	if cfg == nil || cfg.TestSuites == nil {
		return fallback
	}

	// Check L-level tags of first test (tests in same package should have same level)
	for _, tag := range tests[0].Tags {
		if len(tag) >= 2 && tag[0] == '@' && tag[1] == 'L' {
			if suite := cfg.TestSuites.GetSuiteForLTag(tag); suite != "" {
				return suite
			}
		}
	}
	return fallback
}
