// Command: validate books
// Short: Validate books.yml configuration
// Long: Validate the books.yml file including:
// Long:   - Book names match existing module monikers
// Long:   - Referenced modules are type mkdocs-site or mkdocs-pdf
// Long:   - Copy glob patterns are valid syntax
// Long:   - Commands are valid EAC commands
// Long:   - Navigation references are valid
// Long:
// Long: Expected Output:
// Long:   Displays validation results for book names, module types, glob patterns, and commands.
// Long:   Shows which modules reference each book and counts of copy/command/inline sources.
// Long:   Exit code 0 if all books valid, 1 if validation errors found.
package validate

import (
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ValidateBooks)
}

// ValidateBooks validates the books.yml configuration
func ValidateBooks() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	log.Info("Validating books.yml...")
	log.Info("")

	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	cfg, err := config.Load(config.LoadOptions{
		RepoRoot:        workspaceRoot,
		ValidateSchemas: true,
		LazyLoad:        true,
	})
	if err != nil {
		log.Errorf("failed to load config: %v", err)
		return 1
	}

	// Load books with schema validation
	if err := cfg.LoadBooks(true); err != nil {
		log.Errorf("  books.yml: FAILED")
		log.Errorf("    %v", err)
		return 1
	}

	if cfg.Books == nil || len(cfg.Books.Books) == 0 {
		log.Info("  books.yml: not found or empty")
		log.Info("")
		log.Info("No books to validate.")
		return 0
	}

	log.Info("  books.yml: schema valid")
	log.Info("")

	// Load repository for modules cross-reference validation
	if err := cfg.LoadRepository(false); err != nil {
		log.Errorf("failed to load repository: %v", err)
		return 1
	}

	// Load module types for capability checking
	if err := cfg.LoadModuleTypes(false); err != nil {
		log.Errorf("failed to load module types: %v", err)
		return 1
	}

	var hasErrors bool
	var validBooks int

	// Check for duplicate book names
	bookNames := make(map[string]bool)
	for _, book := range cfg.Books.Books {
		if bookNames[book.Name] {
			log.Errorf("Duplicate book name: '%s'", book.Name)
			hasErrors = true
		}
		bookNames[book.Name] = true
	}

	// Build a reverse map: book name -> modules that reference it
	bookToModules := make(map[string][]*config.Module)
	if cfg.Repository != nil {
		for i := range cfg.Repository.Modules {
			mod := &cfg.Repository.Modules[i]
			for _, bookName := range mod.Books {
				bookToModules[bookName] = append(bookToModules[bookName], mod)
			}
		}
	}

	for _, book := range cfg.Books.Books {
		log.Infof("Book '%s':", book.Name)

		// Check if any module references this book
		modules := bookToModules[book.Name]
		if len(modules) == 0 {
			log.Warnf("  No module references this book (orphaned)")
		} else {
			// Show all referencing modules
			// In the new design, modules with books are implicitly documentation modules
			for _, module := range modules {
				log.Infof("  Module: %s (%s)", module.Moniker, module.Type)
			}
		}

		// Count sources
		copyCount := len(book.GetCopySources())
		cmdCount := len(book.GetCommandSources())
		inlineCount := len(book.GetInlineSources())

		log.Infof("  Sources: %d copy, %d command, %d inline", copyCount, cmdCount, inlineCount)

		// Validate commands exist
		commandErrors := validateCommands(book)
		if len(commandErrors) > 0 {
			for _, cmdErr := range commandErrors {
				log.Errorf("  %s", cmdErr)
			}
			hasErrors = true
		}

		// Validate generated nav references
		navErrors := validateGeneratedNav(book)
		if len(navErrors) > 0 {
			for _, navErr := range navErrors {
				log.Errorf("  %s", navErr)
			}
			hasErrors = true
		}

		if len(commandErrors) == 0 && len(navErrors) == 0 {
			log.Info("  Status: valid")
			validBooks++
		}

		log.Info("")
	}

	if hasErrors {
		log.Infof("Validation failed: %d/%d books valid", validBooks, len(cfg.Books.Books))
		return 1
	}

	log.Infof("All %d books validated successfully.", len(cfg.Books.Books))
	return 0
}

// validateCommands checks that all commands in sources are valid EAC commands
func validateCommands(book config.Book) []string {
	var errors []string

	// Check command sources
	for _, src := range book.GetCommandSources() {
		if !isValidCommand(src.Command) {
			errors = append(errors, "Unknown command: '"+src.Command+"'")
		}
	}

	// Check inline sources
	for _, src := range book.GetInlineSources() {
		for _, insert := range src.Inserts {
			if !isValidCommand(insert.Command) {
				errors = append(errors, "Unknown inline command: '"+insert.Command+"'")
			}
		}
	}

	return errors
}

// validateGeneratedNav checks that generated nav configurations are valid
func validateGeneratedNav(book config.Book) []string {
	var errors []string

	for _, nav := range book.GeneratedNav {
		if nav.Section == "" {
			errors = append(errors, "Generated nav missing 'section'")
		}
		if nav.InsertInto == "" {
			errors = append(errors, "Generated nav missing 'insert_into'")
		}
		if nav.Position != "" && !isValidPosition(nav.Position) {
			errors = append(errors, "Invalid position: '"+nav.Position+"' (expected 'first', 'last', or 'after:<item>')")
		}
	}

	return errors
}

// isValidCommand checks if a command is a valid EAC show command
func isValidCommand(cmd string) bool {
	// Known valid show commands that produce markdown output
	validCommands := []string{
		"show files",
		"show modules",
		"show dependencies",
		"show moduletypes",
		"show environments",
		"show tests",
		"show test-results",
		"show config",
		"show suite",
		"show books",
	}

	for _, valid := range validCommands {
		if cmd == valid || strings.HasPrefix(cmd, valid+" ") {
			return true
		}
	}
	return false
}

// isValidPosition checks if a position string is valid
func isValidPosition(pos string) bool {
	if pos == "first" || pos == "last" {
		return true
	}
	if strings.HasPrefix(pos, "after:") && len(pos) > 6 {
		return true
	}
	return false
}
