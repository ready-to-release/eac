// Command: test import-manual
// Short: Import manual test results for a module release
// Args:
// Long: Import manual test results from a JSON file and validate against schemas,
// Long: exported scenarios, and repository configuration.
// Long:
// Long: This command validates manual test results, checks for conflicts, and stores
// Long: them in the repository for later merging into test manifests.
// Long:
// Long: Validation includes:
// Long: - JSON schema compliance
// Long: - Release version matching
// Long: - Module validation
// Long: - Scenario ID cross-validation against exports
// Long: - Conflict detection (existing results)
// Long:
// Long: Expected Output:
// Long:   - Results stored in test-results/<module>/<version>/manual-results.json
// Long:   - Exit code 0 on success, non-zero on error
// Long:
// Long: Example:
// Long:   test import-manual --input results.json --release v1.2.0
// Long:   test import-manual --input results.json --release v1.2.0 --force
// Flag.input: type=string, usage=Path to manual test results JSON file (required)
// Flag.release: type=string, usage=Release version being tested (required)
// Flag.force: type=bool, usage=Overwrite existing results if present
package test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/core/fileutil"
	"github.com/ready-to-release/eac/go/core/config"
)

// ManualTestResults represents the imported manual test results.
type ManualTestResults struct {
	ImportMetadata ImportMetadata     `json:"import_metadata"`
	Results        []ManualTestResult `json:"results"`
}

// ImportMetadata contains metadata about the import.
type ImportMetadata struct {
	TestTime        string  `json:"test_time"`
	Tester          string  `json:"tester"`
	Module          string  `json:"module"`
	ReleaseVersion  string  `json:"release_version"`
	DurationSeconds float64 `json:"duration_seconds,omitempty"`
	SchemaVersion   string  `json:"schema_version"`
}

// ManualTestResult represents a single manual test result.
type ManualTestResult struct {
	ScenarioID      string              `json:"scenario_id"`
	Status          string              `json:"status"`
	DurationSeconds float64             `json:"duration_seconds,omitempty"`
	Notes           string              `json:"notes,omitempty"`
	Error           string              `json:"error,omitempty"`
	Evidence        []EvidenceReference `json:"evidence,omitempty"`
}

