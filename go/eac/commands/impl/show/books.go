// Command: show books
// Description: List all configured books and their sources
// Short: Display all book configurations in a human-readable table
// Long: The show books command displays all books defined in books.yml,
// Long: including each book's name, description, and source counts.
// Long: Books aggregate static content with dynamically-generated content
// Long: from EAC commands for MkDocs documentation sites.
package show

import (
	"fmt"

	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ShowBooks)
}

// ShowBooks displays all configured books in a table format
func ShowBooks() int {
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot, LazyLoad: true})
	if err != nil {
		log.Errorf("failed to load config: %v", err)
		return 1
	}

	if err := cfg.LoadBooks(false); err != nil {
		log.Errorf("failed to load books: %v", err)
		return 1
	}

	if err := cfg.LoadRepository(false); err != nil {
		log.Errorf("failed to load repository: %v", err)
		return 1
	}

	if cfg.Books == nil || len(cfg.Books.Books) == 0 {
		log.Info("No books configured")
		log.Info("")
		log.Info("Create .r2r/eac/books.yml to define books.")
		return 0
	}

	// Build a reverse map: book name -> modules that reference it
	bookToModules := make(map[string][]string)
	if cfg.Repository != nil {
		for _, mod := range cfg.Repository.Modules {
			for _, bookName := range mod.Books {
				bookToModules[bookName] = append(bookToModules[bookName], mod.Moniker)
			}
		}
	}

	tb := render.NewTableBuilder().
		WithHeaders("Name", "Output", "Modules", "Description", "Copy", "Cmd", "Inline")

	for _, book := range cfg.Books.Books {
		copyCount := len(book.GetCopySources())
		cmdCount := len(book.GetCommandSources())
		inlineCount := len(book.GetInlineSources())

		description := book.Description
		if len(description) > 30 {
			description = description[:27] + "..."
		}

		// Get modules that reference this book
		modules := bookToModules[book.Name]
		moduleStr := "-"
		if len(modules) > 0 {
			moduleStr = modules[0]
			if len(modules) > 1 {
				moduleStr = fmt.Sprintf("%s +%d", modules[0], len(modules)-1)
			}
		}

		tb.AddRow(
			book.Name,
			book.GetOutput(),
			moduleStr,
			description,
			fmt.Sprintf("%d", copyCount),
			fmt.Sprintf("%d", cmdCount),
			fmt.Sprintf("%d", inlineCount),
		)
	}

	log.Info(tb.Build())
	return 0
}
