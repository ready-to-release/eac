package create

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TagAnalysis contains the results of analyzing tags
type TagAnalysis struct {
	ExistingTags    []string // Tags in the existing control
	ReferencedTags  []string // Tags referenced in other specs
	OrphanedTags    []string // Tags that would be orphaned
	PreservableTags []string // Tags that should be preserved
}

// analyzeTagOrphans checks if overwriting a control would create orphaned tags
func analyzeTagOrphans(existingControlPath string, newControlContent string, specsDir string) (*TagAnalysis, error) {
	analysis := &TagAnalysis{}

	// 1. Extract tags from existing control
	existingContent, err := os.ReadFile(existingControlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read existing control: %w", err)
	}
	analysis.ExistingTags = extractTags(string(existingContent))

	// 2. Extract tags from new control content
	newTags := extractTags(newControlContent)

	// 3. Find tags that would be removed
	removedTags := difference(analysis.ExistingTags, newTags)
	if len(removedTags) == 0 {
		// No tags being removed, safe to proceed
		return analysis, nil
	}

	// 4. Scan all specs for references to removed tags
	for _, tag := range removedTags {
		isReferenced, err := isTagReferenced(tag, existingControlPath, specsDir)
		if err != nil {
			return nil, err
		}
		if isReferenced {
			analysis.ReferencedTags = append(analysis.ReferencedTags, tag)
			analysis.OrphanedTags = append(analysis.OrphanedTags, tag)
			analysis.PreservableTags = append(analysis.PreservableTags, tag)
		}
	}

	return analysis, nil
}

// extractTags extracts all @risk-control tags from Gherkin content
func extractTags(content string) []string {
	tagRegex := regexp.MustCompile(`@risk-control:[a-z0-9-]+`)
	matches := tagRegex.FindAllString(content, -1)

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, tag := range matches {
		if !seen[tag] {
			seen[tag] = true
			unique = append(unique, tag)
		}
	}

	return unique
}

// isTagReferenced checks if a tag is referenced in any spec file except the control being overwritten
func isTagReferenced(tag string, excludePath string, specsDir string) (bool, error) {
	var referenced bool

	err := filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the control file we're about to overwrite
		absExcludePath, _ := filepath.Abs(excludePath)
		absPath, _ := filepath.Abs(path)
		if absPath == absExcludePath {
			return nil
		}

		// Only check .feature files
		if !info.IsDir() && strings.HasSuffix(path, ".feature") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Check if tag is referenced
			if strings.Contains(string(content), tag) {
				referenced = true
				return filepath.SkipAll // Found it, stop searching
			}
		}

		return nil
	})

	return referenced, err
}

// difference returns elements in a that are not in b
func difference(a, b []string) []string {
	bMap := make(map[string]bool)
	for _, item := range b {
		bMap[item] = true
	}

	var diff []string
	for _, item := range a {
		if !bMap[item] {
			diff = append(diff, item)
		}
	}

	return diff
}

// formatOrphanedTagsError creates a user-friendly error message for orphaned tags
func formatOrphanedTagsError(analysis *TagAnalysis, controlPath string) string {
	var msg strings.Builder

	msg.WriteString("Cannot overwrite control - would create orphaned tags\n\n")
	msg.WriteString(fmt.Sprintf("Control: %s\n\n", controlPath))
	msg.WriteString("The following tags are referenced in other specifications but would be removed:\n")
	for _, tag := range analysis.OrphanedTags {
		msg.WriteString(fmt.Sprintf("  - %s\n", tag))
	}
	msg.WriteString("\nTo fix this:\n")
	msg.WriteString("1. Ensure the new control includes these tags:\n")
	for _, tag := range analysis.PreservableTags {
		msg.WriteString(fmt.Sprintf("     %s\n", tag))
	}
	msg.WriteString("\n2. OR remove references to these tags from other specs\n")
	msg.WriteString("\n3. OR use --allow-orphans to force overwrite (not recommended)\n")

	return msg.String()
}
