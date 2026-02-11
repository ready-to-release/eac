package templates

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&templatesCommand{},
		&templatesInstallCommand{},
	}
}
