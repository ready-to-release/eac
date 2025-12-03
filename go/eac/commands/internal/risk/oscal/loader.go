// Package oscal provides OSCAL document loading and parsing.
package oscal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LoadProfile loads an OSCAL profile from a file.
func LoadProfile(profilePath string) (*Profile, error) {
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile JSON: %w", err)
	}

	return &profile, nil
}

// LoadAssessmentResults loads an OSCAL assessment-results from a file.
func LoadAssessmentResults(arPath string) (*AssessmentResults, error) {
	data, err := os.ReadFile(arPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read assessment-results file: %w", err)
	}

	var ar AssessmentResults
	if err := json.Unmarshal(data, &ar); err != nil {
		return nil, fmt.Errorf("failed to parse assessment-results JSON: %w", err)
	}

	return &ar, nil
}

// DiscoverProfiles finds all profile files in the specs/risk-controls directory.
// Returns a map of module name to profile path.
func DiscoverProfiles(workspaceRoot string) (map[string]string, error) {
	profileDir := filepath.Join(workspaceRoot, "specs", "risk-controls")

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

// DiscoverAssessmentResults finds all assessment-results files in the out/risk directory.
// Returns a map of module name to assessment-results path.
func DiscoverAssessmentResults(workspaceRoot string) (map[string]string, error) {
	riskDir := filepath.Join(workspaceRoot, "out", "risk")

	assessments := make(map[string]string)

	entries, err := os.ReadDir(riskDir)
	if err != nil {
		if os.IsNotExist(err) {
			return assessments, nil // No risk directory yet
		}
		return nil, fmt.Errorf("failed to read risk directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		moduleName := entry.Name()
		arPath := filepath.Join(riskDir, moduleName, "assessment-results.json")

		if _, err := os.Stat(arPath); err == nil {
			assessments[moduleName] = arPath
		}
	}

	return assessments, nil
}

// GetProfilePath returns the default path for a module's profile.
func GetProfilePath(workspaceRoot, moduleName string) string {
	if moduleName == "" {
		moduleName = "common"
	}
	return filepath.Join(workspaceRoot, "specs", "risk-controls", moduleName+".profile.json")
}

// GetAssessmentResultsPath returns the default path for a module's assessment-results.
func GetAssessmentResultsPath(workspaceRoot, moduleName string) string {
	return filepath.Join(workspaceRoot, "out", "risk", moduleName, "assessment-results.json")
}

// LoadAllAssessmentResults loads all assessment-results files for aggregated reporting.
func LoadAllAssessmentResults(workspaceRoot string) ([]*AssessmentResults, error) {
	arMap, err := DiscoverAssessmentResults(workspaceRoot)
	if err != nil {
		return nil, err
	}

	if len(arMap) == 0 {
		return nil, fmt.Errorf("no assessment-results found in out/risk/")
	}

	var results []*AssessmentResults
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
func LoadAllProfiles(workspaceRoot string) ([]*Profile, error) {
	profileMap, err := DiscoverProfiles(workspaceRoot)
	if err != nil {
		return nil, err
	}

	if len(profileMap) == 0 {
		return nil, fmt.Errorf("no profiles found in specs/risk-controls/")
	}

	var profiles []*Profile
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
