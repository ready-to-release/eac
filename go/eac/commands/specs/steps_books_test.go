// Package specs contains godog step implementations for eac-commands.
//
// This file contains books command step definitions for book preprocessing features.
package specs

import (
	"fmt"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/eac/godog"
)

// registerBooksSteps registers step definitions for books command features.
func registerBooksSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
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
		return eacgodog.OutputContainsAny(ctx, "|", "Name", "Copy", "Command")
	})
	sc.Step(`^I should see "([^"]*)" in the output$`, func(text string) error {
		return eacgodog.OutputContains(ctx, text)
	})
}

// booksRemoveBooksYaml removes books.yml for negative tests (requires isolation).
func booksRemoveBooksYaml(ctx *eacgodog.TestContext) error {
	if ctx.IsolatedDir == "" {
		return fmt.Errorf("books.yml removal requires isolated test environment")
	}
	return eacgodog.RemoveFile(ctx, ".r2r/eac/books.yml")
}

// booksCreateWithModule creates books.yml with a command that references a non-existent module.
// Since books don't directly reference modules (modules reference books), we test this by
// creating a command source that would fail when the module doesn't exist.
func booksCreateWithModule(ctx *eacgodog.TestContext, module string) error {
	if ctx.IsolatedDir == "" {
		return fmt.Errorf("books.yml creation requires isolated test environment")
	}
	// Create a book with a command that requires the module to exist
	// The "show artifacts" command requires a valid module moniker
	content := fmt.Sprintf(`books:
  - name: test-book
    description: Test book referencing module via command
    sources:
      - type: command
        command: "show artifacts %s"
        target: "artifacts.md"
`, module)
	return eacgodog.CreateFile(ctx, ".r2r/eac/books.yml", content)
}

// booksCreateWithInlineCommand creates books.yml with an inline command source (requires isolation).
func booksCreateWithInlineCommand(ctx *eacgodog.TestContext, cmd string) error {
	if ctx.IsolatedDir == "" {
		return fmt.Errorf("books.yml creation requires isolated test environment")
	}
	content := fmt.Sprintf(`books:
  - name: docs
    description: Test book with inline command
    sources:
      - type: copy
        from: "docs/**/*.md"
        to: "./"
      - type: inline
        target: "index.md"
        inserts:
          - marker: "test-marker"
            command: "%s"
`, cmd)
	return eacgodog.CreateFile(ctx, ".r2r/eac/books.yml", content)
}
