// Command: validate control-tags
// Short: Validate @control tags against OSCAL catalog
// Long: Validates that all @control and @controls tags in feature files reference
// Long: valid control IDs from the OSCAL catalog.
// Long:
// Long: This ensures:
// Long:   - Control IDs use correct format (e.g., ac-2, au-3(1))
// Long:   - Control IDs exist in the OSCAL catalog
// Long:   - Evidence collection can find these controls during assessment
// Long:
// Long: The validation:
// Long:   - Discovers all Gherkin feature files in the repository
// Long:   - Extracts all @control: and @controls: tags
// Long:   - Loads the OSCAL catalog from templates/specs/risk-catalog/
// Long:   - Checks that each control ID exists in the catalog
// Long:   - Reports invalid or missing control IDs with file locations
// Long:
// Long: Expected Output:
// Long:   Displays invalid or missing control IDs with file locations (path:line).
// Long:   Groups errors by control ID. Exit code 0 if all tags valid, 1 if invalid tags found.
// Long:
// Long: Example:
// Long:   validate control-tags
package validate

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

var ctlog = logging.C()

func init() {
	registry.Register(ControlTags)
}

var (
	controlTagPattern  = regexp.MustCompile(`@control:([a-z]{2,4}-[0-9]+(?:\([0-9]+\))?)`)
	controlsTagPattern = regexp.MustCompile(`@controls:((?:[a-z]{2,4}-[0-9]+(?:\([0-9]+\))?,)*[a-z]{2,4}-[0-9]+(?:\([0-9]+\))?)`)
)

// ControlTagUsage tracks where a control tag is used
type ControlTagUsage struct {
	ControlID string
	FilePath  string
	LineNum   int
}

// ControlTags validates @control and @controls tags against OSCAL catalog
func ControlTags() int {
	args := os.Args[3:] // Skip program name, "validate", and "control-tags"

	// Define expected flags (none for this command)
	commandFlags := []flags.FlagDefinition{}

	// Validate flags
	if err := flags.ValidateFlags(args, commandFlags); err != nil {
		ctlog.Errorf("%v", err)
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		ctlog.Errorf("Error: %v", err)
		return 1
	}

	// Load configuration
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot})
	if err != nil {
		ctlog.Errorf("Error: failed to load config: %v", err)
		return 1
	}

	ctlog.Info("Validating OSCAL control tags...")

	// Load OSCAL catalog
	catalogPath := cfg.Repository.RiskCatalogPathAbs(workspaceRoot)
	if _, err := os.Stat(catalogPath); os.IsNotExist(err) {
		ctlog.Errorf("Error: OSCAL catalog not found at %s", catalogPath)
		ctlog.Error("Please ensure the catalog exists before validating control tags")
		return 1
	}

	catalog, err := oscal.LoadCatalog(catalogPath)
	if err != nil {
		ctlog.Errorf("Error loading catalog: %v", err)
		return 1
	}

	// Extract all valid control IDs from catalog for validation
	validControlIDs, err := oscal.ExtractControlIDs(catalog)
	if err != nil {
		ctlog.Errorf("Error extracting control IDs from catalog: %v", err)
		return 1
	}

	// Build lookup map for efficient validation
	validControls := make(map[string]bool)
	for _, id := range validControlIDs {
		validControls[strings.ToLower(id)] = true
	}

	// Discover all feature files
	specsDir := filepath.Join(workspaceRoot, cfg.Repository.Paths.SpecsRoot)
	var featureFiles []string

	err = filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".feature") {
			featureFiles = append(featureFiles, path)
		}
		return nil
	})

	if err != nil {
		ctlog.Errorf("Error: failed to discover feature files: %v", err)
		return 1
	}

	if len(featureFiles) == 0 {
		ctlog.Info("No feature files found to validate")
		return 0
	}

	// Extract and validate control tags from all files
	var invalidUsages []ControlTagUsage
	totalTags := 0

	for _, featurePath := range featureFiles {
		usages, err := extractControlTags(featurePath, workspaceRoot)
		if err != nil {
			ctlog.Errorf("Warning: Failed to process %s: %v", featurePath, err)
			continue
		}

		totalTags += len(usages)

		// Validate each control ID
		for _, usage := range usages {
			normalizedID := strings.ToLower(usage.ControlID)
			if !validControls[normalizedID] {
				invalidUsages = append(invalidUsages, usage)
			}
		}
	}

	// Report results
	if len(invalidUsages) == 0 {
		ctlog.Info("")
		ctlog.Infof("✅ All control tags are valid (%d tags validated)", totalTags)
		return 0
	}

	// Report invalid tags
	ctlog.Error("")
	ctlog.Errorf("❌ Found %d invalid control tag(s):", len(invalidUsages))
	ctlog.Error("")

	// Group by control ID for better reporting
	byControlID := make(map[string][]ControlTagUsage)
	for _, usage := range invalidUsages {
		byControlID[usage.ControlID] = append(byControlID[usage.ControlID], usage)
	}

	// Sort control IDs
	var controlIDs []string
	for id := range byControlID {
		controlIDs = append(controlIDs, id)
	}
	sort.Strings(controlIDs)

	// Print grouped errors
	for _, controlID := range controlIDs {
		usages := byControlID[controlID]
		ctlog.Errorf("  Control '%s' not found in catalog:", controlID)
		for _, usage := range usages {
			relPath, _ := filepath.Rel(workspaceRoot, usage.FilePath)
			ctlog.Errorf("    - %s:%d", relPath, usage.LineNum)
		}
	}

	ctlog.Error("")
	ctlog.Error("Troubleshooting:")
	ctlog.Error("  1. Check catalog at templates/specs/risk-catalog/controls.catalog.json")
	ctlog.Error("  2. Verify control ID format (e.g., ac-2, au-3, ia-5(1))")
	ctlog.Error("  3. Add missing controls to catalog if needed")
	ctlog.Error("  4. Update tags in feature files to use valid control IDs")

	return 1
}

// extractControlTags extracts all control tags from a feature file
func extractControlTags(featurePath string, workspaceRoot string) ([]ControlTagUsage, error) {
	content, err := os.ReadFile(featurePath)
	if err != nil {
		return nil, err
	}

	var usages []ControlTagUsage
	lines := strings.Split(string(content), "\n")

	for lineNum, line := range lines {
		// Find @control:<id> tags
		matches := controlTagPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				usages = append(usages, ControlTagUsage{
					ControlID: match[1],
					FilePath:  featurePath,
					LineNum:   lineNum + 1,
				})
			}
		}

		// Find @controls:<id1>,<id2> tags
		multiMatches := controlsTagPattern.FindAllStringSubmatch(line, -1)
		for _, match := range multiMatches {
			if len(match) > 1 {
				// Split comma-separated IDs
				ids := strings.Split(match[1], ",")
				for _, id := range ids {
					usages = append(usages, ControlTagUsage{
						ControlID: strings.TrimSpace(id),
						FilePath:  featurePath,
						LineNum:   lineNum + 1,
					})
				}
			}
		}
	}

	return usages, nil
}
