package services

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
)

// environmentsAdapter wraps config.EnvironmentsConfig
type environmentsAdapter struct {
	envs *config.EnvironmentsConfig
}

func (a *environmentsAdapter) GetEnvironment(name string) (core.EnvironmentPort, bool) {
	env, ok := a.envs.GetEnvironment(name)
	if !ok {
		return nil, false
	}
	return &environmentAdapter{e: *env}, true
}

func (a *environmentsAdapter) ListEnvironments() []string {
	names := make([]string, len(a.envs.Environments))
	for i := range a.envs.Environments {
		names[i] = a.envs.Environments[i].Moniker
	}
	return names
}

type environmentAdapter struct {
	e config.Environment
}

func (a *environmentAdapter) GetName() string               { return a.e.Moniker }
func (a *environmentAdapter) GetVariables() map[string]string { return nil } // Environments don't have Variables field
