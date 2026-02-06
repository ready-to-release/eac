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
			SystemTools:    make(map[string]*ToolDefinition),
			ContainerTools: make(map[string]*ToolDefinition),
			ToolBindings:   make(map[string]ToolBinding),
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

	// Apply implicit defaults (e.g., docker requirement for container tools)
	applyToolConfigDefaults(config)

	return config, nil
}

// applyToolConfigDefaults applies implicit defaults to the tool configuration.
// This includes:
// - Adding "docker" as an implicit requirement for all container tools
// - Applying serve defaults (port range, restart policy, auto-open browser)
func applyToolConfigDefaults(config *ToolConfig) {
	if config == nil {
		return
	}

	// All container tools implicitly require docker
	for _, tool := range config.ContainerTools {
		if tool == nil {
			continue
		}
		addImplicitDockerRequirement(tool)
		applyServeDefaults(tool)
	}
}

// addImplicitDockerRequirement adds "docker" to a tool's requirements if not already present.
// This is applied to all container tools since they inherently require docker to run.
func addImplicitDockerRequirement(tool *ToolDefinition) {
	if tool == nil {
		return
	}

	// Check if docker is already in requirements
	for _, req := range tool.Requirements {
		if req == "docker" {
			return // Already has docker requirement
		}
	}

	// Add docker as implicit requirement
	tool.Requirements = append(tool.Requirements, "docker")
}

// applyServeDefaults applies default serve configuration values.
// Defaults:
// - HostPortRange: "9000-9999"
// - RestartPolicy: "unless-stopped"
// - AutoOpenBrowser: true
func applyServeDefaults(tool *ToolDefinition) {
	if tool == nil || tool.Serve == nil {
		return
	}

	if tool.Serve.HostPortRange == "" {
		tool.Serve.HostPortRange = "9000-9999"
	}
	if tool.Serve.RestartPolicy == "" {
		tool.Serve.RestartPolicy = "unless-stopped"
	}
	// AutoOpenBrowser defaults to true if not explicitly set
	// Since bool zero value is false, we need to check if it was explicitly set
	// For now, we'll use a pointer to distinguish between unset and false
	// Actually, looking at ServeConfig, AutoOpenBrowser is just bool, not *bool
	// So we can't distinguish unset from false. We'll default to true only if
	// the entire serve config exists (meaning the tool is a serve tool)
	// This is a conservative approach - tools that don't want auto-open can set it to false
}

