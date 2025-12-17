// migrate-flags.go - Automated flag validation migration script
// Migrates commands from hardcoded FlagDefinition arrays to registry-based validation
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run migrate-flags.go <file_path>")
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
		fmt.Println("No changes needed")
		return
	}

	// Write migrated content
	if err := os.WriteFile(filePath, []byte(migrated), 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Migrated: %s\n", filepath.Base(filePath))
}

func migrateFile(content string) string {
	// Step 1: Remove FlagDefinition variable declarations
	// Pattern: var somethingFlags = []flags.FlagDefinition{ ... }
	// Also handle: commandFlags, *Flags patterns
	flagDefPattern := regexp.MustCompile(`(?s)var\s+\w+\s*=\s*\[\]flags\.FlagDefinition\{[^}]*\}\s*\n`)
	content = flagDefPattern.ReplaceAllString(content, "")

	// Also handle inline declarations: flagDefs := []flags.FlagDefinition{...}
	inlinePattern := regexp.MustCompile(`(?s)\w+\s*:=\s*\[\]flags\.FlagDefinition\{[^}]*\}\s*\n`)
	content = inlinePattern.ReplaceAllString(content, "")

	// Step 2: Replace ValidateFlags calls with ValidateFlagsFromRegistry
	// Pattern: flags.ValidateFlags(os.Args[2:], anyVariable)
	validatePattern := regexp.MustCompile(`flags\.ValidateFlags\(os\.Args\[2:\],\s*\w+\)`)
	content = validatePattern.ReplaceAllString(content, "flags.ValidateFlagsFromRegistry(os.Args[2:])")

	return content
}
