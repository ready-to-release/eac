// Package oscal provides OSCAL document loading and parsing.
package oscal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/core/paths"
)

// LoadProfile loads an OSCAL profile from a file using go-oscal types.
func LoadProfile(profilePath string) (*oscalTypes.Profile, error) {
	data, err := os.ReadFile(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("profile file not found: %s", profilePath)
		}
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	var oscalDoc oscalTypes.OscalModels
	if err := json.Unmarshal(data, &oscalDoc); err != nil {
		return nil, fmt.Errorf("failed to parse profile JSON: %w", err)
	}

	if oscalDoc.Profile == nil {
		return nil, fmt.Errorf("file does not contain a profile document")
	}

	return oscalDoc.Profile, nil
}

// LoadAssessmentResults loads an OSCAL assessment-results from a file using go-oscal types.
func LoadAssessmentResults(arPath string) (*oscalTypes.AssessmentResults, error) {
	data, err := os.ReadFile(arPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read assessment-results file: %w", err)
	}

	var oscalDoc oscalTypes.OscalModels
	if err := json.Unmarshal(data, &oscalDoc); err != nil {
		return nil, fmt.Errorf("failed to parse assessment-results JSON: %w", err)
	}

	if oscalDoc.AssessmentResults == nil {
		return nil, fmt.Errorf("file does not contain an assessment-results document")
	}

	return oscalDoc.AssessmentResults, nil
}

// DiscoverProfiles finds all profile files in the specs/risk-controls directory.
// Returns a map of module name to profile path.
func DiscoverProfiles(workspaceRoot string) (map[string]string, error) {
	profileDir := filepath.Join(workspaceRoot, paths.SpecsDir, paths.RiskControlsDir)

	profiles := make(map[string]string)

	entries, err := os.ReadDir(profileDir)
	if err != nil {
		if os.IsNotExist(err) {
			return profiles, nil // No profiles directory yet
		}
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".profile.json") {
			continue
		}

		// Extract module name from filename (e.g., "billing-service.profile.json" -> "billing-service")
		moduleName := strings.TrimSuffix(name, ".profile.json")
		profiles[moduleName] = filepath.Join(profileDir, name)
	}

	return profiles, nil
}

// findLatestAssessmentResults finds the most recent assessment-results file in a module directory.
// Returns the full path to the latest file, or empty string if none found.
func findLatestAssessmentResults(moduleDir string) (string, error) {
	entries, err := os.ReadDir(moduleDir)
	if err != nil {
		return "", err
	}

	var assessmentFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Match timestamped format (assessment-results-YYYYMMDD-HHMMSS.json)
		if strings.HasPrefix(name, "assessment-results-") && strings.HasSuffix(name, ".json") {
			assessmentFiles = append(assessmentFiles, name)
		}
	}

	if len(assessmentFiles) == 0 {
		return "", fmt.Errorf("no assessment-results files found")
	}

	// Sort in reverse order (latest first) - timestamps are lexicographically sortable
	sort.Sort(sort.Reverse(sort.StringSlice(assessmentFiles)))

	// Return the latest file
	return filepath.Join(moduleDir, assessmentFiles[0]), nil
}

