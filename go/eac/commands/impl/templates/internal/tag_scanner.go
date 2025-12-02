package internal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// TagInfo contains information about a tag and where it's used
type TagInfo struct {
	Tag       string            // The tag (e.g., "@industry:PHARMA")
	Locations []TagLocation     // All locations where this tag appears
	Metadata  map[string]string // Parsed metadata (prefix, suffix for parameterized tags)
}

// TagLocation represents where a tag was found
type TagLocation struct {
	FilePath string // Relative path to the file
	LineNum  int    // Line number (1-based)
	Context  string // "feature", "scenario", "comment"
}

// TagScanResult contains the results of scanning templates for tags
type TagScanResult struct {
	Tags              []TagInfo                // All unique tags found
	FeatureFiles      int                      // Number of feature files scanned
	TotalTagInstances int                      // Total number of tag usages
	ByCategory        map[string][]TagInfo     // Tags grouped by category (prefix)
	ByFile            map[string]*FileTagInfo  // Tags grouped by file path
	ByDirectory       map[string]*DirTagInfo   // Tags grouped by directory
}

// FileTagInfo contains tag information for a single file
type FileTagInfo struct {
	FilePath       string            // Relative path to the file
	Tags           []string          // All tags in this file (deduplicated)
	TagsByType     map[string][]string // Tags grouped by type (comment vs gherkin)
	Metadata       map[string]string // Parsed metadata from comment tags
	RiskControlID  string            // Feature-level risk control ID (if present)
	ScenarioCount  int               // Number of scenarios
	Industries     []string          // Industry codes from metadata
	Compliance     []string          // Compliance standards from metadata
	Severity       string            // Severity level
	ControlType    string            // Control type (preventive, detective, corrective)
}

// DirTagInfo contains aggregated tag information for a directory
type DirTagInfo struct {
	DirPath      string   // Relative directory path
	FileCount    int      // Number of feature files
	TotalTags    int      // Total tag instances
	UniqueTags   []string // Unique tags across all files
	RiskControls []string // Risk control IDs in this directory
}

// TagScanner scans template feature files and extracts Gherkin tags
type TagScanner struct {
	templateDir string
	tags        map[string]*TagInfo    // tag -> info
	files       map[string]*FileTagInfo // file path -> file info
}

// NewTagScanner creates a new scanner for the given template directory
func NewTagScanner(templateDir string) *TagScanner {
	return &TagScanner{
		templateDir: templateDir,
		tags:        make(map[string]*TagInfo),
		files:       make(map[string]*FileTagInfo),
	}
}

// Scan walks the template directory and extracts all tags from feature files
func (s *TagScanner) Scan() (*TagScanResult, error) {
	// Verify template directory exists
	info, err := os.Stat(s.templateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("template directory does not exist: %s", s.templateDir)
		}
		return nil, fmt.Errorf("cannot access template directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("template path is not a directory: %s", s.templateDir)
	}

	featureCount := 0
	totalInstances := 0

	// Walk the directory and scan feature files
	err = filepath.WalkDir(s.templateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process feature files
		if !strings.HasSuffix(path, ".feature") {
			return nil
		}

		featureCount++
		count, err := s.scanFeatureFile(path)
		if err != nil {
			return fmt.Errorf("failed to scan %s: %w", path, err)
		}
		totalInstances += count

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan templates: %w", err)
	}

	// Build result
	result := &TagScanResult{
		FeatureFiles:      featureCount,
		TotalTagInstances: totalInstances,
		ByCategory:        make(map[string][]TagInfo),
		ByFile:            s.files,
		ByDirectory:       make(map[string]*DirTagInfo),
	}

	// Convert map to sorted slice and categorize
	for _, tagInfo := range s.tags {
		result.Tags = append(result.Tags, *tagInfo)

		// Categorize by prefix
		category := categorizeTag(tagInfo.Tag)
		result.ByCategory[category] = append(result.ByCategory[category], *tagInfo)
	}

	// Build directory aggregations
	for filePath, fileInfo := range s.files {
		dirPath := filepath.Dir(filePath)
		if result.ByDirectory[dirPath] == nil {
			result.ByDirectory[dirPath] = &DirTagInfo{
				DirPath: dirPath,
			}
		}
		dirInfo := result.ByDirectory[dirPath]
		dirInfo.FileCount++
		dirInfo.TotalTags += len(fileInfo.Tags)

		// Add unique tags
		tagSet := make(map[string]bool)
		for _, t := range dirInfo.UniqueTags {
			tagSet[t] = true
		}
		for _, t := range fileInfo.Tags {
			if !tagSet[t] {
				dirInfo.UniqueTags = append(dirInfo.UniqueTags, t)
				tagSet[t] = true
			}
		}

		// Add risk controls
		if fileInfo.RiskControlID != "" {
			dirInfo.RiskControls = append(dirInfo.RiskControls, fileInfo.RiskControlID)
		}
	}

	// Sort tags alphabetically
	sort.Slice(result.Tags, func(i, j int) bool {
		return result.Tags[i].Tag < result.Tags[j].Tag
	})

	// Sort each category
	for cat := range result.ByCategory {
		sort.Slice(result.ByCategory[cat], func(i, j int) bool {
			return result.ByCategory[cat][i].Tag < result.ByCategory[cat][j].Tag
		})
	}

	// Sort directory unique tags
	for _, dirInfo := range result.ByDirectory {
		sort.Strings(dirInfo.UniqueTags)
		sort.Strings(dirInfo.RiskControls)
	}

	return result, nil
}

