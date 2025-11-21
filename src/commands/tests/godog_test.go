package tests

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	coretesting "github.com/ready-to-release/eac/src/core/testing"
	"github.com/ready-to-release/eac/src/core/repository"
)

func TestFeatures(t *testing.T) {
	// Initialize the original repository root before any tests change directories
	// This ensures we can always find the go.mod file for running commands
	var err error
	originalRepoRoot, err = repository.GetRepositoryRoot("")
	if err != nil {
		t.Fatalf("failed to get repository root: %v", err)
	}

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

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options:             opts,
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
