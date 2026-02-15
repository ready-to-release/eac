// Package templates contains godog step implementations for eac.
package templates

import (
	"context"
	"os"

	"github.com/cucumber/godog"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

// registryLookup adapts the new CommandRegistryPort to the CommandLookupFunc signature.
func registryLookup(cmdName string) (func() int, bool) {
	reg := registry.Global()
	if reg == nil {
		return nil, false
	}
	cmd, ok := reg.Get(cmdName)
	if !ok {
		return nil, false
	}
	simple, ok := cmd.(core.SimpleCommandPort)
	if !ok {
		return nil, false
	}
	return func() int {
		return simple.Execute(context.Background(), &core.CommandRequest{
			Args:   os.Args[1:],
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		})
	}, true
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
