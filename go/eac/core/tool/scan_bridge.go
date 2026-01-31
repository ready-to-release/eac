// Package tool provides a scan bridge that integrates the tool system.
package tool

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/domain/modules"
)

// ScanFunc is the signature for scanner functions.
// Parameters: workspace root, module root, output dir, log writer, scan options
// Returns: findings (JSON-serializable), error
type ScanFunc func(workspaceRoot, moduleRoot, outputDir string, logWriter io.Writer, opts ScanOptions) (interface{}, error)

// ScanBridge provides a unified interface for resolving security scan handlers.
// All scanners are resolved dynamically from tool-config.yml definitions.
// Scanner tools are detected by their GetScannerCategory() capability.
type ScanBridge struct {
	mu sync.RWMutex

	// Tool system integration
	registry      Registry
	resolver      *DefaultResolver
	executor      Executor
	toolSystemSet bool // tracks if SetToolSystem was called
}

// NewScanBridge creates a new scan bridge.
func NewScanBridge() *ScanBridge {
	return &ScanBridge{}
}

// SetToolSystem configures the tool system for tool-config.yml defined scanners.
func (b *ScanBridge) SetToolSystem(registry Registry, resolver *DefaultResolver, executor Executor) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.registry = registry
	b.resolver = resolver
	b.executor = executor
	b.toolSystemSet = true
}

// getRegistry returns the registry to use, falling back to GlobalRegistry.
// Only falls back if SetToolSystem was not explicitly called.
func (b *ScanBridge) getRegistry() Registry {
	if b.registry != nil {
		return b.registry
	}
	// Only fall back to global if SetToolSystem was not explicitly called
	if !b.toolSystemSet {
		return GlobalRegistry()
	}
	return nil
}

// getExecutor returns the executor to use, falling back to GlobalExecutor.
// Only falls back if SetToolSystem was not explicitly called.
func (b *ScanBridge) getExecutor() Executor {
	if b.executor != nil {
		return b.executor
	}
	// Only fall back to global if SetToolSystem was not explicitly called
	if !b.toolSystemSet {
		return GlobalExecutor()
	}
	return nil
}

// GetScannerByToolID returns the scanner function for a tool ID from tool-config.yml.
func (b *ScanBridge) GetScannerByToolID(toolID string) ScanFunc {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry := b.getRegistry()
	executor := b.getExecutor()

	if registry != nil && executor != nil {
		if tool, ok := registry.Get(toolID); ok {
			return b.createToolScanFuncWithExecutor(tool, executor)
		}
	}

	return nil
}

// GetScannerToolByID returns the tool definition by ID.
// Returns nil if the tool doesn't exist.
func (b *ScanBridge) GetScannerToolByID(toolID string) *ToolDefinition {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry := b.getRegistry()
	if registry == nil {
		return nil
	}

	tool, ok := registry.Get(toolID)
	if !ok {
		return nil
	}

	return tool
}

// HasScannerByToolID checks if a scanner tool exists.
func (b *ScanBridge) HasScannerByToolID(toolID string) bool {
	return b.GetScannerToolByID(toolID) != nil
}

// GetScannersForComponentType returns all scanner tools assigned to a component type.
// Uses the resolver to look up scanners from component-tools config.
func (b *ScanBridge) GetScannersForComponentType(componentType string) []*ToolDefinition {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.resolver == nil {
		return nil
	}

	// Get scanner tools from resolver
	tools, err := b.resolver.ResolveMultiple(componentType, OperationScan)
	if err != nil || len(tools) == 0 {
		return nil
	}

	return tools
}

// GetAllScannerTools returns all tools that can be used as scanners.
// A scanner tool is identified by having a scanner category (via GetScannerCategory).
func (b *ScanBridge) GetAllScannerTools() []*ToolDefinition {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry := b.getRegistry()
	if registry == nil {
		return nil
	}

	var scanners []*ToolDefinition
	for _, tool := range registry.GetAll() {
		if tool.GetScannerCategory() != "" {
			scanners = append(scanners, tool)
		}
	}
	return scanners
}

// GetScannersByCategory returns all scanner tools of a specific category.
// Categories include: sbom, vuln, secrets, iac, compliance, sast, zap
func (b *ScanBridge) GetScannersByCategory(category string) []*ToolDefinition {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry := b.getRegistry()
	if registry == nil {
		return nil
	}

	var scanners []*ToolDefinition
	for _, tool := range registry.GetAll() {
		if tool.GetScannerCategory() == category {
			scanners = append(scanners, tool)
		}
	}
	return scanners
}

