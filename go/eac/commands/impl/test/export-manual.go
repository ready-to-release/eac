// Command: test export-manual
// Short: Export manual test scenarios for human execution
// Args:
// Long: Export manual test scenarios (tagged with @Manual) from Gherkin specifications
// Long: for human execution and evidence collection.
// Long:
// Long: This command scans module specifications, extracts scenarios tagged with @Manual,
// Long: generates stable scenario IDs, and exports them in JSON, CSV, or Markdown format.
// Long:
// Long: The exported file includes scenario metadata (name, tags, steps), feature context,
// Long: and release information for traceability.
// Long:
// Long: Expected Output:
// Long:   - manual-test-scenarios.{json,csv,md} file created
// Long:   - Scenarios validated against manual-test-export.schema.json
// Long:   - Exit code 0 on success, non-zero on error
// Long:
// Long: Example:
// Long:   test export-manual --module eac-commands --release v1.2.0 --format json
// Long:   test export-manual --module eac-commands --release v1.2.0 --format csv
// Long:   test export-manual --module eac-commands --release v1.2.0 --format markdown
// Flag.module: type=string, usage=Module moniker to export manual tests from (required)
// Flag.release: type=string, usage=Release version being tested (required)
// Flag.format: type=string, usage=Export format: json, csv, markdown (default: json)
package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/fileutil"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/specs/export/formats"
	"github.com/ready-to-release/eac/go/eac/core/specs/gherkin"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func init() {
	registry.Register(ExportManual)
}

// Type aliases for backward compatibility
type ManualTestExport = formats.ManualTestExport
type ExportMetadata = formats.ExportMetadata
type ExportedScenario = formats.ExportedScenario

