package validate

import (
	"os"
	"path/filepath"
	"testing"
)

// TestValidateDocs_ObsoleteReferences tests detection of obsolete file references
func TestValidateDocs_ObsoleteReferences(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	docsDir := filepath.Join(tempDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("failed to create docs dir: %v", err)
	}

	tests := []struct {
		name           string
		content        string
		obsoleteFiles  []string
		expectedErrors int
	}{
		{
			name: "no obsolete references",
			content: `# Configuration
Configuration lives in repository.yml.
`,
			obsoleteFiles:  []string{"module-types.yml", "system-dependencies.yml"},
			expectedErrors: 0,
		},
		{
			name: "single obsolete reference",
			content: `# Configuration
Configuration lives in .eac/module-types.yml.
`,
			obsoleteFiles:  []string{"module-types.yml", "system-dependencies.yml"},
			expectedErrors: 1,
		},
		{
			name: "multiple obsolete references",
			content: `# Configuration
Types are in module-types.yml.
Dependencies are in system-dependencies.yml.
`,
			obsoleteFiles:  []string{"module-types.yml", "system-dependencies.yml"},
			expectedErrors: 2,
		},
		{
			name:           "reference in code block should still be flagged",
			content:        "# Example\n```yaml\n# .eac/module-types.yml\ntypes:\n  - name: test\n```\n",
			obsoleteFiles:  []string{"module-types.yml"},
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write test file
			testFile := filepath.Join(docsDir, "test.md")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}
			defer os.Remove(testFile)

			// Create validator
			validator := NewDocsValidator(DocsValidatorOptions{
				ObsoleteFiles: tt.obsoleteFiles,
			})

			// Validate
			results, err := validator.ValidateFile(testFile)
			if err != nil {
				t.Fatalf("validation failed: %v", err)
			}

			if len(results) != tt.expectedErrors {
				t.Errorf("expected %d errors, got %d: %v", tt.expectedErrors, len(results), results)
			}
		})
	}
}

// TestValidateDocs_ValidateDirectory tests directory validation
func TestValidateDocs_ValidateDirectory(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	docsDir := filepath.Join(tempDir, "docs")
	subDir := filepath.Join(docsDir, "reference")

	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	// Create files with and without issues
	cleanFile := filepath.Join(docsDir, "clean.md")
	if err := os.WriteFile(cleanFile, []byte("# Clean\nNo issues here."), 0644); err != nil {
		t.Fatalf("failed to write clean file: %v", err)
	}

	issueFile := filepath.Join(subDir, "issue.md")
	if err := os.WriteFile(issueFile, []byte("# Issue\nReferences module-types.yml."), 0644); err != nil {
		t.Fatalf("failed to write issue file: %v", err)
	}

	// Validate directory
	validator := NewDocsValidator(DocsValidatorOptions{
		ObsoleteFiles: []string{"module-types.yml"},
	})

	results, err := validator.ValidateDirectory(docsDir)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	// Should find 1 issue in the subdirectory file
	totalIssues := 0
	for _, fileResults := range results {
		totalIssues += len(fileResults)
	}

	if totalIssues != 1 {
		t.Errorf("expected 1 issue, got %d", totalIssues)
	}
}

// TestDefaultObsoleteFiles tests that default obsolete files are configured
func TestDefaultObsoleteFiles(t *testing.T) {
	opts := DefaultDocsValidatorOptions()

	// Should include the known obsolete files
	expectedFiles := []string{"module-types.yml", "system-dependencies.yml"}
	for _, expected := range expectedFiles {
		found := false
		for _, actual := range opts.ObsoleteFiles {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected obsolete file %q not found in defaults", expected)
		}
	}
}
