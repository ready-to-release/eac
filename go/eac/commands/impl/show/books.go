// Command: show books
// Description: List all configured books and their sources
// Short: Display all book configurations in a human-readable table
// Long: The show books command displays all books defined in books.yml,
// Long: including each book's name, description, and source counts.
// Long: Books aggregate static content with dynamically-generated content
// Long: from EAC commands for MkDocs documentation sites.
// Long: Use --format json for machine-readable output.
// Long:
// Long: Expected Output:
// Long: - Table with columns: Name, Output, Modules, Description, Copy (count), Cmd (count), Inline (count)
// Long: - Each row shows a book with its source counts and which modules reference it
// Flag.format: type=string, default=table, usage=Output format (table, json)
// Usage: show books [--format <format>]
package show

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ShowBooks)
}

// BookInfo represents book information for JSON output
type BookInfo struct {
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Output      string   `json:"output"`
	Filename    string   `json:"filename"`
	Modules     []string `json:"modules"`
}

// ShowBooks displays all configured books in a table format
func ShowBooks() int {
	// Parse flags
	format := "table"
	args := os.Args[3:] // Skip program name, "show", and "books"
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--format" {
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		} else if len(arg) > 9 && arg[:9] == "--format=" {
			format = arg[9:]
		}
	}

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
		if format == "json" {
			fmt.Println("[]")
			return 0
		}
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

	// JSON output
	if format == "json" {
		return outputBooksJSON(cfg, bookToModules)
	}

	// Table output (default)
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

// outputBooksJSON outputs books in JSON format
func outputBooksJSON(cfg *config.EACConfig, bookToModules map[string][]string) int {
	var books []BookInfo

	for _, book := range cfg.Books.Books {
		modules := bookToModules[book.Name]
		if modules == nil {
			modules = []string{}
		}

		// Determine filename based on output format
		filename := book.Name
		output := book.GetOutput()
		if output == "site" {
			filename = "" // HTML site, no single file
		} else if output == "pdf-dark" || output == "pdf-light" {
			// PDF output: {name}-{theme}.pdf
			theme := "dark"
			if output == "pdf-light" {
				theme = "light"
			}
			filename = fmt.Sprintf("%s-%s.pdf", book.Name, theme)
		}

		books = append(books, BookInfo{
			Name:        book.Name,
			Title:       book.Title,
			Description: book.Description,
			Output:      output,
			Filename:    filename,
			Modules:     modules,
		})
	}

	jsonBytes, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		log.Errorf("failed to marshal books to JSON: %v", err)
		return 1
	}

	fmt.Println(string(jsonBytes))
	return 0
}
