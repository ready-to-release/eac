// File: src/commands/impl/specs/validate/validate_test.go
package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/src/core/contracts"
)

// Intent: Test specification validation command core functionality
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Table-driven tests with clear test case names
//   - Each test focuses on a single behavior
//
// Easy to change:
//   - Test data is clearly separated from test logic
//   - Reusable test fixtures and helpers
//
// Hard to break:
//   - Tests cover happy path and error cases
//   - File system operations use t.TempDir() for isolation
//   - Comprehensive edge case coverage

func TestValidateGherkinFile_ValidFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErrs int
	}{
		{
			name: "valid specification with all required elements",
			content: `@deps:go @ov
Feature: src-commands_test-feature

  As a developer
  I want to test validation
  So that I can ensure quality

  Rule: Test rule

    @L2 @ov
    Scenario: Test scenario
      Given a precondition
      When an action occurs
      Then a result is expected
`,
			wantErrs: 0,
		},
		{
			name: "valid specification with multiple rules and scenarios",
			content: `@deps:go @ov
Feature: src-commands_multi-test

  As a developer
  I want multiple scenarios
  So that I can test thoroughly

  Rule: First rule

    @L2 @ov
    Scenario: First scenario
      Given precondition one
      When action one
      Then result one

    @L2 @iv
    Scenario: Second scenario
      Given precondition two
      When action two
      Then result two

  Rule: Second rule

    @L2 @pv
    Scenario: Third scenario
      Given precondition three
      When action three
      Then result three
`,
			wantErrs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			setupContractFiles(t, tmpDir)

			// Create test file
			testFile := filepath.Join(tmpDir, "specs", "test.feature")
			if err := os.MkdirAll(filepath.Dir(testFile), 0755); err != nil {
				t.Fatalf("failed to create test directory: %v", err)
			}
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			// Run validation
			errors, err := validateGherkinFile(testFile, tmpDir)
			if err != nil {
				t.Fatalf("validateGherkinFile() unexpected error: %v", err)
			}

			// Count critical errors
			criticalCount := contracts.CountCriticalErrors(errors)
			if criticalCount != tt.wantErrs {
				t.Errorf("validateGherkinFile() got %d critical errors, want %d", criticalCount, tt.wantErrs)
				if len(errors) > 0 {
					t.Logf("Errors: %s", contracts.FormatValidationErrors(errors))
				}
			}
		})
	}
}

func TestValidateGherkinFile_InvalidFile(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantErrorCode string
	}{
		{
			name: "missing feature declaration",
			content: `@deps:go @ov
Rule: Test rule

  @L2 @ov
  Scenario: Test scenario
    Given a precondition
`,
			wantErrorCode: "MISSING_FEATURE",
		},
		{
			name: "missing rule declaration",
			content: `@deps:go @ov
Feature: src-commands_test-feature

  As a developer
  I want to test validation
  So that I can ensure quality

  @L2 @ov
  Scenario: Test scenario
    Given a precondition
`,
			wantErrorCode: "MISSING_RULE",
		},
		{
			name: "missing scenario declaration",
			content: `@deps:go @ov
Feature: src-commands_test-feature

  As a developer
  I want to test validation
  So that I can ensure quality

  Rule: Test rule
`,
			wantErrorCode: "MISSING_SCENARIO",
		},
		{
			name: "invalid feature naming",
			content: `@deps:go @ov
Feature: InvalidName

  As a developer
  I want to test validation
  So that I can ensure quality

  Rule: Test rule

    @L2 @ov
    Scenario: Test scenario
      Given a precondition
`,
			wantErrorCode: "INVALID_FEATURE_NAMING",
		},
		{
			name: "missing verification tag",
			content: `@deps:go @ov
Feature: src-commands_test-feature

  As a developer
  I want to test validation
  So that I can ensure quality

  Rule: Test rule

    @L2
    Scenario: Test scenario
      Given a precondition
`,
			wantErrorCode: "MISSING_VERIFICATION_TAG",
		},
		{
			name: "rule before feature",
			content: `@deps:go @ov
Rule: Test rule

Feature: src-commands_test-feature

  @L2 @ov
  Scenario: Test scenario
    Given a precondition
`,
			wantErrorCode: "RULE_BEFORE_FEATURE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			setupContractFiles(t, tmpDir)

			// Create test file
			testFile := filepath.Join(tmpDir, "specs", "test.feature")
			if err := os.MkdirAll(filepath.Dir(testFile), 0755); err != nil {
				t.Fatalf("failed to create test directory: %v", err)
			}
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			// Run validation
			errors, err := validateGherkinFile(testFile, tmpDir)
			if err != nil {
				t.Fatalf("validateGherkinFile() unexpected error: %v", err)
			}

			// Check for expected error code
			found := false
			for _, e := range errors {
				if e.Code == tt.wantErrorCode {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("validateGherkinFile() missing expected error code %q", tt.wantErrorCode)
				t.Logf("Got errors: %s", contracts.FormatValidationErrors(errors))
			}
		})
	}
}

