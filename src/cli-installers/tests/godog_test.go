// Godog test runner for src-cli-installers feature specifications.
//
// This file executes BDD scenarios from specs/src-cli-installers/ using the
// cucumber/godog framework. It supports suite-based filtering via environment
// variables set by the test orchestrator.
//
// # Environment Variables
//
//   - GODOG_SUITE_TAGS: Tag filter expression (e.g., "@L0,@L1,@L2" for commit suite)
//   - GODOG_OUTPUT_DIR: Directory for test reports
//   - GODOG_REPORT_FORMAT: "cucumber" (JSON) or "junit" (XML)
//   - GODOG_REPORT_NAME: Custom report filename
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
)

// scenarioCount tracks scenarios executed for validation.
// Used to detect when godog silently returns zero scenarios due to invalid tag syntax.
var scenarioCount int32

func TestFeatures(t *testing.T) {
	outputDir := os.Getenv("GODOG_OUTPUT_DIR")
	reportFormat := os.Getenv("GODOG_REPORT_FORMAT")
	if reportFormat == "" {
		reportFormat = "cucumber"
	}

	// Get suite tag filter from test orchestrator
	// The orchestrator builds the filter with skip tags integrated into each selector
	tagFilter := os.Getenv("GODOG_SUITE_TAGS")
	if tagFilter == "" {
		log.Fatalf("GODOG_SUITE_TAGS environment variable not set - test orchestrator should set this")
	}

	opts := &godog.Options{
		Format:   "pretty",
		Paths:    []string{"../../../specs/src-cli-installers"},
		TestingT: t,
		Tags:     tagFilter,
	}

	// If output directory is set, add report formatter
	if outputDir != "" {
		var reportPath string
		var formatterName string

		reportName := os.Getenv("GODOG_REPORT_NAME")
		if reportName == "" {
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
			reportPath = filepath.Join(outputDir, reportName)
			formatterName = "cucumber"
		}

		reportFormatted := filepath.ToSlash(reportPath)
		opts.Format = fmt.Sprintf("pretty,%s:%s", formatterName, reportFormatted)

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
