package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/cucumber/godog"
	coretesting "github.com/ready-to-release/eac/src/core/testing"
)

func TestLoggingFeatures(t *testing.T) {
	outputDir := os.Getenv("GODOG_OUTPUT_DIR")
	reportFormat := os.Getenv("GODOG_REPORT_FORMAT")
	if reportFormat == "" {
		reportFormat = "cucumber"
	}

	consoleFormat := os.Getenv("GODOG_FORMAT")
	if consoleFormat == "" {
		consoleFormat = "pretty"
	}

	// Load tag contract to get skip reasons
	contract, err := coretesting.LoadTagContract()
	if err != nil {
		log.Fatalf("Failed to load tag contract: %v", err)
	}

	// Build tag filter from contract skip reasons
	skipFilter := contract.BuildGodogSkipTagFilter()
	tagFilter := skipFilter + " && ~@pending"

	paths := os.Getenv("GODOG_PATHS")
	if paths == "" {
		paths = "../../../../specs/src-core/logging"
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
			reportName = "junit-logging.xml"
		} else {
			reportName = "cucumber-logging.json"
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
	InitializeLoggingSteps(sc)

	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		resetLoggingContext()
		return ctx, nil
	})
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
