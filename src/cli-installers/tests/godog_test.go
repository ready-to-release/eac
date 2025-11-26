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
)

func TestFeatures(t *testing.T) {
	outputDir := os.Getenv("GODOG_OUTPUT_DIR")
	reportFormat := os.Getenv("GODOG_REPORT_FORMAT")
	if reportFormat == "" {
		reportFormat = "cucumber"
	}

	// Load tag contract to get skip reasons
	contract, err := coretesting.LoadTagContract()
	if err != nil {
		log.Fatalf("Failed to load tag contract: %v", err)
	}

	// Build tag filter from contract skip reasons
	skipFilter := contract.BuildGodogSkipTagFilter()
	tagFilter := skipFilter + " && ~@pending"

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
