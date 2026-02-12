// Usage: show books
package show

import (
	"context"
	"fmt"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/repository"
)

type showBooksCommand struct{}

var _ core.SimpleCommandPort = (*showBooksCommand)(nil)

func (c *showBooksCommand) Name() string { return "show books" }

func (c *showBooksCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-books",
		Short:         "Display all book configurations in a human-readable table",
		Long:          "The show books command displays all books defined in books.yml,\nincluding each book's name, description, and source counts.\nBooks aggregate static content with dynamically-generated content\nfrom EAC commands for MkDocs documentation sites.\n\nExpected Output:\n- Table with columns: Name, Output, Modules, Description, Copy (count), Cmd (count), Inline (count)\n- Each row shows a book with its source counts and which modules reference it",
	}
}

func (c *showBooksCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowBooks()
}

// ShowBooks displays all configured books in a table format.
func ShowBooks() int {
	return ExecuteShowCommand(showBooksImpl)
}

func showBooksImpl() int {
	args := os.Args[3:] // Skip program name, "show", and "books"

	// Parse flags
	format := flags.GetFlagValue(args, "--format")
	if format == "" {
		format = "table"
	}

	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot, LazyLoad: true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
		return 1
	}

	// Load repository first - books need modules for from_modules generator
	if err := cfg.LoadRepository(false); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load repository: %v\n", err)
		return 1
	}

	if err := cfg.LoadBooks(false); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load books: %v\n", err)
		return 1
	}

	if cfg.Books == nil || len(cfg.Books.Books) == 0 {
		fmt.Println("No books configured")
		fmt.Println("")
		fmt.Println("Create .eac/books.yml to define books.")
		return 0
	}

	// Build a reverse map: book name -> modules that reference it
	bookToModules := make(map[string][]string)
	if cfg.Repository != nil {
		for _, mod := range cfg.Repository.Modules {
			for _, bookName := range mod.GetBooks() {
				bookToModules[bookName] = append(bookToModules[bookName], mod.Moniker)
			}
		}
	}

	// Table output
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

	fmt.Println(tb.Build())
	return 0
}
