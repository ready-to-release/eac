package work

import "github.com/ready-to-release/eac/go/clibase/registry"

func init() {
	registry.RegisterAll(
		Commit,
		Create,
		CreatePR,
		Merge,
		Pull,
		Remove,
		ShowWorkspaces,
		Work,
	)
}
