package get

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/repository"
)

type getBookDescriptionCommand struct{}

var _ core.SimpleCommandPort = (*getBookDescriptionCommand)(nil)

func (c *getBookDescriptionCommand) Name() string { return "get book-description" }

func (c *getBookDescriptionCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-book-description",
		Short:         "Get description for a book by filename",
		Long: "Returns the description for a book PDF by matching its filename.\n" +
			"\n" +
			"This replaces the jq pattern:\n" +
			"  echo \"$BOOKS_JSON\" | jq -r --arg filename \"X\" '.[] | select(.filename == $filename) | .description'\n" +
			"\n" +
			"Exit codes:\n" +
			"  0 - Book found, outputs description\n" +
			"  1 - Book not found or error\n" +
			"\n" +
			"Example:\n" +
			"  get book-description user-guide-dark.pdf    # Outputs: User Guide\n" +
			"  get book-description --default \"Docs\" X.pdf # Outputs default if not found",
		Flags: []core.FlagSpec{
			{Name: "default", Type: "string", Usage: "Default value if book not found"},
		},
	}
}

func (c *getBookDescriptionCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetBookDescription()
}

func GetBookDescription() int {
	// Parse arguments
	filename := ""
	defaultValue := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--default" && i+1 < len(os.Args):
			defaultValue = os.Args[i+1]
			i++
		case !strings.HasPrefix(arg, "--") && filename == "":
			filename = arg
		}
	}

	if filename == "" {
		fmt.Fprintln(os.Stderr, "Error: filename required")
		fmt.Fprintln(os.Stderr, "Usage: get book-description <filename> [--default <value>]")
		return 1
	}

	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return 1
	}

	// Strip common suffixes to match book name
	// user-guide-dark.pdf -> user-guide
	// user-guide-light.pdf -> user-guide
	// user-guide.pdf -> user-guide
	bookName := filename
	bookName = strings.TrimSuffix(bookName, ".pdf")
	bookName = strings.TrimSuffix(bookName, "-dark")
	bookName = strings.TrimSuffix(bookName, "-light")

	// Find the book by name (derived from filename)
	// PDF filenames follow pattern: {name}-dark.pdf or {name}-light.pdf
	for _, book := range cfg.Books.Books {
		if book.Name == bookName {
			if book.Description != "" {
				fmt.Println(book.Description)
			} else if book.Title != "" {
				fmt.Println(book.Title)
			} else {
				fmt.Println(book.Name)
			}
			return 0
		}
	}

	// Book not found
	if defaultValue != "" {
		fmt.Println(defaultValue)
		return 0
	}

	fmt.Fprintf(os.Stderr, "Error: book '%s' not found\n", filename)
	return 1
}