func TestValidateDirectory_AllValid(t *testing.T) {
	tmpDir := t.TempDir()
	setupContractFiles(t, tmpDir)

	// Create multiple valid files
	validContent := `@deps:go @ov
Feature: src-commands_test-feature

  As a developer
  I want to test validation
  So that I can ensure quality

  Rule: Test rule

    @L2 @ov
    Scenario: Test scenario
      Given a precondition
      When an action occurs
      Then a result is expected
`

	specsDir := filepath.Join(tmpDir, "specs")
	files := []string{
		"test1/spec.feature",
		"test2/spec.feature",
		"test3/spec.feature",
	}

	for _, f := range files {
		filePath := filepath.Join(specsDir, f)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		if err := os.WriteFile(filePath, []byte(validContent), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	// Run validation
	results, err := validateDirectory(specsDir, tmpDir, false)
	if err != nil {
		t.Fatalf("validateDirectory() error = %v", err)
	}

	// Check results
	if len(results) != len(files) {
		t.Errorf("validateDirectory() processed %d files, want %d", len(results), len(files))
	}

	for _, result := range results {
		if !result.Valid {
			t.Errorf("file %s failed validation: %v", result.Path, result.Errors)
		}
	}
}

func TestValidateDirectory_MixedResults(t *testing.T) {
	tmpDir := t.TempDir()
	setupContractFiles(t, tmpDir)

	specsDir := filepath.Join(tmpDir, "specs")

	// Create valid file
	validContent := `@deps:go @ov
Feature: src-commands_valid-test

  As a developer
  I want to test validation
  So that I can ensure quality

  Rule: Test rule

    @L2 @ov
    Scenario: Test scenario
      Given a precondition
      When an action occurs
      Then a result is expected
`

	validFile := filepath.Join(specsDir, "valid", "spec.feature")
	if err := os.MkdirAll(filepath.Dir(validFile), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(validFile, []byte(validContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Create invalid file (missing Rule)
	invalidContent := `@deps:go @ov
Feature: src-commands_invalid-test

  As a developer
  I want to test validation
  So that I can ensure quality

  @L2 @ov
  Scenario: Test scenario
    Given a precondition
`

	invalidFile := filepath.Join(specsDir, "invalid", "spec.feature")
	if err := os.MkdirAll(filepath.Dir(invalidFile), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(invalidFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Run validation
	results, err := validateDirectory(specsDir, tmpDir, false)
	if err != nil {
		t.Fatalf("validateDirectory() error = %v", err)
	}

	// Check results
	if len(results) != 2 {
		t.Fatalf("validateDirectory() processed %d files, want 2", len(results))
	}

	validCount := 0
	invalidCount := 0
	for _, result := range results {
		if result.Valid {
			validCount++
		} else {
			invalidCount++
		}
	}

	if validCount != 1 {
		t.Errorf("got %d valid files, want 1", validCount)
	}
	if invalidCount != 1 {
		t.Errorf("got %d invalid files, want 1", invalidCount)
	}
}

func TestValidateDirectory_SkipNonFeatureFiles(t *testing.T) {
	tmpDir := t.TempDir()
	setupContractFiles(t, tmpDir)

	specsDir := filepath.Join(tmpDir, "specs")

	// Create .feature file
	featureContent := `@deps:go @ov
Feature: src-commands_test-feature

  As a developer
  I want to test validation
  So that I can ensure quality

  Rule: Test rule

    @L2 @ov
    Scenario: Test scenario
      Given a precondition
`

	featureFile := filepath.Join(specsDir, "test.feature")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(featureFile, []byte(featureContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Create non-feature files
	mdFile := filepath.Join(specsDir, "README.md")
	if err := os.WriteFile(mdFile, []byte("# README"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	txtFile := filepath.Join(specsDir, "notes.txt")
	if err := os.WriteFile(txtFile, []byte("notes"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Run validation
	results, err := validateDirectory(specsDir, tmpDir, false)
	if err != nil {
		t.Fatalf("validateDirectory() error = %v", err)
	}

	// Should only process .feature file
	if len(results) != 1 {
		t.Errorf("validateDirectory() processed %d files, want 1", len(results))
	}

	if !strings.HasSuffix(results[0].Path, ".feature") {
		t.Errorf("validateDirectory() processed non-feature file: %s", results[0].Path)
	}
}

func TestValidatePath_Security(t *testing.T) {
	tmpDir := t.TempDir()

	// Use a path definitely outside tmpDir for the absolute path test
	// Get the parent directory of tmpDir and add a different subdirectory
	outsidePath := filepath.Join(filepath.Dir(filepath.Dir(tmpDir)), "outside-repo", "file.txt")

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid relative path",
			path:    "specs/test.feature",
			wantErr: false,
		},
		{
			name:    "path traversal attempt",
			path:    "../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "absolute path outside repo",
			path:    outsidePath,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path, tmpDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatValidationResult_SingleFile(t *testing.T) {
	tests := []struct {
		name   string
		result *ValidationResult
		want   []string // strings that should be in output
	}{
		{
			name: "valid file",
			result: &ValidationResult{
				Path:   "specs/test.feature",
				Valid:  true,
				Errors: []contracts.ValidationError{},
			},
			want: []string{"✅", "Validation passed", "specs/test.feature"},
		},
		{
			name: "invalid file with errors",
			result: &ValidationResult{
				Path:  "specs/test.feature",
				Valid: false,
				Errors: []contracts.ValidationError{
					{Code: "MISSING_RULE", Message: "Missing Rule: declaration", Severity: "error", Line: 0},
				},
			},
			want: []string{"❌", "Validation failed", "MISSING_RULE", "Missing Rule: declaration"},
		},
		{
			name: "file with warnings",
			result: &ValidationResult{
				Path:  "specs/test.feature",
				Valid: true,
				Errors: []contracts.ValidationError{
					{Code: "WARN_NAMING", Message: "Naming could be improved", Severity: "warning", Line: 5},
				},
			},
			want: []string{"✅", "Validation passed", "warning", "Line 5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := formatValidationResult(tt.result)

			for _, wantStr := range tt.want {
				if !strings.Contains(output, wantStr) {
					t.Errorf("formatValidationResult() output missing %q\nGot: %s", wantStr, output)
				}
			}
		})
	}
}

func TestFormatValidationSummary(t *testing.T) {
	tests := []struct {
		name    string
		results []*ValidationResult
		want    []string // strings that should be in output
	}{
		{
			name: "all passed",
			results: []*ValidationResult{
				{Path: "specs/test1.feature", Valid: true},
				{Path: "specs/test2.feature", Valid: true},
			},
			want: []string{"2 passed", "0 failed"},
		},
		{
			name: "mixed results",
			results: []*ValidationResult{
				{Path: "specs/test1.feature", Valid: true},
				{Path: "specs/test2.feature", Valid: false},
				{Path: "specs/test3.feature", Valid: false},
			},
			want: []string{"1 passed", "2 failed"},
		},
		{
			name:    "no files",
			results: []*ValidationResult{},
			want:    []string{"No specification files found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := formatValidationSummary(tt.results)

			for _, wantStr := range tt.want {
				if !strings.Contains(output, wantStr) {
					t.Errorf("formatValidationSummary() output missing %q\nGot: %s", wantStr, output)
				}
			}
		})
	}
}

// ==============================================================================
// parseValidateConfig Tests
// ==============================================================================

// withArgs temporarily sets os.Args for testing and restores it after
func withArgs(args []string, fn func()) {
	oldArgs := os.Args
	os.Args = args
	defer func() { os.Args = oldArgs }()
	fn()
}

func TestParseValidateConfig_Flags(t *testing.T) {
	// Create a temp directory to act as a git repository
	tmpDir := t.TempDir()

	// Create .git directory to simulate a git repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Create a test feature file
	specsDir := filepath.Join(tmpDir, "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs directory: %v", err)
	}
	testFile := filepath.Join(specsDir, "test.feature")
	if err := os.WriteFile(testFile, []byte("Feature: test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Change to temp directory so repository.GetRepositoryRoot works
	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	tests := []struct {
		name        string
		args        []string
		wantQuiet   bool
		wantVerbose bool
		wantFormat  string
		wantStrict  bool
		wantFix     bool
		wantErr     bool
		errContains string
	}{
		{
			name:       "simple path",
			args:       []string{"r2r", "specs", "validate", "specs/test.feature"},
			wantFormat: "text",
			wantErr:    false,
		},
		{
			name:      "quiet flag short",
			args:      []string{"r2r", "specs", "validate", "-q", "specs/test.feature"},
			wantQuiet: true,
			wantFormat: "text",
			wantErr:   false,
		},
		{
			name:      "quiet flag long",
			args:      []string{"r2r", "specs", "validate", "--quiet", "specs/test.feature"},
			wantQuiet: true,
			wantFormat: "text",
			wantErr:   false,
		},
		{
			name:        "verbose flag short",
			args:        []string{"r2r", "specs", "validate", "-v", "specs/test.feature"},
			wantVerbose: true,
			wantFormat:  "text",
			wantErr:     false,
		},
		{
			name:        "verbose flag long",
			args:        []string{"r2r", "specs", "validate", "--verbose", "specs/test.feature"},
			wantVerbose: true,
			wantFormat:  "text",
			wantErr:     false,
		},
		{
			name:       "format flag short json",
			args:       []string{"r2r", "specs", "validate", "-f", "json", "specs/test.feature"},
			wantFormat: "json",
			wantErr:    false,
		},
		{
			name:       "format flag long text",
			args:       []string{"r2r", "specs", "validate", "--format", "text", "specs/test.feature"},
			wantFormat: "text",
			wantErr:    false,
		},
		{
			name:       "strict flag",
			args:       []string{"r2r", "specs", "validate", "--strict", "specs/test.feature"},
			wantStrict: true,
			wantFormat: "text",
			wantErr:    false,
		},
		{
			name:       "fix flag",
			args:       []string{"r2r", "specs", "validate", "--fix", "specs/test.feature"},
			wantFix:    true,
			wantFormat: "text",
			wantErr:    false,
		},
		{
			name:        "multiple flags combined",
			args:        []string{"r2r", "specs", "validate", "-q", "-v", "-f", "json", "--strict", "specs/test.feature"},
			wantQuiet:   true,
			wantVerbose: true,
			wantFormat:  "json",
			wantStrict:  true,
			wantErr:     false,
		},
		{
			name:        "missing path",
			args:        []string{"r2r", "specs", "validate"},
			wantErr:     true,
			errContains: "file path or directory is required",
		},
		{
			name:        "invalid format",
			args:        []string{"r2r", "specs", "validate", "-f", "xml", "specs/test.feature"},
			wantErr:     true,
			errContains: "invalid format",
		},
		{
			name:        "format flag missing argument",
			args:        []string{"r2r", "specs", "validate", "-f"},
			wantErr:     true,
			errContains: "--format requires an argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config *ValidateConfig
			var err error

			withArgs(tt.args, func() {
				config, err = parseValidateConfig()
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("parseValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && err != nil {
					if !strings.Contains(err.Error(), tt.errContains) {
						t.Errorf("parseValidateConfig() error = %v, want error containing %q", err, tt.errContains)
					}
				}
				return
			}

			if config.Quiet != tt.wantQuiet {
				t.Errorf("parseValidateConfig() Quiet = %v, want %v", config.Quiet, tt.wantQuiet)
			}
			if config.Verbose != tt.wantVerbose {
				t.Errorf("parseValidateConfig() Verbose = %v, want %v", config.Verbose, tt.wantVerbose)
			}
			if config.Format != tt.wantFormat {
				t.Errorf("parseValidateConfig() Format = %q, want %q", config.Format, tt.wantFormat)
			}
			if config.Strict != tt.wantStrict {
				t.Errorf("parseValidateConfig() Strict = %v, want %v", config.Strict, tt.wantStrict)
			}
			if config.Fix != tt.wantFix {
				t.Errorf("parseValidateConfig() Fix = %v, want %v", config.Fix, tt.wantFix)
			}
		})
	}
}

func TestParseValidateConfig_CheckTags(t *testing.T) {
	// Create a temp directory to act as a git repository
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	specsDir := filepath.Join(tmpDir, "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs directory: %v", err)
	}
	testFile := filepath.Join(specsDir, "test.feature")
	if err := os.WriteFile(testFile, []byte("Feature: test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	tests := []struct {
		name          string
		args          []string
		wantCheckTags bool
	}{
		{
			name:          "default has check tags enabled",
			args:          []string{"r2r", "specs", "validate", "specs/test.feature"},
			wantCheckTags: true,
		},
		{
			name:          "explicit check-tags flag",
			args:          []string{"r2r", "specs", "validate", "--check-tags", "specs/test.feature"},
			wantCheckTags: true,
		},
		{
			name:          "no-check-tags disables",
			args:          []string{"r2r", "specs", "validate", "--no-check-tags", "specs/test.feature"},
			wantCheckTags: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config *ValidateConfig
			var err error

			withArgs(tt.args, func() {
				config, err = parseValidateConfig()
			})

			if err != nil {
				t.Fatalf("parseValidateConfig() unexpected error: %v", err)
			}

			if config.CheckTags != tt.wantCheckTags {
				t.Errorf("parseValidateConfig() CheckTags = %v, want %v", config.CheckTags, tt.wantCheckTags)
			}
		})
	}
}

func TestOutputJSON(t *testing.T) {
	tests := []struct {
		name    string
		results []*ValidationResult
		want    []string // strings that should be in JSON output
	}{
		{
			name: "single valid result",
			results: []*ValidationResult{
				{Path: "specs/test.feature", Valid: true, Errors: []contracts.ValidationError{}},
			},
			want: []string{`"valid": true`, `"passed": 1`, `"failed": 0`},
		},
		{
			name: "single invalid result",
			results: []*ValidationResult{
				{
					Path:  "specs/test.feature",
					Valid: false,
					Errors: []contracts.ValidationError{
						{Code: "MISSING_RULE", Message: "Missing Rule", Severity: "error"},
					},
				},
			},
			want: []string{`"valid": false`, `"passed": 0`, `"failed": 1`, `"MISSING_RULE"`},
		},
		{
			name: "mixed results",
			results: []*ValidationResult{
				{Path: "specs/valid.feature", Valid: true, Errors: []contracts.ValidationError{}},
				{Path: "specs/invalid.feature", Valid: false, Errors: []contracts.ValidationError{}},
			},
			want: []string{`"total": 2`, `"passed": 1`, `"failed": 1`},
		},
		{
			name:    "empty results",
			results: []*ValidationResult{},
			want:    []string{`"total": 0`, `"passed": 0`, `"failed": 0`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			outputJSON(tt.results)

			w.Close()
			os.Stdout = oldStdout

			buf := make([]byte, 4096)
			n, _ := r.Read(buf)
			output := string(buf[:n])

			for _, wantStr := range tt.want {
				if !strings.Contains(output, wantStr) {
					t.Errorf("outputJSON() output missing %q\nGot: %s", wantStr, output)
				}
			}
		})
	}
}

// Test helpers

// setupContractFiles creates minimal contract files needed for testing
func setupContractFiles(t *testing.T, tmpDir string) {
	t.Helper()

	contractDir := filepath.Join(tmpDir, "contracts", "ai", "specifications", "0.1.0")
	if err := os.MkdirAll(contractDir, 0755); err != nil {
		t.Fatalf("failed to create contract directory: %v", err)
	}

	// Create contract.yml
	contractYAML := `name: Gherkin Specification Structure
version: 0.1.0
description: Contract for Gherkin specification files

structure:
  feature:
    required: true
    naming_pattern: "^[a-z][a-z0-9-]*_[a-z][a-z0-9-]*$"

  rule:
    required: true
    description: "acceptance criteria"

  scenario:
    required: true
    description: "behavior examples"
    required_tags:
      - "@ov"
      - "@iv"
      - "@pv"
      - "@piv"
      - "@ppv"
`

	contractFile := filepath.Join(contractDir, "contract.yml")
	if err := os.WriteFile(contractFile, []byte(contractYAML), 0644); err != nil {
		t.Fatalf("failed to write contract.yml: %v", err)
	}

	// Create anti-corruption.yml
	antiCorruptionYAML := "version: 0.1.0\n" +
		"description: Anti-corruption layer rules for specifications\n" +
		"\n" +
		"remove_patterns:\n" +
		"  - pattern: '^\\s*```\\w*\\s*$'\n" +
		"    description: \"Remove markdown code fences\"\n" +
		"\n" +
		"  - pattern: '^\\s*---+\\s*$'\n" +
		"    description: \"Remove horizontal rules\"\n" +
		"\n" +
		"remove_prefixes:\n" +
		"  - \"**Initialized\"\n" +
		"  - \"I'll help\"\n" +
		"  - \"Here is\"\n" +
		"  - \"Here's\"\n" +
		"\n" +
		"content_start_marker: \"Feature:\"\n"

	antiCorruptionFile := filepath.Join(contractDir, "anti-corruption.yml")
	if err := os.WriteFile(antiCorruptionFile, []byte(antiCorruptionYAML), 0644); err != nil {
		t.Fatalf("failed to write anti-corruption.yml: %v", err)
	}
}
