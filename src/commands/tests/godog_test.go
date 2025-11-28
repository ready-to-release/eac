// Godog test runner for src-commands feature specifications.
//
// This file executes BDD scenarios from specs/src-commands/ using the
// cucumber/godog framework. It supports suite-based filtering via environment
// variables set by the test orchestrator.
//
// # Environment Variables
//
//   - GODOG_SUITE_TAGS: Tag filter expression (e.g., "@L0,@L1,@L2" for commit suite)
//   - GODOG_OUTPUT_DIR: Directory for test reports
//   - GODOG_REPORT_FORMAT: "cucumber" (JSON) or "junit" (XML)
//   - GODOG_REPORT_NAME: Custom report filename
//   - GODOG_FORMAT: Console output format (pretty, progress, etc.)
//   - GODOG_PATHS: Override feature file paths
//
// # Tag Filter Syntax (CRITICAL)
//
// Godog's tag parser silently fails with incorrect syntax, returning zero
// scenarios without any error. This can cause CI to falsely pass!
//
// Correct:  @tag1,@tag2 (OR), @tag1 && @tag2 (AND), ~@tag (NOT)
// WRONG:    (@tag1 || @tag2), @tag1 || @tag2, @tag1 or @tag2
//
// See BuildGodogTagFilter in src/core/testing/suite.go for details.
package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/core/config"
	"github.com/ready-to-release/eac/src/core/repository"
)

// scenarioCount tracks scenarios executed for validation.
// Used to detect when godog silently returns zero scenarios due to invalid tag syntax.
var scenarioCount int32

func TestFeatures(t *testing.T) {
	// Cache repository root before tests change working directory.
	// Commands need this to locate go.mod for "go run" execution.
	var err error
	originalRepoRoot, err = repository.GetRepositoryRoot("")
	if err != nil {
		t.Fatalf("failed to get repository root: %v", err)
	}

	outputDir := os.Getenv("GODOG_OUTPUT_DIR")
	reportFormat := os.Getenv("GODOG_REPORT_FORMAT")
	if reportFormat == "" {
		reportFormat = "cucumber"
	}

	// Console output format: pretty (verbose), progress (dots), etc.
	consoleFormat := os.Getenv("GODOG_FORMAT")
	if consoleFormat == "" {
		consoleFormat = "pretty"
	}

	// Load config for skip reasons (@skip:wip, @skip:broken, etc.)
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Build base tag filter: exclude all @skip:<reason> tags and @pending
	// Example: "~@skip:wip && ~@skip:broken && ~@skip:flaky && ... && ~@pending"
	skipFilter := cfg.TestingTags.BuildGodogSkipTagFilter()
	tagFilter := skipFilter + " && ~@pending"

	// Append suite tag filter if provided by test orchestrator.
	// CRITICAL: Do NOT wrap in parentheses - godog's parser breaks silently!
	// Example: "@L0,@L1,@L2" for commit suite, "@iv,@ov,@pv && ~@L0 && ~@L1 && ~@L2" for acceptance
	suiteTagFilter := os.Getenv("GODOG_SUITE_TAGS")
	if suiteTagFilter != "" {
		tagFilter = tagFilter + " && " + suiteTagFilter
	}

	// Allow paths to be customized via environment variable (for individual feature file execution)
	paths := os.Getenv("GODOG_PATHS")
	if paths == "" {
		paths = "../../../specs/src-commands" // Default: all src-commands specs
	}

	opts := &godog.Options{
		Format:   consoleFormat,
		Paths:    []string{paths},
		TestingT: t,
		Tags:     tagFilter, // Skip scenarios tagged with @skip:<reason> (from contract) or @pending
		Strict:   true,      // Fail on undefined or pending steps
	}

	// If output directory is set, add report formatter
	// Format: "formatter1:path1,formatter2:path2"
	// This is supported natively by Godog since v0.12.0
	if outputDir != "" {
		var reportPath string
		var formatterName string

		// Allow custom report name via environment variable (for parallel execution)
		// This prevents multiple test packages from overwriting each other's reports
		reportName := os.Getenv("GODOG_REPORT_NAME")
		if reportName == "" {
			// Default names if not specified
			if reportFormat == "junit" {
				reportName = "junit.xml"
			} else {
				reportName = "cucumber.json"
			}
		}

		if reportFormat == "junit" {
			reportPath = filepath.Join(outputDir, reportName)
			formatterName = "junit"
		} else {
			// Default: cucumber
			reportPath = filepath.Join(outputDir, reportName)
			formatterName = "cucumber"
		}

		// Convert Windows paths to forward slashes for Godog
		reportFormatted := filepath.ToSlash(reportPath)

		// Construct multi-formatter string: console format + report file
		opts.Format = fmt.Sprintf("%s,%s:%s", consoleFormat, formatterName, reportFormatted)

		fmt.Printf("Registering formatters:\n")
		fmt.Printf("  - Pretty (console)\n")
		fmt.Printf("  - %s: %s\n", strings.Title(formatterName), reportFormatted)
	}

	// Reset scenario counter before test run
	atomic.StoreInt32(&scenarioCount, 0)

	// Wrap scenario initializer to count executed scenarios.
	// This is essential for detecting silent tag filter failures.
	wrappedInitializer := func(sc *godog.ScenarioContext) {
		sc.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
			atomic.AddInt32(&scenarioCount, 1)
			return ctx, nil
		})
		InitializeScenario(sc)
	}

	suite := godog.TestSuite{
		ScenarioInitializer: wrappedInitializer,
		Options:             opts,
	}

	exitCode := suite.Run()

	// CRITICAL VALIDATION: Fail if zero scenarios executed.
	//
	// Godog has a dangerous behavior: invalid tag filter syntax causes it to
	// silently return success (exit code 0) with "No scenarios" instead of
	// reporting an error. This can cause CI pipelines to pass when no tests
	// actually ran!
	//
	// Common causes of zero scenarios:
	//   - Parentheses in tag filter: "(@tag1 || @tag2)" breaks the parser
	//   - Using "||" instead of comma for OR
	//   - Using "or"/"and" keywords instead of "," and "&&"
	//   - Tag filter that legitimately matches nothing (check suite config)
	count := atomic.LoadInt32(&scenarioCount)
	if count == 0 {
		t.Fatalf("CRITICAL: No scenarios were executed! This likely indicates "+
			"an invalid tag filter syntax. Godog silently fails with bad syntax.\n"+
			"Tag filter: %q\n"+
			"See src/core/testing/suite.go BuildGodogTagFilter for correct syntax.", tagFilter)
	}

	if exitCode != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
