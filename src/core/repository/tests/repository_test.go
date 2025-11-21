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

func TestRepositoryFeatures(t *testing.T) {
	outputDir := os.Getenv("GODOG_OUTPUT_DIR")
	reportFormat := os.Getenv("GODOG_REPORT_FORMAT")
	if reportFormat == "" {
		reportFormat = "cucumber" // Default format
	}

	// Allow format to be customized via environment variable
	// Supported formats: pretty, progress, cucumber, junit, events, undefined
	consoleFormat := os.Getenv("GODOG_FORMAT")
	if consoleFormat == "" {
		consoleFormat = "pretty" // Default: verbose output
	}

	// Load tag contract to get skip reasons
	contract, err := coretesting.LoadTagContract()
	if err != nil {
		log.Fatalf("Failed to load tag contract: %v", err)
	}

	// Build tag filter from contract skip reasons
	skipFilter := contract.BuildGodogSkipTagFilter()
	tagFilter := skipFilter + " && ~@pending"

	// Allow paths to be customized via environment variable (for individual feature file execution)
	paths := os.Getenv("GODOG_PATHS")
	if paths == "" {
		paths = "../../../../specs/repository" // Default: all repository specs
	}

	opts := &godog.Options{
		Format:   consoleFormat,
		Paths:    []string{paths},
		TestingT: t,
		Tags:     tagFilter, // Skip scenarios tagged with @skip:<reason> (from contract) or @pending
		Strict:   true,      // Fail on undefined or pending steps
	}

	// If output directory is set, add report formatter
	if outputDir != "" {
		var reportName string
		if reportFormat == "junit" {
			reportName = "junit-repository.xml"
		} else {
			reportName = "cucumber-repository.json"
		}

		reportPath := fmt.Sprintf("%s/%s", outputDir, reportName)
		opts.Format = fmt.Sprintf("%s,%s:%s", consoleFormat, reportFormat, reportPath)

		fmt.Printf("Registering formatters:\n")
		fmt.Printf("  - Pretty (console)\n")
		fmt.Printf("  - %s: %s\n", reportFormat, reportPath)
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
	// Repository validation steps
	InitializeRepositoryGoModulesTidyScenario(sc)
	InitializeRepositoryNoUnorderedFilesScenario(sc)
	InitializeRepositoryOneModulePerFileScenario(sc)
	InitializeModuleHierarchyScenario(sc)

	// Cleanup after each scenario
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		resetGoModuleTidyContext()
		resetNoUnorderedFilesContext()
		resetOneModulePerFileContext()
		return ctx, nil
	})
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
