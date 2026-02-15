//go:build !lite

// Update commands that depend on build infrastructure.
package main

import (
	_ "github.com/ready-to-release/eac/go/commands/update/docs"
	_ "github.com/ready-to-release/eac/go/commands/update/evidence"
)
