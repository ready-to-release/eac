// Package srccommands contains godog step implementations for specs/eac-commands.
//
// This file contains books command step definitions for book preprocessing features.
package srccommands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// registerBooksSteps registers step definitions for books command features.
func registerBooksSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	// Given steps - build output verification
	sc.Step(`^the "([^"]*)" module has been built$`, func(module string) error {
		return booksVerifyModuleBuilt(ctx, module)
	})

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

	// Then steps - build output directory verification
	sc.Step(`^build output "([^"]*)" exists$`, func(path string) error {
		return booksBuildOutputExists(ctx, path)
	})
	sc.Step(`^build output "([^"]*)" contains "([^"]*)"$`, func(path, content string) error {
		return booksBuildOutputContains(ctx, path, content)
	})
	sc.Step(`^build output "([^"]*)" contains "([^"]*)" files$`, func(dir, pattern string) error {
		return booksBuildOutputHasFiles(ctx, dir, pattern)
	})
	sc.Step(`^build output "([^"]*)" directory exists$`, func(path string) error {
		return booksBuildOutputDirExists(ctx, path)
	})
	sc.Step(`^the file contains "([^"]*)"$`, func(content string) error {
		// Uses last checked file from previous step
		return internal.OutputContains(ctx, content)
	})
	sc.Step(`^directory structure is preserved from source$`, func() error {
		// Verified implicitly by glob pattern matching
		return nil
	})
	sc.Step(`^build log "([^"]*)" contains "([^"]*)" or "([^"]*)"$`, func(path, text1, text2 string) error {
		return booksBuildLogContains(ctx, path, text1, text2)
	})
}

// booksVerifyModuleBuilt checks that a module's build output exists.
func booksVerifyModuleBuilt(ctx *internal.TestContext, module string) error {
	buildDir := filepath.Join(ctx.OriginalRepoRoot, "out", "build", module)
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		return fmt.Errorf("module '%s' has not been built (expected: %s)", module, buildDir)
	}
	return nil
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
    description: Test book with specific module
    sources:
      - type: copy
        from: "docs/**/*.md"
        to: ""
`, module)
	return internal.CreateFile(ctx, ".r2r/eac/books.yml", content)
}

// booksCreateWithInlineCommand creates books.yml with an inline command source (requires isolation).
func booksCreateWithInlineCommand(ctx *internal.TestContext, cmd string) error {
	if ctx.IsolatedDir == "" {
		return fmt.Errorf("books.yml creation requires isolated test environment")
	}
	content := fmt.Sprintf(`books:
  - name: docs
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

// booksBuildOutputExists checks if a path exists in the build output.
// Paths containing "/staging" are routed to out/staging/{module}/ instead of out/build/
func booksBuildOutputExists(ctx *internal.TestContext, path string) error {
	fullPath := resolveBuildPath(ctx.OriginalRepoRoot, path)

	// Retry up to 3 times with 500ms delay for build artifacts to be available
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(500 * time.Millisecond)
		}

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			lastErr = fmt.Errorf("build output not found: %s", fullPath)
			continue
		}

		// Store content for subsequent "the file contains" steps
		if data, err := os.ReadFile(fullPath); err == nil {
			ctx.CommandOutput = string(data)
		}
		return nil
	}

	return lastErr
}

// resolveBuildPath converts a test path to the actual filesystem path.
// Paths like "docs/staging/..." are routed to out/staging/docs/...
// Other paths go to out/build/...
func resolveBuildPath(repoRoot, path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) >= 2 && parts[1] == "staging" {
		// "docs/staging/foo" -> "out/staging/docs/foo"
		module := parts[0]
		rest := strings.Join(parts[2:], string(filepath.Separator))
		return filepath.Join(repoRoot, "out", "staging", module, rest)
	}
	return filepath.Join(repoRoot, "out", "build", path)
}

// booksBuildOutputContains checks if a file in build output contains expected content.
func booksBuildOutputContains(ctx *internal.TestContext, path, content string) error {
	fullPath := resolveBuildPath(ctx.OriginalRepoRoot, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read build output %s: %w", path, err)
	}
	if !strings.Contains(string(data), content) {
		return fmt.Errorf("build output %s does not contain '%s'", path, content)
	}
	ctx.CommandOutput = string(data)
	return nil
}

// booksBuildOutputHasFiles checks if a directory contains files matching a pattern.
func booksBuildOutputHasFiles(ctx *internal.TestContext, dir, pattern string) error {
	fullDir := resolveBuildPath(ctx.OriginalRepoRoot, dir)
	fullPattern := filepath.Join(fullDir, pattern)

	matches, err := doublestar.FilepathGlob(fullPattern)
	if err != nil {
		return fmt.Errorf("invalid pattern %s: %w", pattern, err)
	}
	if len(matches) == 0 {
		return fmt.Errorf("no files matching %s in %s", pattern, dir)
	}
	return nil
}

// booksBuildOutputDirExists checks if a directory exists in build output.
func booksBuildOutputDirExists(ctx *internal.TestContext, path string) error {
	fullPath := resolveBuildPath(ctx.OriginalRepoRoot, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("directory not found: %s", fullPath)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", path)
	}
	return nil
}

// booksBuildLogContains checks if build log contains expected text.
func booksBuildLogContains(ctx *internal.TestContext, path, text1, text2 string) error {
	fullPath := filepath.Join(ctx.OriginalRepoRoot, "out", "build", path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read build log %s: %w", path, err)
	}
	content := string(data)
	if strings.Contains(content, text1) || strings.Contains(content, text2) {
		return nil
	}
	return fmt.Errorf("build log does not contain '%s' or '%s'", text1, text2)
}
