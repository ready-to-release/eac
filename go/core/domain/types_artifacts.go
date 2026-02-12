package domain

import "github.com/ready-to-release/eac/go/core/config"

// --- BaseContract artifact and build methods ---
// Split from types.go to keep logical groups manageable.

// HasBuildArtifacts returns true if any component has build artifacts defined.
func (b *BaseContract) HasBuildArtifacts() bool {
	for _, comp := range b.Components {
		if comp != nil && comp.Build != nil && len(comp.Build.Artifacts) > 0 {
			return true
		}
	}
	return false
}

// GetBuildArtifacts returns all build artifacts from all components.
func (b *BaseContract) GetBuildArtifacts() []config.ModuleArtifact {
	var result []config.ModuleArtifact
	for _, comp := range b.Components {
		if comp != nil && comp.Build != nil {
			result = append(result, comp.Build.Artifacts...)
		}
	}
	return result
}

// HasExecutableArtifacts returns true if any component has executable artifacts.
func (b *BaseContract) HasExecutableArtifacts() bool {
	for _, comp := range b.Components {
		if comp != nil && comp.Build != nil {
			for _, a := range comp.Build.Artifacts {
				if a.Type == "executable" {
					return true
				}
			}
		}
	}
	return false
}

// HasTestArtifacts returns true if any component has test artifacts.
func (b *BaseContract) HasTestArtifacts() bool {
	for _, comp := range b.Components {
		if comp != nil && comp.Build != nil {
			for _, a := range comp.Build.Artifacts {
				if a.Type == "test" {
					return true
				}
			}
		}
	}
	return false
}

// GetBuildHandler returns the build handler from the first component that has one.
// This is used for module-level handler override detection.
func (b *BaseContract) GetBuildHandler() string {
	for _, comp := range b.Components {
		if comp != nil && comp.Build != nil && comp.Build.Handler != "" {
			return comp.Build.Handler
		}
	}
	return ""
}

// GetArtifactsByType returns all artifacts of the specified type from all components.
func (b *BaseContract) GetArtifactsByType(artifactType string) []config.ModuleArtifact {
	var result []config.ModuleArtifact
	for _, comp := range b.Components {
		if comp != nil && comp.Build != nil {
			for _, a := range comp.Build.Artifacts {
				if a.Type == artifactType {
					result = append(result, a)
				}
			}
		}
	}
	return result
}
