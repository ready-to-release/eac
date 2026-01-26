// Package srccommands contains godog step implementations for specs/eac-commands.
//
// This file contains security command step definitions.
// Features: specs/eac-commands/security/
//
// This implements step definitions for security scanner features including:
// - Evidence file verification
// - JSON schema validation
// - Log file checks
// - Security scanner output validation
package srccommands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// evidenceFile represents the structure of a security evidence file.
type evidenceFile struct {
	Module    string          `json:"module"`
	Scanner   string          `json:"scanner"`
	Timestamp string          `json:"timestamp"`
	SHA256    string          `json:"sha256"`
	Findings  json.RawMessage `json:"findings"`
}

// readEvidenceFile reads and parses an evidence file.
func readEvidenceFile(filePath string) (*evidenceFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read evidence file: %w", err)
	}

	var evidence evidenceFile
	if err := json.Unmarshal(data, &evidence); err != nil {
		return nil, fmt.Errorf("failed to parse evidence file: %w", err)
	}

	return &evidence, nil
}

// securityTestState holds state for security tests.
type securityTestState struct {
	lastCheckedDirectory string
}

// registerSecuritySteps registers step definitions for security command features.
func registerSecuritySteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	state := &securityTestState{}

	// Evidence file steps
	sc.Step(`^evidence files should exist in directory "([^"]*)"$`, func(directory string) error {
		return evidenceFilesExistInDirectory(directory, ctx, state)
	})
	sc.Step(`^the latest evidence file should have JSON field "([^"]*)" with value "([^"]*)"$`, func(field, value string) error {
		return theLatestEvidenceFileHasJSONField(field, value, ctx, state)
	})
	sc.Step(`^the latest evidence file should have JSON field "([^"]*)" matching (.+)$`, func(field, format string) error {
		return theLatestEvidenceFileHasJSONFieldMatchingFormat(field, format, ctx, state)
	})
	sc.Step(`^the latest evidence file should have JSON field "([^"]*)" with (\d+) character (.+)$`, func(field string, length int, hashType string) error {
		return theLatestEvidenceFileHasJSONFieldWithCharacterHash(field, length, hashType, ctx, state)
	})
	sc.Step(`^the latest evidence file should have JSON field "([^"]*)" with non-empty data$`, func(field string) error {
		return theLatestEvidenceFileHasJSONFieldWithNonEmptyData(field, ctx, state)
	})

	// Log file steps
	sc.Step(`^a log file should exist in directory "([^"]*)"$`, func(directory string) error {
		return aLogFileExistsInDirectory(directory, ctx)
	})
}

// ============================================================================
// Evidence File Verification Steps
// ============================================================================

// evidenceFilesExistInDirectory checks that evidence files exist in the specified directory.
func evidenceFilesExistInDirectory(directory string, ctx *internal.TestContext, state *securityTestState) error {
	// Save the directory for use in subsequent steps
	state.lastCheckedDirectory = directory

	// Use isolated test directory if available, otherwise use repository root
	workspaceRoot := ctx.IsolatedDir
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

// getLatestEvidenceFile finds the most recent evidence file in a directory.
func getLatestEvidenceFile(directory string, ctx *internal.TestContext) (string, error) {
	// Use isolated test directory if available, otherwise use repository root
	workspaceRoot := ctx.IsolatedDir
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

// theLatestEvidenceFileHasJSONField checks that the latest evidence file has a specific JSON field with value.
func theLatestEvidenceFileHasJSONField(field, value string, ctx *internal.TestContext, state *securityTestState) error {
	latestFile, err := getLatestEvidenceFile(state.lastCheckedDirectory, ctx)
	if err != nil {
		return err
	}

	// Read the evidence file
	evidence, err := readEvidenceFile(latestFile)
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

// theLatestEvidenceFileHasJSONFieldMatchingFormat checks field matches a format.
func theLatestEvidenceFileHasJSONFieldMatchingFormat(field, format string, ctx *internal.TestContext, state *securityTestState) error {
	latestFile, err := getLatestEvidenceFile(state.lastCheckedDirectory, ctx)
	if err != nil {
		return err
	}

	evidence, err := readEvidenceFile(latestFile)
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

// theLatestEvidenceFileHasJSONFieldWithCharacterHash checks hash field length.
func theLatestEvidenceFileHasJSONFieldWithCharacterHash(field string, length int, hashType string, ctx *internal.TestContext, state *securityTestState) error {
	latestFile, err := getLatestEvidenceFile(state.lastCheckedDirectory, ctx)
	if err != nil {
		return err
	}

	evidence, err := readEvidenceFile(latestFile)
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

// theLatestEvidenceFileHasJSONFieldWithNonEmptyData checks findings is not empty.
func theLatestEvidenceFileHasJSONFieldWithNonEmptyData(field string, ctx *internal.TestContext, state *securityTestState) error {
	latestFile, err := getLatestEvidenceFile(state.lastCheckedDirectory, ctx)
	if err != nil {
		return err
	}

	evidence, err := readEvidenceFile(latestFile)
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

// aLogFileExistsInDirectory checks that a log file exists in the specified directory.
func aLogFileExistsInDirectory(directory string, ctx *internal.TestContext) error {
	// Use isolated test directory if available, otherwise use repository root
	workspaceRoot := ctx.IsolatedDir
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
