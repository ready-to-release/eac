// Package srccommands contains godog step implementations for specs/eac-commands.
//
// This file contains init command step definitions.
package srccommands

import (
	"fmt"
	"os"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// registerInitSteps registers step definitions for init command features.
func registerInitSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	// Special init command steps
	// Note: Generic "I run ..." steps are handled by internal/steps.go
	sc.Step(`^I run "init" without any flags$`, func() error {
		return ctx.RunCommand("init")
	})

	// Then steps - verification
	sc.Step(`^the \.r2r/eac directory is created$`, func() error {
		return initDirExists(ctx, ".r2r/eac")
	})
	sc.Step(`^a \.r2r/eac/ai-provider\.yml file is created$`, func() error {
		return initFileExists(ctx, ".r2r/eac/ai-provider.yml")
	})
	sc.Step(`^the \.r2r/eac/ai-provider\.yml file contains "([^"]*)"$`, func(content string) error {
		return initFileContains(ctx, ".r2r/eac/ai-provider.yml", content)
	})
	sc.Step(`^stdout contains provider selection confirmation$`, func() error {
		return initOutputContainsAny(ctx, "claude", "openai", "provider", "Initialized", "gemini")
	})
	sc.Step(`^stdout contains API key instructions$`, func() error {
		return initOutputContainsAny(ctx, "API", "key", "ANTHROPIC", "OPENAI")
	})
	sc.Step(`^stdout contains link to get API key$`, func() error {
		return initOutputContainsAny(ctx, "http", "API key", "api-key", "console.anthropic", "platform.openai")
	})

	// Pre-existing configuration steps
	sc.Step(`^a \.r2r/eac/ai-provider\.yml file exists with ([^"]*)$`, func(provider string) error {
		content := fmt.Sprintf("provider: %s\n", provider)
		return internal.CreateFile(ctx, ".r2r/eac/ai-provider.yml", content)
	})
	sc.Step(`^no \.r2r/eac/ai-provider\.yml file exists$`, func() error {
		return internal.RemoveFile(ctx, ".r2r/eac/ai-provider.yml")
	})
}

// initDirExists checks if a directory exists in the isolated environment.
func initDirExists(ctx *internal.TestContext, dir string) error {
	fullPath := internal.ResolvePath(ctx, dir)
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory %s was not created", dir)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", dir)
	}
	return nil
}

// initFileExists checks if a file exists in the isolated environment.
func initFileExists(ctx *internal.TestContext, path string) error {
	fullPath := internal.ResolvePath(ctx, path)
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file %s was not created", path)
	}
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s exists but is a directory, not a file", path)
	}
	return nil
}

// initFileContains checks if a file contains expected content.
func initFileContains(ctx *internal.TestContext, path, expected string) error {
	fullPath := internal.ResolvePath(ctx, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}
	if !strings.Contains(string(content), expected) {
		return fmt.Errorf("file %s does not contain %q\nActual content:\n%s", path, expected, string(content))
	}
	return nil
}

// initOutputContainsAny checks if command output contains any of the given strings.
func initOutputContainsAny(ctx *internal.TestContext, texts ...string) error {
	output := strings.ToLower(ctx.CommandOutput)
	for _, text := range texts {
		if strings.Contains(output, strings.ToLower(text)) {
			return nil
		}
	}
	return fmt.Errorf("output does not contain any of %v\nOutput:\n%s", texts, ctx.CommandOutput)
}
