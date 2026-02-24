//go:build !lite

// AI provider adapters and their command dependencies.
package main

import (
	// AI provider adapters (self-register via init)
	_ "github.com/ready-to-release/eac/go/adapters/ai-test"
	_ "github.com/ready-to-release/eac/go/adapters/claude"
	_ "github.com/ready-to-release/eac/go/adapters/claude-cli"
	_ "github.com/ready-to-release/eac/go/adapters/gemini"
	_ "github.com/ready-to-release/eac/go/adapters/openai"
)
