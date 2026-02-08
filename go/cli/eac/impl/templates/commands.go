package templates

import "github.com/ready-to-release/eac/go/clibase/registry"

func init() {
	registry.RegisterAll(
		Templates,
		TemplatesInstall,
	)
}
