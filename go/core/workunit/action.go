package workunit

import core "github.com/ready-to-release/eac/contracts/core/0.1.0"

// ActionType aliases and constants re-exported from contracts/core for convenience.
type ActionType = core.ActionType

const (
	ActionBuild = core.ActionBuild
	ActionTest  = core.ActionTest
	ActionLint  = core.ActionLint
	ActionScan  = core.ActionScan
)