// ExportManual exports manual test scenarios for human execution.
func ExportManual() int {
	// Parse command line arguments
	args := os.Args[2:] // Skip "test" and "export-manual"

	var moduleFlag, releaseFlag, formatFlag string
	formatFlag = "json" // default

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--module":
			if i+1 < len(args) {
				moduleFlag = args[i+1]
				i++
			}
		case "--release":
			if i+1 < len(args) {
				releaseFlag = args[i+1]
				i++
			}
		case "--format":
			if i+1 < len(args) {
				formatFlag = args[i+1]
				i++
			}
		}
	}

	// Validate required flags
	if moduleFlag == "" {
		log.Errorf("--module flag is required")
		return 1
	}
	if releaseFlag == "" {
		log.Errorf("--release flag is required")
		return 1
	}

	// Load repository config
	workspaceRoot, err := os.Getwd()
	if err != nil {
		log.Errorf("getting working directory: %v", err)
		return 1
	}

	repoCfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot})
	if err != nil {
		log.Errorf("loading repository config: %v", err)
		return 1
	}

	// Validate module exists
	_, moduleExists := repoCfg.Repository.GetModule(moduleFlag)
	if !moduleExists {
		log.Errorf("unknown module: %s", moduleFlag)
		return 1
	}

	// Get git commit SHA
	gitCommit := getGitCommitSHA()

	// Find feature files for module
	specsDir := filepath.Join("specs", moduleFlag)
	featureFiles, err := findFeatureFiles(specsDir)
	if err != nil {
		log.Errorf("finding feature files: %v", err)
		return 1
	}

	// Parse scenarios from all feature files using new parser
	var allScenarios []gherkin.ScenarioDetail
	for _, featureFile := range featureFiles {
		scenarios, err := gherkin.ParseFile(featureFile)
		if err != nil {
			log.Warnf("parsing %s: %v (skipping)", featureFile, err)
			continue
		}
		allScenarios = append(allScenarios, scenarios...)
	}

	// Filter for @Manual scenarios
	manualScenarios := gherkin.FilterScenariosByTag(allScenarios, "@Manual")

	if len(manualScenarios) == 0 {
		log.Errorf("no @Manual scenarios found for module %s", moduleFlag)
		return 1
	}

	// Detect and report scenario ID collisions
	collisions := gherkin.DetectIDCollisions(manualScenarios)
	if len(collisions) > 0 {
		log.Errorf("Found %d scenario ID collision(s):", len(collisions))
		for id, scenarios := range collisions {
			log.Errorf("  ID '%s' is shared by %d scenarios:", id, len(scenarios))
			for _, s := range scenarios {
				log.Errorf("    - %s:%d: %s", s.FilePath, s.LineNumber, s.Name)
			}
		}
		log.Errorf("Scenario IDs must be unique. Please rename scenarios or features to avoid collisions.")
		return 1
	}

	// Build export structure
	export := &ManualTestExport{
		ExportMetadata: ExportMetadata{
			ExportTime:     time.Now().UTC().Format(time.RFC3339),
			Module:         moduleFlag,
			ReleaseVersion: releaseFlag,
			GitCommit:      gitCommit,
			SchemaVersion:  "1.0",
		},
		Scenarios: make([]ExportedScenario, 0, len(manualScenarios)),
	}

	for _, scenario := range manualScenarios {
		export.Scenarios = append(export.Scenarios, ExportedScenario{
			ScenarioID:   gherkin.GenerateScenarioID(scenario),
			FeatureName:  scenario.FeatureName,
			ScenarioName: scenario.Name,
			Tags:         scenario.Tags, // Tags already combined by parser
			Steps:        scenario.Steps,
			Description:  scenario.Description,
			FilePath:     scenario.FilePath,
		})
	}

	// Get formatter
	formatter, err := formats.GetFormatter(formatFlag)
	if err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Write export file atomically
	outputFile := fmt.Sprintf("manual-test-scenarios.%s", formatter.FileExtension())
	var buf bytes.Buffer
	if err := formatter.Write(export, &buf); err != nil {
		log.Errorf("formatting export: %v", err)
		return 1
	}
	if err := fileutil.AtomicWrite(outputFile, buf.Bytes(), 0644); err != nil {
		log.Errorf("writing output file: %v", err)
		return 1
	}

	// Validate against schema (JSON only)
	if formatFlag == "json" {
		schemaPath := filepath.Join(workspaceRoot, "contracts/eac-core/0.1.0/manual-test-export.schema.json")
		if err := validateAgainstSchema(outputFile, schemaPath); err != nil {
			log.Errorf("schema validation failed: %v", err)
			return 1
		}
	}

	log.Infof("Exported %d manual scenarios to %s", len(manualScenarios), outputFile)
	return 0
}

// getGitCommitSHA returns the current git commit SHA (40 chars).
func getGitCommitSHA() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		log.Warnf("getting git commit SHA: %v (using placeholder)", err)
		return "0000000000000000000000000000000000000000"
	}
	return strings.TrimSpace(string(output))
}

// findFeatureFiles finds all .feature files in a directory.
func findFeatureFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".feature") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// validateAgainstSchema validates a JSON file against a schema.
func validateAgainstSchema(dataFile, schemaFile string) error {
	// Load and compile schema
	schemaData, err := os.ReadFile(schemaFile)
	if err != nil {
		return fmt.Errorf("reading schema file: %w", err)
	}

	var schemaDoc any
	if err := json.Unmarshal(schemaData, &schemaDoc); err != nil {
		return fmt.Errorf("parsing schema: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	schemaURL := "schema.json"
	if err := compiler.AddResource(schemaURL, schemaDoc); err != nil {
		return fmt.Errorf("adding schema resource: %w", err)
	}

	schema, err := compiler.Compile(schemaURL)
	if err != nil {
		return fmt.Errorf("compiling schema: %w", err)
	}

	// Load and parse data file
	data, err := os.ReadFile(dataFile)
	if err != nil {
		return fmt.Errorf("reading data file: %w", err)
	}

	var dataDoc any
	if err := json.Unmarshal(data, &dataDoc); err != nil {
		return fmt.Errorf("parsing data: %w", err)
	}

	// Validate
	if err := schema.Validate(dataDoc); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}
