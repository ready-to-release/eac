// Package tool provides tool lookup utilities for accessing tool configurations.
package tool

// GetToolImage returns the Docker image for a tool from tool-config.yml.
// Returns the full image reference (image:tag) or empty string if not found.
// This provides a simple way for packages to look up tool images without
// using the full bridge pattern.
func GetToolImage(toolID string) string {
	// Try to get from the build bridge's registry (it has the tool system)
	bridge := GlobalBuildBridge()
	if bridge == nil {
		return ""
	}

	bridge.mu.RLock()
	defer bridge.mu.RUnlock()

	if bridge.registry == nil {
		return ""
	}

	tool, ok := bridge.registry.Get(toolID)
	if !ok {
		return ""
	}

	return tool.FullImage()
}

// GetToolDefinition returns the full tool definition from tool-config.yml.
// Returns nil if not found.
func GetToolDefinition(toolID string) *ToolDefinition {
	bridge := GlobalBuildBridge()
	if bridge == nil {
		return nil
	}

	bridge.mu.RLock()
	defer bridge.mu.RUnlock()

	if bridge.registry == nil {
		return nil
	}

	tool, ok := bridge.registry.Get(toolID)
	if !ok {
		return nil
	}

	return tool
}

// GetToolImageWithDefault returns the Docker image for a tool, or the default if not found.
func GetToolImageWithDefault(toolID, defaultImage string) string {
	image := GetToolImage(toolID)
	if image == "" {
		return defaultImage
	}
	return image
}
