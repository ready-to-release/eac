package tests

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/core/config"
)

func TestFeatures(t *testing.T) {
	outputDir := os.Getenv("GODOG_OUTPUT_DIR")
	reportFormat := os.Getenv("GODOG_REPORT_FORMAT")
	if reportFormat == "" {
		reportFormat = "cucumber" // Default format
	}

	// Load config to get skip reasons
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Build tag filter from config skip reasons
	skipFilter := cfg.TestingTags.BuildGodogSkipTagFilter()
	tagFilter := skipFilter + " && ~@pending"

	// Add suite tag filter if provided (e.g., "@L0,@L1,@L2" for commit suite)
	// CRITICAL: Do NOT wrap in parentheses - godog's parser breaks silently!
	suiteTagFilter := os.Getenv("GODOG_SUITE_TAGS")
	if suiteTagFilter != "" {
		tagFilter = tagFilter + " && " + suiteTagFilter
	}

	opts := &godog.Options{
		Format:   "pretty",
		Paths:    []string{"../../../specs/src-cli"},
		TestingT: t,
		Tags:     tagFilter, // Skip scenarios tagged with @skip:<reason> (from contract) or @pending
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

		// Construct multi-formatter string: pretty (console) + report file
		opts.Format = fmt.Sprintf("pretty,%s:%s", formatterName, reportFormatted)

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