// scanFeatureFile extracts tags from a single feature file
func (s *TagScanner) scanFeatureFile(filePath string) (int, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	// Get relative path for display
	relPath, err := filepath.Rel(s.templateDir, filePath)
	if err != nil {
		relPath = filePath
	}
	relPath = filepath.ToSlash(relPath)

	// Initialize file info
	fileInfo := &FileTagInfo{
		FilePath:   relPath,
		Tags:       []string{},
		TagsByType: make(map[string][]string),
		Metadata:   make(map[string]string),
		Industries: []string{},
		Compliance: []string{},
	}
	tagSet := make(map[string]bool)

	lines := strings.Split(string(content), "\n")
	instanceCount := 0

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Count scenarios
		if strings.HasPrefix(trimmed, "Scenario:") || strings.HasPrefix(trimmed, "Scenario Outline:") {
			fileInfo.ScenarioCount++
		}

		// Determine context
		var context string
		if strings.HasPrefix(trimmed, "#") {
			// Comment line with tags
			context = "comment"
			// Remove leading # and any spaces
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
		} else if strings.HasPrefix(trimmed, "@") {
			// Tag line before Feature/Scenario/Rule
			context = "gherkin"
		} else {
			continue
		}

		// Extract tags from the line
		tags := extractTagsFromLine(trimmed)
		for _, tag := range tags {
			instanceCount++
			s.addTag(tag, TagLocation{
				FilePath: relPath,
				LineNum:  lineNum + 1,
				Context:  context,
			})

			// Add to file-level tracking
			if !tagSet[tag] {
				fileInfo.Tags = append(fileInfo.Tags, tag)
				tagSet[tag] = true
			}
			fileInfo.TagsByType[context] = append(fileInfo.TagsByType[context], tag)

			// Extract metadata from tags
			s.extractFileMetadata(tag, fileInfo)
		}
	}

	// Store file info
	s.files[relPath] = fileInfo

	return instanceCount, nil
}

