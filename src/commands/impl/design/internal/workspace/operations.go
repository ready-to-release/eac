package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// getRepoRoot returns the repository root directory
func getRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree to find the repository root
	dir := cwd
	for {
		// Check if we're at a directory that has "specs" as a subdirectory
		specsPath := filepath.Join(dir, "specs")
		if stat, err := os.Stat(specsPath); err == nil && stat.IsDir() {
			// Found the root - this directory has a specs subdirectory
			return dir, nil
		}

		// Check if we're in a src subdirectory structure
		base := filepath.Base(dir)
		parent := filepath.Dir(dir)

		// If we're in src/commands/design, src/commands, or src/cli
		if base == "design" || base == "commands" || base == "cli" {
			grandparent := filepath.Dir(parent)
			if filepath.Base(parent) == "src" {
				return grandparent, nil
			}
		}

		// If we're in src, go up one level
		if base == "src" {
			return parent, nil
		}

		// Move up one directory
		nextDir := filepath.Dir(dir)
		if nextDir == dir {
			// Reached the root of the filesystem
			return "", fmt.Errorf("could not find repository root (no specs/ directory found)")
		}
		dir = nextDir
	}
}

// GetWorkspaceRoot returns the specs directory (workspace root for all modules)
func GetWorkspaceRoot() (string, error) {
	root, err := getRepoRoot()
	if err != nil {
		return "", err
	}
	// Workspace root is the specs directory
	return filepath.Join(root, "specs"), nil
}

// GetModulePath returns the design directory for a specific module
// Pattern: specs/<module>/design/
func GetModulePath(module string) (string, error) {
	root, err := GetWorkspaceRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, module, "design"), nil
}

// GetWorkspacePath returns the workspace.dsl file path for a module
// Pattern: specs/<module>/design/workspace.dsl
func GetWorkspacePath(module string) (string, error) {
	modulePath, err := GetModulePath(module)
	if err != nil {
		return "", err
	}
	return filepath.Join(modulePath, "workspace.dsl"), nil
}

// Create creates a new workspace for a module
func Create(module, name, description string, force bool) (string, error) {
	if module == "" || name == "" || description == "" {
		return "", fmt.Errorf("module, name, and description are required")
	}

	modulePath, err := GetModulePath(module)
	if err != nil {
		return "", err
	}

	// Create specs/<module>/design/ directory
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create module directory: %w", err)
	}

	dslPath := filepath.Join(modulePath, "workspace.dsl")

	// Check if workspace already exists
	if _, err := os.Stat(dslPath); err == nil && !force {
		return "", fmt.Errorf("workspace already exists at %s (use --force to overwrite)", dslPath)
	}

	dslContent := GenerateBaseDSL(name, description)

	if err := os.WriteFile(dslPath, []byte(dslContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write DSL file: %w", err)
	}

	return fmt.Sprintf("Created workspace '%s' at %s", name, dslPath), nil
}

// AddContainer adds a container to an existing workspace
func AddContainer(module, name, technology, description string) (string, error) {
	if module == "" || name == "" {
		return "", fmt.Errorf("module and name are required")
	}

	dslPath, err := GetWorkspacePath(module)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(dslPath); os.IsNotExist(err) {
		return "", fmt.Errorf("workspace not found for module '%s' at %s", module, dslPath)
	}

	dsl, err := os.ReadFile(dslPath)
	if err != nil {
		return "", fmt.Errorf("failed to read DSL file: %w", err)
	}

	containerID := SanitizeID(name)
	containerDef := fmt.Sprintf("\n            %s = container \"%s\" \"%s\" \"%s\"\n",
		containerID, name, description, technology)

	systemIdx := strings.Index(string(dsl), "system = softwareSystem")
	if systemIdx == -1 {
		return "", fmt.Errorf("system not found in workspace")
	}

	insertIdx := strings.Index(string(dsl)[systemIdx:], "# Containers will be added here")
	if insertIdx == -1 {
		insertIdx = strings.Index(string(dsl)[systemIdx:], "}")
	}
	insertIdx += systemIdx

	newDSL := string(dsl[:insertIdx]) + containerDef + string(dsl[insertIdx:])

	if err := os.WriteFile(dslPath, []byte(newDSL), 0644); err != nil {
		return "", fmt.Errorf("failed to write DSL file: %w", err)
	}

	return fmt.Sprintf("Added container '%s' to %s module", name, module), nil
}

// AddRelationship adds a relationship to an existing workspace
func AddRelationship(module, source, destination, description, technology string) (string, error) {
	if module == "" || source == "" || destination == "" {
		return "", fmt.Errorf("module, source, and destination are required")
	}

	dslPath, err := GetWorkspacePath(module)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(dslPath); os.IsNotExist(err) {
		return "", fmt.Errorf("workspace not found for module '%s' at %s", module, dslPath)
	}

	dsl, err := os.ReadFile(dslPath)
	if err != nil {
		return "", fmt.Errorf("failed to read DSL file: %w", err)
	}

	var relationshipDef string
	if technology != "" {
		relationshipDef = fmt.Sprintf("\n        %s -> %s \"%s\" \"%s\"\n", source, destination, description, technology)
	} else {
		relationshipDef = fmt.Sprintf("\n        %s -> %s \"%s\"\n", source, destination, description)
	}

	insertMarker := "# Define relationships here"
	insertIdx := strings.Index(string(dsl), insertMarker)
	if insertIdx == -1 {
		insertIdx = strings.Index(string(dsl), "    }\n\n    views {")
		if insertIdx == -1 {
			return "", fmt.Errorf("could not find insertion point for relationship")
		}
	} else {
		insertIdx += len(insertMarker)
	}

	newDSL := string(dsl[:insertIdx]) + relationshipDef + string(dsl[insertIdx:])

	if err := os.WriteFile(dslPath, []byte(newDSL), 0644); err != nil {
		return "", fmt.Errorf("failed to write DSL file: %w", err)
	}

	return fmt.Sprintf("Added relationship: %s -> %s in %s module", source, destination, module), nil
}

// Export reads and returns the workspace DSL content
func Export(module string) (string, int, error) {
	if module == "" {
		return "", 0, fmt.Errorf("module is required")
	}

	dslPath, err := GetWorkspacePath(module)
	if err != nil {
		return "", 0, err
	}

	if _, err := os.Stat(dslPath); os.IsNotExist(err) {
		return "", 0, fmt.Errorf("workspace not found for module '%s' at %s", module, dslPath)
	}

	content, err := os.ReadFile(dslPath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read DSL file: %w", err)
	}

	return string(content), len(content), nil
}
