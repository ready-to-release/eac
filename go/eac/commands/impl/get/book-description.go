// Command: get book-description
// Short: Get description for a book by filename
// Long: Returns the description for a book PDF by matching its filename.
// Long:
// Long: This replaces the jq pattern:
// Long:   echo "$BOOKS_JSON" | jq -r --arg filename "X" '.[] | select(.filename == $filename) | .description'
// Long:
// Long: Exit codes:
// Long:   0 - Book found, outputs description
// Long:   1 - Book not found or error
// Long:
// Long: Example:
// Long:   get book-description user-guide-dark.pdf    # Outputs: User Guide
// Long:   get book-description --default "Docs" X.pdf # Outputs default if not found
// Flag.default: type=string, usage=Default value if book not found
package get

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetBookDescription)
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
