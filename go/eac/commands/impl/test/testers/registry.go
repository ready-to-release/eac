// Package testers provides test handlers using the pluggable tool system.
//
// All testers are defined in tool-config.yml and resolved via tool.GlobalTestBridge().
// This package provides convenience functions for accessing test handlers.
package testers

import (
	"github.com/ready-to-release/eac/go/eac/core/domain/modules"
	"github.com/ready-to-release/eac/go/eac/core/tool"
)

// Type alias - delegate to tool package type.
type TestFunc = tool.TestFunc

// GetTestFunc returns the appropriate test function for a module from tool-config.yml.
// It matches module component types to test handlers.
func GetTestFunc(module *modules.ModuleContract) TestFunc {
	return tool.GlobalTestBridge().GetTestFunc(module)
}

// HasHandler checks if a handler exists for the given name.
func HasHandler(name string) bool {
	return tool.GlobalTestBridge().HasHandler(name)
}
