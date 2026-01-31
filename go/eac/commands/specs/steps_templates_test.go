// Package specs contains godog step implementations for eac-commands.
package specs

import (
	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/eac/godog"
)

// registerTemplatesSteps registers step definitions for templates feature.
// Note: Common steps (command should succeed, output should contain, file should exist, etc.)
// are already registered by eacgodog.RegisterCommonSteps - only templates-specific steps here.
func registerTemplatesSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// When steps - templates-specific command execution
	sc.Step(`^I run the templates command "([^"]*)"$`, func(cmdLine string) error {
		return ctx.RunCommand(cmdLine)
	})
}
