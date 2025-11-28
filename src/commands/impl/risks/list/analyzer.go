package list

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ControlLinkage represents a risk control and its linkages
type ControlLinkage struct {
	Path         string      `json:"path"`
	Tags         []string    `json:"tags"`
	ReferencedBy []Reference `json:"referenced_by"`
	Status       string      `json:"status"` // "linked" or "orphaned"
}

// Reference represents a spec that references a control
type Reference struct {
	SpecPath string `json:"path"`
	Line     int    `json:"line"`
}

// LinkageReport contains the full linkage analysis
type LinkageReport struct {
	Controls     []ControlLinkage `json:"controls"`
	Summary      Summary          `json:"summary"`
	MissingLinks []MissingLink    `json:"missing_links,omitempty"`
}

// Summary contains overall statistics
type Summary struct {
	Total    int `json:"total"`
	Linked   int `json:"linked"`
	Orphaned int `json:"orphaned"`
}

// MissingLink represents a spec that should have risk control links
type MissingLink struct {
	SpecPath         string `json:"spec"`
	SuggestedControl string `json:"suggested_control,omitempty"`
}

// analyzeLinkages analyzes all risk control linkages
func analyzeLinkages(workspaceRoot string) (*LinkageReport, error) {
	report := &LinkageReport{}

	// 1. Find all risk controls
	controlsDir := filepath.Join(workspaceRoot, "specs", "risk-controls")
	controls, err := findRiskControls(controlsDir)
	if err != nil {
		return nil, err
	}

	// 2. Find all specs
	specsDir := filepath.Join(workspaceRoot, "specs")
	specs, err := findAllSpecs(specsDir)
	if err != nil {
		return nil, err
	}

	// 3. Build tag index (which specs reference which tags)
	tagIndex := buildTagIndex(specs)

	// 4. Analyze each control
	for _, controlPath := range controls {
		linkage, err := analyzeControl(controlPath, tagIndex)
		if err != nil {
			return nil, err
		}
		report.Controls = append(report.Controls, linkage)
	}

	// 5. Calculate summary
	report.Summary = calculateSummary(report.Controls)

	// 6. Find specs missing risk control links
	report.MissingLinks = findMissingLinks(specs, controls)

	return report, nil
}

// findRiskControls finds all risk control feature files
func findRiskControls(controlsDir string) ([]string, error) {
	var controls []string

	// Check if directory exists
	if _, err := os.Stat(controlsDir); os.IsNotExist(err) {
		return controls, nil
	}

	err := filepath.Walk(controlsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".feature") {
			controls = append(controls, path)
		}
		return nil
	})

	return controls, err
}

// findAllSpecs finds all specification files (excluding risk-controls)
func findAllSpecs(specsDir string) ([]string, error) {
	var specs []string
	controlsDir := filepath.Join(specsDir, "risk-controls")

	err := filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip risk-controls directory
		if info.IsDir() && path == controlsDir {
			return filepath.SkipDir
		}

		if !info.IsDir() && strings.HasSuffix(path, ".feature") {
			specs = append(specs, path)
		}
		return nil
	})

	return specs, err
}

// buildTagIndex creates a map of tag -> []Reference
func buildTagIndex(specs []string) map[string][]Reference {
	tagIndex := make(map[string][]Reference)
	tagRegex := regexp.MustCompile(`@risk-control:[a-z0-9-]+`)

	for _, specPath := range specs {
		content, err := os.ReadFile(specPath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for lineNum, line := range lines {
			tags := tagRegex.FindAllString(line, -1)
			for _, tag := range tags {
				tagIndex[tag] = append(tagIndex[tag], Reference{
					SpecPath: specPath,
					Line:     lineNum + 1, // 1-indexed
				})
			}
		}
	}

	return tagIndex
}

// analyzeControl analyzes a single control
func analyzeControl(controlPath string, tagIndex map[string][]Reference) (ControlLinkage, error) {
	linkage := ControlLinkage{
		Path: controlPath,
	}

	// Extract tags from control
	content, err := os.ReadFile(controlPath)
	if err != nil {
		return linkage, err
	}

	tagRegex := regexp.MustCompile(`@risk-control:[a-z0-9-]+`)
	linkage.Tags = tagRegex.FindAllString(string(content), -1)

	// Deduplicate tags
	seen := make(map[string]bool)
	var uniqueTags []string
	for _, tag := range linkage.Tags {
		if !seen[tag] {
			seen[tag] = true
			uniqueTags = append(uniqueTags, tag)
		}
	}
	linkage.Tags = uniqueTags

	// Find references for each tag
	seenSpecs := make(map[string]bool)
	for _, tag := range linkage.Tags {
		if refs, ok := tagIndex[tag]; ok {
			for _, ref := range refs {
				// Skip self-references (the control file itself)
				if ref.SpecPath == controlPath {
					continue
				}

				// Deduplicate by spec path
				if !seenSpecs[ref.SpecPath] {
					linkage.ReferencedBy = append(linkage.ReferencedBy, ref)
					seenSpecs[ref.SpecPath] = true
				}
			}
		}
	}

	// Determine status
	if len(linkage.ReferencedBy) > 0 {
		linkage.Status = "linked"
	} else {
		linkage.Status = "orphaned"
	}

	return linkage, nil
}

// calculateSummary calculates summary statistics
func calculateSummary(controls []ControlLinkage) Summary {
	summary := Summary{
		Total: len(controls),
	}

	for _, control := range controls {
		if control.Status == "linked" {
			summary.Linked++
		} else {
			summary.Orphaned++
		}
	}

	return summary
}

// findMissingLinks identifies specs that should have risk control links
func findMissingLinks(specs []string, controls []string) []MissingLink {
	var missing []MissingLink

	// Security-related keywords
	securityKeywords := []string{"auth", "security", "encryption", "validation", "access", "permission", "token"}

	for _, specPath := range specs {
		// Check if spec has any risk-control tags
		content, err := os.ReadFile(specPath)
		if err != nil {
			continue
		}

		tagRegex := regexp.MustCompile(`@risk-control:[a-z0-9-]+`)
		if len(tagRegex.FindAllString(string(content), -1)) > 0 {
			// Already has risk control links
			continue
		}

		// Check if spec path suggests it needs risk controls
		specPathLower := strings.ToLower(specPath)
		contentLower := strings.ToLower(string(content))
		needsControl := false

		for _, keyword := range securityKeywords {
			if strings.Contains(specPathLower, keyword) || strings.Contains(contentLower, keyword) {
				needsControl = true
				break
			}
		}

		if needsControl {
			// Suggest a control based on path
			suggested := suggestControl(specPath, controls)
			missing = append(missing, MissingLink{
				SpecPath:         specPath,
				SuggestedControl: suggested,
			})
		}
	}

	return missing
}

// suggestControl suggests a risk control for a spec
func suggestControl(specPath string, controls []string) string {
	// Extract domain from spec path (e.g., specs/auth/login.feature -> auth)
	parts := strings.Split(filepath.ToSlash(specPath), "/")
	if len(parts) < 2 {
		return ""
	}

	domain := parts[len(parts)-2] // Parent directory

	// Find matching control
	for _, control := range controls {
		if strings.Contains(strings.ToLower(control), strings.ToLower(domain)) {
			return control
		}
	}

	return ""
}
