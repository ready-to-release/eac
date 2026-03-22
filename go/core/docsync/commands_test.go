//go:build L0

package docsync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateDocStub(t *testing.T) {
	tests := []struct {
		name            string
		command         string
		wantTitle       string
		wantMarker      string
		wantMarkerExact string
	}{
		{
			name:            "single word command",
			command:         "build",
			wantTitle:       "# Build",
			wantMarker:      "<!-- book:cmd build -->",
			wantMarkerExact: "<!-- book:cmd build -->",
		},
		{
			name:            "two word command",
			command:         "get modules",
			wantTitle:       "# Get Modules",
			wantMarker:      "<!-- book:cmd get modules -->",
			wantMarkerExact: "<!-- book:cmd get modules -->",
		},
		{
			name:            "three word command",
			command:         "pipeline ci dispatch-and-wait",
			wantTitle:       "# Pipeline Ci Dispatch-and-wait",
			wantMarker:      "<!-- book:cmd pipeline ci dispatch-and-wait -->",
			wantMarkerExact: "<!-- book:cmd pipeline ci dispatch-and-wait -->",
		},
		{
			name:            "command with hyphen",
			command:         "release-this",
			wantTitle:       "# Release-this",
			wantMarker:      "<!-- book:cmd release-this -->",
			wantMarkerExact: "<!-- book:cmd release-this -->",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateDocStub(tt.command)

			// Verify title format
			assert.Contains(t, got, tt.wantTitle, "should contain proper title")

			// Verify marker format
			assert.Contains(t, got, tt.wantMarker, "should contain book:cmd marker")

			// Verify exact marker syntax
			assert.Contains(t, got, tt.wantMarkerExact, "marker should have exact format")

			// Verify structure: should start with title, contain marker
			lines := splitLines(got)
			require.GreaterOrEqual(t, len(lines), 3, "should have at least 3 lines")
			assert.True(t, lines[0] == tt.wantTitle || lines[0]+"" == tt.wantTitle,
				"first line should be title")
		})
	}
}

func TestGenerateDocStub_MarkerFormat(t *testing.T) {
	t.Run("marker has proper HTML comment syntax", func(t *testing.T) {
		stub := GenerateDocStub("test command")

		// Verify it uses proper HTML comment markers
		assert.Contains(t, stub, "<!--", "should start marker with <!--")
		assert.Contains(t, stub, "-->", "should end marker with -->")

		// Verify marker is complete and well-formed
		assert.Contains(t, stub, "<!-- book:cmd test command -->",
			"marker should be properly formatted")
	})

	t.Run("stub ends with blank line for content", func(t *testing.T) {
		stub := GenerateDocStub("build")

		// Should end with newlines to allow appending content
		assert.True(t, len(stub) > 2 && stub[len(stub)-2:] == "\n\n",
			"should end with double newline for content")
	})
}

