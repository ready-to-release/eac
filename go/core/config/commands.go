package config

import (
	"os"
	"path/filepath"
	"strings"
)

// CommandsFileName is the name of the commands configuration file.
const CommandsFileName = "commands.yml"

// CommandsConfig holds CLI command documentation mapping configuration.
// Loaded from .r2r/eac/commands.yml.
type CommandsConfig struct {
	Defaults   CommandsDefaults           `yaml:"defaults"`
	Categories map[string]CategoryConfig  `yaml:"categories"`
	Overrides  map[string]CommandOverride `yaml:"overrides"`
}

// CommandsDefaults holds default settings for command documentation.
type CommandsDefaults struct {
	DocsBase            string `yaml:"docs_base"`
	UncategorizedFolder string `yaml:"uncategorized_folder"`
}

// CategoryConfig defines a command category.
type CategoryConfig struct {
	Description string `yaml:"description"`
}

// CommandOverride defines per-command overrides.
type CommandOverride struct {
	Path     string `yaml:"path"`
	SkipDocs bool   `yaml:"skip_docs"`
}

// DefaultCommandsConfig returns sensible defaults when no config file exists.
func DefaultCommandsConfig() *CommandsConfig {
	return &CommandsConfig{
		Defaults: CommandsDefaults{
			DocsBase:            "docs/reference/eac/commands",
			UncategorizedFolder: "core",
		},
		Categories: map[string]CategoryConfig{
			"get":       {Description: "Retrieve data in structured format"},
			"show":      {Description: "Display data in human-readable format"},
			"create":    {Description: "Generate new artifacts"},
			"validate":  {Description: "Check correctness against schemas"},
			"update":    {Description: "Modify existing artifacts"},
			"release":   {Description: "Release management operations"},
			"pipeline":  {Description: "CI/CD orchestration"},
			"scan":      {Description: "Security scanning"},
			"serve":     {Description: "Start local servers"},
			"test":      {Description: "Execute test suites"},
			"work":      {Description: "Workspace management"},
			"templates": {Description: "Template management"},
			"core":      {Description: "Core commands"},
		},
		Overrides: make(map[string]CommandOverride),
	}
}

// Initialize is a no-op but kept for interface consistency.
func (c *CommandsConfig) Initialize() error {
	return nil
}

// ShouldSkipDocs returns true if a command should not require documentation.
func (c *CommandsConfig) ShouldSkipDocs(command string) bool {
	if override, ok := c.Overrides[command]; ok {
		return override.SkipDocs
	}
	return false
}

// GetDocPath returns the expected documentation path for a command.
// The repoRoot is needed to check if category directories exist.
func (c *CommandsConfig) GetDocPath(command, sourcePath, repoRoot string) string {
	// Check explicit path override first
	if override, ok := c.Overrides[command]; ok && override.Path != "" {
		return override.Path
	}

	parts := strings.Split(command, " ")
	docsBase := c.Defaults.DocsBase
	if docsBase == "" {
		docsBase = "docs/reference/eac/commands"
	}

	if len(parts) == 1 {
		// Single-word command - check if it's a known category
		if _, isCategory := c.Categories[parts[0]]; isCategory {
			// It's a category root command (e.g., "get" -> "get/get.md")
			return filepath.Join(docsBase, parts[0], parts[0]+".md")
		}

		// Also check if directory exists on filesystem
		if repoRoot != "" {
			categoryDir := filepath.Join(repoRoot, docsBase, parts[0])
			if info, err := os.Stat(categoryDir); err == nil && info.IsDir() {
				return filepath.Join(docsBase, parts[0], parts[0]+".md")
			}
		}

		// Not a category - goes to uncategorized folder
		uncategorized := c.Defaults.UncategorizedFolder
		if uncategorized == "" {
			uncategorized = "core"
		}
		return filepath.Join(docsBase, uncategorized, parts[0]+".md")
	}

	// Multi-word command: first word is category, rest joined with hyphen
	// e.g., "get files" -> "get/files.md"
	// e.g., "pipeline ci dispatch-and-wait" -> "pipeline/ci-dispatch-and-wait.md"
	category := parts[0]
	subcommand := strings.Join(parts[1:], "-")
	return filepath.Join(docsBase, category, subcommand+".md")
}

// GetCategories returns the list of valid category names.
func (c *CommandsConfig) GetCategories() []string {
	categories := make([]string, 0, len(c.Categories))
	for name := range c.Categories {
		categories = append(categories, name)
	}
	return categories
}

// IsValidCategory checks if a category name is valid.
func (c *CommandsConfig) IsValidCategory(name string) bool {
	_, ok := c.Categories[name]
	return ok
}
