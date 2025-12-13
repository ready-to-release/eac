// Command: templates tags
// Description: Extract and validate tags from template feature files
// Short: Extract template tags
// Long: The templates tags command scans template feature files and extracts all Gherkin tags.
// Long: Tags are extracted from both Gherkin lines (@tag) and comment metadata (# @tag).
// Long: Use --validate to check tags against the template-tags.yml configuration.
// Long: Use --format to control output format (table, json, summary).
// Long:
// Long: Expected Output:
// Long:   - Gherkin tags extracted from template files
// Long:   - Validation results if --validate is enabled
// Long:   - Exit code 0 if all tags valid, 1 if validation fails
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug mode
// Flag.template: type=string, usage=Git repository URL or local directory to scan (default: templates)
// Flag.validate: type=bool, default=true, usage=Validate tags against template-tags.yml
// Flag.format: type=string, default=table, usage=Output format (table, json, summary)
// Usage: templates tags [--template <source>] [--validate] [--format <format>] [--debug]
package tags

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"

	"github.com/ready-to-release/eac/go/eac/commands/impl/templates/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	eacConfig "github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

var log = logging.C()

func init() {
	registry.Register(TemplatesTags)
}

// TemplatesTags extracts and validates tags from template feature files
func TemplatesTags() int {
	config, err := parseConfig()
	if err != nil {
		log.Errorf("%v", err)
		return 1
	}
	defer config.Logger.Sync()

	config.Logger.Info("Starting templates tags command",
		zap.String("source", config.TemplateSource),
		zap.Bool("validate", config.Validate),
		zap.String("format", config.OutputFormat))

	// Resolve template directory
	templateDir, cleanup, err := resolveTemplateDirectory(config)
	if err != nil {
		config.Logger.Error("Failed to resolve template directory", zap.Error(err))
		log.Errorf("%v", err)
		return 1
	}
	defer cleanup()

	// Scan for tags
	scanner := internal.NewTagScanner(templateDir)
	result, err := scanner.Scan()
	if err != nil {
		config.Logger.Error("Failed to scan templates", zap.Error(err))
		log.Errorf("Failed to scan templates: %v", err)
		return 1
	}

	config.Logger.Info("Scan completed",
		zap.Int("featureFiles", result.FeatureFiles),
		zap.Int("uniqueTags", len(result.Tags)),
		zap.Int("totalInstances", result.TotalTagInstances))

	// Validate if requested
	var validationReport *internal.ValidationReport
	if config.Validate {
		tagsConfig, err := internal.LoadTemplateTagsConfig(config.WorkspaceRoot)
		if err != nil {
			config.Logger.Warn("Failed to load template-tags.yml, skipping validation", zap.Error(err))
			log.Warnf("⚠ Validation skipped: %v", err)
		} else {
			validationReport = tagsConfig.ValidateTags(result.Tags)
		}
	}

	// Output based on depth and format
	switch config.Depth {
	case "file":
		return outputByFile(result, validationReport, config.OutputFormat)
	case "category":
		return outputByCategory(result, validationReport, config.OutputFormat)
	case "coverage":
		return outputCoverage(result, config.OutputFormat)
	case "consistency":
		return outputConsistency(result, validationReport, config.OutputFormat)
	default:
		// Default summary mode
		switch config.OutputFormat {
		case "json":
			return outputJSON(result, validationReport)
		case "summary":
			return outputSummary(result, validationReport)
		default:
			return outputTable(result, validationReport)
		}
	}
}

func outputTable(result *internal.TagScanResult, report *internal.ValidationReport) int {
	log.Info("Template Tags Extraction Report")
	log.Info("================================\n")

	// Summary
	log.Infof("Feature files scanned: %d", result.FeatureFiles)
	log.Infof("Unique tags found:     %d", len(result.Tags))
	log.Infof("Total tag instances:   %d\n", result.TotalTagInstances)

	// By category
	categories := result.GetCategories()
	log.Info("Tags by Category:")
	log.Info("-----------------")

	for _, category := range categories {
		tags := result.ByCategory[category]
		log.Infof("\n  %s (%d tags):", strings.ToUpper(category), len(tags))
		for _, tag := range tags {
			status := ""
			if report != nil {
				for _, r := range report.Results {
					if r.Tag == tag.Tag {
						if r.IsValid {
							status = " ✓"
						} else {
							status = " ✗"
						}
						break
					}
				}
			}
			log.Infof("    %s%s  [%d locations]", tag.Tag, status, len(tag.Locations))
		}
	}

	// Validation summary if available
	if report != nil {
		log.Info("\n\nValidation Summary:")
		log.Info("-------------------")
		log.Infof("  Valid tags:   %d", report.ValidTags)
		log.Infof("  Invalid tags: %d", report.InvalidTags)

		if report.InvalidTags > 0 {
			log.Info("\nInvalid Tags:")
			for _, r := range report.Results {
				if !r.IsValid {
					log.Infof("  ✗ %s", r.Tag)
					if r.Error != "" {
						log.Infof("      Error: %s", r.Error)
					}
					if len(r.Suggestions) > 0 {
						log.Infof("      Suggestions: %s", strings.Join(r.Suggestions[:min(3, len(r.Suggestions))], ", "))
					}
				}
			}
			return 1
		}
	}

	log.Info("\n✅ Tag extraction completed successfully")
	return 0
}

