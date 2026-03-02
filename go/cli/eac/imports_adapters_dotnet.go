//go:build !lite

// .NET test runner adapters (self-register via init).
package main

import (
	_ "github.com/ready-to-release/eac/go/adapters/dotnet"
	_ "github.com/ready-to-release/eac/go/adapters/reqnroll"
)
