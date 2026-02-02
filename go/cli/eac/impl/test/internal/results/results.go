// Package results provides test result parsing and aggregation utilities
package results

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/test/internal/ctrf"
	"github.com/ready-to-release/eac/go/cli/eac/impl/test/internal/cucumber"
	"github.com/ready-to-release/eac/go/core/logging"
)

var log = logging.C()

// CucumberTestResult holds parsed test result data from a cucumber.json file.
type CucumberTestResult struct {
	ScenarioName string
	FeaturePath  string
	Status       string
	DurationMs   int64
	Tags         []string
}

// ParseCucumberResults walks the module test directory and parses all cucumber.json files
// to extract actual test results with durations.
// Walks component subdirectories (out/test/<module>/<component>/) to find cucumber files.
func ParseCucumberResults(moduleTestDir string) []CucumberTestResult {
	var results []CucumberTestResult

	//nolint:errcheck // Walk errors are non-fatal; we collect what we can find
	_ = filepath.Walk(moduleTestDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return nil
		}

		// Skip manifest file
		if strings.HasSuffix(path, "test.manifest.json") {
			return nil
		}

		// Only process cucumber JSON files (various naming patterns)
		// Patterns: *.cucumber.json, cucumber.json, cucumber-*.json
		fileName := info.Name()
		if !isCucumberFile(fileName) {
			return nil
		}

		// Parse the cucumber report
		report, parseErr := cucumber.ParseFile(path)
		if parseErr != nil {
			log.Debugf("Failed to parse cucumber file %s: %v", path, parseErr)
			return nil
		}

		// Extract results from each feature and scenario
		for i := range report {
			feature := &report[i]
			featurePath := feature.URI

			for j := range feature.Elements {
				scenario := &feature.Elements[j]
				// Skip background elements
				if scenario.Type == "background" {
					continue
				}

				// Extract tags as strings
				var tags []string
				for _, tag := range scenario.Tags {
					tags = append(tags, tag.Name)
				}

				results = append(results, CucumberTestResult{
					ScenarioName: scenario.Name,
					FeaturePath:  featurePath,
					Status:       scenario.GetStatus(),
					DurationMs:   scenario.GetDurationMs(),
					Tags:         tags,
				})
			}
		}

		return nil
	})

	return results
}

// isCucumberFile checks if a filename matches cucumber JSON file patterns.
// Patterns: *.cucumber.json, cucumber.json, cucumber-*.json
func isCucumberFile(fileName string) bool {
	return strings.HasSuffix(fileName, ".cucumber.json") ||
		fileName == "cucumber.json" ||
		(strings.HasPrefix(fileName, "cucumber-") && strings.HasSuffix(fileName, ".json"))
}

// AggregateCucumberReports collects all cucumber.json files from module test directories
// and aggregates them into a single cucumber.json file at testRunDir.
// Walks component subdirectories (out/test/<module>/<component>/) to find cucumber files.
// Returns the path to the aggregated file, or empty string if no cucumber files found.
func AggregateCucumberReports(testRunDir string) string {
	var allFeatures cucumber.CucumberReport

	// Walk through all module directories in out/test/
	entries, err := os.ReadDir(testRunDir)
	if err != nil {
		log.Debugf("Failed to read test directory: %v", err)
		return ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip hidden directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		moduleDir := filepath.Join(testRunDir, entry.Name())

		// Walk through module directory (component subdirectories) to find cucumber files
		//nolint:errcheck // Walk errors are non-fatal; we collect what we can find
		_ = filepath.Walk(moduleDir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil || info.IsDir() {
				return nil
			}

			// Skip manifest file
			if strings.HasSuffix(path, "test.manifest.json") {
				return nil
			}

			// Match cucumber JSON files (various naming patterns)
			if !isCucumberFile(info.Name()) {
				return nil
			}

			// Parse and aggregate
			report, parseErr := cucumber.ParseFile(path)
			if parseErr != nil {
				log.Debugf("Failed to parse cucumber file %s: %v", path, parseErr)
				return nil
			}

			allFeatures = append(allFeatures, report...)
			return nil
		})
	}

	if len(allFeatures) == 0 {
		return ""
	}

	// Write aggregated report
	aggregatedPath := filepath.Join(testRunDir, "cucumber.json")
	data, err := json.MarshalIndent(allFeatures, "", "  ")
	if err != nil {
		log.Warnf("Failed to marshal aggregated cucumber report: %v", err)
		return ""
	}

	if err := os.WriteFile(aggregatedPath, data, 0o644); err != nil {
		log.Warnf("Failed to write aggregated cucumber report: %v", err)
		return ""
	}

	log.Debugf("Aggregated %d cucumber features to %s", len(allFeatures), aggregatedPath)
	return aggregatedPath
}

// AggregateCTRFReports collects all unit.json (CTRF) files from module test directories
// and aggregates them into a single unit.json file at testRunDir.
// Walks component subdirectories (out/test/<module>/<component>/) to find unit.json files.
// Returns the path to the aggregated file, or empty string if no CTRF files found.
func AggregateCTRFReports(testRunDir string) string {
	aggregatedReport := ctrf.NewEmptyReport("aggregated")

	// Walk through all module directories in out/test/
	entries, err := os.ReadDir(testRunDir)
	if err != nil {
		log.Debugf("Failed to read test directory: %v", err)
		return ""
	}

	foundCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip hidden directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		moduleDir := filepath.Join(testRunDir, entry.Name())

		// Walk through module directory (component subdirectories) to find unit.json files
		//nolint:errcheck // Walk errors are non-fatal; we collect what we can find
		_ = filepath.Walk(moduleDir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil || info.IsDir() {
				return nil
			}

			// Skip manifest file
			if strings.HasSuffix(path, "test.manifest.json") {
				return nil
			}

			// Match unit.json files
			if info.Name() != "unit.json" {
				return nil
			}

			// Parse and aggregate
			report, parseErr := ctrf.ParseFile(path)
			if parseErr != nil {
				log.Debugf("Failed to parse CTRF file %s: %v", path, parseErr)
				return nil
			}

			aggregatedReport.Merge(report)
			foundCount++
			return nil
		})
	}

	if foundCount == 0 {
		return ""
	}

	// Write aggregated report
	aggregatedPath := filepath.Join(testRunDir, "unit.json")
	data, err := aggregatedReport.ToJSON()
	if err != nil {
		log.Warnf("Failed to marshal aggregated CTRF report: %v", err)
		return ""
	}

	if err := os.WriteFile(aggregatedPath, data, 0o644); err != nil {
		log.Warnf("Failed to write aggregated CTRF report: %v", err)
		return ""
	}

	log.Debugf("Aggregated %d CTRF reports (%d tests) to %s", foundCount, aggregatedReport.Results.Summary.Tests, aggregatedPath)
	return aggregatedPath
}