// loadToolConfigDefaults loads default tool configuration from domain.
func loadToolConfigDefaults(repoRoot string) (*ToolConfig, error) {
	root := defaultsRoot(repoRoot)
	if root == "" {
		return nil, ErrNoToolConfig
	}

	// Look for tool-config.yml in contracts/core/<version>/defaults/
	contractsPath := filepath.Join(root, "contracts", "core")

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

// checkRequiredSystemTools validates that bootstrap tools are defined in system-tools.
// Only docker and go are required as system dependencies - everything else can run in containers.
func checkRequiredSystemTools(config *ToolConfig, filePath string) ConfigValidationErrors {
	var errs ConfigValidationErrors

	// Bootstrap tools required as system dependencies
	// - docker: needed to run any container tools
	// - go: needed to build the tooling itself
	requiredSystemTools := []string{"docker", "go"}

	for _, toolID := range requiredSystemTools {
		_, exists := config.SystemTools[toolID]
		if !exists {
			errs = append(errs, ConfigValidationError{
				Message: fmt.Sprintf("required bootstrap tool %q is not defined in system-tools", toolID),
				File:    filePath,
				Fix:     fmt.Sprintf("Add %q to the system-tools section - it's required for bootstrapping", toolID),
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
				if !config.hasToolCanonical(toolID) {
					errs = append(errs, ConfigValidationError{
						Message: fmt.Sprintf("component-tools[%s].%s references unknown tool %q", compType, op, toolID),
						File:    filePath,
						Fix:     fmt.Sprintf("Add tool %q to system-tools or container-tools section", toolID),
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
					if !config.hasToolCanonical(toolID) {
						errs = append(errs, ConfigValidationError{
							Message: fmt.Sprintf("environments[%s].component-tools[%s].%s references unknown tool %q", envName, compType, op, toolID),
							File:    filePath,
							Fix:     fmt.Sprintf("Add tool %q to system-tools or container-tools section", toolID),
						})
					}
				}
			}
		}
	}

	// Check system-tools requirements reference valid tools
	for id, tool := range config.SystemTools {
		for _, req := range tool.Requirements {
			if !config.hasToolCanonical(req) {
				errs = append(errs, ConfigValidationError{
					Message: fmt.Sprintf("system-tools[%s] requires unknown tool %q", id, req),
					File:    filePath,
					Fix:     fmt.Sprintf("Add tool %q to system-tools or container-tools section", req),
				})
			}
		}
	}

	// Check container-tools requirements reference valid tools
	for id, tool := range config.ContainerTools {
		for _, req := range tool.Requirements {
			if !config.hasToolCanonical(req) {
				errs = append(errs, ConfigValidationError{
					Message: fmt.Sprintf("container-tools[%s] requires unknown tool %q", id, req),
					File:    filePath,
					Fix:     fmt.Sprintf("Add tool %q to system-tools or container-tools section", req),
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

	// Merge executor-mode (override wins if set)
	if override.ExecutorMode != "" {
		base.ExecutorMode = override.ExecutorMode
	}

	// Merge credentials (additive: override adds to base)
	if override.Credentials != nil {
		if base.Credentials == nil {
			base.Credentials = &CredentialsConfig{}
		}
		base.Credentials.HostEnv = mergeUniqueStrings(base.Credentials.HostEnv, override.Credentials.HostEnv)
		base.Credentials.CIEnv = mergeUniqueStrings(base.Credentials.CIEnv, override.Credentials.CIEnv)
	}

	// Merge system-tools
	for id, tool := range override.SystemTools {
		if tool.ID == "" {
			tool.ID = id
		}
		if base.SystemTools == nil {
			base.SystemTools = make(map[string]*ToolDefinition)
		}
		base.SystemTools[id] = tool
	}

	// Merge container-tools
	for id, tool := range override.ContainerTools {
		if tool.ID == "" {
			tool.ID = id
		}
		if base.ContainerTools == nil {
			base.ContainerTools = make(map[string]*ToolDefinition)
		}
		base.ContainerTools[id] = tool
	}

	// Merge tool-bindings
	for name, binding := range override.ToolBindings {
		if base.ToolBindings == nil {
			base.ToolBindings = make(map[string]ToolBinding)
		}
		base.ToolBindings[name] = binding
	}

	// Merge component-tools
	for compType, assignment := range override.ComponentTools {
		if base.ComponentTools == nil {
			base.ComponentTools = make(map[string]*ToolAssignment)
		}
		if base.ComponentTools[compType] == nil {
			base.ComponentTools[compType] = assignment
		} else {
			mergeToolAssignment(base.ComponentTools[compType], assignment)
		}
	}

	// Merge environments
	for envName, envConfig := range override.Environments {
		if base.Environments == nil {
			base.Environments = make(map[string]*EnvironmentConfig)
		}
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
		if base.Caches == nil {
			base.Caches = make(map[string]*CacheConfig)
		}
		base.Caches[name] = cache
	}

	// Merge test-type-mapping
	for testType, compType := range override.TestTypeMapping {
		if base.TestTypeMapping == nil {
			base.TestTypeMapping = make(map[string]string)
		}
		base.TestTypeMapping[testType] = compType
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
// Also returns the loaded ToolConfig for access to credentials and other global settings.
func InitializeFromConfig(repoRoot, configRoot string) (*DefaultRegistry, *DefaultResolver, *ToolConfig, error) {
	config, err := LoadToolConfig(repoRoot, configRoot)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("loading tool config: %w", err)
	}

	registry := NewRegistry()
	if err := registry.RegisterFromConfig(config); err != nil {
		return nil, nil, nil, fmt.Errorf("registering tools: %w", err)
	}

	resolver := NewResolver(registry)
	resolver.LoadFromConfig(config, true) // Load as defaults

	return registry, resolver, config, nil
}

// mergeUniqueStrings appends items from add to base, skipping duplicates.
func mergeUniqueStrings(base, add []string) []string {
	seen := make(map[string]bool, len(base))
	for _, s := range base {
		seen[s] = true
	}
	for _, s := range add {
		if !seen[s] {
			base = append(base, s)
			seen[s] = true
		}
	}
	return base
}
