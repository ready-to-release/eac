package gotidy

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/cli/eac/impl/update/internal/gowork"
	"github.com/ready-to-release/eac/go/clibase/goexec"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
)

type updateGoTidyCommand struct{}

var _ core.SimpleCommandPort = (*updateGoTidyCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&updateGoTidyCommand{},
	}
}

func (c *updateGoTidyCommand) Name() string { return "update go-tidy" }

func (c *updateGoTidyCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "update-go-tidy",
		Short:         "Run go mod tidy across all workspace modules",
		Long:          "Ensures all Go modules in the workspace have tidy go.mod and go.sum files.\n\nParses go.work to discover all workspace modules, then runs 'go mod tidy'\nin each module directory. Also runs 'go work sync' at the repo root.\n\nThis is the fix counterpart to 'validate go-tidy' (which only checks).\n\nExamples:\n  update go-tidy                  Tidy all modules\n  update go-tidy --dry-run        Show what would change without modifying files\n  update go-tidy --verbose        Show go mod tidy output for each module",
		Flags: []core.FlagSpec{
			{Name: "dry-run", Type: "bool", DefaultValue: "false", Usage: "Show what would change without modifying files"},
			{Name: "verbose", Shorthand: "v", Type: "bool", DefaultValue: "false", Usage: "Show go mod tidy output for each module"},
		},
	}
}

func (c *updateGoTidyCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return UpdateGoTidy()
}

var log = logging.C()
// UpdateGoTidy runs go mod tidy across all workspace modules.
func UpdateGoTidy() int {
	args := os.Args[3:] // Skip program name, "update", "go-tidy"

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
		fmt.Printf("\nTidying %d modules...\n", len(modules))
	}

	for _, modulePath := range modules {
		relPath, relErr := filepath.Rel(repoRoot, modulePath)
		if relErr != nil {
			relPath = modulePath
		}

		if dryRun {
			diff, err := runGoModTidyDiff(modulePath)
			if err != nil {
				fmt.Printf("  ❌ %s: %v\n", relPath, err)
				errors = append(errors, relPath)
				continue
			}
			if diff != "" {
				modified++
				fmt.Printf("  ⚠️  %s (would change)\n", relPath)
				if verbose {
					fmt.Println(gowork.IndentLines(diff, "    "))
				}
			} else {
				fmt.Printf("  ✅ %s (tidy)\n", relPath)
			}
		} else {
			changed, err := runGoModTidy(modulePath, verbose)
			if err != nil {
				fmt.Printf("  ❌ %s: %v\n", relPath, err)
				errors = append(errors, relPath)
				continue
			}
			if changed {
				modified++
				fmt.Printf("  ⚠️  %s (modified)\n", relPath)
			} else {
				fmt.Printf("  ✅ %s\n", relPath)
			}
		}
	}

	// Summary
	fmt.Println()
	if dryRun {
		fmt.Printf("Summary: %d modules checked, %d would change, %d errors\n",
			len(modules), modified, len(errors))
	} else {
		fmt.Printf("Summary: %d modules processed, %d modified, %d errors\n",
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

// runGoModTidyDiff runs "go mod tidy -diff" and returns the diff output.
func runGoModTidyDiff(modulePath string) (string, error) {
	output, exitCode, err := goexec.RunCombined(context.Background(), modulePath, "mod", "tidy", "-diff")
	if err != nil {
		return "", fmt.Errorf("go mod tidy -diff failed: %w", err)
	}

	diff := strings.TrimSpace(string(output))
	if exitCode != 0 {
		if diff != "" {
			return diff, nil // Non-zero exit with diff output means "not tidy"
		}
		return "", fmt.Errorf("go mod tidy -diff failed (exit %d)", exitCode)
	}
	return diff, nil
}

// runGoModTidy runs "go mod tidy" and reports whether files changed.
func runGoModTidy(modulePath string, verbose bool) (bool, error) {
	// Snapshot go.mod and go.sum before
	goModBefore, err := os.ReadFile(filepath.Join(modulePath, "go.mod"))
	if err != nil {
		return false, fmt.Errorf("failed to read go.mod: %w", err)
	}
	goSumBefore, _ := os.ReadFile(filepath.Join(modulePath, "go.sum")) // go.sum may not exist yet

	output, exitCode, runErr := goexec.RunCombined(context.Background(), modulePath, "mod", "tidy")
	if verbose && len(output) > 0 {
		fmt.Print(string(output))
	}
	if runErr != nil {
		return false, fmt.Errorf("go mod tidy failed: %w", runErr)
	}
	if exitCode != 0 {
		return false, fmt.Errorf("go mod tidy failed (exit %d): %s", exitCode, strings.TrimSpace(string(output)))
	}

	// Compare after
	goModAfter, _ := os.ReadFile(filepath.Join(modulePath, "go.mod"))
	goSumAfter, _ := os.ReadFile(filepath.Join(modulePath, "go.sum"))

	changed := !bytes.Equal(goModBefore, goModAfter) || !bytes.Equal(goSumBefore, goSumAfter)
	return changed, nil
}