// DiscoverAssessmentResults finds all assessment-results files in the out/risk directory.
// Returns a map of module name to assessment-results path.
// If multiple assessment-results exist for a module, returns the most recent one.
func DiscoverAssessmentResults(workspaceRoot string) (map[string]string, error) {
	riskDir := filepath.Join(workspaceRoot, paths.OutDir, paths.RiskDir)

	assessments := make(map[string]string)

	entries, err := os.ReadDir(riskDir)
	if err != nil {
		if os.IsNotExist(err) {
			return assessments, nil // No risk directory yet
		}
		return nil, fmt.Errorf("failed to read risk directory: %w", err)
	}

	// Find the latest timestamp directory (format: 2025-12-05T10-26-39)
	var latestTimestamp string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if directory name looks like a timestamp (has 'T' character)
		// This distinguishes between timestamp dirs and old-style module dirs
		if strings.Contains(entry.Name(), "T") {
			if entry.Name() > latestTimestamp {
				latestTimestamp = entry.Name()
			}
		}
	}

	// If we found a timestamp directory, look for modules inside it
	if latestTimestamp != "" {
		timestampDir := filepath.Join(riskDir, latestTimestamp)
		moduleEntries, err := os.ReadDir(timestampDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read timestamp directory: %w", err)
		}

		for _, entry := range moduleEntries {
			if !entry.IsDir() {
				continue
			}

			moduleName := entry.Name()
			moduleDir := filepath.Join(timestampDir, moduleName)

			// Look for assessment-results.json
			arFile := filepath.Join(moduleDir, "assessment-results.json")
			if _, err := os.Stat(arFile); err == nil {
				assessments[moduleName] = arFile
			}
		}
	} else {
		// Fallback: Look for old-style structure (direct module directories)
		for _, entry := range entries {
			if !entry.IsDir() || strings.Contains(entry.Name(), "T") {
				continue
			}

			moduleName := entry.Name()
			moduleDir := filepath.Join(riskDir, moduleName)

			// Find latest assessment-results file
			latestFile, err := findLatestAssessmentResults(moduleDir)
			if err == nil && latestFile != "" {
				assessments[moduleName] = latestFile
			}
		}
	}

	return assessments, nil
}

// GetProfilePath returns the default path for a module's profile.
// Uses centralized repository path conventions.
func GetProfilePath(workspaceRoot, moduleName string) string {
	return paths.RiskProfilePath(workspaceRoot, moduleName)
}

// GetAssessmentResultsPath returns the default path for a module's assessment-results.
// Uses centralized repository path conventions.
func GetAssessmentResultsPath(workspaceRoot, moduleName string) string {
	return paths.RiskAssessmentResultsPath(workspaceRoot, moduleName)
}

// LoadAllAssessmentResults loads all assessment-results files for aggregated reporting.
func LoadAllAssessmentResults(workspaceRoot string) ([]*oscalTypes.AssessmentResults, error) {
	arMap, err := DiscoverAssessmentResults(workspaceRoot)
	if err != nil {
		return nil, err
	}

	if len(arMap) == 0 {
		return nil, fmt.Errorf("no assessment-results found in %s/%s/", paths.OutDir, paths.RiskDir)
	}

	var results []*oscalTypes.AssessmentResults
	for _, arPath := range arMap {
		ar, err := LoadAssessmentResults(arPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", arPath, err)
		}
		results = append(results, ar)
	}

	return results, nil
}

// LoadAllProfiles loads all profile files for a given module or all modules.
func LoadAllProfiles(workspaceRoot string) ([]*oscalTypes.Profile, error) {
	profileMap, err := DiscoverProfiles(workspaceRoot)
	if err != nil {
		return nil, err
	}

	if len(profileMap) == 0 {
		return nil, fmt.Errorf("no profiles found in %s/%s/", paths.SpecsDir, paths.RiskControlsDir)
	}

	var profiles []*oscalTypes.Profile
	for _, profilePath := range profileMap {
		profile, err := LoadProfile(profilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", profilePath, err)
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// GetModuleNamesFromAssessments returns sorted list of module names with assessment-results.
func GetModuleNamesFromAssessments(workspaceRoot string) ([]string, error) {
	arMap, err := DiscoverAssessmentResults(workspaceRoot)
	if err != nil {
		return nil, err
	}

	var names []string
	for name := range arMap {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// DetectOSCALDocumentType detects if a file is a profile, assessment-results, or catalog.
// Returns "profile", "assessment-results", "catalog", or empty string if unknown.
func DetectOSCALDocumentType(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Try to detect based on JSON structure
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	if _, ok := raw["profile"]; ok {
		return "profile", nil
	}

	if _, ok := raw["assessment-results"]; ok {
		return "assessment-results", nil
	}

	// Also check for hyphenated version
	if _, ok := raw["assessment_results"]; ok {
		return "assessment-results", nil
	}

	if _, ok := raw["catalog"]; ok {
		return "catalog", nil
	}

	return "", nil
}