func outputSummary(result *internal.TagScanResult, report *internal.ValidationReport) int {
	log.Infof("Scanned %d feature files, found %d unique tags (%d instances)",
		result.FeatureFiles, len(result.Tags), result.TotalTagInstances)

	categories := result.GetCategories()
	for _, category := range categories {
		tags := result.ByCategory[category]
		log.Infof("  %s: %d tags", category, len(tags))
	}

	if report != nil {
		if report.InvalidTags > 0 {
			log.Infof("\n❌ %d invalid tags found", report.InvalidTags)
			return 1
		}
		log.Infof("\n✅ All %d tags are valid", report.ValidTags)
	}

	return 0
}

func outputJSON(result *internal.TagScanResult, report *internal.ValidationReport) int {
	output := struct {
		Summary struct {
			FeatureFiles      int `json:"feature_files"`
			UniqueTags        int `json:"unique_tags"`
			TotalTagInstances int `json:"total_instances"`
		} `json:"summary"`
		Tags       []internal.TagInfo            `json:"tags"`
		ByCategory map[string][]internal.TagInfo `json:"by_category"`
		Validation *internal.ValidationReport    `json:"validation,omitempty"`
	}{
		Tags:       result.Tags,
		ByCategory: result.ByCategory,
		Validation: report,
	}
	output.Summary.FeatureFiles = result.FeatureFiles
	output.Summary.UniqueTags = len(result.Tags)
	output.Summary.TotalTagInstances = result.TotalTagInstances

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Errorf("Failed to marshal JSON: %v", err)
		return 1
	}

	fmt.Println(string(jsonBytes))

	if report != nil && report.InvalidTags > 0 {
		return 1
	}
	return 0
}

func resolveTemplateDirectory(config *Config) (string, func(), error) {
	var templateDir string
	var cleanup func()

	if internal.IsGitRepository(config.TemplateSource) {
		// Git URL provided - use local templates from appropriate root
		var root string
		if containerRoot := repository.GetContainerRoot(); containerRoot != "" {
			// Running in container - use container root
			root = containerRoot
			config.Logger.Info("Running in container, using local templates",
				zap.String("containerRoot", containerRoot))
		} else {
			// Not in container - use workspace root
			root = config.WorkspaceRoot
			config.Logger.Info("Using local templates from repository",
				zap.String("workspaceRoot", root))
		}

		// Load EAC config for template path
		cfg, err := eacConfig.Load(eacConfig.LoadOptions{RepoRoot: root})
		if err != nil {
			return "", nil, fmt.Errorf("failed to load config: %w", err)
		}

		templateDir = filepath.Join(root, cfg.Repository.Paths.Templates)
		cleanup = func() {}

		config.Logger.Debug("Local templates validated",
			zap.String("dir", templateDir))
		log.Infof("Using templates from %s", templateDir)
	} else {
		// Local directory - resolve relative to appropriate root
		if !filepath.IsAbs(config.TemplateSource) {
			var root string
			if containerRoot := repository.GetContainerRoot(); containerRoot != "" {
				// Running in container - use container root
				root = containerRoot
				config.Logger.Info("Running in container, using local templates",
					zap.String("containerRoot", containerRoot))
			} else {
				// Not in container - use workspace root
				root = config.WorkspaceRoot
				config.Logger.Info("Using local templates from repository",
					zap.String("workspaceRoot", root))
			}
			templateDir = filepath.Join(root, config.TemplateSource)
		} else {
			templateDir = config.TemplateSource
			config.Logger.Info("Using local templates directory",
				zap.String("path", templateDir))
		}
		cleanup = func() {}
	}

	// Verify directory exists
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return "", cleanup, fmt.Errorf("template directory not found: %s", templateDir)
	}

	return templateDir, cleanup, nil
}

