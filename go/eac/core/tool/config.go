package tool

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ToolConfigFileName is the config file for tool definitions.
const ToolConfigFileName = "tool-config.yml"

// ErrNoToolConfig is returned when no tool configuration exists.
var ErrNoToolConfig = errors.New("no tool configuration found")

// LoadToolConfig loads the complete tool configuration.
// It loads defaults from contracts, then merges with project-level overrides.
func LoadToolConfig(repoRoot, configRoot string) (*ToolConfig, error) {
	// Load defaults first
	config, err := loadToolConfigDefaults(repoRoot)
	if err != nil && !errors.Is(err, ErrNoToolConfig) {
		return nil, err
	}

	if config == nil {
		config = &ToolConfig{
			Tools:          make(map[string]*ToolDefinition),
			ComponentTools: make(map[string]*ToolAssignment),
			Environments:   make(map[string]*EnvironmentConfig),
			Caches:         make(map[string]*CacheConfig),
		}
	}

	// Load project-level overrides (optional)
	overridePath := filepath.Join(configRoot, ToolConfigFileName)
	if data, err := os.ReadFile(overridePath); err == nil {
		var override ToolConfig
		if err := yaml.Unmarshal(data, &override); err != nil {
			return nil, fmt.Errorf("parsing tool-config override: %w", err)
		}
		mergeToolConfig(config, &override)
	}

	return config, nil
}

// loadToolConfigDefaults loads default tool configuration from contracts.
func loadToolConfigDefaults(repoRoot string) (*ToolConfig, error) {
	root := defaultsRoot(repoRoot)
	if root == "" {
		return nil, ErrNoToolConfig
	}

	// Look for tool-config.yml in contracts/eac-core/<version>/defaults/
	contractsPath := filepath.Join(root, "contracts", "eac-core")

	// Find the highest version
	entries, err := os.ReadDir(contractsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoToolConfig
		}
		return nil, fmt.Errorf("reading contracts directory: %w", err)
	}

	var version string
	for _, e := range entries {
		if e.IsDir() {
			version = e.Name() // Will get the last (highest) version due to alphabetical ordering
		}
	}

	if version == "" {
		return nil, ErrNoToolConfig
	}

	configPath := filepath.Join(contractsPath, version, "defaults", ToolConfigFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoToolConfig
		}
		return nil, fmt.Errorf("reading tool-config defaults: %w", err)
	}

	var config ToolConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing tool-config defaults: %w", err)
	}

	// Backfill IDs from map keys
	for id, tool := range config.Tools {
		if tool.ID == "" {
			tool.ID = id
		}
	}

	return &config, nil
}

// defaultsRoot returns the root directory for loading contract defaults.
// Container-aware: uses R2R_CONTAINER_ROOT when running in container.
func defaultsRoot(repoRoot string) string {
	if containerRoot := os.Getenv("R2R_CONTAINER_ROOT"); containerRoot != "" {
		return containerRoot
	}
	return repoRoot
}

// mergeToolConfig merges override into base config.
// Tools and assignments from override take precedence.
func mergeToolConfig(base, override *ToolConfig) {
	if override == nil {
		return
	}

	// Merge tools
	for id, tool := range override.Tools {
		if tool.ID == "" {
			tool.ID = id
		}
		base.Tools[id] = tool
	}

	// Merge component-tools
	for compType, assignment := range override.ComponentTools {
		if base.ComponentTools[compType] == nil {
			base.ComponentTools[compType] = assignment
		} else {
			mergeToolAssignment(base.ComponentTools[compType], assignment)
		}
	}

	// Merge environments
	for envName, envConfig := range override.Environments {
		if base.Environments[envName] == nil {
			base.Environments[envName] = envConfig
		} else {
			// Merge component-tools within environment
			for compType, assignment := range envConfig.ComponentTools {
				if base.Environments[envName].ComponentTools == nil {
					base.Environments[envName].ComponentTools = make(map[string]*ToolAssignment)
				}
				if base.Environments[envName].ComponentTools[compType] == nil {
					base.Environments[envName].ComponentTools[compType] = assignment
				} else {
					mergeToolAssignment(base.Environments[envName].ComponentTools[compType], assignment)
				}
			}
		}
	}

	// Merge caches
	for name, cache := range override.Caches {
		base.Caches[name] = cache
	}
}

// mergeToolAssignment merges override into base assignment.
// Non-empty fields from override take precedence.
func mergeToolAssignment(base, override *ToolAssignment) {
	if override.Builder != "" {
		base.Builder = override.Builder
	}
	if override.Linter != "" {
		base.Linter = override.Linter
	}
	if override.Scanner != "" {
		base.Scanner = override.Scanner
	}
	if override.Tester != "" {
		base.Tester = override.Tester
	}
	if override.Server != "" {
		base.Server = override.Server
	}
	if len(override.Linters) > 0 {
		base.Linters = override.Linters
	}
	if len(override.Scanners) > 0 {
		base.Scanners = override.Scanners
	}
	if len(override.Servers) > 0 {
		base.Servers = override.Servers
	}
}

// InitializeFromConfig creates and initializes a Registry and Resolver from configuration.
// This is the main entry point for using the tool system.
func InitializeFromConfig(repoRoot, configRoot string) (*DefaultRegistry, *DefaultResolver, error) {
	config, err := LoadToolConfig(repoRoot, configRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("loading tool config: %w", err)
	}

	registry := NewRegistry()
	if err := registry.RegisterFromConfig(config); err != nil {
		return nil, nil, fmt.Errorf("registering tools: %w", err)
	}

	resolver := NewResolver(registry)
	resolver.LoadFromConfig(config, true) // Load as defaults

	return registry, resolver, nil
}
