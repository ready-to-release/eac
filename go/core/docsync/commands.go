// Package docsync provides shared logic for scanning and synchronizing CLI command documentation.
package docsync

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
	"gopkg.in/yaml.v3"
)

// CommandInfo represents a command from get valid-commands.
type CommandInfo struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
}

// CommandDocStatus represents the sync status of a command's documentation.
type CommandDocStatus struct {
	Command     string // CLI command name (e.g., "get modules")
	ExpectedDoc string // Expected doc path relative to repo root
	Exists      bool   // Whether the doc file exists
}

// CommandDocSyncResult holds the results of scanning command documentation.
type CommandDocSyncResult struct {
	ValidCommands   int                // Total valid CLI commands
	DocumentedCount int                // Commands with existing docs
	MissingDocs     []CommandDocStatus // Commands without docs
	OrphanedDocs    []string           // Doc files for non-existent commands (relative paths)
}

// ScanCommandDocs compares valid CLI commands against existing documentation.
// It uses CommandsConfig.GetDocPath() for path mapping.
// cmdBinary is the path to the commands binary to run `get valid-commands`.
// repoRoot is the repository root for resolving paths.
// cmdConfig provides path mapping rules.
// Returns the scan result or error.
func ScanCommandDocs(cmdBinary, repoRoot string, cmdConfig *config.CommandsConfig) (*CommandDocSyncResult, error) {
	commands, err := getValidCommands(cmdBinary, repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid commands: %w", err)
	}

	return ScanCommandDocsWithCommands(commands, repoRoot, cmdConfig)
}

// ScanCommandDocsWithCommands is the testable version that accepts commands directly.
func ScanCommandDocsWithCommands(commands []CommandInfo, repoRoot string, cmdConfig *config.CommandsConfig) (*CommandDocSyncResult, error) {
	result := &CommandDocSyncResult{
		ValidCommands:   len(commands),
		DocumentedCount: 0,
		MissingDocs:     []CommandDocStatus{},
		OrphanedDocs:    []string{},
	}

	// Build set of valid commands and expected doc paths
	validCommands := make(map[string]bool)
	expectedDocPaths := make(map[string]bool)

	for _, cmd := range commands {
		// Skip commands configured to not require docs
		if cmdConfig.ShouldSkipDocs(cmd.Command) {
			result.ValidCommands-- // Don't count skipped commands
			continue
		}

		validCommands[cmd.Command] = true

		docPath := cmdConfig.GetDocPath(cmd.Command, "", repoRoot)
		normalizedPath := filepath.ToSlash(docPath)
		expectedDocPaths[normalizedPath] = true

		fullPath := filepath.Join(repoRoot, docPath)
		exists := fileExists(fullPath)

		status := CommandDocStatus{
			Command:     cmd.Command,
			ExpectedDoc: docPath,
			Exists:      exists,
		}

		if exists {
			result.DocumentedCount++
		} else {
			result.MissingDocs = append(result.MissingDocs, status)
		}
	}

	// Scan for orphaned docs (files with book:cmd markers for non-existent commands)
	docsBase := cmdConfig.Defaults.DocsBase
	if docsBase == "" {
		docsBase = "docs/reference/eac/commands"
	}

	orphaned, err := findOrphanedDocs(repoRoot, docsBase, validCommands)
	if err != nil {
		// Non-fatal: continue with empty orphan list
		orphaned = []string{}
	}
	result.OrphanedDocs = orphaned

	return result, nil
}

// GenerateDocStub creates the content for a new command doc file.
// command is the full command name (e.g., "get components")
// Returns the markdown content for the doc stub.
func GenerateDocStub(command string) string {
	// Generate title from command name
	// "get components" -> "Get Components"
	parts := strings.Split(command, " ")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	title := strings.Join(parts, " ")

	return fmt.Sprintf(`# %s

<!-- book:cmd %s -->

`, title, command)
}

// getValidCommands runs the CLI binary to get the list of valid commands.
func getValidCommands(cmdBinary, repoRoot string) ([]CommandInfo, error) {
	cmd := exec.Command(cmdBinary, "get", "valid-commands")
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "NO_COLOR=1")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run get valid-commands: %s", stderr.String())
	}

	var commands []CommandInfo
	if err := yaml.Unmarshal(stdout.Bytes(), &commands); err != nil {
		return nil, fmt.Errorf("failed to parse valid-commands output: %w", err)
	}

	return commands, nil
}

// fileExists checks if a file exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// cmdMarkerPattern matches <!-- book:cmd X --> markers in markdown files.
// Captures the full marker content for validation.
var cmdMarkerPattern = regexp.MustCompile(`<!--\s*book:cmd\s+([^>]+?)\s*-->`)

// findOrphanedDocs scans the docs directory for markdown files that have
// <!-- book:cmd X --> markers for commands that don't exist.
// Files without markers are NOT considered orphaned (they're manual docs).
// Files with markers containing flags (--) are also skipped (variant docs).
func findOrphanedDocs(repoRoot, docsBase string, validCommands map[string]bool) ([]string, error) {
	var orphaned []string

	docsDir := filepath.Join(repoRoot, docsBase)
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		return orphaned, nil
	}

	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		matches := cmdMarkerPattern.FindSubmatch(content)
		if matches == nil {
			return nil
		}

		cmdName := strings.TrimSpace(string(matches[1]))

		// Skip markers with flags (e.g., "scan --scanner compliance")
		// These are variant docs, not command docs
		if strings.Contains(cmdName, "--") {
			return nil
		}

		if !validCommands[cmdName] {
			relPath, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return nil
			}
			orphaned = append(orphaned, filepath.ToSlash(relPath))
		}

		return nil
	})

	return orphaned, err
}