func parseConfig() (*Config, error) {
	args := []string{}
	if len(os.Args) > 3 {
		args = os.Args[3:]
	}

	var templateSource string
	var format string
	var depth string
	validate := true
	debug := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--debug", "-d":
			debug = true
		case "--template":
			if i+1 < len(args) {
				templateSource = args[i+1]
				i++
			}
		case "--format":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "--depth":
			if i+1 < len(args) {
				depth = args[i+1]
				i++
			}
		case "--validate":
			validate = true
		case "--no-validate":
			validate = false
		case "--help", "-h":
			printUsage()
			os.Exit(0)
		}
	}

	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	config := NewConfig(workspaceRoot, debug, nil)

	var loggerInst *logging.Logger
	if debug {
		loggerInst, err = logging.NewWithDebug("templates", workspaceRoot)
	} else {
		loggerInst, err = logging.NewDefault("templates", workspaceRoot)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	config.Logger = loggerInst

	if templateSource != "" {
		config.TemplateSource = templateSource
	}
	if format != "" {
		config.OutputFormat = format
	}
	if depth != "" {
		config.Depth = depth
	}
	config.Validate = validate

	return config, nil
}

func printUsage() {
	log.Info("Extract and validate tags from template feature files")
	log.Info("")
	log.Info("Usage: r2r templates tags [--depth <mode>] [--format <format>] [--template <source>] [--no-validate] [--debug]")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --depth <mode>        Analysis depth (default: summary)")
	log.Info("                        summary    - Tag counts by category")
	log.Info("                        file       - Per-file tag breakdown")
	log.Info("                        category   - Tags grouped by catalog directory")
	log.Info("                        coverage   - Compliance/industry coverage matrix")
	log.Info("                        consistency- Check for missing required tags")
	log.Info("  --format <format>     Output format: table, json, summary (default: table)")
	log.Info("  --template <source>   Git repository URL or local directory to scan")
	log.Info("                        (default: templates)")
	log.Info("  --no-validate         Skip validation against template-tags.yml")
	log.Info("  --debug, -d           Enable debug mode")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Default tag summary")
	log.Info("  r2r templates tags")
	log.Info("")
	log.Info("  # Per-file analysis")
	log.Info("  r2r templates tags --depth file")
	log.Info("")
	log.Info("  # Compliance coverage matrix")
	log.Info("  r2r templates tags --depth coverage")
	log.Info("")
	log.Info("  # Check tag consistency")
	log.Info("  r2r templates tags --depth consistency")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// sortedKeys returns sorted keys from a map
func sortedKeys(m map[string][]internal.TagInfo) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ============================================================================
// Depth-specific output functions
// ============================================================================

// outputByFile shows per-file tag breakdown
func outputByFile(result *internal.TagScanResult, report *internal.ValidationReport, format string) int {
	if format == "json" {
		return outputByFileJSON(result)
	}

	log.Info("Template Tags - Per-File Analysis")
	log.Info("==================================\n")

	// Sort files by path
	var filePaths []string
	for path := range result.ByFile {
		filePaths = append(filePaths, path)
	}
	sort.Strings(filePaths)

	for _, filePath := range filePaths {
		fileInfo := result.ByFile[filePath]
		log.Infof("\n📄 %s", filePath)
		log.Infof("   Risk Control: %s", orDefault(fileInfo.RiskControlID, "(none)"))
		log.Infof("   Scenarios:    %d", fileInfo.ScenarioCount)
		log.Infof("   Severity:     %s", orDefault(fileInfo.Severity, "(none)"))
		log.Infof("   Control Type: %s", orDefault(fileInfo.ControlType, "(none)"))

		if len(fileInfo.Industries) > 0 {
			log.Infof("   Industries:   %s", strings.Join(fileInfo.Industries, ", "))
		}
		if len(fileInfo.Compliance) > 0 {
			log.Infof("   Compliance:   %s", strings.Join(fileInfo.Compliance, ", "))
		}

		log.Infof("   Tags (%d):", len(fileInfo.Tags))
		// Show comment tags
		if tags, ok := fileInfo.TagsByType["comment"]; ok && len(tags) > 0 {
			uniqueTags := uniqueStrings(tags)
			log.Infof("     Comment:  %s", strings.Join(uniqueTags[:min(10, len(uniqueTags))], " "))
		}
		// Show gherkin tags
		if tags, ok := fileInfo.TagsByType["gherkin"]; ok && len(tags) > 0 {
			uniqueTags := uniqueStrings(tags)
			log.Infof("     Gherkin:  %s", strings.Join(uniqueTags[:min(10, len(uniqueTags))], " "))
		}
	}

	log.Infof("\n\nTotal: %d files analyzed", len(filePaths))
	return 0
}

func outputByFileJSON(result *internal.TagScanResult) int {
	output := struct {
		Files map[string]*internal.FileTagInfo `json:"files"`
		Total int                              `json:"total_files"`
	}{
		Files: result.ByFile,
		Total: len(result.ByFile),
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Errorf("Failed to marshal JSON: %v", err)
		return 1
	}
	fmt.Println(string(jsonBytes))
	return 0
}

// outputByCategory shows tags grouped by catalog directory
func outputByCategory(result *internal.TagScanResult, report *internal.ValidationReport, format string) int {
	if format == "json" {
		return outputByCategoryJSON(result)
	}

	log.Info("Template Tags - By Catalog Category")
	log.Info("====================================\n")

	// Sort directories
	var dirPaths []string
	for path := range result.ByDirectory {
		dirPaths = append(dirPaths, path)
	}
	sort.Strings(dirPaths)

	for _, dirPath := range dirPaths {
		dirInfo := result.ByDirectory[dirPath]
		log.Infof("\n📁 %s", dirPath)
		log.Infof("   Files:         %d", dirInfo.FileCount)
		log.Infof("   Risk Controls: %d", len(dirInfo.RiskControls))
		log.Infof("   Total Tags:    %d", dirInfo.TotalTags)

		if len(dirInfo.RiskControls) > 0 {
			log.Info("   Controls:")
			for _, ctrl := range dirInfo.RiskControls[:min(10, len(dirInfo.RiskControls))] {
				log.Infof("     - %s", ctrl)
			}
			if len(dirInfo.RiskControls) > 10 {
				log.Infof("     ... and %d more", len(dirInfo.RiskControls)-10)
			}
		}
	}

	log.Infof("\n\nTotal: %d directories", len(dirPaths))
	return 0
}

func outputByCategoryJSON(result *internal.TagScanResult) int {
	output := struct {
		Directories map[string]*internal.DirTagInfo `json:"directories"`
		Total       int                             `json:"total_directories"`
	}{
		Directories: result.ByDirectory,
		Total:       len(result.ByDirectory),
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Errorf("Failed to marshal JSON: %v", err)
		return 1
	}
	fmt.Println(string(jsonBytes))
	return 0
}

// outputCoverage shows compliance/industry coverage matrix
func outputCoverage(result *internal.TagScanResult, format string) int {
	// Build coverage matrix
	complianceCount := make(map[string]int)
	industryCount := make(map[string]int)
	severityCount := make(map[string]int)
	controlTypeCount := make(map[string]int)

	for _, fileInfo := range result.ByFile {
		for _, c := range fileInfo.Compliance {
			complianceCount[c]++
		}
		for _, i := range fileInfo.Industries {
			industryCount[i]++
		}
		if fileInfo.Severity != "" {
			severityCount[fileInfo.Severity]++
		}
		if fileInfo.ControlType != "" {
			controlTypeCount[fileInfo.ControlType]++
		}
	}

	if format == "json" {
		output := struct {
			Compliance  map[string]int `json:"compliance_coverage"`
			Industry    map[string]int `json:"industry_coverage"`
			Severity    map[string]int `json:"severity_distribution"`
			ControlType map[string]int `json:"control_type_distribution"`
			TotalFiles  int            `json:"total_files"`
		}{
			Compliance:  complianceCount,
			Industry:    industryCount,
			Severity:    severityCount,
			ControlType: controlTypeCount,
			TotalFiles:  len(result.ByFile),
		}
		jsonBytes, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonBytes))
		return 0
	}

	log.Info("Template Tags - Coverage Analysis")
	log.Info("==================================\n")
	log.Infof("Total risk control templates: %d\n", len(result.ByFile))

	// Compliance coverage
	log.Info("Compliance Standards Coverage:")
	log.Info("------------------------------")
	complianceKeys := sortedMapKeys(complianceCount)
	for _, std := range complianceKeys {
		count := complianceCount[std]
		pct := float64(count) / float64(len(result.ByFile)) * 100
		bar := strings.Repeat("█", int(pct/5))
		log.Infof("  %-15s %3d (%5.1f%%) %s", std, count, pct, bar)
	}

	// Industry coverage
	log.Info("\nIndustry Coverage:")
	log.Info("------------------")
	industryKeys := sortedMapKeys(industryCount)
	for _, ind := range industryKeys {
		count := industryCount[ind]
		pct := float64(count) / float64(len(result.ByFile)) * 100
		bar := strings.Repeat("█", int(pct/5))
		log.Infof("  %-10s %3d (%5.1f%%) %s", ind, count, pct, bar)
	}

	// Severity distribution
	log.Info("\nSeverity Distribution:")
	log.Info("----------------------")
	severityKeys := sortedMapKeys(severityCount)
	for _, sev := range severityKeys {
		count := severityCount[sev]
		pct := float64(count) / float64(len(result.ByFile)) * 100
		log.Infof("  %-10s %3d (%5.1f%%)", sev, count, pct)
	}

	// Control type distribution
	log.Info("\nControl Type Distribution:")
	log.Info("--------------------------")
	ctrlKeys := sortedMapKeys(controlTypeCount)
	for _, ct := range ctrlKeys {
		count := controlTypeCount[ct]
		pct := float64(count) / float64(len(result.ByFile)) * 100
		log.Infof("  %-12s %3d (%5.1f%%)", ct, count, pct)
	}

	return 0
}

