// Package servers provides serve handlers using the pluggable tool system.
//
// All servers are defined in tool-config.yml and resolved via tool.GlobalServeBridge().
// This package provides convenience functions for accessing server handlers.
//
// Server tools are detected dynamically by their IsServable() capability
// (having a Serve configuration with container_port in tool-config.yml).
package servers

import (
	"github.com/ready-to-release/eac/go/eac/core/tool"
)

// Type aliases - delegate to tool package types.
type (
	// ServeFunc is the signature for server functions.
	ServeFunc = tool.ServeFunc

	// ServeOptions contains options for serve execution.
	ServeOptions = tool.ServeOptions

	// ServeResult captures the outcome of starting a server.
	ServeResult = tool.ServeResult
)

// GetServer returns the server function for a tool ID from tool-config.yml.
// The tool must be servable (have a Serve configuration).
func GetServer(toolID string) ServeFunc {
	return tool.GlobalServeBridge().GetServerByToolID(toolID)
}

// GetServerForComponent returns the server function for a component type.
// Uses the resolver to look up the server tool from component-tools config.
func GetServerForComponent(componentType string) ServeFunc {
	return tool.GlobalServeBridge().GetServerForComponent(componentType)
}

// HasServer checks if a server tool exists and is servable.
func HasServer(toolID string) bool {
	return tool.GlobalServeBridge().HasServerByToolID(toolID)
}

// HasServerForComponent checks if a server is available for the given component type.
func HasServerForComponent(componentType string) bool {
	return tool.GlobalServeBridge().HasServerForComponent(componentType)
}

// GetAllServableTools returns all tools that can be used as servers.
func GetAllServableTools() []*tool.ToolDefinition {
	return tool.GlobalServeBridge().GetAllServableTools()
}

// GetServerTool returns the tool definition for a component type's server.
func GetServerTool(componentType string) *tool.ToolDefinition {
	return tool.GlobalServeBridge().GetServerTool(componentType)
}

// GetServerToolByID returns the tool definition by ID if it's servable.
func GetServerToolByID(toolID string) *tool.ToolDefinition {
	return tool.GlobalServeBridge().GetServerToolByID(toolID)
}
