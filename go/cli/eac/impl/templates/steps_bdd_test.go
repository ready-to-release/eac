// Package templates contains godog step implementations for eac-cli.
package templates

import (
	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/clibase/registry"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

// registryLookup adapts registry.GetCommand to the CommandLookupFunc signature.
func registryLookup(cmdName string) (func() int, bool) {
	reg := registry.GetCommand(cmdName)
	if reg == nil {
		return nil, false
	}
	return reg.Func, true
}

// registerSteps registers step definitions for templates feature.
// Note: Common steps (command should succeed, output should contain, file should exist, etc.)
// are already registered by eacgodog.RegisterCommonSteps - only templates-specific steps here.
func registerSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// Wire in-process command dispatch to avoid subprocess overhead
	ctx.CommandDispatcher = eacgodog.MakeInProcessDispatcher(ctx, registryLookup)
	// When steps - templates-specific command execution
	sc.Step(`^I run the templates command "([^"]*)"$`, func(cmdLine string) error {
		return ctx.RunCommand(cmdLine)
	})
}