// outputConsistency checks for missing required tags
func outputConsistency(result *internal.TagScanResult, report *internal.ValidationReport, format string) int {
	type ConsistencyIssue struct {
		FilePath string   `json:"file_path"`
		Issues   []string `json:"issues"`
	}

	var issues []ConsistencyIssue

	for filePath, fileInfo := range result.ByFile {
		var fileIssues []string

		// Skip non-risk-control files (like specification.feature)
		if !strings.Contains(filePath, "risk-controls/catalog") {
			continue
		}

		// Check for required metadata
		if fileInfo.RiskControlID == "" {
			fileIssues = append(fileIssues, "Missing feature-level @risk-control tag")
		}
		if fileInfo.Severity == "" {
			fileIssues = append(fileIssues, "Missing @severity tag")
		}
		if fileInfo.ControlType == "" {
			fileIssues = append(fileIssues, "Missing @control-type tag")
		}
		if len(fileInfo.Industries) == 0 {
			fileIssues = append(fileIssues, "Missing @industry tag(s)")
		}
		if len(fileInfo.Compliance) == 0 {
			fileIssues = append(fileIssues, "Missing compliance standard tag(s)")
		}
		if fileInfo.ScenarioCount == 0 {
			fileIssues = append(fileIssues, "No scenarios defined")
		}

		// Check scenario-level risk-control tags
		gherkinTags := fileInfo.TagsByType["gherkin"]
		hasScenarioRiskControl := false
		for _, tag := range gherkinTags {
			if strings.HasPrefix(tag, "@risk-control:") && strings.Contains(tag, "-") {
				parts := strings.Split(tag, "-")
				lastPart := parts[len(parts)-1]
				if len(lastPart) == 2 && isAllDigits(lastPart) {
					hasScenarioRiskControl = true
					break
				}
			}
		}
		if fileInfo.ScenarioCount > 0 && !hasScenarioRiskControl {
			fileIssues = append(fileIssues, "Scenarios missing @risk-control:<name>-NN tags")
		}

		if len(fileIssues) > 0 {
			issues = append(issues, ConsistencyIssue{
				FilePath: filePath,
				Issues:   fileIssues,
			})
		}
	}

	// Sort by file path
	sort.Slice(issues, func(i, j int) bool {
		return issues[i].FilePath < issues[j].FilePath
	})

	if format == "json" {
		output := struct {
			TotalFiles   int                `json:"total_files"`
			FilesWithIssues int             `json:"files_with_issues"`
			Issues       []ConsistencyIssue `json:"issues"`
		}{
			TotalFiles:      len(result.ByFile),
			FilesWithIssues: len(issues),
			Issues:          issues,
		}
		jsonBytes, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonBytes))
		if len(issues) > 0 {
			return 1
		}
		return 0
	}

	log.Info("Template Tags - Consistency Check")
	log.Info("==================================\n")

	if len(issues) == 0 {
		log.Info("✅ All risk control templates have required metadata tags")
		return 0
	}

	log.Infof("❌ Found %d files with consistency issues:\n", len(issues))

	for _, issue := range issues {
		log.Infof("📄 %s", issue.FilePath)
		for _, iss := range issue.Issues {
			log.Infof("   ⚠ %s", iss)
		}
		log.Info("")
	}

	return 1
}

// Helper functions

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func uniqueStrings(strs []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range strs {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

func sortedMapKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
