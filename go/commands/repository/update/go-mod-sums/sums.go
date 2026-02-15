package gomodsums

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/update/internal/gowork"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

type updateGoModSumsCommand struct{}

var _ core.SimpleCommandPort = (*updateGoModSumsCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&updateGoModSumsCommand{},
	}
}

func (c *updateGoModSumsCommand) Name() string { return "update go-mod-sums" }

func (c *updateGoModSumsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "update-go-mod-sums",
		Short:         "Sync go.sum files across all workspace modules",
		Long:          "Downloads all declared dependencies and refreshes go.sum checksums\nwithout modifying go.mod files.\n\nParses go.work to discover all workspace modules, then runs 'go mod download'\nin each module directory. Also runs 'go work sync' at the repo root.\n\nUse 'update go-tidy' instead if you need to add/remove dependencies in go.mod.\n\nExamples:\n  update go-mod-sums                  Sync all go.sum files\n  update go-mod-sums --dry-run        Show modules with stale go.sum\n  update go-mod-sums --verbose        Show download output for each module",
		Flags: []core.FlagSpec{
			{Name: "dry-run", Type: "bool", DefaultValue: "false", Usage: "Show which go.sum files would change without modifying them"},
			{Name: "verbose", Shorthand: "v", Type: "bool", DefaultValue: "false", Usage: "Show go mod download output for each module"},
		},
	}
}

func (c *updateGoModSumsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return UpdateGoModSums()
}

var log = logging.C()
// UpdateGoModSums downloads dependencies and refreshes go.sum across all workspace modules.
func UpdateGoModSums() int {
	args := os.Args[3:] // Skip program name, "update", "go-mod-sums"

	dryRun := false
	verbose := false

	for _, arg := range args {
		switch arg {
		case "--dry-run":
			dryRun = true
		case "-v", "--verbose":
			verbose = true
		case "-h", "--help":
			return 0 // Registry-based help handles this
		}
	}

	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to get repository root: %v", err)
		return 1
	}

	modules, err := gowork.ParseModules(repoRoot)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	if len(modules) == 0 {
		fmt.Println("No modules found in go.work")
		return 0
	}

	var errors []string
	modified := 0

	if dryRun {
		fmt.Printf("Checking %d modules (dry run)...\n", len(modules))
	} else {
		// Run go work sync first
		fmt.Println("Syncing workspace...")
		if err := runGoWorkSync(repoRoot, verbose); err != nil {
			fmt.Printf("  ❌ go work sync: %v\n", err)
			errors = append(errors, "go work sync")
		} else {
			fmt.Println("  ✅ go work sync")
		}
		fmt.Printf("\nDownloading dependencies for %d modules...\n", len(modules))
	}

	for _, modulePath := range modules {
		relPath, relErr := filepath.Rel(repoRoot, modulePath)
		if relErr != nil {
			relPath = modulePath
		}

		if dryRun {
			changed, err := runGoModDownloadDryRun(modulePath)
			if err != nil {
				fmt.Printf("  ❌ %s: %v\n", relPath, err)
				errors = append(errors, relPath)
				continue
			}
			if changed {
				modified++
				fmt.Printf("  ⚠️  %s (go.sum would change)\n", relPath)
			} else {
				fmt.Printf("  ✅ %s\n", relPath)
			}
		} else {
			changed, err := runGoModDownload(modulePath, verbose)
			if err != nil {
				fmt.Printf("  ❌ %s: %v\n", relPath, err)
				errors = append(errors, relPath)
				continue
			}
			if changed {
				modified++
				fmt.Printf("  ⚠️  %s (go.sum updated)\n", relPath)
			} else {
				fmt.Printf("  ✅ %s\n", relPath)
			}
		}
	}

	// Summary
	fmt.Println()
	if dryRun {
		fmt.Printf("Summary: %d modules checked, %d go.sum would change, %d errors\n",
			len(modules), modified, len(errors))
	} else {
		fmt.Printf("Summary: %d modules processed, %d go.sum updated, %d errors\n",
			len(modules), modified, len(errors))
	}

	if len(errors) > 0 {
		return 1
	}
	return 0
}

// runGoWorkSync runs "go work sync" at the repo root.
func runGoWorkSync(repoRoot string, verbose bool) error {
	ts := tool.GlobalToolSystem()
	output, exitCode, err := ts.RunToolCombined(context.Background(), "go", repoRoot, "work", "sync")
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	if verbose && len(output) > 0 {
		fmt.Print(string(output))
	}
	return nil
}

// runGoModDownload runs "go mod download" and reports whether go.sum changed.
func runGoModDownload(modulePath string, verbose bool) (bool, error) {
	ts := tool.GlobalToolSystem()

	// Snapshot go.sum before (may not exist)
	goSumBefore, _ := os.ReadFile(filepath.Join(modulePath, "go.sum"))

	output, exitCode, err := ts.RunToolCombined(context.Background(), "go", modulePath, "mod", "download")
	if verbose && len(output) > 0 {
		fmt.Print(string(output))
	}
	if err != nil {
		return false, fmt.Errorf("go mod download failed: %w", err)
	}
	if exitCode != 0 {
		return false, fmt.Errorf("go mod download failed: %s", strings.TrimSpace(string(output)))
	}

	// Compare go.sum after
	goSumAfter, _ := os.ReadFile(filepath.Join(modulePath, "go.sum"))
	changed := !bytes.Equal(goSumBefore, goSumAfter)
	return changed, nil
}

// runGoModDownloadDryRun checks whether go.sum would change, then restores the original.
func runGoModDownloadDryRun(modulePath string) (bool, error) {
	ts := tool.GlobalToolSystem()
	sumPath := filepath.Join(modulePath, "go.sum")

	// Snapshot go.sum before
	goSumBefore, beforeErr := os.ReadFile(sumPath)
	existedBefore := beforeErr == nil

	output, exitCode, err := ts.RunToolCombined(context.Background(), "go", modulePath, "mod", "download")
	if err != nil {
		restoreFile(sumPath, goSumBefore, existedBefore)
		return false, fmt.Errorf("go mod download failed: %w", err)
	}
	if exitCode != 0 {
		restoreFile(sumPath, goSumBefore, existedBefore)
		return false, fmt.Errorf("go mod download failed: %s", strings.TrimSpace(string(output)))
	}

	// Compare
	goSumAfter, _ := os.ReadFile(sumPath)
	changed := !bytes.Equal(goSumBefore, goSumAfter)

	// Restore original go.sum
	restoreFile(sumPath, goSumBefore, existedBefore)

	return changed, nil
}

// restoreFile restores a file to its previous state.
func restoreFile(path string, content []byte, existed bool) {
	if existed {
		os.WriteFile(path, content, 0644) //nolint:errcheck
	} else {
		os.Remove(path) //nolint:errcheck
	}
}
