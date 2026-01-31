// Package scanners provides scan handlers using the pluggable tool system.
//
// All scanners are defined in tool-config.yml and resolved via tool.GlobalScanBridge().
// This package provides convenience functions for accessing scanner handlers.
//
// Scanner tools are detected dynamically by their GetScannerCategory() capability
// (derived from tool ID patterns like "trivy-sbom" → "sbom" or explicit metadata).
package scanners

import (
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/domain/modules"
	"github.com/ready-to-release/eac/go/eac/core/tool"
)

// Type aliases - delegate to tool package types.
type (
	// ScanFunc is the signature for scanner functions.
	ScanFunc = tool.ScanFunc

	// ScanOptions contains options for scan execution.
	ScanOptions = tool.ScanOptions

	// ScanHandler is the interface for scan handlers.
	ScanHandler = tool.ScanHandler
)

// GetScanner returns the scanner function for a tool ID from tool-config.yml.
func GetScanner(toolID string) ScanFunc {
	return tool.GlobalScanBridge().GetScannerByToolID(toolID)
}

// HasScanner checks if a scanner tool exists.
func HasScanner(toolID string) bool {
	return tool.GlobalScanBridge().HasScannerByToolID(toolID)
}

// GetAllScannerTools returns all tools that can be used as scanners.
func GetAllScannerTools() []*tool.ToolDefinition {
	return tool.GlobalScanBridge().GetAllScannerTools()
}

// GetScannersByCategory returns all scanner tools of a specific category.
// Categories include: sbom, vuln, secrets, iac, compliance, sast, zap
func GetScannersByCategory(category string) []*tool.ToolDefinition {
	return tool.GlobalScanBridge().GetScannersByCategory(category)
}

// GetScannersForModule returns all applicable scanner tool IDs for a module's components.
// Uses component-types.yml to determine which scanners apply based on
// the component types present in the module.
func GetScannersForModule(module *modules.ModuleContract, componentTypes *config.ComponentTypesConfig) []string {
	return tool.GlobalScanBridge().GetScannersForModule(module, componentTypes)
}

// GetScannerToolsForModule returns tool definitions for all applicable scanners.
func GetScannerToolsForModule(module *modules.ModuleContract, componentTypes *config.ComponentTypesConfig) []*tool.ToolDefinition {
	return tool.GlobalScanBridge().GetScannerToolsForModule(module, componentTypes)
}

// GetScannerToolByID returns the tool definition by ID.
func GetScannerToolByID(toolID string) *tool.ToolDefinition {
	return tool.GlobalScanBridge().GetScannerToolByID(toolID)
}

// GetScanHandler returns a ScanHandler interface for a tool ID.
// This provides compatibility with code expecting the ScanHandler interface.
func GetScanHandler(toolID string) ScanHandler {
	return tool.GlobalScanBridge().GetScanHandler(toolID)
}

// IsContainer returns true if the scanner tool runs in a container.
func IsContainer(toolID string) bool {
	return tool.GlobalScanBridge().IsContainer(toolID)
}
