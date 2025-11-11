package render

import (
	"fmt"
	"testing"
)

// These tests generate example output for documentation
func TestExampleOutputs(t *testing.T) {
	t.Run("SimpleTable", func(t *testing.T) {
		headers := []string{"Name", "Age", "City"}
		rows := [][]interface{}{
			{"Alice", 30, "NYC"},
			{"Bob", 25, "LA"},
			{"Charlie", 35, "Chicago"},
		}
		result := SimpleMarkdownTable(headers, rows)
		fmt.Println("\n=== Simple Table ===")
		fmt.Println(result)
	})

	t.Run("BuilderWithAutoIndex", func(t *testing.T) {
		result := NewTableBuilder().
			WithHeaders("Module", "Type", "Status").
			WithAutoIndex().
			AddRow("cli", "application", "active").
			AddRow("contracts", "library", "active").
			AddRow("mcp", "server", "active").
			WithFooter("", "Total", "3 modules").
			Build()
		fmt.Println("\n=== Builder with Auto-Index ===")
		fmt.Println(result)
	})

	t.Run("KeyValueTable", func(t *testing.T) {
		data := map[string]interface{}{
			"Name":    "MyProject",
			"Version": "1.0.0",
			"Author":  "Team",
			"License": "MIT",
		}
		result := RenderKeyValueTable("Property", "Value", data)
		fmt.Println("\n=== Key-Value Table ===")
		fmt.Println(result)
	})

	t.Run("CompactList", func(t *testing.T) {
		items := []string{"Initialize database", "Load configuration", "Start server"}
		result := RenderCompactList("Startup Tasks", items)
		fmt.Println("\n=== Compact List ===")
		fmt.Println(result)
	})
}
