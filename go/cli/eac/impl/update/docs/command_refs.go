package docs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/docsync"
	"github.com/ready-to-release/eac/go/core/paths"
)

// CommandRefsResult holds the result of a command-refs update operation.
type CommandRefsResult struct {
	ValidCommands   int
	DocumentedCount int
	MissingDocs     int
	OrphanedDocs    int
	Created         int
	Removed         int
}

// runCommandRefsUpdate handles the command-refs area update.
// It syncs command reference documentation: creates missing stubs and removes orphans.
func runCommandRefsUpdate(repoRoot string, opts UpdateOptions, logWriter io.Writer) (CommandRefsResult, error) {
	result := CommandRefsResult{}

	fmt.Fprintln(logWriter, "Syncing command reference documentation...")

	// Load config
	cfg, err := config.Load(config.LoadOptions{
		RepoRoot:        repoRoot,
		ValidateSchemas: false, // Skip validation for faster loading
	})
	if err != nil {
		return result, fmt.Errorf("failed to load config: %w", err)
	}

	// Find commands binary
	cmdBinary := findCommandsBinary(repoRoot)
	if cmdBinary == "" {
		return result, fmt.Errorf("commands binary not found. Run 'go build ./go/cli/eac' first")
	}

	// Scan current state
	scanResult, err := docsync.ScanCommandDocs(cmdBinary, repoRoot, cfg.Commands)
	if err != nil {
		return result, fmt.Errorf("scanning command docs: %w", err)
	}

	result.ValidCommands = scanResult.ValidCommands
	result.DocumentedCount = scanResult.DocumentedCount
	result.MissingDocs = len(scanResult.MissingDocs)
	result.OrphanedDocs = len(scanResult.OrphanedDocs)

	// Report current state
	fmt.Fprintf(logWriter, "Commands: %d total, %d documented\n",
		result.ValidCommands, result.DocumentedCount)

	if result.MissingDocs > 0 {
		fmt.Fprintf(logWriter, "Missing docs: %d\n", result.MissingDocs)
	}
	if result.OrphanedDocs > 0 {
		fmt.Fprintf(logWriter, "Orphaned docs: %d\n", result.OrphanedDocs)
	}

	// Nothing to do?
	if result.MissingDocs == 0 && result.OrphanedDocs == 0 {
		fmt.Fprintln(logWriter, "All command docs in sync.")
		return result, nil
	}

	// Dry run - log what would happen
	if opts.DryRun {
		if result.MissingDocs > 0 {
			fmt.Fprintln(logWriter, "\nWould create:")
			for _, doc := range scanResult.MissingDocs {
				fmt.Fprintf(logWriter, "  + %s\n", doc.ExpectedDoc)
			}
		}
		if result.OrphanedDocs > 0 {
			fmt.Fprintln(logWriter, "\nWould remove (marker references non-existent command):")
			for _, doc := range scanResult.OrphanedDocs {
				fmt.Fprintf(logWriter, "  - %s\n", doc)
			}
		}
		return result, nil
	}

	// Apply changes
	fmt.Fprintln(logWriter)

	// Create missing docs
	for _, doc := range scanResult.MissingDocs {
		fullPath := filepath.Join(repoRoot, doc.ExpectedDoc)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Errorf("Failed to create directory %s: %v", dir, err)
			continue
		}
		content := docsync.GenerateDocStub(doc.Command)
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			log.Errorf("Failed to write %s: %v", doc.ExpectedDoc, err)
			continue
		}
		result.Created++
		fmt.Fprintf(logWriter, "  + %s\n", doc.ExpectedDoc)
	}

	// Remove orphaned docs
	for _, doc := range scanResult.OrphanedDocs {
		fullPath := filepath.Join(repoRoot, doc)
		if err := os.Remove(fullPath); err != nil {
			if !os.IsNotExist(err) {
				log.Errorf("Failed to remove %s: %v", doc, err)
			}
			continue
		}
		result.Removed++
		fmt.Fprintf(logWriter, "  - %s\n", doc)
	}

	if result.Created > 0 || result.Removed > 0 {
		fmt.Fprintf(logWriter, "\nCreated %d, removed %d\n", result.Created, result.Removed)
	}

	return result, nil
}

// findCommandsBinary locates the commands binary.
// Checks multiple locations in order of preference.
func findCommandsBinary(repoRoot string) string {
	// Check environment variable first
	if envPath := os.Getenv("COMMANDS_PATH"); envPath != "" {
		return envPath
	}

	// Standard location
	cmdBinary := paths.CommandsBinaryPath(repoRoot)
	if _, err := os.Stat(cmdBinary); err == nil {
		return cmdBinary
	}

	// Try with .exe extension on Windows
	if runtime.GOOS == "windows" {
		cmdBinaryExe := cmdBinary + ".exe"
		if _, err := os.Stat(cmdBinaryExe); err == nil {
			return cmdBinaryExe
		}
	}

	// Fallback locations
	candidates := []string{
		filepath.Join(repoRoot, "out", "tools", "eac"),
	}

	// Add .exe variants on Windows
	if runtime.GOOS == "windows" {
		candidates = append(candidates,
			filepath.Join(repoRoot, "out", "tools", "eac.exe"),
		)
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

