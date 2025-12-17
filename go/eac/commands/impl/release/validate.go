// Command: validate release
// Short: Validate changelog format and structure
// Long: Validates that changelog files follow the Keep a Changelog format and
// Long: contain valid version entries.
// Long:
// Long: Checks performed:
// Long:   - File exists at the path defined in module contract
// Long:   - Valid Keep a Changelog header format
// Long:   - Version entries have valid format ([x.y.z] - YYYY-MM-DD)
// Long:   - Version numbers are valid semver or calver
// Long:   - Versions are in descending order (newest first)
// Long:   - No duplicate version numbers
// Long:
// Long: Examples:
// Long:   validate release r2r-cli           # Validate single module
// Long:   validate release --all             # Validate all modules with changelogs
// Flag.all: type=bool, usage=Validate all modules with changelogs
// Flag.json: type=bool, usage=Output result in JSON format
// Args: modules
package release

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/changelog"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var validateLog = logging.C("release")

func init() {
	registry.Register(ReleaseValidate)
}

// ValidationResult contains the result of changelog validation
type ValidationResult struct {
	Module   string   `json:"module"`
	Valid    bool     `json:"valid"`
	Path     string   `json:"path"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// ValidationReport contains results for multiple modules
type ValidationReport struct {
	Results  []ValidationResult `json:"results"`
	AllValid bool               `json:"all_valid"`
}

func ReleaseValidate() int {
	// Parse flags
	module := ""
	checkAll := false
	asJSON := false

	args := os.Args[3:] // Skip binary, "release", "validate"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--help" || arg == "-h":
			// Help is handled by the framework via command comments
			return 0
		case arg == "--all":
			checkAll = true
		case arg == "--json":
			asJSON = true
		default:
			if !strings.HasPrefix(arg, "--") && module == "" {
				module = arg
			}
		}
	}

	if !checkAll && module == "" {
		validateLog.Error("module moniker required (or use --all)")
		validateLog.Info("Usage: release validate <module> [--json]")
		validateLog.Info("       release validate --all [--json]")
		return 1
	}

	// Load workspace root
	workspaceRoot, err := registry.GetWorkspaceRoot()
	if err != nil {
		validateLog.Errorf("failed to get workspace root: %v", err)
		return 1
	}

	// Load module contracts
	moduleRegistry, err := modules.LoadFromWorkspace("")
	if err != nil {
		validateLog.Errorf("failed to load modules: %v", err)
		return 1
	}

	// Determine which modules to validate
	var modulesToValidate []string
	if checkAll {
		modulesToValidate = findModulesWithChangelogs(workspaceRoot)
	} else {
		modulesToValidate = []string{module}
	}

	// Validate each module
	report := ValidationReport{
		Results:  make([]ValidationResult, 0, len(modulesToValidate)),
		AllValid: true,
	}

	for _, mod := range modulesToValidate {
		moduleContract, exists := moduleRegistry.Get(mod)
		if !exists {
			validateLog.Warnf("module '%s' not found", mod)
			continue
		}
		result := validateChangelog(mod, moduleContract, workspaceRoot)
		report.Results = append(report.Results, result)
		if !result.Valid {
			report.AllValid = false
		}
	}

	// Output results
	if asJSON {
		var output interface{}
		if len(report.Results) == 1 {
			output = report.Results[0]
		} else {
			output = report
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			validateLog.Errorf("failed to marshal JSON: %v", err)
			return 1
		}
		validateLog.Info(string(jsonBytes))
	} else {
		for _, result := range report.Results {
			if result.Valid {
				validateLog.Infof("✅ %s: valid", result.Module)
			} else {
				validateLog.Infof("❌ %s: invalid", result.Module)
				for _, e := range result.Errors {
					validateLog.Infof("   - %s", e)
				}
			}
			for _, w := range result.Warnings {
				validateLog.Infof("   ⚠️  %s", w)
			}
		}

		if len(report.Results) > 1 {
			validateLog.Info("")
			if report.AllValid {
				validateLog.Info("All changelogs valid.")
			} else {
				validateLog.Info("Some changelogs have errors.")
			}
		}
	}

	if !report.AllValid {
		return 1
	}
	return 0
}

func validateChangelog(module string, moduleContract *modules.ModuleContract, workspaceRoot string) ValidationResult {
	changelogPath := moduleContract.GetChangelogPath()
	result := ValidationResult{
		Module: module,
		Path:   changelogPath,
		Valid:  true,
	}

	fullPath := filepath.Join(workspaceRoot, changelogPath)

	// Check file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("changelog not found at %s", result.Path))
		return result
	}

	// Parse changelog
	cl, err := changelog.Parse(fullPath)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("failed to parse changelog: %v", err))
		return result
	}

	// Validate header
	if cl.Title == "" {
		result.Warnings = append(result.Warnings, "missing title")
	}

	// Validate versions
	if len(cl.Versions) == 0 {
		result.Warnings = append(result.Warnings, "no version entries found")
	}

	// Check for duplicate versions
	versionSet := make(map[string]bool)
	for _, v := range cl.Versions {
		if versionSet[v.Number] {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("duplicate version: %s", v.Number))
		}
		versionSet[v.Number] = true
	}

	// Validate version format based on type
	semverRegex := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	calverRegex := regexp.MustCompile(`^\d{4}\.\d{2}\.\d{2}(\.\d+)?$`)

	for _, v := range cl.Versions {
		if v.Number == "" {
			result.Valid = false
			result.Errors = append(result.Errors, "found version with empty number")
			continue
		}

		// Check version format
		isSemver := semverRegex.MatchString(v.Number)
		isCalver := calverRegex.MatchString(v.Number)

		if !isSemver && !isCalver {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("invalid version format: %s (expected semver x.y.z or calver YYYY.MM.DD)", v.Number))
		}

		// Check date
		if v.Date.IsZero() {
			result.Warnings = append(result.Warnings, fmt.Sprintf("version %s missing date", v.Number))
		}

		// Check for empty version (no entries)
		if !v.HasEntries() {
			result.Warnings = append(result.Warnings, fmt.Sprintf("version %s has no entries", v.Number))
		}
	}

	// Check version order (should be descending)
	for i := 1; i < len(cl.Versions); i++ {
		prev := cl.Versions[i-1].Number
		curr := cl.Versions[i].Number

		// Simple string comparison works for both semver and calver if formatted consistently
		// But for proper ordering, we'd need version-aware comparison
		// For now, just check dates if available
		if !cl.Versions[i-1].Date.IsZero() && !cl.Versions[i].Date.IsZero() {
			if cl.Versions[i].Date.After(cl.Versions[i-1].Date) {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("version order may be incorrect: %s (%s) comes before %s (%s)",
						prev, cl.Versions[i-1].Date.Format("2006-01-02"),
						curr, cl.Versions[i].Date.Format("2006-01-02")))
			}
		}
	}

	return result
}
