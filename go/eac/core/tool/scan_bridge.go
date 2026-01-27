// Package tool provides a scan bridge that integrates the tool system.
package tool

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

// ScannerType represents the type of security scanner.
// This mirrors internal.ScannerType to avoid import cycles between
// the tool package and scan/internal package.
type ScannerType string

const (
	ScannerSBOM       ScannerType = "sbom"
	ScannerVuln       ScannerType = "vuln"
	ScannerSecrets    ScannerType = "secrets"
	ScannerCompliance ScannerType = "compliance"
	ScannerIaC        ScannerType = "iac"
	ScannerSAST       ScannerType = "sast"
	ScannerDAST       ScannerType = "zap"
)

// ParseScannerType converts a string to ScannerType.
func ParseScannerType(s string) (ScannerType, bool) {
	switch s {
	case "sbom":
		return ScannerSBOM, true
	case "vuln":
		return ScannerVuln, true
	case "secrets":
		return ScannerSecrets, true
	case "compliance":
		return ScannerCompliance, true
	case "iac":
		return ScannerIaC, true
	case "sast":
		return ScannerSAST, true
	case "zap":
		return ScannerDAST, true
	default:
		return "", false
	}
}

// ScanFunc is the signature for scanner functions.
// Parameters: workspace root, module root, output dir, log writer, scan options
// Returns: findings (JSON-serializable), error
type ScanFunc func(workspaceRoot, moduleRoot, outputDir string, logWriter io.Writer, opts ScanOptions) (interface{}, error)

// ScanBridge provides a unified interface for resolving security scan handlers.
// All scanners are resolved from tool-config.yml definitions.
type ScanBridge struct {
	mu sync.RWMutex

	// Scanner type to tool ID mapping (from tool-config.yml)
	scannerTools map[ScannerType]string

	// Tool system integration
	registry      Registry
	resolver      *DefaultResolver
	executor      Executor
	toolSystemSet bool // tracks if SetToolSystem was called
}

// NewScanBridge creates a new scan bridge.
func NewScanBridge() *ScanBridge {
	return &ScanBridge{
		scannerTools: map[ScannerType]string{
			// Default mappings (from tool-config.yml)
			ScannerSBOM:       "trivy-sbom",
			ScannerVuln:       "trivy-vuln",
			ScannerSecrets:    "trivy-secrets",
			ScannerIaC:        "trivy-iac",
			ScannerCompliance: "trivy-compliance",
			ScannerSAST:       "semgrep-sast",
			ScannerDAST:       "zap-dast",
		},
	}
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

// SetScannerToolMapping sets a custom scanner-to-tool mapping.
// This allows overriding the default tool ID for a scanner type.
func (b *ScanBridge) SetScannerToolMapping(scannerType ScannerType, toolID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.scannerTools[scannerType] = toolID
}

// GetScannerToolID returns the tool ID mapped to a scanner type.
func (b *ScanBridge) GetScannerToolID(scannerType ScannerType) string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.scannerTools[scannerType]
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

// GetScanner returns the scanner function for a scanner type from tool-config.yml.
func (b *ScanBridge) GetScanner(scannerType ScannerType) ScanFunc {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry := b.getRegistry()
	executor := b.getExecutor()

	if registry != nil && executor != nil {
		toolID := b.scannerTools[scannerType]
		if tool, ok := registry.Get(toolID); ok {
			return b.createToolScanFuncWithExecutor(tool, executor)
		}
	}

	return nil
}

// HasScanner checks if a scanner is available for the given type.
func (b *ScanBridge) HasScanner(scannerType ScannerType) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry := b.getRegistry()
	if registry != nil {
		toolID := b.scannerTools[scannerType]
		if _, ok := registry.Get(toolID); ok {
			return true
		}
	}

	return false
}

// GetAllScannerTypes returns all available scanner types.
func (b *ScanBridge) GetAllScannerTypes() []ScannerType {
	b.mu.RLock()
	defer b.mu.RUnlock()

	scannerSet := make(map[ScannerType]bool)

	// Add scanners available via tool registry
	registry := b.getRegistry()
	if registry != nil {
		for st, toolID := range b.scannerTools {
			if _, ok := registry.Get(toolID); ok {
				scannerSet[st] = true
			}
		}
	}

	result := make([]ScannerType, 0, len(scannerSet))
	for st := range scannerSet {
		result = append(result, st)
	}
	return result
}

// GetScannersForModule returns all applicable scanners for a module's components.
// Uses component-types.yml to determine which scanners apply based on
// the component types present in the module.
func (b *ScanBridge) GetScannersForModule(module *modules.ModuleContract, componentTypes *config.ComponentTypesConfig) []ScannerType {
	if module == nil || componentTypes == nil {
		return nil
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	scannerSet := make(map[ScannerType]bool)

	for _, compName := range module.GetEnabledComponents() {
		compTypeName := module.Components.GetComponentType(compName)
		compType := componentTypes.Get(compTypeName)
		if compType == nil {
			continue
		}

		// Get scanners defined for this component type
		scanners := compType.GetScanners()
		for _, s := range scanners {
			if st, ok := ParseScannerType(s); ok {
				scannerSet[st] = true
			}
		}
	}

	result := make([]ScannerType, 0, len(scannerSet))
	for st := range scannerSet {
		result = append(result, st)
	}
	return result
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

// GetScanHandler returns a ScanHandler interface for a scanner type.
// This provides compatibility with code expecting the ScanHandler interface.
func (b *ScanBridge) GetScanHandler(scannerType ScannerType) ScanHandler {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry := b.getRegistry()
	executor := b.getExecutor()

	if registry != nil && executor != nil {
		toolID := b.scannerTools[scannerType]
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
