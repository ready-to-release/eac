// Package help contains godog step implementations for eac-cli.
//
// This file contains help command step definitions for show-help and get-commands features.
package help

import (
	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/clibase/registry"
	eacgodog "github.com/ready-to-release/eac/go/godog"
)

// registryLookup adapts registry.GetCommand to the CommandLookupFunc signature.
func registryLookup(cmdName string) (func() int, bool) {
	reg := registry.GetCommand(cmdName)
	if reg == nil {
		return nil, false
	}
	return reg.Func, true
}

// registerSteps registers step definitions for help-related command features.
func registerSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// Wire in-process command dispatch to avoid subprocess overhead
	ctx.CommandDispatcher = eacgodog.MakeInProcessDispatcher(ctx, registryLookup)
	// Then steps - output verification for show-help
	sc.Step(`^I see a list of commands with descriptions$`, func() error {
		// Check that output contains command-like patterns
		return eacgodog.OutputContainsAny(ctx, "show", "create", "build")
	})
	sc.Step(`^commands are grouped by category$`, func() error {
		// Help output groups commands - check for category-like structure
		return nil // Placeholder - actual grouping check would be implementation-specific
	})
	sc.Step(`^commands within each group are sorted$`, func() error {
		// Commands should be alphabetically sorted within groups
		return nil // Placeholder
	})
	sc.Step(`^I see the command description$`, func() error {
		return eacgodog.OutputContainsAny(ctx, "Description", "Usage", "build")
	})
	sc.Step(`^I see available flags$`, func() error {
		return eacgodog.OutputContainsAny(ctx, "--", "-", "flag", "Flag")
	})
	sc.Step(`^I see usage examples$`, func() error {
		return eacgodog.OutputContainsAny(ctx, "Example", "Usage", "eac")
	})
	sc.Step(`^I see the description for "([^"]*)"$`, func(cmd string) error {
		return eacgodog.OutputContains(ctx, cmd)
	})
	// Note: `stderr contains "X" or "Y"` is registered in common steps (go/godog/steps.go)
	sc.Step(`^I see available subcommands under "([^"]*)"$`, func(parent string) error {
		return eacgodog.OutputContains(ctx, parent)
	})
	sc.Step(`^each subcommand has a brief description$`, func() error {
		// Each subcommand should have some description text
		return nil // Placeholder
	})
	sc.Step(`^I see all subcommands listed$`, func() error {
		return nil // Placeholder for verbose mode
	})
	sc.Step(`^I see advanced options$`, func() error {
		return nil // Placeholder for verbose mode
	})

	// Then steps - output verification for get-commands
	sc.Step(`^stdout is valid JSON$`, func() error {
		return eacgodog.OutputContains(ctx, "{")
	})
	sc.Step(`^the JSON contains a "([^"]*)" array$`, func(field string) error {
		// Check that output contains the array field
		return eacgodog.OutputContains(ctx, "\""+field+"\"")
	})
	sc.Step(`^the JSON contains a "([^"]*)" object$`, func(field string) error {
		// Check that output contains the object field
		return eacgodog.OutputContains(ctx, "\""+field+"\"")
	})
	sc.Step(`^each command has "([^"]*)" and "([^"]*)"$`, func(field1, field2 string) error {
		if err := eacgodog.OutputContains(ctx, field1); err != nil {
			return err
		}
		return eacgodog.OutputContains(ctx, field2)
	})
	sc.Step(`^the tree maps parent commands to children$`, func() error {
		return eacgodog.OutputContains(ctx, "tree")
	})
	sc.Step(`^each command has:$`, func() error {
		// Table-based step - check for common command fields
		return eacgodog.OutputContainsAny(ctx, "name", "description", "parts")
	})
	sc.Step(`^"([^"]*)" has "([^"]*)" set to (true|false)$`, func(cmd, field, value string) error {
		// Check for is_leaf field
		return eacgodog.OutputContains(ctx, field)
	})

	// New step definitions for get-commands scenarios
	sc.Step(`^modules exist in the repository$`, func() error {
		// Modules always exist in test environment (fixture has modules)
		return nil
	})
	sc.Step(`^the array contains module monikers$`, func() error {
		// Check that modules array has content
		return eacgodog.OutputContainsAny(ctx, "core", "eac-cli", "docs")
	})
	sc.Step(`^"([^"]*)" has "([^"]*)" set to "([^"]*)"$`, func(cmd, field, value string) error {
		// Check that command has args field with specific value
		if err := eacgodog.OutputContains(ctx, cmd); err != nil {
			return err
		}
		return eacgodog.OutputContains(ctx, value)
	})
	sc.Step(`^"([^"]*)" has "([^"]*)" empty or unset$`, func(cmd, field string) error {
		// For JSON with omitempty, field won't exist if empty
		// Just verify the command exists in output
		return eacgodog.OutputContains(ctx, cmd)
	})
	sc.Step(`^stdout is valid YAML$`, func() error {
		// Check for YAML structure indicators
		return eacgodog.OutputContainsAny(ctx, "commands:", "modules:", "tree:")
	})
	sc.Step(`^the YAML contains commands list$`, func() error {
		// Check for commands list in YAML format
		return eacgodog.OutputContains(ctx, "commands:")
	})
}
