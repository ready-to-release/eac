package formats

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func getTestExport() *ManualTestExport {
	return &ManualTestExport{
		ExportMetadata: ExportMetadata{
			ExportTime:     time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Module:         "test-module",
			ReleaseVersion: "v1.0.0",
			GitCommit:      "abc123def456",
			SchemaVersion:  "1.0",
		},
		Scenarios: []ExportedScenario{
			{
				ScenarioID:   "test-module:feature-1:scenario-1",
				FeatureName:  "Test Feature",
				ScenarioName: "Test Scenario 1",
				Tags:         []string{"@Manual", "@P0"},
				Steps: []string{
					"Given a test condition",
					"When an action is performed",
					"Then a result is observed",
				},
				Description: "Test description",
				FilePath:    "specs/test-module/feature-1.feature",
			},
		},
	}
}

func TestJSONFormatter_Write(t *testing.T) {
	formatter := NewJSONFormatter()
	export := getTestExport()

	var buf bytes.Buffer
	err := formatter.Write(export, &buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify it's valid JSON
	var result ManualTestExport
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Verify content
	if result.ExportMetadata.Module != "test-module" {
		t.Errorf("Module: got %q, want %q", result.ExportMetadata.Module, "test-module")
	}
	if len(result.Scenarios) != 1 {
		t.Errorf("Scenarios: got %d, want 1", len(result.Scenarios))
	}
}

func TestJSONFormatter_FileExtension(t *testing.T) {
	formatter := NewJSONFormatter()
	if formatter.FileExtension() != "json" {
		t.Errorf("FileExtension: got %q, want %q", formatter.FileExtension(), "json")
	}
}

func TestCSVFormatter_Write(t *testing.T) {
	formatter := NewCSVFormatter()
	export := getTestExport()

	var buf bytes.Buffer
	err := formatter.Write(export, &buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	// Verify header
	if !strings.Contains(output, "scenario_id,feature_name,scenario_name") {
		t.Error("CSV should contain header row")
	}

	// Verify data
	if !strings.Contains(output, "test-module:feature-1:scenario-1") {
		t.Error("CSV should contain scenario ID")
	}
	if !strings.Contains(output, "Test Feature") {
		t.Error("CSV should contain feature name")
	}
	if !strings.Contains(output, "@Manual @P0") {
		t.Error("CSV should contain tags")
	}
}

func TestCSVFormatter_FileExtension(t *testing.T) {
	formatter := NewCSVFormatter()
	if formatter.FileExtension() != "csv" {
		t.Errorf("FileExtension: got %q, want %q", formatter.FileExtension(), "csv")
	}
}

func TestMarkdownFormatter_Write(t *testing.T) {
	formatter := NewMarkdownFormatter()
	export := getTestExport()

	var buf bytes.Buffer
	err := formatter.Write(export, &buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	// Verify markdown structure
	if !strings.Contains(output, "# Manual Test Scenarios") {
		t.Error("Markdown should contain title")
	}
	if !strings.Contains(output, "## 1. Test Scenario 1") {
		t.Error("Markdown should contain scenario heading")
	}
	if !strings.Contains(output, "test-module:feature-1:scenario-1") {
		t.Error("Markdown should contain scenario ID")
	}
	if !strings.Contains(output, "Test Feature") {
		t.Error("Markdown should contain feature name")
	}
	if !strings.Contains(output, "Given a test condition") {
		t.Error("Markdown should contain steps")
	}
}

func TestMarkdownFormatter_FileExtension(t *testing.T) {
	formatter := NewMarkdownFormatter()
	if formatter.FileExtension() != "md" {
		t.Errorf("FileExtension: got %q, want %q", formatter.FileExtension(), "md")
	}
}

func TestGetFormatter_JSON(t *testing.T) {
	formatter, err := GetFormatter("json")
	if err != nil {
		t.Fatalf("GetFormatter failed: %v", err)
	}
	if _, ok := formatter.(*JSONFormatter); !ok {
		t.Error("Expected JSONFormatter")
	}
}

func TestGetFormatter_CSV(t *testing.T) {
	formatter, err := GetFormatter("csv")
	if err != nil {
		t.Fatalf("GetFormatter failed: %v", err)
	}
	if _, ok := formatter.(*CSVFormatter); !ok {
		t.Error("Expected CSVFormatter")
	}
}

func TestGetFormatter_Markdown(t *testing.T) {
	testCases := []string{"markdown", "md", "Markdown", "MD"}
	for _, format := range testCases {
		t.Run(format, func(t *testing.T) {
			formatter, err := GetFormatter(format)
			if err != nil {
				t.Fatalf("GetFormatter(%q) failed: %v", format, err)
			}
			if _, ok := formatter.(*MarkdownFormatter); !ok {
				t.Errorf("GetFormatter(%q): expected MarkdownFormatter", format)
			}
		})
	}
}

func TestGetFormatter_Unsupported(t *testing.T) {
	_, err := GetFormatter("xml")
	if err == nil {
		t.Fatal("Expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Error message should mention unsupported format, got: %v", err)
	}
}
