// Package main container initialization.
// Wires up the Docker adapter as the default container provider.
package main

import (
	"github.com/ready-to-release/eac/go/adapters/docker"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	// Wire the Docker adapter as the default container provider.
	// This enables container-based tool execution throughout the application.
	tool.SetDefaultContainerProvider(docker.GlobalContainer)
}
