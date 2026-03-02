//go:build !lite

// Python test runner adapters (self-register via init).
package main

import (
	_ "github.com/ready-to-release/eac/go/adapters/behave"
	_ "github.com/ready-to-release/eac/go/adapters/pytest"
)