func TestScanCommandDocsWithCommands_MissingDocs(t *testing.T) {
	// Setup temp directory structure
	repoRoot := t.TempDir()

	// Create minimal docs directory structure (no command docs)
	docsBase := filepath.Join(repoRoot, "docs", "reference", "eac", "commands")
	require.NoError(t, os.MkdirAll(filepath.Join(docsBase, "get"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(docsBase, "show"), 0o755))

	// Create config with default settings
	cmdConfig := config.DefaultCommandsConfig()

	// Mock commands list
	commands := []CommandInfo{
		{Command: "get modules", Description: "Get modules"},
		{Command: "get files", Description: "Get files"},
		{Command: "show modules", Description: "Show modules"},
		{Command: "build", Description: "Build artifacts"},
	}

	result, err := ScanCommandDocsWithCommands(commands, repoRoot, cmdConfig)
	require.NoError(t, err)

	// Verify counts
	assert.Equal(t, 4, result.ValidCommands, "should have 4 valid commands")
	assert.Equal(t, 0, result.DocumentedCount, "should have 0 documented commands")
	assert.Len(t, result.MissingDocs, 4, "should have 4 missing docs")

	// Verify each command is in missing docs
	missingCommands := make(map[string]CommandDocStatus)
	for _, m := range result.MissingDocs {
		missingCommands[m.Command] = m
	}

	assert.Contains(t, missingCommands, "get modules")
	assert.Contains(t, missingCommands, "get files")
	assert.Contains(t, missingCommands, "show modules")
	assert.Contains(t, missingCommands, "build")

	// Verify expected paths are correct
	assert.Equal(t, "docs/reference/eac/commands/get/modules.md",
		filepath.ToSlash(missingCommands["get modules"].ExpectedDoc))
	assert.Equal(t, "docs/reference/eac/commands/get/files.md",
		filepath.ToSlash(missingCommands["get files"].ExpectedDoc))
}

func TestScanCommandDocsWithCommands_AllDocumented(t *testing.T) {
	// Setup temp directory structure
	repoRoot := t.TempDir()

	// Create docs directory structure with all required docs
	docsBase := filepath.Join(repoRoot, "docs", "reference", "eac", "commands")
	require.NoError(t, os.MkdirAll(filepath.Join(docsBase, "get"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(docsBase, "show"), 0o755))
	// Create the doc files
	docFiles := []string{
		filepath.Join(docsBase, "get", "modules.md"),
		filepath.Join(docsBase, "get", "files.md"),
		filepath.Join(docsBase, "show", "modules.md"),
		filepath.Join(docsBase, "build.md"), // Single-word uncategorized goes to docsBase root
	}

	for _, docFile := range docFiles {
		require.NoError(t, os.WriteFile(docFile, []byte("# Doc\n"), 0o644))
	}

	// Create config with default settings
	cmdConfig := config.DefaultCommandsConfig()

	// Mock commands list matching the created docs
	commands := []CommandInfo{
		{Command: "get modules", Description: "Get modules"},
		{Command: "get files", Description: "Get files"},
		{Command: "show modules", Description: "Show modules"},
		{Command: "build", Description: "Build artifacts"},
	}

	result, err := ScanCommandDocsWithCommands(commands, repoRoot, cmdConfig)
	require.NoError(t, err)

	// Verify counts
	assert.Equal(t, 4, result.ValidCommands, "should have 4 valid commands")
	assert.Equal(t, 4, result.DocumentedCount, "should have 4 documented commands")
	assert.Empty(t, result.MissingDocs, "should have no missing docs")
}

func TestScanCommandDocsWithCommands_PartiallyDocumented(t *testing.T) {
	// Setup temp directory structure
	repoRoot := t.TempDir()

	// Create docs directory structure with some docs
	docsBase := filepath.Join(repoRoot, "docs", "reference", "eac", "commands")
	require.NoError(t, os.MkdirAll(filepath.Join(docsBase, "get"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(docsBase, "show"), 0o755))

	// Only create some doc files
	require.NoError(t, os.WriteFile(
		filepath.Join(docsBase, "get", "modules.md"),
		[]byte("# Get Modules\n"),
		0o644))
	require.NoError(t, os.WriteFile(
		filepath.Join(docsBase, "show", "modules.md"),
		[]byte("# Show Modules\n"),
		0o644))

	// Create config with default settings
	cmdConfig := config.DefaultCommandsConfig()

	// Mock commands list (some documented, some not)
	commands := []CommandInfo{
		{Command: "get modules", Description: "Get modules"},     // documented
		{Command: "get files", Description: "Get files"},         // missing
		{Command: "show modules", Description: "Show modules"},   // documented
		{Command: "show files", Description: "Show file listing"}, // missing
	}

	result, err := ScanCommandDocsWithCommands(commands, repoRoot, cmdConfig)
	require.NoError(t, err)

	// Verify counts
	assert.Equal(t, 4, result.ValidCommands, "should have 4 valid commands")
	assert.Equal(t, 2, result.DocumentedCount, "should have 2 documented commands")
	assert.Len(t, result.MissingDocs, 2, "should have 2 missing docs")

	// Verify correct commands are missing
	missingCommands := make([]string, 0)
	for _, m := range result.MissingDocs {
		missingCommands = append(missingCommands, m.Command)
	}
	assert.Contains(t, missingCommands, "get files")
	assert.Contains(t, missingCommands, "show files")
}

func TestScanCommandDocsWithCommands_OrphanedDocs(t *testing.T) {
	// Setup temp directory structure
	repoRoot := t.TempDir()

	// Create docs directory structure
	docsBase := filepath.Join(repoRoot, "docs", "reference", "eac", "commands")
	require.NoError(t, os.MkdirAll(filepath.Join(docsBase, "get"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(docsBase, "show"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(docsBase, "obsolete"), 0o755))

	// Create valid doc files (with markers for existing commands)
	require.NoError(t, os.WriteFile(
		filepath.Join(docsBase, "get", "modules.md"),
		[]byte("# Get Modules\n\n<!-- book:cmd get modules -->\n"),
		0o644))
	require.NoError(t, os.WriteFile(
		filepath.Join(docsBase, "show", "modules.md"),
		[]byte("# Show Modules\n\n<!-- book:cmd show modules -->\n"),
		0o644))

	// Create orphaned doc files (with markers for NON-EXISTENT commands)
	require.NoError(t, os.WriteFile(
		filepath.Join(docsBase, "get", "old-command.md"),
		[]byte("# Old Command\n\n<!-- book:cmd get old-command -->\n"),
		0o644))
	require.NoError(t, os.WriteFile(
		filepath.Join(docsBase, "obsolete", "deprecated.md"),
		[]byte("# Deprecated\n\n<!-- book:cmd obsolete deprecated -->\n"),
		0o644))
	require.NoError(t, os.WriteFile(
		filepath.Join(docsBase, "show", "removed-feature.md"),
		[]byte("# Removed Feature\n\n<!-- book:cmd show removed-feature -->\n"),
		0o644))

	// Create a manual doc file (no marker - should NOT be orphaned)
	require.NoError(t, os.WriteFile(
		filepath.Join(docsBase, "overview.md"),
		[]byte("# Overview\n\nThis is manual documentation.\n"),
		0o644))

	// Create config with default settings
	cmdConfig := config.DefaultCommandsConfig()

	// Mock commands list (only has matching commands for valid docs)
	commands := []CommandInfo{
		{Command: "get modules", Description: "Get modules"},
		{Command: "show modules", Description: "Show modules"},
	}

	result, err := ScanCommandDocsWithCommands(commands, repoRoot, cmdConfig)
	require.NoError(t, err)

	// Verify counts
	assert.Equal(t, 2, result.ValidCommands, "should have 2 valid commands")
	assert.Equal(t, 2, result.DocumentedCount, "should have 2 documented commands")
	assert.Empty(t, result.MissingDocs, "should have no missing docs")

	// Verify orphaned docs are detected (only files with markers for non-existent commands)
	assert.Len(t, result.OrphanedDocs, 3, "should have 3 orphaned docs")

	// Verify orphaned doc paths
	orphanedPaths := make(map[string]bool)
	for _, p := range result.OrphanedDocs {
		orphanedPaths[p] = true
	}

	assert.True(t, orphanedPaths["docs/reference/eac/commands/get/old-command.md"],
		"should detect get/old-command.md as orphaned")
	assert.True(t, orphanedPaths["docs/reference/eac/commands/obsolete/deprecated.md"],
		"should detect obsolete/deprecated.md as orphaned")
	assert.True(t, orphanedPaths["docs/reference/eac/commands/show/removed-feature.md"],
		"should detect show/removed-feature.md as orphaned")
}

func TestScanCommandDocsWithCommands_IndexFilesNotOrphaned(t *testing.T) {
	// Setup temp directory structure
	repoRoot := t.TempDir()

	// Create docs directory structure
	docsBase := filepath.Join(repoRoot, "docs", "reference", "eac", "commands")
	require.NoError(t, os.MkdirAll(filepath.Join(docsBase, "get"), 0o755))

	// Create index files (should NOT be considered orphaned)
	indexFiles := []string{
		filepath.Join(docsBase, "index.md"),
		filepath.Join(docsBase, "_index.md"),
		filepath.Join(docsBase, "get", "index.md"),
	}

	// Create valid command doc
	require.NoError(t, os.WriteFile(
		filepath.Join(docsBase, "get", "modules.md"),
		[]byte("# Get Modules\n"),
		0o644))

	for _, indexFile := range indexFiles {
		require.NoError(t, os.WriteFile(indexFile, []byte("# Index\n"), 0o644))
	}

	// Create config with default settings
	cmdConfig := config.DefaultCommandsConfig()

	// Mock commands list
	commands := []CommandInfo{
		{Command: "get modules", Description: "Get modules"},
	}

	result, err := ScanCommandDocsWithCommands(commands, repoRoot, cmdConfig)
	require.NoError(t, err)

	// Verify index files are not in orphaned list
	assert.Empty(t, result.OrphanedDocs, "index files should not be considered orphaned")
}

func TestScanCommandDocsWithCommands_SkippedDocs(t *testing.T) {
	// Setup temp directory structure
	repoRoot := t.TempDir()

	// Create docs directory (minimal setup)
	docsBase := filepath.Join(repoRoot, "docs", "reference", "eac", "commands")
	require.NoError(t, os.MkdirAll(filepath.Join(docsBase, "get"), 0o755))

	// Create one doc file
	require.NoError(t, os.WriteFile(
		filepath.Join(docsBase, "get", "modules.md"),
		[]byte("# Get Modules\n"),
		0o644))

	// Create config with skip_docs override for one command
	cmdConfig := config.DefaultCommandsConfig()
	cmdConfig.Overrides = map[string]config.CommandOverride{
		"help": {SkipDocs: true},
	}

	// Mock commands list including a skipped command
	commands := []CommandInfo{
		{Command: "get modules", Description: "Get modules"},
		{Command: "help", Description: "Show help"}, // Should be skipped
	}

	result, err := ScanCommandDocsWithCommands(commands, repoRoot, cmdConfig)
	require.NoError(t, err)

	// Verify skipped command is not counted
	assert.Equal(t, 1, result.ValidCommands, "should have 1 valid command (help is skipped)")
	assert.Equal(t, 1, result.DocumentedCount, "should have 1 documented command")
	assert.Empty(t, result.MissingDocs, "should have no missing docs")
}

func TestScanCommandDocsWithCommands_CustomDocPath(t *testing.T) {
	// Setup temp directory structure
	repoRoot := t.TempDir()

	// Create docs directory with custom path
	customPath := filepath.Join(repoRoot, "docs", "cli", "reference", "special-command.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(customPath), 0o755))
	require.NoError(t, os.WriteFile(customPath, []byte("# Special Command\n"), 0o644))

	// Create config with path override
	cmdConfig := config.DefaultCommandsConfig()
	cmdConfig.Overrides = map[string]config.CommandOverride{
		"special": {Path: "docs/cli/reference/special-command.md"},
	}

	// Mock commands list
	commands := []CommandInfo{
		{Command: "special", Description: "A special command with custom path"},
	}

	result, err := ScanCommandDocsWithCommands(commands, repoRoot, cmdConfig)
	require.NoError(t, err)

	// Verify the custom path is used
	assert.Equal(t, 1, result.ValidCommands, "should have 1 valid command")
	assert.Equal(t, 1, result.DocumentedCount, "should have 1 documented command")
	assert.Empty(t, result.MissingDocs, "should have no missing docs")
}

func TestScanCommandDocsWithCommands_CustomDocPath_Missing(t *testing.T) {
	// Setup temp directory structure
	repoRoot := t.TempDir()

	// Don't create the custom path file

	// Create config with path override
	cmdConfig := config.DefaultCommandsConfig()
	cmdConfig.Overrides = map[string]config.CommandOverride{
		"special": {Path: "docs/cli/reference/special-command.md"},
	}

	// Mock commands list
	commands := []CommandInfo{
		{Command: "special", Description: "A special command with custom path"},
	}

	result, err := ScanCommandDocsWithCommands(commands, repoRoot, cmdConfig)
	require.NoError(t, err)

	// Verify the missing doc has the custom expected path
	assert.Equal(t, 1, result.ValidCommands, "should have 1 valid command")
	assert.Equal(t, 0, result.DocumentedCount, "should have 0 documented commands")
	assert.Len(t, result.MissingDocs, 1, "should have 1 missing doc")
	assert.Equal(t, "docs/cli/reference/special-command.md", result.MissingDocs[0].ExpectedDoc)
}

func TestScanCommandDocsWithCommands_EmptyCommandList(t *testing.T) {
	repoRoot := t.TempDir()
	cmdConfig := config.DefaultCommandsConfig()

	// Empty commands list
	commands := []CommandInfo{}

	result, err := ScanCommandDocsWithCommands(commands, repoRoot, cmdConfig)
	require.NoError(t, err)

	assert.Equal(t, 0, result.ValidCommands)
	assert.Equal(t, 0, result.DocumentedCount)
	assert.Empty(t, result.MissingDocs)
	assert.Empty(t, result.OrphanedDocs)
}

func TestScanCommandDocsWithCommands_NoDocsDirectory(t *testing.T) {
	repoRoot := t.TempDir()
	cmdConfig := config.DefaultCommandsConfig()

	// Don't create any docs directory - it should handle gracefully

	commands := []CommandInfo{
		{Command: "get modules", Description: "Get modules"},
	}

	result, err := ScanCommandDocsWithCommands(commands, repoRoot, cmdConfig)
	require.NoError(t, err)

	assert.Equal(t, 1, result.ValidCommands)
	assert.Equal(t, 0, result.DocumentedCount)
	assert.Len(t, result.MissingDocs, 1)
	assert.Empty(t, result.OrphanedDocs, "should be empty when docs dir doesn't exist")
}

func TestScanCommandDocsWithCommands_CategoryAsCategoryDoc(t *testing.T) {
	// Test that single-word commands that are also categories
	// get the correct path: category/category.md

	repoRoot := t.TempDir()

	// Create docs directory with category doc
	docsBase := filepath.Join(repoRoot, "docs", "reference", "eac", "commands")
	require.NoError(t, os.MkdirAll(filepath.Join(docsBase, "get"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(docsBase, "get", "get.md"),
		[]byte("# Get\n"),
		0o644))

	cmdConfig := config.DefaultCommandsConfig()

	// The "get" command is a category
	commands := []CommandInfo{
		{Command: "get", Description: "Root get command"},
	}

	result, err := ScanCommandDocsWithCommands(commands, repoRoot, cmdConfig)
	require.NoError(t, err)

	// Verify it found the category doc
	assert.Equal(t, 1, result.ValidCommands)
	assert.Equal(t, 1, result.DocumentedCount)
	assert.Empty(t, result.MissingDocs)
}

func TestCommandDocStatus_Fields(t *testing.T) {
	status := CommandDocStatus{
		Command:     "get modules",
		ExpectedDoc: "docs/reference/eac/commands/get/modules.md",
		Exists:      true,
	}

	assert.Equal(t, "get modules", status.Command)
	assert.Equal(t, "docs/reference/eac/commands/get/modules.md", status.ExpectedDoc)
	assert.True(t, status.Exists)
}

func TestCommandDocSyncResult_Fields(t *testing.T) {
	result := CommandDocSyncResult{
		ValidCommands:   10,
		DocumentedCount: 8,
		MissingDocs: []CommandDocStatus{
			{Command: "test1", ExpectedDoc: "path1.md", Exists: false},
			{Command: "test2", ExpectedDoc: "path2.md", Exists: false},
		},
		OrphanedDocs: []string{"orphan1.md", "orphan2.md"},
	}

	assert.Equal(t, 10, result.ValidCommands)
	assert.Equal(t, 8, result.DocumentedCount)
	assert.Len(t, result.MissingDocs, 2)
	assert.Len(t, result.OrphanedDocs, 2)
}

func TestCommandInfo_Fields(t *testing.T) {
	cmd := CommandInfo{
		Command:     "get modules",
		Description: "Get the list of modules",
	}

	assert.Equal(t, "get modules", cmd.Command)
	assert.Equal(t, "Get the list of modules", cmd.Description)
}

// Helper function to split string into lines
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
