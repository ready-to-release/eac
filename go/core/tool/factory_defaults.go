package tool

import (
	"fmt"
	"sync"

	clone "github.com/huandu/go-clone/generic"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"gopkg.in/yaml.v3"
)

// factoryToolConfigDefaults holds the parsed-once tool config from embedded contracts.
var (
	factoryToolConfigDefaults     *ToolConfig
	factoryToolConfigDefaultsOnce sync.Once
	factoryToolConfigDefaultsErr  error
)

// getFactoryToolConfigDefaults returns the factory default ToolConfig.
// Parsed once from embedded contracts, never mutated.
func getFactoryToolConfigDefaults() (*ToolConfig, error) {
	factoryToolConfigDefaultsOnce.Do(func() {
		factoryToolConfigDefaults, factoryToolConfigDefaultsErr = parseToolConfigDefaults()
	})
	return factoryToolConfigDefaults, factoryToolConfigDefaultsErr
}

// cloneFactoryToolConfigDefaults returns a deep copy of the factory tool config.
func cloneFactoryToolConfigDefaults() (*ToolConfig, error) {
	factory, err := getFactoryToolConfigDefaults()
	if err != nil {
		return nil, err
	}
	return clone.Clone(factory), nil
}

// parseToolConfigDefaults does the actual embedded FS read + parse + validate.
// Called exactly once via sync.Once.
func parseToolConfigDefaults() (*ToolConfig, error) {
	data, err := core.FS.ReadFile(core.DefaultPath(ToolConfigFileName))
	if err != nil {
		return nil, ErrNoToolConfig
	}

	var config ToolConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing tool-config defaults: %w", err)
	}

	// Backfill IDs from map keys
	for id, tool := range config.SystemTools {
		if tool.ID == "" {
			tool.ID = id
		}
	}
	for id, tool := range config.ContainerTools {
		if tool.ID == "" {
			tool.ID = id
		}
	}

	// Validate configuration before returning
	if errs := validateToolConfigWithDuplicates(data, "embedded:"+ToolConfigFileName, &config); len(errs) > 0 {
		return nil, errs
	}

	return &config, nil
}

// ResetFactoryToolConfigForTesting resets the factory tool config singleton.
func ResetFactoryToolConfigForTesting() {
	factoryToolConfigDefaultsOnce = sync.Once{}
	factoryToolConfigDefaults = nil
	factoryToolConfigDefaultsErr = nil
}
