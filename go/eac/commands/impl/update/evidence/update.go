// Command: update evidence
// Short: Build evidence documentation for a module
// Long: Generates evidence PDFs from a module's evidence_books configuration.
// Long: Evidence books are markdown-based documentation packages that aggregate
// Long: test results, security scans, and other compliance artifacts.
// Long:
// Long: Unlike regular books built via 'build', evidence books are built independently
// Long: using this command, producing PDFs in out/evidence/<module>/.
// Long:
// Long: Expected Output:
// Long:   - PDF files in out/evidence/<module>/<book-name>-dark.pdf
// Long:   - One PDF per evidence book configured for the module
// Flag.all: type=bool, default=false, usage=Build evidence for all modules with evidence_books
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show detailed progress
// Usage: update evidence <module>
package evidence

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/build/builders"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

var log = logging.C()

func init() {
	registry.Register(UpdateEvidence)
}

// UpdateEvidence builds evidence documentation for modules.
func UpdateEvidence() int {
	// Parse flags
	// Skip program name, "update", and "evidence" (3 args for the two-word command)
	args := os.Args[3:]
	buildAll := false
	verbose := false
	var moduleName string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--all":
			buildAll = true
		case "-v", "--verbose":
			verbose = true
		default:
			if !strings.HasPrefix(args[i], "-") && moduleName == "" {
				moduleName = args[i]
			}
		}
	}

	// Validate arguments
	if !buildAll && moduleName == "" {
		log.Errorf("Error: module name required (or use --all)")
		log.Errorf("Usage: update evidence <module>")
		log.Errorf("       update evidence --all")
		return 1
	}

	// Get repo root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Load config
	cfg, err := config.Load(config.LoadOptions{RepoRoot: repoRoot})
	if err != nil {
		log.Errorf("Error loading config: %v", err)
		return 1
	}
	_ = cfg.LoadBooks(false) //nolint:errcheck // best-effort config load

	// Determine which modules to build
	var modulesToBuild []string
	if buildAll {
		modulesToBuild = cfg.GetModulesWithEvidenceBooks()
		if len(modulesToBuild) == 0 {
			log.Errorf("No modules have evidence_books configured")
			return 1
		}
		fmt.Printf("Building evidence for %d module(s)...\n", len(modulesToBuild))
	} else {
		// Validate module exists and has evidence_books
		evidenceBooks := cfg.GetEvidenceBooksByModule(moduleName)
		if len(evidenceBooks) == 0 {
			log.Errorf("Module '%s' has no evidence_books configured", moduleName)

			// List modules that do have evidence_books
			withEvidence := cfg.GetModulesWithEvidenceBooks()
			if len(withEvidence) > 0 {
				log.Errorf("\nModules with evidence_books:")
				for _, m := range withEvidence {
					log.Errorf("  - %s", m)
				}
			}
			return 1
		}
		modulesToBuild = []string{moduleName}
	}

	// Build evidence for each module
	var logWriter io.Writer = os.Stdout
	if !verbose {
		// In non-verbose mode, we still output key messages but suppress detailed logs
		logWriter = os.Stdout
	}

	failedModules := []string{}

	for _, moniker := range modulesToBuild {
		fmt.Printf("\n=== Building evidence: %s ===\n", moniker)

		evidenceBooks := cfg.GetEvidenceBooksByModule(moniker)

		// Use out/evidence/<module> as base - this matches the pattern of out/build/<module>
		// The builder expects: moduleOutputDir for staging, bookOutputDir for final output
		moduleOutputDir := paths.EvidenceOutputPath(repoRoot, moniker)

		// Clean and create the output directory structure
		if err := os.RemoveAll(moduleOutputDir); err != nil && !os.IsNotExist(err) {
			log.Errorf("Failed to clean output directory: %v", err)
			failedModules = append(failedModules, moniker)
			continue
		}
		if err := os.MkdirAll(moduleOutputDir, 0o755); err != nil {
			log.Errorf("Failed to create output directory: %v", err)
			failedModules = append(failedModules, moniker)
			continue
		}

		// Create a minimal module contract for the builder
		// The builder only uses this for logging and Docker config (which defaults for PDF)
		minimalModule := modules.NewModuleContract(contracts.BaseContract{
			Moniker: moniker,
			Type:    "evidence",
		}, repoRoot)

		// Build each evidence book
		for _, book := range evidenceBooks {
			fmt.Printf("  Building book: %s (%s)\n", book.Name, book.GetOutput())

			// For PDF books, output goes to a subdirectory named after the book
			// For site books, output goes directly to moduleOutputDir
			var bookOutputDir string
			if book.GetOutput() == "site" {
				bookOutputDir = moduleOutputDir
			} else {
				bookOutputDir = filepath.Join(moduleOutputDir, book.Name)
			}

			if err := os.MkdirAll(bookOutputDir, 0o755); err != nil {
				log.Errorf("Failed to create book output directory: %v", err)
				failedModules = append(failedModules, moniker)
				continue
			}

			// Debug: show paths being used
			if verbose {
				fmt.Printf("  [DEBUG] repoRoot (workspaceRoot): %s\n", repoRoot)
				fmt.Printf("  [DEBUG] moduleOutputDir: %s\n", moduleOutputDir)
				fmt.Printf("  [DEBUG] bookOutputDir: %s\n", bookOutputDir)
				fmt.Printf("  [DEBUG] Expected staging: %s\n", filepath.Join(moduleOutputDir, "staging", book.Name))
			}

			// Build the book using the exported BuildSingleBook function
			// Parameters: module, book, workspaceRoot, moduleOutputDir (for staging), bookOutputDir (for output)
			exitCode := builders.BuildSingleBook(minimalModule, book, repoRoot, moduleOutputDir, bookOutputDir, logWriter)
			if exitCode != 0 {
				log.Errorf("Failed to build evidence book '%s' for module '%s'", book.Name, moniker)
				failedModules = append(failedModules, moniker)
				continue
			}

			// Move PDF to module output root (same as buildModuleBooks does)
			bookOutput := book.GetOutput()
			if bookOutput != "site" {
				themes := []string{}
				switch bookOutput {
				case "pdf-dark":
					themes = []string{"dark"}
				case "pdf-light":
					themes = []string{"light"}
				case "pdf-all":
					themes = []string{"dark", "light"}
				}

				for _, theme := range themes {
					srcPdf := filepath.Join(bookOutputDir, "site", "pdf", fmt.Sprintf("%s-%s.pdf", book.Name, theme))
					dstPdf := filepath.Join(moduleOutputDir, fmt.Sprintf("%s-%s.pdf", book.Name, theme))
					if err := copyFile(srcPdf, dstPdf); err != nil {
						if verbose {
							fmt.Printf("  ⚠️  Could not move PDF: %v\n", err)
						}
					} else {
						os.Remove(srcPdf)
						fmt.Printf("  📄 %s-%s.pdf\n", book.Name, theme)
					}
				}
			}
		}

		fmt.Printf("Evidence output: %s\n", moduleOutputDir)
	}

	if len(failedModules) > 0 {
		log.Errorf("\n❌ Failed to build evidence for: %v", failedModules)
		return 1
	}

	fmt.Printf("\n✅ Evidence build complete\n")
	return 0
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
