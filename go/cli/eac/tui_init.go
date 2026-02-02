// Package main TUI initialization.
// Ensures TUI factories are registered via init() functions.
package main

import (
	// Register parallel TUI for build/test/lint/scan commands.
	// This enables the visual TUI console for parallel execution visualization.
	_ "github.com/ready-to-release/eac/go/adapters/tui/parallel"

	// Register selector TUI for interactive command selection.
	// This enables the visual selector for work/get/show/validate commands.
	_ "github.com/ready-to-release/eac/go/adapters/tui/selector"
)