// extractFileMetadata extracts structured metadata from a tag
func (s *TagScanner) extractFileMetadata(tag string, fileInfo *FileTagInfo) {
	// Feature-level risk control (no numeric suffix like -01, -02)
	if strings.HasPrefix(tag, "@risk-control:") {
		// Check if it ends with a numeric suffix like -01, -02
		parts := strings.Split(tag, "-")
		lastPart := parts[len(parts)-1]
		// Feature-level tags don't end with 2-digit numeric suffix
		if len(lastPart) != 2 || !isNumeric(lastPart) {
			fileInfo.RiskControlID = tag
		}
	}

	// Industry codes
	if strings.HasPrefix(tag, "@industry:") {
		industry := strings.TrimPrefix(tag, "@industry:")
		fileInfo.Industries = append(fileInfo.Industries, industry)
	}

	// Severity
	if strings.HasPrefix(tag, "@severity:") {
		fileInfo.Severity = strings.TrimPrefix(tag, "@severity:")
	}

	// Control type
	if strings.HasPrefix(tag, "@control-type:") {
		fileInfo.ControlType = strings.TrimPrefix(tag, "@control-type:")
	}

	// Compliance standards (standalone tags)
	complianceStandards := []string{
		"@iso27001", "@hipaa", "@pci-dss", "@fda-21cfr11", "@nist-800-53",
		"@sox", "@owasp", "@gdpr", "@ccpa", "@fedramp", "@soc2",
		"@mdr", "@ivdr", "@iso13485", "@iec62304", "@fda-premarket",
		"@imdrf", "@iso14971", "@gxp", "@gamp5", "@alcoa-plus",
		"@eu-ai-act", "@nist-ai-rmf", "@iso42001", "@lgpd", "@eidas",
		"@esign-act", "@csa-ccm", "@fips-140-2", "@fips-140-3",
		"@iso27017", "@iso27018", "@slsa", "@nist-ssdf",
		"@nist-sp800-124", "@owasp-masvs", "@iso31000",
	}
	for _, std := range complianceStandards {
		if tag == std {
			fileInfo.Compliance = append(fileInfo.Compliance, strings.TrimPrefix(tag, "@"))
			break
		}
	}
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// addTag adds or updates a tag entry
func (s *TagScanner) addTag(tag string, location TagLocation) {
	if existing, ok := s.tags[tag]; ok {
		existing.Locations = append(existing.Locations, location)
	} else {
		s.tags[tag] = &TagInfo{
			Tag:       tag,
			Locations: []TagLocation{location},
			Metadata:  parseTagMetadata(tag),
		}
	}
}

// extractTagsFromLine extracts all @tag tokens from a line
func extractTagsFromLine(line string) []string {
	var tags []string
	parts := strings.Fields(line)

	for _, part := range parts {
		if strings.HasPrefix(part, "@") {
			// Clean up any trailing punctuation (including parens from lists)
			tag := strings.TrimRight(part, ",.;:)(")
			if tag != "@" && tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	return tags
}

// parseTagMetadata extracts structured data from parameterized tags
func parseTagMetadata(tag string) map[string]string {
	metadata := make(map[string]string)

	if strings.Contains(tag, ":") {
		parts := strings.SplitN(tag, ":", 2)
		metadata["prefix"] = parts[0]
		if len(parts) > 1 {
			metadata["value"] = parts[1]
		}
	}

	return metadata
}

// categorizeTag determines the category of a tag based on its prefix
func categorizeTag(tag string) string {
	if strings.Contains(tag, ":") {
		parts := strings.SplitN(tag, ":", 2)
		prefix := strings.TrimPrefix(parts[0], "@")
		return prefix
	}
	return "other"
}

// GetTagCount returns the number of unique tags found
func (s *TagScanner) GetTagCount() int {
	return len(s.tags)
}

// GetCategories returns all unique tag categories found
func (result *TagScanResult) GetCategories() []string {
	var categories []string
	for cat := range result.ByCategory {
		categories = append(categories, cat)
	}
	sort.Strings(categories)
	return categories
}

// GetTagsByPrefix returns all tags with a specific prefix
func (result *TagScanResult) GetTagsByPrefix(prefix string) []TagInfo {
	return result.ByCategory[prefix]
}

// GetUniqueValues returns unique values for parameterized tags with a given prefix
func (result *TagScanResult) GetUniqueValues(prefix string) []string {
	valuesSet := make(map[string]bool)
	for _, tagInfo := range result.Tags {
		if tagInfo.Metadata["prefix"] == "@"+prefix {
			if val, ok := tagInfo.Metadata["value"]; ok && val != "" {
				valuesSet[val] = true
			}
		}
	}

	var values []string
	for v := range valuesSet {
		values = append(values, v)
	}
	sort.Strings(values)
	return values
}
