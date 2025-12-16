// Package srccommands contains godog step implementations for specs/eac-commands.
package srccommands

import (
	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// registerTemplatesSteps registers step definitions for templates feature.
// Note: Common steps (command should succeed, output should contain, file should exist, etc.)
// are already registered by internal.RegisterCommonSteps - only templates-specific steps here.
func registerTemplatesSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	// When steps - templates-specific command execution
	sc.Step(`^I run the templates command "([^"]*)"$`, func(cmdLine string) error {
		return ctx.RunCommand(cmdLine)
	})
}
