package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCountLintIssues(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "empty file returns zero issues",
			content:   "",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "no issues in JSON",
			content:   `{"Issues": []}`,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "single issue",
			content:   `{"Issues": [{"Text": "unused variable"}]}`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "multiple issues",
			content: `{"Issues": [
				{"Text": "unused variable"},
				{"Text": "missing return"},
				{"Text": "shadowed variable"}
			]}`,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "null issues field",
			content:   `{"Issues": null}`,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "missing issues field",
			content:   `{"Report": {}}`,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "invalid JSON returns error",
			content:   `{not valid json`,
			wantCount: 0,
			wantErr:   true,
		},
		{
			name: "issues with extra fields",
			content: `{
				"Issues": [
					{"Text": "err1", "Severity": "warning", "Pos": {"Line": 10}},
					{"Text": "err2", "Severity": "error", "Pos": {"Line": 20}}
				],
				"Report": {"Linters": []}
			}`,
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			jsonPath := filepath.Join(tmpDir, "lint.json")

			if err := os.WriteFile(jsonPath, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			got, err := countLintIssues(jsonPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("countLintIssues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantCount {
				t.Errorf("countLintIssues() = %d, want %d", got, tt.wantCount)
			}
		})
	}
}

func TestCountLintIssues_NonexistentFile(t *testing.T) {
	_, err := countLintIssues("/nonexistent/path/lint.json")
	if err == nil {
		t.Error("countLintIssues() with nonexistent file should return error")
	}
}
