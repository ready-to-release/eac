package logging

import (
	"sync"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"gopkg.in/yaml.v3"
)

// factoryLoggingConfig holds the parsed-once logging defaults.
var (
	factoryLoggingConfig     LoggingConfig
	factoryLoggingConfigOnce sync.Once
	factoryLoggingConfigSet  bool
)

// getFactoryLoggingDefaults returns a deep copy of the factory default LoggingConfig.
// Parsed once from embedded contracts. Returns built-in defaults if
// the embedded file does not exist or is invalid.
func getFactoryLoggingDefaults() LoggingConfig {
	factoryLoggingConfigOnce.Do(func() {
		data, err := core.FS.ReadFile(core.DefaultPath("logging.yml"))
		if err != nil {
			factoryLoggingConfig = DefaultLoggingConfig()
			factoryLoggingConfigSet = true
			return
		}

		var cfg LoggingConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			factoryLoggingConfig = DefaultLoggingConfig()
			factoryLoggingConfigSet = true
			return
		}

		factoryLoggingConfig = applyDefaults(cfg)
		factoryLoggingConfigSet = true
	})

	// Return a copy with a fresh Targets map to prevent mutation.
	return cloneLoggingConfig(factoryLoggingConfig)
}

// cloneLoggingConfig returns a copy with a fresh Targets map.
func cloneLoggingConfig(src LoggingConfig) LoggingConfig {
	clone := src
	if src.Targets != nil {
		clone.Targets = make(map[string]TargetConfig, len(src.Targets))
		for k, v := range src.Targets {
			clone.Targets[k] = v
		}
	}
	return clone
}

// ResetFactoryLoggingConfigForTesting resets the factory logging singleton.
func ResetFactoryLoggingConfigForTesting() {
	factoryLoggingConfigOnce = sync.Once{}
	factoryLoggingConfig = LoggingConfig{}
	factoryLoggingConfigSet = false
}
