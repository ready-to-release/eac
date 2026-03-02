//go:build !lite

// Web/JS test runner adapters (self-register via init).
package main

import (
	_ "github.com/ready-to-release/eac/go/adapters/cucumber"
	_ "github.com/ready-to-release/eac/go/adapters/mocha"
)
