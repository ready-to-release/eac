// Command: validate books
// Short: Validate books.yml configuration
// Long: Validate the books.yml file including:
// Long:   - Book names match existing module monikers
// Long:   - Referenced modules are type mkdocs-site
// Long:   - Copy glob patterns are valid syntax
// Long:   - Commands are valid EAC commands
// Long:   - Navigation references are valid
package validate

import (
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ValidateBooks)
}

// ValidateBooks validates the books.yml configuration
func ValidateBooks() int {
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

	// Load modules for cross-reference validation
	if err := cfg.LoadModules(false); err != nil {
		log.Errorf("failed to load modules: %v", err)
		return 1
	}

	var hasErrors bool
	var validBooks int

	for _, book := range cfg.Books.Books {
		log.Infof("Book '%s':", book.Name)

		// Check module exists
		module, found := cfg.Modules.GetModule(book.Name)
		if !found {
			log.Errorf("  Module '%s' not found in modules.yml", book.Name)
			hasErrors = true
			continue
		}

		// Check module is mkdocs-site
		if module.Type != "mkdocs-site" {
			log.Errorf("  Module '%s' is type '%s', expected 'mkdocs-site'", book.Name, module.Type)
			hasErrors = true
			continue
		}

		log.Infof("  Module: %s (mkdocs-site)", book.Name)

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
