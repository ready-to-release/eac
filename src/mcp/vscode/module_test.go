package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetermineFileModule(t *testing.T) {
	testCases := []struct {
		filePath string
		expected string
	}{
		// Automation modules - extracted by name
		{"automation/sh-vscode/script.sh", "sh-vscode"},
		{"automation/pwsh-build/build.ps1", "pwsh-build"},
		{"automation/sh-deploy/deploy.sh", "sh-deploy"},

		// Container modules - extracted by name
		{"containers/mkdocs/Dockerfile", "mkdocs"},
		{"containers/nginx-proxy/config.conf", "nginx-proxy"},

		// MCP servers - mcp-<service>
		{"src/mcp/pwsh/main.go", "src-mcp-pwsh"},
		{"src/mcp/vscode/main.go", "src-mcp-vscode"},
		{"src/mcp/docs/server.go", "src-mcp-docs"},
		{"src/mcp/github/api.go", "src-mcp-github"},

		// VSCode extension
		{".vscode/extensions/claude-mcp-vscode/src/extension.ts", "vscode-ext-claude-commit"},
		{".vscode/extensions/claude-mcp-vscode/package.json", "vscode-ext-claude-commit"},

		// Contracts - contracts-<name>
		{"contracts/repository/0.1.0/definitions.yml", "contracts-repository"},
		{"contracts/deployable-units/0.1.0/src-mcp-pwsh.yml", "contracts-deployable-units"},

		// Documentation
		{"docs/reference/trunk/versioning.md", "docs"},
		{"README.md", "docs"},
		{"QUICKSTART.md", "docs"},

		// Configuration
		{".claude/agents/vscode-ext-claude-commitension-commit-button.md", "claude-config"},
		{".vscode/settings.json", "vscode-config"},
		{".gitignore", "repo-config"},
		{"mkdocs.yml", "repo-config"},
		{"package.json", "repo-config"},
	}

	for _, tc := range testCases {
		result := determineFileModule(tc.filePath)
		if result != tc.expected {
			t.Errorf("For path %q: expected %q but got %q", tc.filePath, tc.expected, result)
		} else {
			fmt.Printf("✓ %s → %s\n", tc.filePath, result)
		}
	}
}

// FindAllRepositoryFiles returns all files in the repository that aren't gitignored
// Uses git to list both tracked and untracked files
func FindAllRepositoryFiles(repoRoot string) ([]string, error) {
	var allFiles []string

	// Get all tracked files (respects .gitignore automatically)
	trackedCmd := exec.Command("git", "ls-files")
	trackedCmd.Dir = repoRoot
	trackedOutput, err := trackedCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files failed: %w", err)
	}

	// Parse tracked files
	trackedLines := strings.Split(string(trackedOutput), "\n")
	for _, line := range trackedLines {
		line = strings.TrimSpace(line)
		if line != "" {
			allFiles = append(allFiles, line)
		}
	}

	// Get all untracked files (excluding those matched by .gitignore)
	untrackedCmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	untrackedCmd.Dir = repoRoot
	untrackedOutput, err := untrackedCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files --others failed: %w", err)
	}

	// Parse untracked files
	untrackedLines := strings.Split(string(untrackedOutput), "\n")
	for _, line := range untrackedLines {
		line = strings.TrimSpace(line)
		if line != "" {
			allFiles = append(allFiles, line)
		}
	}

	return allFiles, nil
}

// TestAllRepositoryFiles verifies that ALL non-gitignored files can be properly classified
func TestAllRepositoryFiles(t *testing.T) {
	// Get repository root (three levels up from src/mcp/vscode)
	repoRoot, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatalf("Failed to get repository root: %v", err)
	}

	fmt.Printf("\nRepository root: %s\n\n", repoRoot)

	// Find all files in the repository
	files, err := FindAllRepositoryFiles(repoRoot)
	if err != nil {
		t.Fatalf("Failed to find repository files: %v", err)
	}

	fmt.Printf("Found %d files in repository (tracked + untracked, excluding gitignored)\n\n", len(files))

	// Track statistics
	moduleStats := make(map[string]int)
	unknownFiles := []string{}
	totalFiles := 0

	// Test each file
	for _, file := range files {
		if file == "" {
			continue
		}

		totalFiles++
		module := determineFileModule(file)
		moduleStats[module]++

		// Track unknown files - these indicate gaps in our detection
		if module == "unknown" {
			unknownFiles = append(unknownFiles, file)
		}
	}

	// Print statistics
	fmt.Println("=== Module Detection Statistics ===")
	fmt.Println()
	for module, count := range moduleStats {
		if module == "unknown" {
			fmt.Printf("❌ %-30s: %d files\n", module, count)
		} else {
			fmt.Printf("✓  %-30s: %d files\n", module, count)
		}
	}
	fmt.Printf("\nTotal files processed: %d\n", totalFiles)

	// Report unknown files
	if len(unknownFiles) > 0 {
		fmt.Println()
		fmt.Printf("=== ❌ WARNING: %d files could not be classified ===\n", len(unknownFiles))
		fmt.Println()
		for _, file := range unknownFiles {
			fmt.Printf("  - %s\n", file)
		}
		fmt.Println()
		fmt.Println("These files need pattern rules added to determineFileModule()")
		t.Errorf("Found %d files with unknown module classification", len(unknownFiles))
	} else {
		fmt.Println()
		fmt.Println("✅ SUCCESS: All files were properly classified!")
	}
}

