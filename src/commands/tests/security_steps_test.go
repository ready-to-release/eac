// Godog BDD step definitions - Security command steps
//
// Features: specs/src-commands/security/
//
// This file implements step definitions for security scanner features including:
// - Evidence file verification
// - JSON schema validation
// - Log file checks
// - Security scanner output validation
package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/commands/impl/security"
	"github.com/ready-to-release/eac/src/core/repository"
)

// ============================================================================
// Evidence File Verification Steps
// ============================================================================

// lastCheckedDirectory tracks the last directory checked for evidence files
var lastCheckedDirectory string

// evidenceFilesExistInDirectory checks that evidence files exist in the specified directory
func evidenceFilesExistInDirectory(directory string) error {
	// Save the directory for use in subsequent steps
	lastCheckedDirectory = directory

	// Use isolated test directory if available, otherwise use repository root
	workspaceRoot := isolatedTestProjectDir
	if workspaceRoot == "" {
		root, err := repository.GetRepositoryRoot("")
		if err != nil {
			return fmt.Errorf("failed to get workspace root: %w", err)
		}
		workspaceRoot = root
	}

	fullPath := filepath.Join(workspaceRoot, directory)

	// Check if directory exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", fullPath)
	}

	// Check for JSON files in directory
	files, err := filepath.Glob(filepath.Join(fullPath, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to glob files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no evidence files found in directory: %s", fullPath)
	}

	return nil
}

// getLatestEvidenceFile finds the most recent evidence file in a directory
func getLatestEvidenceFile(directory string) (string, error) {
	// Use isolated test directory if available, otherwise use repository root
	workspaceRoot := isolatedTestProjectDir
	if workspaceRoot == "" {
		root, err := repository.GetRepositoryRoot("")
		if err != nil {
			return "", fmt.Errorf("failed to get workspace root: %w", err)
		}
		workspaceRoot = root
	}

	fullPath := filepath.Join(workspaceRoot, directory)

	files, err := filepath.Glob(filepath.Join(fullPath, "*.json"))
	if err != nil {
		return "", fmt.Errorf("failed to glob files: %w", err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no evidence files found in directory: %s", fullPath)
	}

	// Find most recent file by modification time
	var latestFile string
	var latestTime time.Time

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		if latestFile == "" || info.ModTime().After(latestTime) {
			latestFile = file
			latestTime = info.ModTime()
		}
	}

	return latestFile, nil
}

// theLatestEvidenceFileHasJSONField checks that the latest evidence file has a specific JSON field with value
func theLatestEvidenceFileHasJSONField(field, value string) error {
	latestFile, err := getLatestEvidenceFile(lastCheckedDirectory)
	if err != nil {
		return err
	}

	// Read the evidence file
	evidence, err := security.ReadEvidence(latestFile)
	if err != nil {
		return fmt.Errorf("failed to read evidence file: %w", err)
	}

	// Check the field
	switch field {
	case "module":
		if evidence.Module != value {
			return fmt.Errorf("module field = %s, want %s", evidence.Module, value)
		}
	case "scanner":
		if evidence.Scanner != value {
			return fmt.Errorf("scanner field = %s, want %s", evidence.Scanner, value)
		}
	default:
		return fmt.Errorf("unknown field: %s", field)
	}

	return nil
}

// theLatestEvidenceFileHasJSONFieldMatchingFormat checks field matches a format
func theLatestEvidenceFileHasJSONFieldMatchingFormat(field, format string) error {
	latestFile, err := getLatestEvidenceFile(lastCheckedDirectory)
	if err != nil {
		return err
	}

	evidence, err := security.ReadEvidence(latestFile)
	if err != nil {
		return fmt.Errorf("failed to read evidence file: %w", err)
	}

	switch field {
	case "timestamp":
		if format == "RFC3339 format" {
			// Check if timestamp contains T and Z (RFC3339 markers)
			if !strings.Contains(evidence.Timestamp, "T") || !strings.Contains(evidence.Timestamp, "Z") {
				return fmt.Errorf("timestamp does not appear to be RFC3339 format: %s", evidence.Timestamp)
			}
		}
	default:
		return fmt.Errorf("unknown field: %s", field)
	}

	return nil
}

