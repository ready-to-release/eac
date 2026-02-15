//go:build L1 && ov
// +build L1,ov

// Feature: commands_templates
// Unit tests for templates install claude command
package claude_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/commands/repository/templates/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeTemplatesInstallation(t *testing.T) {
	// Create temp directories (automatically cleaned up)
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates", "claude")
	outputDir := filepath.Join(tmpDir, ".claude")

	// Create template structure
	agentsDir := filepath.Join(templateDir, "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	// Create test template file
	tmplFile := filepath.Join(agentsDir, "architect.md")
	tmplContent := "# Architect Agent\nUses MCP commands: get-modules"
	require.NoError(t, os.WriteFile(tmplFile, []byte(tmplContent), 0644))

	// Create renderer (nil values = copy without replacement)
	renderer := internal.NewRenderer(templateDir, outputDir, nil)

	// Render templates
	err := renderer.RenderTemplates()
	require.NoError(t, err)

	// Verify output
	outputFile := filepath.Join(outputDir, "agents", "architect.md")
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	// Should preserve content exactly
	assert.Equal(t, tmplContent, string(content))
	assert.Contains(t, string(content), "get-modules")
}

func TestClaudeTemplatesPreservesExistingFiles(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates", "claude")
	outputDir := filepath.Join(tmpDir, ".claude")

	// Create existing file in output
	existingDir := filepath.Join(outputDir, "agents")
	require.NoError(t, os.MkdirAll(existingDir, 0755))
	existingFile := filepath.Join(existingDir, "go-architect.md")
	existingContent := "# Go Architect (existing)"
	require.NoError(t, os.WriteFile(existingFile, []byte(existingContent), 0644))

	// Create template with different file
	agentsDir := filepath.Join(templateDir, "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	tmplFile := filepath.Join(agentsDir, "architect.md")
	require.NoError(t, os.WriteFile(tmplFile, []byte("# Generic Architect"), 0644))

	// Render templates
	renderer := internal.NewRenderer(templateDir, outputDir, nil)
	err := renderer.RenderTemplates()
	require.NoError(t, err)

	// Verify existing file still exists
	content, err := os.ReadFile(existingFile)
	require.NoError(t, err)
	assert.Equal(t, existingContent, string(content), "Existing files should be preserved")

	// Verify new file also exists
	newFile := filepath.Join(outputDir, "agents", "architect.md")
	assert.FileExists(t, newFile, "New template should be installed")
}

func TestClaudeTemplatesDirectoryNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates", "claude")
	outputDir := filepath.Join(tmpDir, ".claude")

	// Don't create template directory
	renderer := internal.NewRenderer(templateDir, outputDir, nil)
	err := renderer.RenderTemplates()

	// Should fail gracefully
	assert.Error(t, err)
}

func TestClaudeTemplatesMultipleDirectories(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates", "claude")
	outputDir := filepath.Join(tmpDir, ".claude")

	// Create template structure with multiple subdirectories
	agentsDir := filepath.Join(templateDir, "agents")
	commandsDir := filepath.Join(templateDir, "commands")
	skillsDir := filepath.Join(templateDir, "skills")

	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	require.NoError(t, os.MkdirAll(commandsDir, 0755))
	require.NoError(t, os.MkdirAll(skillsDir, 0755))

	// Create test files in each directory
	require.NoError(t, os.WriteFile(filepath.Join(agentsDir, "architect.md"), []byte("# Architect"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(commandsDir, "plan.md"), []byte("# Plan"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(skillsDir, "feature-workflow.md"), []byte("# Workflow"), 0644))

	// Render templates
	renderer := internal.NewRenderer(templateDir, outputDir, nil)
	err := renderer.RenderTemplates()
	require.NoError(t, err)

	// Verify all files installed
	assert.FileExists(t, filepath.Join(outputDir, "agents", "architect.md"))
	assert.FileExists(t, filepath.Join(outputDir, "commands", "plan.md"))
	assert.FileExists(t, filepath.Join(outputDir, "skills", "feature-workflow.md"))
}

func TestClaudeTemplatesPreservesPlaceholders(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates", "claude")
	outputDir := filepath.Join(tmpDir, ".claude")

	// Create template with placeholders
	agentsDir := filepath.Join(templateDir, "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	tmplContent := "# {{ .AgentName }}\n\nMCP Commands: {{ .Commands }}"
	require.NoError(t, os.WriteFile(filepath.Join(agentsDir, "test.md"), []byte(tmplContent), 0644))

	// Render templates with nil values (should preserve placeholders)
	renderer := internal.NewRenderer(templateDir, outputDir, nil)
	err := renderer.RenderTemplates()
	require.NoError(t, err)

	// Verify placeholders preserved
	content, err := os.ReadFile(filepath.Join(outputDir, "agents", "test.md"))
	require.NoError(t, err)
	assert.Equal(t, tmplContent, string(content))
	assert.Contains(t, string(content), "{{ .AgentName }}")
	assert.Contains(t, string(content), "{{ .Commands }}")
}