// Helper function to print examples (not a test)
func printExamples() {
	examples := []string{
		"automation/sh-vscode/install.sh",
		"automation/pwsh-build/build.ps1",
		"containers/mkdocs/Dockerfile",
		"src/mcp/pwsh/main.go",
		".vscode/extensions/claude-mcp-vscode/src/extension.ts",
		"contracts/repository/0.1.0/definitions.yml",
		"docs/reference/trunk/semantic-commits.md",
		".claude/agents/vscode-ext-claude-commitension-commit-button.md",
		"README.md",
		".gitignore",
	}

	fmt.Println("\n=== Module Detection Examples ===")
	for _, path := range examples {
		module := determineFileModule(path)
		fmt.Printf("%-60s → %s\n", path, module)
	}
}

// TestGetModuleGlobPattern verifies GitHub Actions glob pattern generation
func TestGetModuleGlobPattern(t *testing.T) {
	testCases := []struct {
		module        string
		expectedGlobs []string
		description   string
	}{
		// MCP servers
		{
			module:        "src-mcp-vscode",
			expectedGlobs: []string{"src/mcp/vscode/**"},
			description:   "MCP VSCode server",
		},
		{
			module:        "src-mcp-pwsh",
			expectedGlobs: []string{"src/mcp/pwsh/**"},
			description:   "MCP PowerShell server",
		},
		{
			module:        "src-mcp-docs",
			expectedGlobs: []string{"src/mcp/docs/**"},
			description:   "MCP Docs server",
		},
		{
			module:        "src-mcp-github",
			expectedGlobs: []string{"src/mcp/github/**"},
			description:   "MCP GitHub server",
		},

		// Automation modules
		{
			module:        "sh-vscode",
			expectedGlobs: []string{"automation/sh-vscode/**"},
			description:   "Shell automation",
		},
		{
			module:        "pwsh-build",
			expectedGlobs: []string{"automation/pwsh-build/**"},
			description:   "PowerShell automation",
		},

		// Container modules
		{
			module:        "mkdocs",
			expectedGlobs: []string{"containers/mkdocs/**"},
			description:   "MkDocs container",
		},

		// VSCode extension
		{
			module:        "vscode-ext-claude-commit",
			expectedGlobs: []string{".vscode/extensions/claude-mcp-vscode/**"},
			description:   "VSCode extension",
		},

		// Contracts
		{
			module:        "contracts-repository",
			expectedGlobs: []string{"contracts/repository/**"},
			description:   "Repository contracts",
		},
		{
			module:        "contracts-deployable-units",
			expectedGlobs: []string{"contracts/deployable-units/**"},
			description:   "Deployable unit contracts",
		},

		// Configuration modules
		{
			module:        "claude-config",
			expectedGlobs: []string{".claude/**"},
			description:   "Claude configuration",
		},
		{
			module: "vscode-config",
			expectedGlobs: []string{
				".vscode/*.json",
				".vscode/*.md",
				".vscode/settings.*.json",
			},
			description: "VSCode configuration",
		},

		// Documentation
		{
			module:        "docs",
			expectedGlobs: []string{"docs/**", "*.md"},
			description:   "Documentation",
		},

		// CLI module
		{
			module:        "cli",
			expectedGlobs: []string{"src/cli/**"},
			description:   "CLI module",
		},

		// Repo config (multiple patterns)
		{
			module: "repo-config",
			expectedGlobs: []string{
				"*.json",
				"*.yml",
				"*.yaml",
				".gitignore",
				".gitattributes",
				"LICENSE",
				"*.lock",
			},
			description: "Repository configuration",
		},
	}

	fmt.Println("\n=== GitHub Actions Glob Pattern Generation ===")

	for _, tc := range testCases {
		globs := getModuleGlobPattern(tc.module)

		// Check if we got the expected number of patterns
		if len(globs) != len(tc.expectedGlobs) {
			t.Errorf("Module %q: expected %d glob patterns, got %d", tc.module, len(tc.expectedGlobs), len(globs))
			continue
		}

		// Check if all expected patterns are present
		passed := true
		for i, expected := range tc.expectedGlobs {
			if i >= len(globs) || globs[i] != expected {
				t.Errorf("Module %q: expected glob[%d] = %q, got %q", tc.module, i, expected, globs[i])
				passed = false
			}
		}

		// Print result
		if passed {
			fmt.Printf("✓ %-30s: %s\n", tc.module, tc.description)
			fmt.Printf("  GitHub Actions paths:\n")
			for _, glob := range globs {
				fmt.Printf("    - '%s'\n", glob)
			}
			fmt.Println()
		}
	}
}

// TestModuleGlobPatternIntegration tests the full workflow
func TestModuleGlobPatternIntegration(t *testing.T) {
	// Simulate file changes
	testFiles := []string{
		"src/mcp/vscode/main.go",
		"src/mcp/vscode/module_test.go",
		".vscode/extensions/claude-mcp-vscode/src/extension.ts",
		".vscode/extensions/claude-mcp-vscode/package.json",
		"docs/reference/trunk/semantic-commits.md",
		"automation/sh-vscode/install.sh",
	}

	fmt.Println("\n=== Integration Test: File Changes → Modules → Globs ===")

	// Group by module
	moduleFiles := make(map[string][]string)
	for _, file := range testFiles {
		module := determineFileModule(file)
		moduleFiles[module] = append(moduleFiles[module], file)
	}

	// For each module, show the glob pattern
	for module, files := range moduleFiles {
		globs := getModuleGlobPattern(module)

		fmt.Printf("Module: %s\n", module)
		fmt.Printf("Files changed:\n")
		for _, file := range files {
			fmt.Printf("  - %s\n", file)
		}
		fmt.Printf("GitHub Actions trigger:\n")
		fmt.Printf("```yaml\n")
		fmt.Printf("paths:\n")
		for _, glob := range globs {
			fmt.Printf("  - '%s'\n", glob)
		}
		fmt.Printf("```\n\n")
	}
}
