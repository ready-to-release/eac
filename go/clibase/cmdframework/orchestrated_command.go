package cmdframework

import core "github.com/ready-to-release/eac/contracts/core/0.1.0"

// OrchestratedCommandPort extends CommandPort for commands that use the
// full cmdframework lifecycle (init, resolve, verify, execute, summary).
type OrchestratedCommandPort interface {
	core.CommandPort
	Config() *CommandConfig
	Worker() UnitWorkerFunc
	Hooks() *Hooks
}
