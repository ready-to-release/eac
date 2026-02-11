// Adapter imports that self-register via init()
package main

import (
	// Test runner adapters (self-register via init)
	_ "github.com/ready-to-release/eac/go/adapters/behave"    // registers BehaveRunner + behave descriptor
	_ "github.com/ready-to-release/eac/go/adapters/cucumber"
	_ "github.com/ready-to-release/eac/go/adapters/dotnet"    // registers DotnetTestRunner + dotnet descriptor
	_ "github.com/ready-to-release/eac/go/adapters/godog"     // registers godog descriptor
	_ "github.com/ready-to-release/eac/go/adapters/gotest"    // registers GoTestRunner + gotest descriptor
	_ "github.com/ready-to-release/eac/go/adapters/mocha"
	_ "github.com/ready-to-release/eac/go/adapters/pytest"    // registers PytestRunner + pytest descriptor
	_ "github.com/ready-to-release/eac/go/adapters/reqnroll"  // registers ReqnrollRunner + reqnroll descriptor
)
