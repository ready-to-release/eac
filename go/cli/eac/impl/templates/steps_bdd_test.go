// Package templates contains godog step implementations for eac-cli.
package templates

import (
	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/godog"
)

// registerSteps registers step definitions for templates feature.
// Note: Common steps (command should succeed, output should contain, file should exist, etc.)
// are already registered by eacgodog.RegisterCommonSteps - only templates-specific steps here.
func registerSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// When steps - templates-specific command execution
	sc.Step(`^I run the templates command "([^"]*)"$`, func(cmdLine string) error {
		return ctx.RunCommand(cmdLine)
	})
}
