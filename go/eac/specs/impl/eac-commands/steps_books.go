// Package srccommands contains godog step implementations for specs/eac-commands.
//
// This file contains books command step definitions for book preprocessing features.
package srccommands

import (
	"fmt"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// registerBooksSteps registers step definitions for books command features.
func registerBooksSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	// Given steps - configuration manipulation (uses isolation)
	sc.Step(`^books\.yml does not exist$`, func() error {
		return booksRemoveBooksYaml(ctx)
	})
	sc.Step(`^books\.yml references module "([^"]*)"$`, func(module string) error {
		return booksCreateWithModule(ctx, module)
	})
	sc.Step(`^books\.yml has inline source with command "([^"]*)"$`, func(cmd string) error {
		return booksCreateWithInlineCommand(ctx, cmd)
	})

	// Then steps - output verification
	sc.Step(`^I should see a table with book information$`, func() error {
		return internal.OutputContainsAny(ctx, "|", "Name", "Copy", "Command")
	})
	sc.Step(`^I should see "([^"]*)" in the output$`, func(text string) error {
		return internal.OutputContains(ctx, text)
	})
}

// booksRemoveBooksYaml removes books.yml for negative tests (requires isolation).
func booksRemoveBooksYaml(ctx *internal.TestContext) error {
	if ctx.IsolatedDir == "" {
		return fmt.Errorf("books.yml removal requires isolated test environment")
	}
	return internal.RemoveFile(ctx, ".r2r/eac/books.yml")
}

// booksCreateWithModule creates books.yml referencing a specific module (requires isolation).
func booksCreateWithModule(ctx *internal.TestContext, module string) error {
	if ctx.IsolatedDir == "" {
		return fmt.Errorf("books.yml creation requires isolated test environment")
	}
	content := fmt.Sprintf(`books:
  - name: %s
    module: %s
    description: Test book with specific module
    sources:
      - type: copy
        from: "docs/**/*.md"
        to: ""
`, module, module)
	return internal.CreateFile(ctx, ".r2r/eac/books.yml", content)
}

// booksCreateWithInlineCommand creates books.yml with an inline command source (requires isolation).
func booksCreateWithInlineCommand(ctx *internal.TestContext, cmd string) error {
	if ctx.IsolatedDir == "" {
		return fmt.Errorf("books.yml creation requires isolated test environment")
	}
	content := fmt.Sprintf(`books:
  - name: docs
    module: docs
    description: Test book with inline command
    sources:
      - type: copy
        from: "docs/**/*.md"
        to: ""
      - type: inline
        target: "index.md"
        inserts:
          - marker: "test-marker"
            command: "%s"
`, cmd)
	return internal.CreateFile(ctx, ".r2r/eac/books.yml", content)
}
