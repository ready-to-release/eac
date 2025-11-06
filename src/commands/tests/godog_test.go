package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	outputDir := os.Getenv("GODOG_OUTPUT_DIR")
	reportFormat := os.Getenv("GODOG_REPORT_FORMAT")
	if reportFormat == "" {
		reportFormat = "cucumber" // Default format
	}

	opts := &godog.Options{
		Format:   "pretty",
		Paths:    []string{"../../../specs/src-commands"},
		TestingT: t,
		Tags:     "~@skip && ~@pending", // Skip scenarios tagged with @skip or @pending
	}

	// If output directory is set, add report formatter
	// Format: "formatter1:path1,formatter2:path2"
	// This is supported natively by Godog since v0.12.0
	if outputDir != "" {
		var reportPath string
		var formatterName string

		if reportFormat == "junit" {
			reportPath = filepath.Join(outputDir, "junit.xml")
			formatterName = "junit"
		} else {
			// Default: cucumber
			reportPath = filepath.Join(outputDir, "cucumber.json")
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
