package main

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/cli/eac/allcmds"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

// buildCommandRegistry creates and populates a CommandRegistry with all commands.
func buildCommandRegistry() *registry.CommandRegistry {
	reg, err := allcmds.BuildRegistry()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
		os.Exit(2)
	}
	return reg
}
