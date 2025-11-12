package render

import (
	"strings"
	"testing"
)

func TestSimpleMarkdownTable(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	rows := [][]interface{}{
		{"Alice", 30, "NYC"},
		{"Bob", 25, "LA"},
		{"Charlie", 35, "Chicago"},
	}

	result := SimpleMarkdownTable(headers, rows)

	// Verify the result contains headers
	if !strings.Contains(result, "Name") || !strings.Contains(result, "Age") || !strings.Contains(result, "City") {
		t.Errorf("Table missing headers. Got:\n%s", result)
	}

	// Verify the result contains data
	if !strings.Contains(result, "Alice") || !strings.Contains(result, "Bob") || !strings.Contains(result, "Charlie") {
		t.Errorf("Table missing data rows. Got:\n%s", result)
	}

	// Verify markdown table structure (pipes)
	if !strings.Contains(result, "|") {
		t.Errorf("Result doesn't look like a markdown table. Got:\n%s", result)
	}
}

func TestTableBuilderWithAutoIndex(t *testing.T) {
	result := NewTableBuilder().
		WithHeaders("Module", "Type").
		WithAutoIndex().
		AddRow("cli", "application").
		AddRow("contracts", "library").
		AddRow("mcp", "server").
		Build()

	// Verify index column exists
	if !strings.Contains(result, "#") {
		t.Errorf("Auto-index column missing. Got:\n%s", result)
	}

	// Verify data is present
	if !strings.Contains(result, "cli") || !strings.Contains(result, "contracts") {
		t.Errorf("Table missing data. Got:\n%s", result)
	}
}

func TestTableBuilderWithFooter(t *testing.T) {
	result := NewTableBuilder().
		WithHeaders("Item", "Price").
		AddRow("Coffee", 5.00).
		AddRow("Tea", 3.50).
		AddRow("Cake", 7.00).
		WithFooter("Total", 15.50).
		Build()

	// Verify footer is present
	if !strings.Contains(result, "Total") || !strings.Contains(result, "15.5") {
		t.Errorf("Footer missing or incorrect. Got:\n%s", result)
	}
}

func TestRenderMarkdownTable(t *testing.T) {
	config := &MarkdownTableConfig{
		Headers: []string{"ID", "Name", "Status"},
		Rows: [][]interface{}{
			{1, "Task A", "Done"},
			{2, "Task B", "In Progress"},
		},
	}

	result := RenderMarkdownTable(config)

	if result == "" {
		t.Error("RenderMarkdownTable returned empty string")
	}

	// Check for markdown table structure
	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 lines (header, separator, data), got %d", len(lines))
	}
}

func TestRenderKeyValueTable(t *testing.T) {
	data := map[string]interface{}{
		"Name":    "TestProject",
		"Version": "1.0.0",
		"Status":  "Active",
	}

	result := RenderKeyValueTable("Property", "Value", data)

	// Verify headers
	if !strings.Contains(result, "Property") || !strings.Contains(result, "Value") {
		t.Errorf("Key-value table missing headers. Got:\n%s", result)
	}

	// Verify data
	if !strings.Contains(result, "TestProject") || !strings.Contains(result, "1.0.0") {
		t.Errorf("Key-value table missing data. Got:\n%s", result)
	}
}

func TestRenderCompactList(t *testing.T) {
	items := []string{"First", "Second", "Third"}
	result := RenderCompactList("Tasks", items)

	// Verify header
	if !strings.Contains(result, "Tasks") {
		t.Errorf("Compact list missing header. Got:\n%s", result)
	}

	// Verify items
	for _, item := range items {
		if !strings.Contains(result, item) {
			t.Errorf("Compact list missing item %s. Got:\n%s", item, result)
		}
	}
}

func TestTableBuilderChaining(t *testing.T) {
	// Test fluent interface
	result := NewTableBuilder().
		WithHeaders("Col1", "Col2", "Col3").
		AddRow("A", "B", "C").
		AddRow("D", "E", "F").
		Build()

	if result == "" {
		t.Error("Chained table builder returned empty result")
	}

	if !strings.Contains(result, "Col1") || !strings.Contains(result, "A") {
		t.Errorf("Chained builder produced incorrect table. Got:\n%s", result)
	}
}

func TestAddMultipleRows(t *testing.T) {
	rows := [][]interface{}{
		{"Row1", "Value1"},
		{"Row2", "Value2"},
		{"Row3", "Value3"},
	}

	result := NewTableBuilder().
		WithHeaders("Key", "Value").
		AddRows(rows).
		Build()

	for _, row := range rows {
		if !strings.Contains(result, row[0].(string)) {
			t.Errorf("Table missing row %v. Got:\n%s", row, result)
		}
	}
}

func TestEmptyTable(t *testing.T) {
	result := NewTableBuilder().
		WithHeaders("Header1", "Header2").
		Build()

	// Should still produce a valid table structure with headers
	if !strings.Contains(result, "Header1") {
		t.Errorf("Empty table missing headers. Got:\n%s", result)
	}
}

func TestTrimMarkdownTable(t *testing.T) {
	input := "\n\n  | A | B |\n  | - | - |\n  | 1 | 2 |\n\n  "
	result := TrimMarkdownTable(input)

	if strings.HasPrefix(result, "\n") || strings.HasSuffix(result, "\n") {
		t.Errorf("TrimMarkdownTable didn't trim properly. Got: %q", result)
	}
}

func TestAlignedTable(t *testing.T) {
	at := NewAlignedTable()
	at.SetHeaders("Left", "Center", "Right")
	at.AddRow("A", "B", "C")
	at.AddRow("D", "E", "F")

	result := at.RenderMarkdown()

	if !strings.Contains(result, "Left") || !strings.Contains(result, "A") {
		t.Errorf("Aligned table produced incorrect output. Got:\n%s", result)
	}
}
