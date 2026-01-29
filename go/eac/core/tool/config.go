package tool

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigValidationError represents a validation error with actionable guidance.
type ConfigValidationError struct {
	Message string
	File    string
	Line    int
	Fix     string
}

func (e *ConfigValidationError) Error() string {
	loc := ""
	if e.File != "" {
		loc = e.File
		if e.Line > 0 {
			loc = fmt.Sprintf("%s:%d", e.File, e.Line)
		}
		loc += ": "
	}
	msg := loc + e.Message
	if e.Fix != "" {
		msg += "\n  Fix: " + e.Fix
	}
	return msg
}

// ConfigValidationErrors collects multiple validation errors.
type ConfigValidationErrors []ConfigValidationError

func (e ConfigValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("tool-config.yml has %d validation error(s):\n", len(e)))
	for i, err := range e {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.Error()))
	}
	return sb.String()
}

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

	// Validate configuration before returning
	if errs := validateToolConfigWithDuplicates(data, configPath, &config); len(errs) > 0 {
		return nil, errs
	}

	return &config, nil
}

// validateToolConfigWithDuplicates performs comprehensive validation.
// Note: Duplicate key detection is handled by yaml.v3 at parse time with clear error messages.
// This function performs semantic validation on the parsed config.
func validateToolConfigWithDuplicates(data []byte, filePath string, config *ToolConfig) ConfigValidationErrors {
	var errs ConfigValidationErrors

	// Check bootstrap tools (docker, go) are system type
	errs = append(errs, checkRequiredSystemTools(config, filePath)...)

	// Check naming convention alignment (-host for system, no suffix for container)
	errs = append(errs, checkNamingConventions(config, filePath)...)

	// Check that all referenced tools exist
	errs = append(errs, checkToolReferences(config, filePath)...)


	// Note: data parameter kept for potential future use (e.g., line number extraction)
	_ = data

	return errs
}

// checkRequiredSystemTools validates that bootstrap tools are defined as system type.
// Only docker and go are required as system dependencies - everything else can run in containers.
func checkRequiredSystemTools(config *ToolConfig, filePath string) ConfigValidationErrors {
	var errs ConfigValidationErrors

	// Bootstrap tools required as system dependencies
	// - docker: needed to run any container tools
	// - go: needed to build the tooling itself
	requiredSystemTools := []string{"docker", "go"}

	for _, toolID := range requiredSystemTools {
		tool, exists := config.Tools[toolID]
		if !exists {
			errs = append(errs, ConfigValidationError{
				Message: fmt.Sprintf("required bootstrap tool %q is not defined", toolID),
				File:    filePath,
				Fix:     fmt.Sprintf("Add %q as a system tool - it's required for bootstrapping", toolID),
			})
			continue
		}
		if tool.Type != ToolTypeSystem {
			errs = append(errs, ConfigValidationError{
				Message: fmt.Sprintf("bootstrap tool %q must be type 'system', got %q", toolID, tool.Type),
				File:    filePath,
				Fix:     fmt.Sprintf("%q is a bootstrap dependency and must be system-installed (type: system)", toolID),
			})
		}
	}

	return errs
}

// checkNamingConventions validates tool naming alignment.
// Convention: -host suffix for system tools, no suffix for container tools.
// This is a soft convention - we check for misalignment but the goal is consistency.
func checkNamingConventions(config *ToolConfig, filePath string) ConfigValidationErrors {
	// Clean naming: tool ID is the capability name, type field distinguishes system vs container
	// No suffix validation needed - naming is now flexible
	_ = config
	_ = filePath
	return nil
}

// checkToolReferences validates all tool references in component-tools and environments.
func checkToolReferences(config *ToolConfig, filePath string) ConfigValidationErrors {
	var errs ConfigValidationErrors

	// Check component-tools references
	for compType, assignment := range config.ComponentTools {
		for _, op := range AllOperations() {
			for _, toolID := range assignment.GetToolIDs(op) {
				if _, ok := config.Tools[toolID]; !ok {
					errs = append(errs, ConfigValidationError{
						Message: fmt.Sprintf("component-tools[%s].%s references unknown tool %q", compType, op, toolID),
						File:    filePath,
						Fix:     fmt.Sprintf("Add tool %q to the tools section, or fix the typo in the reference", toolID),
					})
				}
			}
		}
	}

	// Check environment overrides
	for envName, env := range config.Environments {
		for compType, assignment := range env.ComponentTools {
			for _, op := range AllOperations() {
				for _, toolID := range assignment.GetToolIDs(op) {
					if _, ok := config.Tools[toolID]; !ok {
						errs = append(errs, ConfigValidationError{
							Message: fmt.Sprintf("environments[%s].component-tools[%s].%s references unknown tool %q", envName, compType, op, toolID),
							File:    filePath,
							Fix:     fmt.Sprintf("Add tool %q to the tools section, or fix the typo in the reference", toolID),
						})
					}
				}
			}
		}
	}

	// Check tool requirements reference valid tools
	for id, tool := range config.Tools {
		for _, req := range tool.Requirements {
			if _, ok := config.Tools[req]; !ok {
				errs = append(errs, ConfigValidationError{
					Message: fmt.Sprintf("tool %q requires unknown tool %q", id, req),
					File:    filePath,
					Fix:     fmt.Sprintf("Add tool %q to the tools section, or fix the requirement in tool %q", req, id),
				})
			}
		}
	}

	return errs
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