// EvidenceReference represents evidence for a test result.
type EvidenceReference struct {
	URL         string `json:"url"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	SHA256      string `json:"sha256,omitempty"`
}

// ImportManual imports manual test results from a JSON file.
func ImportManual() int {
	// Parse command line arguments
	args := os.Args[2:] // Skip "test" and "import-manual"

	var inputFlag, releaseFlag string
	var forceFlag bool

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--input":
			if i+1 < len(args) {
				inputFlag = args[i+1]
				i++
			}
		case "--release":
			if i+1 < len(args) {
				releaseFlag = args[i+1]
				i++
			}
		case "--force":
			forceFlag = true
		}
	}

	// Validate required flags
	if inputFlag == "" {
		log.Errorf("--input flag is required")
		return 1
	}
	if releaseFlag == "" {
		log.Errorf("--release flag is required")
		return 1
	}

	// Load input file
	data, err := os.ReadFile(inputFlag)
	if err != nil {
		log.Errorf("reading input file: %v", err)
		return 1
	}

	// Parse JSON
	var results ManualTestResults
	if err := json.Unmarshal(data, &results); err != nil {
		log.Errorf("parsing JSON: %v", err)
		return 1
	}

	// Get workspace root
	workspaceRoot, err := os.Getwd()
	if err != nil {
		log.Errorf("getting working directory: %v", err)
		return 1
	}

	// Validate against schema
	schemaPath := filepath.Join(workspaceRoot, "contracts/core/0.1.0/schemas/manual-test-results.schema.json")
	if err := validateAgainstSchema(inputFlag, schemaPath); err != nil {
		log.Errorf("schema validation failed: %v", err)
		return 1
	}

	// Additional validation for required fields and formats (jsonschema may not validate formats by default)
	if err := validateImportMetadata(&results.ImportMetadata); err != nil {
		log.Errorf("validation failed: %v", err)
		return 1
	}

	// Validate release version matches
	if results.ImportMetadata.ReleaseVersion != releaseFlag {
		log.Errorf("release version mismatch: file has %s, flag specifies %s",
			results.ImportMetadata.ReleaseVersion, releaseFlag)
		return 1
	}

	// Load repository config
	repoCfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot})
	if err != nil {
		log.Errorf("loading repository config: %v", err)
		return 1
	}

	// Extract and validate module from scenario IDs
	if len(results.Results) == 0 {
		log.Errorf("no test results found in input file")
		return 1
	}

	module := extractModuleFromScenarioID(results.Results[0].ScenarioID)
	if module == "" {
		log.Errorf("invalid scenario ID format: %s", results.Results[0].ScenarioID)
		return 1
	}

	// Validate module exists in repository
	_, exists := repoCfg.Repository.GetModule(module)
	if !exists {
		log.Errorf("unknown module: %s", module)
		return 1
	}

	// Ensure all results belong to the same module
	for _, result := range results.Results {
		m := extractModuleFromScenarioID(result.ScenarioID)
		if m != module {
			log.Errorf("mixed modules in results: found %s and %s", module, m)
			return 1
		}
	}

	// Validate metadata module matches scenario IDs
	if results.ImportMetadata.Module != module {
		log.Errorf("module mismatch: metadata says %s, scenario IDs say %s",
			results.ImportMetadata.Module, module)
		return 1
	}

	// Validate scenario IDs against export (if export exists)
	exportPath := filepath.Join(workspaceRoot, "manual-test-exports", module, releaseFlag+".json")
	if fileExists(exportPath) {
		if err := validateScenarioIDsAgainstExport(results.Results, exportPath); err != nil {
			log.Errorf("%v", err)
			return 1
		}
	}

	// Check for existing results (conflict detection)
	testResultsDir := filepath.Join(workspaceRoot, "test-results", module, releaseFlag)
	outputPath := filepath.Join(testResultsDir, "manual-results.json")

	if fileExists(outputPath) && !forceFlag {
		log.Errorf("manual test results already exist for %s %s", module, releaseFlag)
		log.Errorf("  Location: %s", outputPath)
		log.Errorf("  To overwrite, use --force flag")
		return 1
	}

	// Create output directory
	if err := os.MkdirAll(testResultsDir, 0755); err != nil {
		log.Errorf("creating output directory: %v", err)
		return 1
	}

	// Write results file atomically with locking (prevents concurrent import corruption)
	if err := fileutil.AtomicWriteJSONWithLock(outputPath, results, 0644); err != nil {
		log.Errorf("writing results file: %v", err)
		return 1
	}

	// Count results by status
	passed, failed, skipped := countResultsByStatus(results.Results)

	log.Infof("Imported manual test results for %s %s", module, releaseFlag)
	log.Infof("  Location: %s", outputPath)
	log.Infof("  Tests: %d passed, %d failed, %d skipped", passed, failed, skipped)

	return 0
}

// extractModuleFromScenarioID extracts the module from a scenario ID.
// Format: module/feature/scenario-slug
func extractModuleFromScenarioID(scenarioID string) string {
	parts := strings.Split(scenarioID, "/")
	if len(parts) < 3 {
		return ""
	}
	return parts[0]
}

// countResultsByStatus counts results by their status.
func countResultsByStatus(results []ManualTestResult) (passed, failed, skipped int) {
	for _, result := range results {
		switch result.Status {
		case "passed":
			passed++
		case "failed":
			failed++
		case "skipped":
			skipped++
		}
	}
	return
}

// validateImportMetadata validates import metadata fields that may not be caught by schema validation.
func validateImportMetadata(metadata *ImportMetadata) error {
	// Validate required field: tester
	if metadata.Tester == "" {
		return fmt.Errorf("required field 'tester' is missing or empty")
	}

	// Validate email format using simple regex
	// RFC 5322 compliant regex is complex, so using a simple but practical validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(metadata.Tester) {
		return fmt.Errorf("invalid email format for tester: %s", metadata.Tester)
	}

	return nil
}

// validateScenarioIDsAgainstExport validates that all scenario IDs in results exist in the export.
func validateScenarioIDsAgainstExport(results []ManualTestResult, exportPath string) error {
	// Load export file
	data, err := os.ReadFile(exportPath)
	if err != nil {
		return fmt.Errorf("reading export file: %w", err)
	}

	// Parse export
	var export ManualTestExport
	if err := json.Unmarshal(data, &export); err != nil {
		return fmt.Errorf("parsing export file: %w", err)
	}

	// Build set of valid scenario IDs
	validIDs := make(map[string]bool)
	for _, scenario := range export.Scenarios {
		validIDs[scenario.ScenarioID] = true
	}

	// Validate all result scenario IDs
	for _, result := range results {
		if !validIDs[result.ScenarioID] {
			return fmt.Errorf("scenario not found in export: %s", result.ScenarioID)
		}
	}

	return nil
}
