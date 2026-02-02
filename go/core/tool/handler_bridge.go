// Package tool provides a handler-tool bridge for unified execution.
package tool

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	container "github.com/ready-to-release/eac/contracts/docker-adapter/0.1.0/interfaces"
)

// HandlerToolBridge adapts tool definitions for handler execution.
// Handlers call ExecuteTool() instead of direct Docker commands.
type HandlerToolBridge struct {
	mu        sync.RWMutex
	registry  Registry
	executor  Executor
	container container.ContainerPort
}

// NewHandlerToolBridge creates a bridge with the global tool system.
func NewHandlerToolBridge() *HandlerToolBridge {
	return &HandlerToolBridge{
		registry: GlobalRegistry(),
		executor: NewExecutorWithRegistry(GlobalRegistry()),
	}
}

// SetContainer sets the container port for the bridge.
func (b *HandlerToolBridge) SetContainer(c container.ContainerPort) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.container = c
}

// ToolContext holds context for tool execution.
type ToolContext struct {
	// WorkspaceRoot is the repository root path
	WorkspaceRoot string

	// StagingDir is the preprocessed content directory
	StagingDir string

	// OutputDir is the output directory
	OutputDir string

	// ConfigPath is the generated mkdocs.yml path
	ConfigPath string

	// LogWriter receives execution logs
	LogWriter io.Writer

	// Variables for placeholder substitution
	Variables map[string]string

	// Weight is the resource multiplier for container provisioning.
	// A weight of 4 means 4x the base CPU and memory allocation.
	// Default: 1 (no amplification)
	Weight int
}

// Validate checks if the context has required fields.
func (tc *ToolContext) Validate() error {
	if tc == nil {
		return errors.New("tool context is nil")
	}
	if tc.WorkspaceRoot == "" {
		return errors.New("workspace root is required")
	}
	return nil
}

// ExecuteTool runs a container tool by name with the given context.
// Returns exit code and any error.
func (b *HandlerToolBridge) ExecuteTool(ctx context.Context, toolName string, tc *ToolContext) (int, error) {
	if err := tc.Validate(); err != nil {
		return 1, fmt.Errorf("invalid tool context: %w", err)
	}

	b.mu.RLock()
	registry := b.registry
	c := b.container
	b.mu.RUnlock()

	if registry == nil {
		return 1, errors.New("tool registry not initialized")
	}

	tool, ok := registry.Get(toolName)
	if !ok {
		return 1, fmt.Errorf("tool not found: %s", toolName)
	}

	// Get container from provider if not set
	if c == nil {
		if defaultContainerProvider != nil {
			c = defaultContainerProvider()
		}
		if c == nil {
			return 1, errors.New("container runtime not configured")
		}
	}

	// Build container configuration from tool definition
	config := b.buildContainerConfig(tool, tc)

	// Execute via container port
	result, err := c.Execute(ctx, config)
	if err != nil {
		return 1, fmt.Errorf("container execution failed: %w", err)
	}

	return result.ExitCode, nil
}

// buildContainerConfig converts tool definition to container config.
func (b *HandlerToolBridge) buildContainerConfig(tool *ToolDefinition, tc *ToolContext) *container.ContainerConfig {
	// Resolve image name
	image := resolveImage(tool)

	// Substitute placeholders in command
	cmd := make([]string, len(tool.Command))
	for i, c := range tool.Command {
		cmd[i] = substitutePlaceholders(c, tc)
	}

	// Build mounts
	var mounts []container.MountConfig
	for _, m := range tool.Mounts {
		mounts = append(mounts, container.MountConfig{
			Source:   substitutePlaceholders(m.Source, tc),
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
		})
	}

	// Build environment
	env := make(map[string]string)
	for k, v := range tool.Env {
		env[k] = v
	}

	// Build config
	config := &container.ContainerConfig{
		Image:      image,
		Command:    cmd,
		Mounts:     mounts,
		Env:        env,
		WorkingDir: tool.WorkDir,
		LogWriter:  tc.LogWriter,
	}

	// Add resources if specified, applying weight multiplier
	if tool.Resources != nil {
		weight := tc.Weight
		if weight <= 0 {
			weight = 1
		}

		// Apply weight multiplier to resources
		cpus := float64(tool.Resources.CPUs) * float64(weight)
		memory := scaleMemory(tool.Resources.Memory, weight)
		shmSize := scaleMemory(tool.Resources.ShmSize, weight)

		config.Resources = &container.ResourceConfig{
			CPUs:    cpus,
			Memory:  memory,
			ShmSize: shmSize,
		}
	}

	return config
}

// scaleMemory multiplies a memory string (e.g., "2g", "512m") by a weight factor.
// Returns the scaled memory string, or original if parsing fails.
func scaleMemory(mem string, weight int) string {
	if mem == "" || weight <= 1 {
		return mem
	}

	// Parse memory value and unit
	mem = strings.TrimSpace(mem)
	if len(mem) == 0 {
		return mem
	}

	// Extract numeric part and unit
	var value int64
	var unit string
	for i, c := range mem {
		if c < '0' || c > '9' {
			var err error
			value, err = parseInt64(mem[:i])
			if err != nil {
				return mem
			}
			unit = strings.ToLower(mem[i:])
			break
		}
		if i == len(mem)-1 {
			// All digits, no unit
			var err error
			value, err = parseInt64(mem)
			if err != nil {
				return mem
			}
		}
	}

	// Scale the value
	scaled := value * int64(weight)

	return fmt.Sprintf("%d%s", scaled, unit)
}

// parseInt64 parses a string to int64.
func parseInt64(s string) (int64, error) {
	var result int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid digit: %c", c)
		}
		result = result*10 + int64(c-'0')
	}
	return result, nil
}

// resolveImage determines the Docker image to use for the tool.
func resolveImage(tool *ToolDefinition) string {
	// Local container: use {dirname}:local
	if tool.LocalPath != "" {
		return tool.LocalImageTag()
	}

	// External image with tag
	if tool.Image != "" && tool.Tag != "" {
		return fmt.Sprintf("%s:%s", tool.Image, tool.Tag)
	}

	// External image without tag (shouldn't happen if validated)
	if tool.Image != "" {
		return tool.Image + ":latest"
	}

	return ""
}

// substitutePlaceholders replaces {placeholder} with values from context.
func substitutePlaceholders(s string, tc *ToolContext) string {
	if tc == nil {
		return s
	}

	replacements := map[string]string{
		"{workspace}": tc.WorkspaceRoot,
		"{staging}":   tc.StagingDir,
		"{output}":    tc.OutputDir,
		"{config}":    tc.ConfigPath,
	}

	// Add custom variables
	for k, v := range tc.Variables {
		replacements["{"+k+"}"] = v
	}

	result := s
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// Global handler tool bridge instance.
var (
	globalHandlerBridge     *HandlerToolBridge
	globalHandlerBridgeOnce sync.Once
)

// GlobalHandlerToolBridge returns the global handler tool bridge instance.
func GlobalHandlerToolBridge() *HandlerToolBridge {
	globalHandlerBridgeOnce.Do(func() {
		globalHandlerBridge = NewHandlerToolBridge()
	})
	return globalHandlerBridge
}
