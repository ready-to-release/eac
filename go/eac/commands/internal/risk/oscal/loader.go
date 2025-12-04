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
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

// LoadProfile loads an OSCAL profile from a file using go-oscal types.
func LoadProfile(profilePath string) (*oscalTypes.Profile, error) {
	data, err := os.ReadFile(profilePath)
	if err != nil {
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
	profileDir := filepath.Join(workspaceRoot, repository.SpecsDir, repository.RiskControlsDir)

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
	riskDir := filepath.Join(workspaceRoot, repository.OutDir, repository.RiskDir)

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
// Uses centralized repository path conventions.
func GetProfilePath(workspaceRoot, moduleName string) string {
	return repository.RiskProfilePath(workspaceRoot, moduleName)
}

// GetAssessmentResultsPath returns the default path for a module's assessment-results.
// Uses centralized repository path conventions.
func GetAssessmentResultsPath(workspaceRoot, moduleName string) string {
	return repository.RiskAssessmentResultsPath(workspaceRoot, moduleName)
}

// LoadAllAssessmentResults loads all assessment-results files for aggregated reporting.
func LoadAllAssessmentResults(workspaceRoot string) ([]*oscalTypes.AssessmentResults, error) {
	arMap, err := DiscoverAssessmentResults(workspaceRoot)
	if err != nil {
		return nil, err
	}

	if len(arMap) == 0 {
		return nil, fmt.Errorf("no assessment-results found in %s/%s/", repository.OutDir, repository.RiskDir)
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
		return nil, fmt.Errorf("no profiles found in %s/%s/", repository.SpecsDir, repository.RiskControlsDir)
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
