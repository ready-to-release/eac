package services

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/tool"
)

// ============================================================================
// Tool Registry Adapter
// ============================================================================

// toolRegistryAdapter wraps tool.DefaultRegistry
type toolRegistryAdapter struct {
	registry *tool.DefaultRegistry
}

func (a *toolRegistryAdapter) Get(canonicalName string) (core.ToolDefinitionPort, bool) {
	t, ok := a.registry.Get(canonicalName)
	if !ok {
		return nil, false
	}
	return &toolDefinitionAdapter{t: t}, true
}

func (a *toolRegistryAdapter) GetForComponent(compType string, op core.ActionType) []core.ToolDefinitionPort {
	return nil // Not implemented in tool.DefaultRegistry
}

func (a *toolRegistryAdapter) Has(canonicalName string) bool {
	return a.registry.Has(canonicalName)
}

func (a *toolRegistryAdapter) AllCanonical() []string {
	all := a.registry.GetAll()
	result := make([]string, 0, len(all))
	for id := range all {
		result = append(result, id)
	}
	return result
}

type toolDefinitionAdapter struct {
	t *tool.ToolDefinition
}

func (a *toolDefinitionAdapter) GetID() string          { return a.t.ID }
func (a *toolDefinitionAdapter) GetCanonicalID() string { return a.t.ID }
func (a *toolDefinitionAdapter) GetDescription() string { return a.t.Description }
func (a *toolDefinitionAdapter) GetType() core.ToolType {
	if a.t.Type == tool.ToolTypeSystem {
		return core.ToolTypeSystem
	}
	return core.ToolTypeContainer
}
func (a *toolDefinitionAdapter) GetBinary() string       { return a.t.Binary }
func (a *toolDefinitionAdapter) GetImage() string        { return a.t.Image }
func (a *toolDefinitionAdapter) GetCommand() []string    { return a.t.Command }
func (a *toolDefinitionAdapter) GetEntrypoint() []string { return a.t.Entrypoint }
func (a *toolDefinitionAdapter) GetEnv() map[string]string { return a.t.Env }
func (a *toolDefinitionAdapter) GetWorkDir() string      { return a.t.WorkDir }
func (a *toolDefinitionAdapter) GetResources() core.ToolResourcesPort { return nil }
func (a *toolDefinitionAdapter) GetVerify() core.ToolVerifyPort       { return nil }
func (a *toolDefinitionAdapter) DisplayName() string                        { return a.t.DisplayName() }
func (a *toolDefinitionAdapter) ShortName() string                          { return a.t.ShortName() }
