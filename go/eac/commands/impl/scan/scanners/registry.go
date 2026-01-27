// Package scanners provides scan handlers using the pluggable tool system.
//
// All scanners are defined in tool-config.yml and resolved via tool.GlobalScanBridge().
// This package provides convenience functions for accessing scanner handlers.
package scanners

import (
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/tool"
)

// Type aliases - delegate to tool package types.
type (
	// ScannerType represents the type of security scanner.
	ScannerType = tool.ScannerType

	// ScanFunc is the signature for scanner functions.
	ScanFunc = tool.ScanFunc

	// ScanOptions contains options for scan execution.
	ScanOptions = tool.ScanOptions

	// ScanHandler is the interface for scan handlers.
	ScanHandler = tool.ScanHandler
)

// Scanner type constants - re-export from tool package for convenience.
const (
	ScannerSBOM       = tool.ScannerSBOM
	ScannerVuln       = tool.ScannerVuln
	ScannerSecrets    = tool.ScannerSecrets
	ScannerCompliance = tool.ScannerCompliance
	ScannerIaC        = tool.ScannerIaC
	ScannerSAST       = tool.ScannerSAST
	ScannerDAST       = tool.ScannerDAST
)

// GetScanner returns the scanner function for a scanner type from tool-config.yml.
func GetScanner(scannerType ScannerType) ScanFunc {
	return tool.GlobalScanBridge().GetScanner(scannerType)
}

// HasScanner checks if a scanner is available for the given type.
func HasScanner(scannerType ScannerType) bool {
	return tool.GlobalScanBridge().HasScanner(scannerType)
}

// GetAllScannerTypes returns all available scanner types from tool-config.yml.
func GetAllScannerTypes() []ScannerType {
	return tool.GlobalScanBridge().GetAllScannerTypes()
}

// GetScannersForModule returns all applicable scanners for a module's components.
// Uses component-types.yml to determine which scanners apply based on
// the component types present in the module.
func GetScannersForModule(module *modules.ModuleContract, componentTypes *config.ComponentTypesConfig) []ScannerType {
	return tool.GlobalScanBridge().GetScannersForModule(module, componentTypes)
}

// GetScannerToolID returns the tool ID mapped to a scanner type.
func GetScannerToolID(scannerType ScannerType) string {
	return tool.GlobalScanBridge().GetScannerToolID(scannerType)
}

// GetScanHandler returns a ScanHandler interface for a scanner type.
// This provides compatibility with code expecting the ScanHandler interface.
func GetScanHandler(scannerType ScannerType) ScanHandler {
	return tool.GlobalScanBridge().GetScanHandler(scannerType)
}

// ParseScannerType converts a string to ScannerType.
func ParseScannerType(s string) (ScannerType, bool) {
	return tool.ParseScannerType(s)
}
