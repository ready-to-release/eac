// Package servers provides serve handlers using the pluggable tool system.
//
// All servers are defined in tool-config.yml and resolved via tool.GlobalServeBridge().
// This package provides convenience functions for accessing server handlers.
package servers

import (
	"github.com/ready-to-release/eac/go/eac/core/tool"
)

// Type aliases - delegate to tool package types.
type (
	// ServerType represents the type of server.
	ServerType = tool.ServerType

	// ServeFunc is the signature for server functions.
	ServeFunc = tool.ServeFunc

	// ServeOptions contains options for serve execution.
	ServeOptions = tool.ServeOptions

	// ServeResult captures the outcome of starting a server.
	ServeResult = tool.ServeResult
)

// Server type constants - re-export from tool package for convenience.
const (
	ServerStaticSite  = tool.ServerStaticSite
	ServerMkDocsLive  = tool.ServerMkDocsLive
	ServerStructurizr = tool.ServerStructurizr
)

// GetServer returns the server function for a server type from tool-config.yml.
func GetServer(serverType ServerType) ServeFunc {
	return tool.GlobalServeBridge().GetServer(serverType)
}

// HasServer checks if a server is available for the given type.
func HasServer(serverType ServerType) bool {
	return tool.GlobalServeBridge().HasServer(serverType)
}

// GetAllServerTypes returns all available server types from tool-config.yml.
func GetAllServerTypes() []ServerType {
	return tool.GlobalServeBridge().GetAllServerTypes()
}

// GetServerToolID returns the tool ID mapped to a server type.
func GetServerToolID(serverType ServerType) string {
	return tool.GlobalServeBridge().GetServerToolID(serverType)
}

// ParseServerType converts a string to ServerType.
func ParseServerType(s string) (ServerType, bool) {
	return tool.ParseServerType(s)
}
