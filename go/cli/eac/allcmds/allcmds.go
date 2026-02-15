// Package allcmds builds the command registry from self-registered providers.
// Command modules register themselves via init() → registry.RegisterProvider().
// Which modules are linked is controlled by build-tagged import files in go/cli/eac/.
package allcmds

import (
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

// BuildRegistry creates a CommandRegistry populated with all commands from
// registered providers and sets it as the global registry.
// Used by both main.go and BDD tests.
func BuildRegistry() (*registry.CommandRegistry, error) {
	reg := registry.NewCommandRegistry()

	for _, provider := range registry.Providers() {
		if err := reg.RegisterAll(provider()...); err != nil {
			return nil, err
		}
	}

	registry.SetGlobal(reg)
	flags.SetRegistry(reg)

	return reg, nil
}
