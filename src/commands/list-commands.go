// Command: list commands
// Description: List all available commands
package main

import (
	"fmt"

	"github.com/ready-to-release/eac/src/commands/render"
)

func init() {
	Register("list commands", ListCommands)
}

func ListCommands() int {
	// Get sorted command names
	var names []string
	for name := range commands {
		names = append(names, name)
	}

	// Simple alphabetical sort
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[i] > names[j] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}

	// Render as compact list
	result := render.RenderCompactList("Available Commands", names)
	fmt.Println(result)

	return 0
}