// GetScannersForModule returns all applicable scanner tool IDs for a module's components.
// Uses component-types.yml to determine which scanners apply based on
// the component types present in the module.
func (b *ScanBridge) GetScannersForModule(module *modules.ModuleContract, componentTypes *config.ComponentTypesConfig) []string {
	if module == nil || componentTypes == nil {
		return nil
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	scannerSet := make(map[string]bool)

	for _, compName := range module.GetEnabledComponents() {
		compTypeName := module.Components.GetComponentType(compName)
		compType := componentTypes.Get(compTypeName)
		if compType == nil {
			continue
		}

		// Get scanners defined for this component type (these are tool IDs)
		scanners := compType.GetScanners()
		for _, s := range scanners {
			scannerSet[s] = true
		}
	}

	result := make([]string, 0, len(scannerSet))
	for toolID := range scannerSet {
		result = append(result, toolID)
	}
	return result
}

// GetScannerToolsForModule returns tool definitions for all applicable scanners.
func (b *ScanBridge) GetScannerToolsForModule(module *modules.ModuleContract, componentTypes *config.ComponentTypesConfig) []*ToolDefinition {
	toolIDs := b.GetScannersForModule(module, componentTypes)
	if len(toolIDs) == 0 {
		return nil
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	registry := b.getRegistry()
	if registry == nil {
		return nil
	}

	var tools []*ToolDefinition
	for _, toolID := range toolIDs {
		if tool, ok := registry.Get(toolID); ok {
			tools = append(tools, tool)
		}
	}
	return tools
}

// createToolScanFuncWithExecutor wraps a ToolDefinition as a ScanFunc with an explicit executor.
func (b *ScanBridge) createToolScanFuncWithExecutor(tool *ToolDefinition, executor Executor) ScanFunc {
	return func(workspaceRoot, moduleRoot, outputDir string, logWriter io.Writer, opts ScanOptions) (interface{}, error) {
		adapter := NewScanHandlerAdapter(tool, executor)
		exitCode, output := adapter.Scan(moduleRoot, workspaceRoot, outputDir, logWriter, opts)
		if exitCode != 0 {
			return nil, fmt.Errorf("scanner exited with code %d", exitCode)
		}

		// Parse JSON output if available
		if len(output) > 0 {
			var findings interface{}
			if err := json.Unmarshal(output, &findings); err == nil {
				return findings, nil
			}
		}
		return nil, nil
	}
}

// GetScanHandler returns a ScanHandler interface for a tool ID.
// This provides compatibility with code expecting the ScanHandler interface.
func (b *ScanBridge) GetScanHandler(toolID string) ScanHandler {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry := b.getRegistry()
	executor := b.getExecutor()

	if registry != nil && executor != nil {
		if tool, ok := registry.Get(toolID); ok {
			return NewScanHandlerAdapter(tool, executor)
		}
	}

	return nil
}

// createExecutionContext builds an execution context for scanner tools.
func (b *ScanBridge) createExecutionContext(workspaceRoot, moduleRoot, outputDir string, logWriter io.Writer, opts ScanOptions) *ExecutionContext {
	return &ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    moduleRoot,
		OutputDir:     outputDir,
		LogWriter:     logWriter,
		Operation:     OperationScan,
		Placeholders: map[string]string{
			"{workspace}": workspaceRoot,
			"{module}":    filepath.Join(workspaceRoot, moduleRoot),
			"{output}":    outputDir,
		},
	}
}

// ResolveTool returns the tool definition for a component type and operation.
// Returns nil if no tool is configured or resolver is not available.
func (b *ScanBridge) ResolveTool(componentType string, operation OperationType) *ToolDefinition {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.resolver == nil {
		return nil
	}

	t, err := b.resolver.Resolve(componentType, operation)
	if err != nil {
		return nil
	}

	return t
}

// IsContainer returns true if the scan handler for the given tool ID runs in a container.
func (b *ScanBridge) IsContainer(toolID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry := b.getRegistry()
	if registry == nil {
		return false
	}

	if t, ok := registry.Get(toolID); ok {
		return t.Type == ToolTypeContainer
	}
	return false
}

// Global scan bridge instance.
var (
	globalScanBridge     *ScanBridge
	globalScanBridgeOnce sync.Once
)

// GlobalScanBridge returns the global scan bridge instance.
func GlobalScanBridge() *ScanBridge {
	globalScanBridgeOnce.Do(func() {
		globalScanBridge = NewScanBridge()
	})
	return globalScanBridge
}
