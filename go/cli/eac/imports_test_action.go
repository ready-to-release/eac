//go:build !lite

// Test action command and its test runner adapters.
package main

import (
	_ "github.com/ready-to-release/eac/go/commands/test"

	// Test runner adapters (self-register via init)
	_ "github.com/ready-to-release/eac/go/adapters/behave"
	_ "github.com/ready-to-release/eac/go/adapters/cucumber"
	_ "github.com/ready-to-release/eac/go/adapters/dotnet"
	_ "github.com/ready-to-release/eac/go/adapters/godog"
	_ "github.com/ready-to-release/eac/go/adapters/gotest"
	_ "github.com/ready-to-release/eac/go/adapters/mocha"
	_ "github.com/ready-to-release/eac/go/adapters/pytest"
	_ "github.com/ready-to-release/eac/go/adapters/reqnroll"
)
