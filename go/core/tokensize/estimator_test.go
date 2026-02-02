package tokensize

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEstimateContent(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantTokens int
		wantChars  int
		wantLines  int
		wantBytes  int64
	}{
		{
			name:       "empty file",
			content:    "",
			wantTokens: 0,
			wantChars:  0,
			wantLines:  0,
			wantBytes:  0,
		},
		{
			name:       "simple content",
			content:    "hello world",
			wantTokens: 2, // 11 chars / 4 = 2
			wantChars:  11,
			wantLines:  1,
			wantBytes:  11,
		},
		{
			name:       "multiple lines with newline",
			content:    "line1\nline2\nline3\n",
			wantTokens: 4, // 18 chars / 4 = 4
			wantChars:  18,
			wantLines:  3,
			wantBytes:  18,
		},
		{
			name:       "multiple lines without trailing newline",
			content:    "line1\nline2\nline3",
			wantTokens: 4, // 17 chars / 4 = 4
			wantChars:  17,
			wantLines:  3,
			wantBytes:  17,
		},
		{
			name:       "typical code snippet",
			content:    "func main() {\n\tfmt.Println(\"Hello\")\n}\n",
			wantTokens: 9, // 38 chars / 4 = 9
			wantChars:  38,
			wantLines:  3,
			wantBytes:  38,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateContent("test.go", []byte(tt.content))

			if result.Tokens != tt.wantTokens {
				t.Errorf("Tokens = %d, want %d", result.Tokens, tt.wantTokens)
			}
			if result.Characters != tt.wantChars {
				t.Errorf("Characters = %d, want %d", result.Characters, tt.wantChars)
			}
			if result.Lines != tt.wantLines {
				t.Errorf("Lines = %d, want %d", result.Lines, tt.wantLines)
			}
			if result.Bytes != tt.wantBytes {
				t.Errorf("Bytes = %d, want %d", result.Bytes, tt.wantBytes)
			}
			if result.Method != MethodCharDiv4 {
				t.Errorf("Method = %s, want %s", result.Method, MethodCharDiv4)
			}
			if result.FilePath != "test.go" {
				t.Errorf("FilePath = %s, want test.go", result.FilePath)
			}
		})
	}
}

func TestEstimateFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := "package main\n\nfunc main() {}\n"

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := EstimateFile(testFile)
	if err != nil {
		t.Fatalf("EstimateFile failed: %v", err)
	}

	if result.Characters != 29 {
		t.Errorf("Characters = %d, want 29", result.Characters)
	}
	if result.Tokens != 7 { // 29 / 4 = 7
		t.Errorf("Tokens = %d, want 7", result.Tokens)
	}
	if result.Lines != 3 {
		t.Errorf("Lines = %d, want 3", result.Lines)
	}
}

func TestEstimateFile_NotFound(t *testing.T) {
	_, err := EstimateFile("/nonexistent/file.go")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestExpandGlobPatterns(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create test files
	files := []string{
		filepath.Join(tmpDir, "main.go"),
		filepath.Join(tmpDir, "util.go"),
		filepath.Join(subDir, "helper.go"),
		filepath.Join(tmpDir, "readme.md"),
	}

	for _, f := range files {
		if err := os.WriteFile(f, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", f, err)
		}
	}

	tests := []struct {
		name     string
		patterns []string
		wantLen  int
	}{
		{
			name:     "single file pattern",
			patterns: []string{"main.go"},
			wantLen:  1,
		},
		{
			name:     "wildcard pattern",
			patterns: []string{"*.go"},
			wantLen:  2,
		},
		{
			name:     "doublestar pattern",
			patterns: []string{"**/*.go"},
			wantLen:  3,
		},
		{
			name:     "multiple patterns",
			patterns: []string{"*.go", "*.md"},
			wantLen:  3,
		},
		{
			name:     "no match",
			patterns: []string{"*.txt"},
			wantLen:  0,
		},
		{
			name:     "empty patterns",
			patterns: []string{},
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandGlobPatterns(tmpDir, tt.patterns)
			if err != nil {
				t.Fatalf("ExpandGlobPatterns failed: %v", err)
			}
			if len(result) != tt.wantLen {
				t.Errorf("got %d files, want %d", len(result), tt.wantLen)
			}
		})
	}
}
