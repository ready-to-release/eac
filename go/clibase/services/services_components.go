package services

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
)

// componentKindsAdapter wraps config.ComponentKindsConfig
type componentKindsAdapter struct {
	kinds *config.ComponentKindsConfig
}

func (a *componentKindsAdapter) Get(name string) core.ComponentTypePort {
	ct := a.kinds.Get(name)
	if ct == nil {
		return nil
	}
	return &componentTypeAdapter{ct: ct, name: name}
}

func (a *componentKindsAdapter) List() []string {
	if a.kinds == nil || a.kinds.Kinds == nil {
		return nil
	}
	result := make([]string, 0, len(a.kinds.Kinds))
	for name := range a.kinds.Kinds {
		result = append(result, name)
	}
	return result
}

type componentTypeAdapter struct {
	ct   *config.ComponentType
	name string
}

func (a *componentTypeAdapter) GetName() string { return a.name }
func (a *componentTypeAdapter) GetPatterns() core.ComponentPatternsPort {
	if a.ct.Files == nil {
		return nil
	}
	return &componentPatternsAdapter{f: a.ct.Files}
}
func (a *componentTypeAdapter) HasBuild() bool { return a.ct.IsBuildable() }

type componentPatternsAdapter struct {
	f *config.ComponentTypeFiles
}

func (a *componentPatternsAdapter) GetSource() []string { return a.f.Source }
func (a *componentPatternsAdapter) GetTests() []string  { return a.f.Tests }
func (a *componentPatternsAdapter) GetConfig() []string { return a.f.Config }