// theLatestEvidenceFileHasJSONFieldWithCharacterHash checks hash field length
func theLatestEvidenceFileHasJSONFieldWithCharacterHash(field string, length int, hashType string) error {
	latestFile, err := getLatestEvidenceFile(lastCheckedDirectory)
	if err != nil {
		return err
	}

	evidence, err := security.ReadEvidence(latestFile)
	if err != nil {
		return fmt.Errorf("failed to read evidence file: %w", err)
	}

	switch field {
	case "sha256":
		if len(evidence.SHA256) != length {
			return fmt.Errorf("sha256 length = %d, want %d", len(evidence.SHA256), length)
		}
		// Verify it's hex
		for _, c := range evidence.SHA256 {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return fmt.Errorf("sha256 contains non-hex character: %c", c)
			}
		}
	default:
		return fmt.Errorf("unknown field: %s", field)
	}

	return nil
}

// theLatestEvidenceFileHasJSONFieldWithNonEmptyData checks findings is not empty
func theLatestEvidenceFileHasJSONFieldWithNonEmptyData(field string) error {
	latestFile, err := getLatestEvidenceFile(lastCheckedDirectory)
	if err != nil {
		return err
	}

	evidence, err := security.ReadEvidence(latestFile)
	if err != nil {
		return fmt.Errorf("failed to read evidence file: %w", err)
	}

	switch field {
	case "findings":
		if len(evidence.Findings) == 0 {
			return fmt.Errorf("findings field is empty")
		}
		// Verify it's valid JSON
		var temp interface{}
		if err := json.Unmarshal(evidence.Findings, &temp); err != nil {
			return fmt.Errorf("findings is not valid JSON: %w", err)
		}
	default:
		return fmt.Errorf("unknown field: %s", field)
	}

	return nil
}

// ============================================================================
// Log File Verification Steps
// ============================================================================

// aLogFileExistsInDirectory checks that a log file exists in the specified directory
func aLogFileExistsInDirectory(directory string) error {
	// Use isolated test directory if available, otherwise use repository root
	workspaceRoot := isolatedTestProjectDir
	if workspaceRoot == "" {
		root, err := repository.GetRepositoryRoot("")
		if err != nil {
			return fmt.Errorf("failed to get workspace root: %w", err)
		}
		workspaceRoot = root
	}

	fullPath := filepath.Join(workspaceRoot, directory)

	// Check if directory exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", fullPath)
	}

	// Check for log files in directory
	files, err := filepath.Glob(filepath.Join(fullPath, "*.log"))
	if err != nil {
		return fmt.Errorf("failed to glob log files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no log files found in directory: %s", fullPath)
	}

	return nil
}

// ============================================================================
// Mock Setup
// ============================================================================

// setupSecurityMocks sets up mocked security tool outputs for testing
func setupSecurityMocks() error {
	// Set in-process mocks for security scanners
	// This uses the mock variables in each scanner package (mockTrivyOutput, etc.)
	security.SetupMocks()
	return nil
}

// ============================================================================
// Step Registration
// ============================================================================

// initializeSecuritySteps registers all security-specific step definitions
func initializeSecuritySteps(sc *godog.ScenarioContext) {
	// Evidence file steps
	sc.Step(`^evidence files should exist in directory "([^"]*)"$`, evidenceFilesExistInDirectory)
	sc.Step(`^the latest evidence file should have JSON field "([^"]*)" with value "([^"]*)"$`, theLatestEvidenceFileHasJSONField)
	sc.Step(`^the latest evidence file should have JSON field "([^"]*)" matching (.+)$`, theLatestEvidenceFileHasJSONFieldMatchingFormat)
	sc.Step(`^the latest evidence file should have JSON field "([^"]*)" with (\d+) character (.+)$`, theLatestEvidenceFileHasJSONFieldWithCharacterHash)
	sc.Step(`^the latest evidence file should have JSON field "([^"]*)" with non-empty data$`, theLatestEvidenceFileHasJSONFieldWithNonEmptyData)

	// Log file steps
	sc.Step(`^a log file should exist in directory "([^"]*)"$`, aLogFileExistsInDirectory)
}
