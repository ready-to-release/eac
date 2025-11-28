package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/core/config"
)

func TestTestingFrameworkFeatures(t *testing.T) {
	outputDir := os.Getenv("GODOG_OUTPUT_DIR")
	reportFormat := os.Getenv("GODOG_REPORT_FORMAT")
	if reportFormat == "" {
		reportFormat = "cucumber"
	}

	consoleFormat := os.Getenv("GODOG_FORMAT")
	if consoleFormat == "" {
		consoleFormat = "pretty"
	}

	// Load config to get skip reasons
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Build tag filter from config skip reasons
	skipFilter := cfg.TestingTags.BuildGodogSkipTagFilter()
	tagFilter := skipFilter + " && ~@pending"

	// Add suite tag filter if provided (e.g., "@L0 || @L1 || @L2" for commit suite)
	suiteTagFilter := os.Getenv("GODOG_SUITE_TAGS")
	if suiteTagFilter != "" {
		tagFilter = tagFilter + " && (" + suiteTagFilter + ")"
	}

	paths := os.Getenv("GODOG_PATHS")
	if paths == "" {
		paths = "../../../../specs/src-core/testing-framework"
	}

	opts := &godog.Options{
		Format:   consoleFormat,
		Paths:    []string{paths},
		TestingT: t,
		Tags:     tagFilter,
		Strict:   true,
	}

	if outputDir != "" {
		var reportName string
		if reportFormat == "junit" {
			reportName = "junit-testing-framework.xml"
		} else {
			reportName = "cucumber-testing-framework.json"
		}

		reportPath := fmt.Sprintf("%s/%s", outputDir, reportName)
		opts.Format = fmt.Sprintf("%s,%s:%s", consoleFormat, reportFormat, reportPath)
	}

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options:             opts,
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(sc *godog.ScenarioContext) {
	InitializeTestingFrameworkSteps(sc)

	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		cleanupTestContext()
		return ctx, nil
	})
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
