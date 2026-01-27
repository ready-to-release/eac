// Package tool provides a unified, pluggable tool composition system.
// It enables any tool (container or system binary) to be assigned to any
// component type for any operation (build, lint, scan, test, serve).
package tool

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// ToolType represents the execution type of a tool.
type ToolType string

const (
	// ToolTypeSystem represents a tool executed as a local system binary.
	ToolTypeSystem ToolType = "system"

	// ToolTypeContainer represents a tool executed in a Docker container.
	ToolTypeContainer ToolType = "container"
)

// OperationType represents the type of operation.
type OperationType string

const (
	OperationBuild OperationType = "build"
	OperationLint  OperationType = "lint"
	OperationScan  OperationType = "scan"
	OperationTest  OperationType = "test"
	OperationServe OperationType = "serve"
)

// AllOperations returns all valid operation types.
func AllOperations() []OperationType {
	return []OperationType{
		OperationBuild,
		OperationLint,
		OperationScan,
		OperationTest,
		OperationServe,
	}
}

// ToolDefinition describes any tool that can be executed.
// Resource scheduling uses existing config.Resources and WeightedSemaphore.
// Tools do NOT define their own resources - they inherit from component types.
type ToolDefinition struct {
	// Identity
	ID          string `yaml:"id" json:"id"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Execution type: "system" or "container"
	Type ToolType `yaml:"type" json:"type"`

	// System tool configuration (Type == "system")
	Binary string   `yaml:"binary,omitempty" json:"binary,omitempty"`
	Args   []string `yaml:"args,omitempty" json:"args,omitempty"`

	// Container tool configuration (Type == "container")
	// Option 1: Direct image specification
	Image string `yaml:"image,omitempty" json:"image,omitempty"`
	Tag   string `yaml:"tag,omitempty" json:"tag,omitempty"`
	// Option 2: Reference to repository.yml containers section
	Container string `yaml:"container,omitempty" json:"container,omitempty"`

	// Container execution details
	Command    []string `yaml:"command,omitempty" json:"command,omitempty"`
	Entrypoint []string `yaml:"entrypoint,omitempty" json:"entrypoint,omitempty"`

	// Execution environment
	Env     map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	WorkDir string            `yaml:"workdir,omitempty" json:"workdir,omitempty"`

	// Container-specific settings
	Mounts      []MountConfig `yaml:"mounts,omitempty" json:"mounts,omitempty"`
	Network     string        `yaml:"network,omitempty" json:"network,omitempty"`
	User        string        `yaml:"user,omitempty" json:"user,omitempty"`
	Privileged  bool          `yaml:"privileged,omitempty" json:"privileged,omitempty"`
	MemoryLimit string        `yaml:"memory_limit,omitempty" json:"memory_limit,omitempty"` // Docker memory limit (e.g., "8g")

	// Validation
	Requirements []string `yaml:"requirements,omitempty" json:"requirements,omitempty"`

	// Verification (for checking tool availability)
	Verify  *ToolVerify `yaml:"verify,omitempty" json:"verify,omitempty"`
	Version string      `yaml:"version,omitempty" json:"version,omitempty"` // Required version (e.g., ">=1.21")

	// Output handling
	OutputFormat string `yaml:"output_format,omitempty" json:"output_format,omitempty"` // "json", "text", "stream"

	// Serve-specific configuration (only for OperationServe)
	Serve *ServeConfig `yaml:"serve,omitempty" json:"serve,omitempty"`
}

// ToolVerify defines how to verify a tool is available.
type ToolVerify struct {
	// Command-based verification
	Command string `yaml:"command,omitempty" json:"command,omitempty"` // Command to run (e.g., "go version")
	Pattern string `yaml:"pattern,omitempty" json:"pattern,omitempty"` // Regex to extract version (e.g., "go version go([0-9]+\\.[0-9]+)")

	// Environment variable verification
	EnvVars []string `yaml:"env_vars,omitempty" json:"env_vars,omitempty"` // Required env vars
	Require string   `yaml:"require,omitempty" json:"require,omitempty"`   // "any" or "all" for env vars
}

// IsCommandBased returns true if this verification uses a command.
func (v *ToolVerify) IsCommandBased() bool {
	return v != nil && v.Command != ""
}

// IsEnvBased returns true if this verification uses environment variables.
func (v *ToolVerify) IsEnvBased() bool {
	return v != nil && len(v.EnvVars) > 0
}

// Clone creates a deep copy of the ToolVerify.
func (v *ToolVerify) Clone() *ToolVerify {
	if v == nil {
		return nil
	}
	clone := &ToolVerify{
		Command: v.Command,
		Pattern: v.Pattern,
		Require: v.Require,
	}
	if v.EnvVars != nil {
		clone.EnvVars = make([]string, len(v.EnvVars))
		copy(clone.EnvVars, v.EnvVars)
	}
	return clone
}

// VerifyResult contains the result of a tool verification.
type VerifyResult struct {
	ToolID          string // Tool identifier
	Available       bool   // Whether the tool is available
	Version         string // Detected version
	RequiredVersion string // Required version from config
	Error           error  // Error if verification failed
}

// IsSuccess returns true if the tool is available without errors.
func (r VerifyResult) IsSuccess() bool {
	return r.Available && r.Error == nil
}

// Validate checks if the tool definition is valid.
func (t *ToolDefinition) Validate() error {
	if t.ID == "" {
		return errors.New("tool ID is required")
	}

	switch t.Type {
	case ToolTypeSystem:
		// System tools require a binary, UNLESS they are env-var-only verification tools
		if t.Binary == "" && !t.Verify.IsEnvBased() {
			return fmt.Errorf("system tool %q requires binary", t.ID)
		}
	case ToolTypeContainer:
		// Must have either direct image or container reference
		if t.Image == "" && t.Container == "" {
			return fmt.Errorf("container tool %q requires image or container reference", t.ID)
		}
	case "":
		return fmt.Errorf("tool %q requires type (system or container)", t.ID)
	default:
		return fmt.Errorf("tool %q has invalid type: %s", t.ID, t.Type)
	}

	// Validate mounts
	for i, mount := range t.Mounts {
		if err := mount.Validate(); err != nil {
			return fmt.Errorf("tool %q mount[%d]: %w", t.ID, i, err)
		}
	}

	return nil
}

// FullImage returns the complete Docker image reference.
// Returns empty string if not a container tool or no image specified.
func (t *ToolDefinition) FullImage() string {
	if t.Image == "" {
		return ""
	}
	if t.Tag == "" {
		return t.Image + ":latest"
	}
	return t.Image + ":" + t.Tag
}

// Clone creates a deep copy of the tool definition.
func (t *ToolDefinition) Clone() *ToolDefinition {
	if t == nil {
		return nil
	}

	clone := &ToolDefinition{
		ID:           t.ID,
		Description:  t.Description,
		Type:         t.Type,
		Binary:       t.Binary,
		Image:        t.Image,
		Tag:          t.Tag,
		Container:    t.Container,
		WorkDir:      t.WorkDir,
		Network:      t.Network,
		User:         t.User,
		Privileged:   t.Privileged,
		MemoryLimit:  t.MemoryLimit,
		Version:      t.Version,
		OutputFormat: t.OutputFormat,
	}

	// Copy slices
	if t.Args != nil {
		clone.Args = make([]string, len(t.Args))
		copy(clone.Args, t.Args)
	}
	if t.Command != nil {
		clone.Command = make([]string, len(t.Command))
		copy(clone.Command, t.Command)
	}
	if t.Entrypoint != nil {
		clone.Entrypoint = make([]string, len(t.Entrypoint))
		copy(clone.Entrypoint, t.Entrypoint)
	}
	if t.Requirements != nil {
		clone.Requirements = make([]string, len(t.Requirements))
		copy(clone.Requirements, t.Requirements)
	}

	// Copy maps
	if t.Env != nil {
		clone.Env = make(map[string]string, len(t.Env))
		for k, v := range t.Env {
			clone.Env[k] = v
		}
	}

	// Copy mounts
	if t.Mounts != nil {
		clone.Mounts = make([]MountConfig, len(t.Mounts))
		copy(clone.Mounts, t.Mounts)
	}

	// Copy verify config
	if t.Verify != nil {
		clone.Verify = t.Verify.Clone()
	}

	// Copy serve config
	if t.Serve != nil {
		clone.Serve = t.Serve.Clone()
	}

	return clone
}

// MountConfig describes a volume mount for container tools.
type MountConfig struct {
	// Source is the host path or placeholder (e.g., "{workspace}", "{module}")
	Source string `yaml:"source" json:"source"`

	// Target is the container path
	Target string `yaml:"target" json:"target"`

	// ReadOnly indicates if the mount should be read-only
	ReadOnly bool `yaml:"readonly,omitempty" json:"readonly,omitempty"`

	// Type is the mount type: "bind" (default), "volume", "tmpfs"
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
}

// Validate checks if the mount configuration is valid.
func (m *MountConfig) Validate() error {
	if m.Source == "" {
		return errors.New("mount source is required")
	}
	if m.Target == "" {
		return errors.New("mount target is required")
	}
	return nil
}

// ResolvePlaceholders replaces placeholders in Source with actual paths.
func (m *MountConfig) ResolvePlaceholders(placeholders map[string]string) MountConfig {
	resolved := *m
	for placeholder, value := range placeholders {
		resolved.Source = strings.ReplaceAll(resolved.Source, placeholder, value)
		resolved.Target = strings.ReplaceAll(resolved.Target, placeholder, value)
	}
	return resolved
}

// ServeConfig extends ToolDefinition for serve operations.
// These fields only apply when the tool is used for OperationServe.
type ServeConfig struct {
	// Port Configuration
	ContainerPort int    `yaml:"container_port,omitempty" json:"container_port,omitempty"` // Port inside container
	HostPortRange string `yaml:"host_port_range,omitempty" json:"host_port_range,omitempty"` // "9000-9999"
	HostPortFixed int    `yaml:"host_port_fixed,omitempty" json:"host_port_fixed,omitempty"` // Fixed port (0 = auto)

	// Live Reload
	WatchEnabled  bool     `yaml:"watch_enabled,omitempty" json:"watch_enabled,omitempty"`
	WatchPaths    []string `yaml:"watch_paths,omitempty" json:"watch_paths,omitempty"`
	WatchExclude  []string `yaml:"watch_exclude,omitempty" json:"watch_exclude,omitempty"`
	LiveReloadURL string   `yaml:"livereload_url,omitempty" json:"livereload_url,omitempty"`

	// Process Management
	RestartPolicy  string        `yaml:"restart_policy,omitempty" json:"restart_policy,omitempty"`   // "unless-stopped", "always", "no"
	HealthCheckURL string        `yaml:"healthcheck_url,omitempty" json:"healthcheck_url,omitempty"` // Health endpoint path
	StartupDelay   time.Duration `yaml:"startup_delay,omitempty" json:"startup_delay,omitempty"`
	StopTimeout    time.Duration `yaml:"stop_timeout,omitempty" json:"stop_timeout,omitempty"`

	// Browser Integration
	AutoOpenBrowser bool   `yaml:"auto_open_browser,omitempty" json:"auto_open_browser,omitempty"`
	BrowserPath     string `yaml:"browser_path,omitempty" json:"browser_path,omitempty"` // URL path (default: "/")
}

// Clone creates a deep copy of the serve config.
func (s *ServeConfig) Clone() *ServeConfig {
	if s == nil {
		return nil
	}

	clone := &ServeConfig{
		ContainerPort:   s.ContainerPort,
		HostPortRange:   s.HostPortRange,
		HostPortFixed:   s.HostPortFixed,
		WatchEnabled:    s.WatchEnabled,
		LiveReloadURL:   s.LiveReloadURL,
		RestartPolicy:   s.RestartPolicy,
		HealthCheckURL:  s.HealthCheckURL,
		StartupDelay:    s.StartupDelay,
		StopTimeout:     s.StopTimeout,
		AutoOpenBrowser: s.AutoOpenBrowser,
		BrowserPath:     s.BrowserPath,
	}

	if s.WatchPaths != nil {
		clone.WatchPaths = make([]string, len(s.WatchPaths))
		copy(clone.WatchPaths, s.WatchPaths)
	}
	if s.WatchExclude != nil {
		clone.WatchExclude = make([]string, len(s.WatchExclude))
		copy(clone.WatchExclude, s.WatchExclude)
	}

	return clone
}

// ToolAssignment maps operations to tools for a component type.
type ToolAssignment struct {
	// ComponentType is the name of the component type (filled by resolver)
	ComponentType string `yaml:"component_type,omitempty" json:"component_type,omitempty"`

	// Operations mapped to tool IDs
	Builder string `yaml:"builder,omitempty" json:"builder,omitempty"`
	Linter  string `yaml:"linter,omitempty" json:"linter,omitempty"`
	Scanner string `yaml:"scanner,omitempty" json:"scanner,omitempty"`
	Tester  string `yaml:"tester,omitempty" json:"tester,omitempty"`
	Server  string `yaml:"server,omitempty" json:"server,omitempty"`

	// Multiple tools per operation
	Linters  []string `yaml:"linters,omitempty" json:"linters,omitempty"`
	Scanners []string `yaml:"scanners,omitempty" json:"scanners,omitempty"`
	Servers  []string `yaml:"servers,omitempty" json:"servers,omitempty"`
}

// GetToolID returns the tool ID for a given operation.
// For operations that support multiple tools (linters, scanners, servers),
// this returns the primary tool.
func (a *ToolAssignment) GetToolID(op OperationType) string {
	switch op {
	case OperationBuild:
		return a.Builder
	case OperationLint:
		return a.Linter
	case OperationScan:
		return a.Scanner
	case OperationTest:
		return a.Tester
	case OperationServe:
		return a.Server
	default:
		return ""
	}
}

// GetToolIDs returns all tool IDs for a given operation.
// For single-tool operations, returns a slice with one element.
func (a *ToolAssignment) GetToolIDs(op OperationType) []string {
	switch op {
	case OperationBuild:
		if a.Builder != "" {
			return []string{a.Builder}
		}
	case OperationLint:
		if len(a.Linters) > 0 {
			return a.Linters
		}
		if a.Linter != "" {
			return []string{a.Linter}
		}
	case OperationScan:
		if len(a.Scanners) > 0 {
			return a.Scanners
		}
		if a.Scanner != "" {
			return []string{a.Scanner}
		}
	case OperationTest:
		if a.Tester != "" {
			return []string{a.Tester}
		}
	case OperationServe:
		if len(a.Servers) > 0 {
			return a.Servers
		}
		if a.Server != "" {
			return []string{a.Server}
		}
	}
	return nil
}

// ExecutionContext provides execution parameters for tool execution.
type ExecutionContext struct {
	// Paths
	WorkspaceRoot string // Repository root
	ModuleRoot    string // Module/component root (relative to workspace)
	OutputDir     string // Output directory for artifacts

	// IO
	LogWriter io.Writer // Log output destination

	// Input context
	Files     []string      // Specific files (for lint)
	Operation OperationType // What operation is being performed

	// Runtime overrides
	EnvOverrides  map[string]string // Additional env vars
	ArgsOverrides []string          // Additional arguments

	// Placeholders for mount path resolution
	// Common placeholders: {workspace}, {module}, {output}, {go_cache}, {npm_cache}
	Placeholders map[string]string
}

// ExecutionResult captures the outcome of tool execution.
type ExecutionResult struct {
	ExitCode     int
	Stdout       []byte
	Stderr       []byte
	Duration     time.Duration
	ArtifactPath string          // Path to generated artifact
	Findings     json.RawMessage // Parsed output (if JSON)
}

// Success returns true if the tool exited with code 0.
func (r *ExecutionResult) Success() bool {
	return r.ExitCode == 0
}

// Output returns stdout if non-empty, otherwise stderr.
func (r *ExecutionResult) Output() []byte {
	if len(r.Stdout) > 0 {
		return r.Stdout
	}
	return r.Stderr
}

// CacheConfig defines a cache directory for tool execution.
type CacheConfig struct {
	Path        string `yaml:"path" json:"path"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// EnvironmentConfig defines environment-specific tool overrides.
type EnvironmentConfig struct {
	Description    string                     `yaml:"description,omitempty" json:"description,omitempty"`
	ComponentTools map[string]*ToolAssignment `yaml:"component-tools,omitempty" json:"component-tools,omitempty"`
}

// ToolConfig represents the complete tool configuration file.
type ToolConfig struct {
	// Tool definitions
	Tools map[string]*ToolDefinition `yaml:"tools,omitempty" json:"tools,omitempty"`

	// Component-to-tool assignments
	ComponentTools map[string]*ToolAssignment `yaml:"component-tools,omitempty" json:"component-tools,omitempty"`

	// Environment-specific overrides
	Environments map[string]*EnvironmentConfig `yaml:"environments,omitempty" json:"environments,omitempty"`

	// Cache configurations
	Caches map[string]*CacheConfig `yaml:"caches,omitempty" json:"caches,omitempty"`
}

// Validate checks if the tool configuration is valid.
func (c *ToolConfig) Validate() []error {
	var errs []error

	// Validate all tool definitions
	for id, tool := range c.Tools {
		if tool.ID == "" {
			tool.ID = id // Backfill ID from map key
		}
		if err := tool.Validate(); err != nil {
			errs = append(errs, err)
		}
	}

	// Validate component-tools reference valid tools
	for compType, assignment := range c.ComponentTools {
		for _, op := range AllOperations() {
			for _, toolID := range assignment.GetToolIDs(op) {
				if _, ok := c.Tools[toolID]; !ok {
					errs = append(errs, fmt.Errorf(
						"component-tools[%s].%s references unknown tool: %s",
						compType, op, toolID,
					))
				}
			}
		}
	}

	// Validate environment overrides reference valid tools
	for envName, env := range c.Environments {
		for compType, assignment := range env.ComponentTools {
			for _, op := range AllOperations() {
				for _, toolID := range assignment.GetToolIDs(op) {
					if _, ok := c.Tools[toolID]; !ok {
						errs = append(errs, fmt.Errorf(
							"environments[%s].component-tools[%s].%s references unknown tool: %s",
							envName, compType, op, toolID,
						))
					}
				}
			}
		}
	}

	return errs
}
