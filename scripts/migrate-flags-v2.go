// migrate-flags-v2.go - Improved automated flag validation migration
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run migrate-flags-v2.go <file_path>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	original := string(content)
	migrated := migrateFile(original)

	if original == migrated {
		return // No changes needed
	}

	// Write migrated content
	if err := os.WriteFile(filePath, []byte(migrated), 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Migrated: %s\n", filepath.Base(filePath))
}

func migrateFile(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	i := 0

	for i < len(lines) {
		line := lines[i]

		// Check if this line starts a FlagDefinition declaration
		if strings.Contains(line, "[]flags.FlagDefinition{") {
			// Skip this line and all lines until we find the closing }
			braceCount := strings.Count(line, "{") - strings.Count(line, "}")
			i++
			for i < len(lines) && braceCount > 0 {
				braceCount += strings.Count(lines[i], "{") - strings.Count(lines[i], "}")
				i++
			}
			continue
		}

		// Replace ValidateFlags calls
		if strings.Contains(line, "flags.ValidateFlags(os.Args[2:]") ||
			strings.Contains(line, "flags.ValidateFlags(args") {
			// Replace with ValidateFlagsFromRegistry
			if strings.Contains(line, "flags.ValidateFlags(os.Args[2:],") {
				line = strings.Split(line, ",")[0] + ")" // Remove the second parameter
				line = strings.ReplaceAll(line, "flags.ValidateFlags", "flags.ValidateFlagsFromRegistry")
			} else if strings.Contains(line, "flags.ValidateFlags(args,") {
				// Change args to os.Args[2:]
				indentation := getIndentation(line)
				line = indentation + "if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {"
			}
		}

		result = append(result, line)
		i++
	}

	return strings.Join(result, "\n")
}

func getIndentation(line string) string {
	for i, ch := range line {
		if ch != ' ' && ch != '\t' {
			return line[:i]
		}
	}
	return ""
}
