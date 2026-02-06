// Command: update go-mod-sums
// Short: Sync go.sum files across all workspace modules
// Long: Downloads all declared dependencies and refreshes go.sum checksums
// Long: without modifying go.mod files.
// Long:
// Long: Parses go.work to discover all workspace modules, then runs 'go mod download'
// Long: in each module directory. Also runs 'go work sync' at the repo root.
// Long:
// Long: Use 'update go-tidy' instead if you need to add/remove dependencies in go.mod.
// Long:
// Long: Examples:
// Long:   update go-mod-sums                  Sync all go.sum files
// Long:   update go-mod-sums --dry-run        Show modules with stale go.sum
// Long:   update go-mod-sums --verbose        Show download output for each module
// Flag.dry-run: type=bool, default=false, usage=Show which go.sum files would change without modifying them
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show go mod download output for each module
package gomodsums

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/update/internal/gowork"
	"github.com/ready-to-release/eac/go/clibase/goexec"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
)

var log = logging.C()

func init() {
	registry.Register(UpdateGoModSums)
}

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
	output, exitCode, err := goexec.RunCombined(context.Background(), repoRoot, "work", "sync")
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
	// Snapshot go.sum before (may not exist)
	goSumBefore, _ := os.ReadFile(filepath.Join(modulePath, "go.sum"))

	output, exitCode, err := goexec.RunCombined(context.Background(), modulePath, "mod", "download")
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
	sumPath := filepath.Join(modulePath, "go.sum")

	// Snapshot go.sum before
	goSumBefore, beforeErr := os.ReadFile(sumPath)
	existedBefore := beforeErr == nil

	output, exitCode, err := goexec.RunCombined(context.Background(), modulePath, "mod", "download")
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
