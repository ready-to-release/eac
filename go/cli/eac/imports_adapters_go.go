//go:build !lite

// Go test runner adapters (self-register via init).
package main

import (
	_ "github.com/ready-to-release/eac/go/adapters/godog"
	_ "github.com/ready-to-release/eac/go/adapters/gotest"
)
