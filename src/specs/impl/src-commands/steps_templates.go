// Package srccommands contains godog step implementations for specs/src-commands.
package srccommands

import (
	"os"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/specs/internal"
)

// registerTemplatesSteps registers step definitions for templates feature.
// Note: Common steps (command should succeed, output should contain, file should exist, etc.)
// are already registered by internal.RegisterCommonSteps - only templates-specific steps here.
func registerTemplatesSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	// Given steps - templates-specific setup
	sc.Step(`^I have a template directory "([^"]*)"$`, func(dirPath string) error {
		fullPath := internal.ResolvePath(ctx, dirPath)
		return os.MkdirAll(fullPath, 0755)
	})
	sc.Step(`^I have a template file "([^"]*)" with content:$`, func(filePath string, content *godog.DocString) error {
		return internal.CreateFile(ctx, filePath, content.Content)
	})
	sc.Step(`^I have a values file "([^"]*)" with:$`, func(filePath string, content *godog.DocString) error {
		return internal.CreateFile(ctx, filePath, content.Content)
	})

	// When steps - templates-specific command execution
	sc.Step(`^I run the templates command "([^"]*)"$`, func(cmdLine string) error {
		return ctx.RunCommand(cmdLine)
	})
}
