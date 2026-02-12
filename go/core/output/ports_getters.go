package output

import (
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// =============================================================================
// UoWManifest port implementation - implements UoWManifestPort directly
// =============================================================================

// Ensure UoWManifest implements UoWManifestPort.
var _ core.UoWManifestPort = (*UoWManifest)(nil)

// GetAction implements UoWManifestPort.GetAction.
func (m *UoWManifest) GetAction() string {
	return string(m.Action)
}

// GetModule implements UoWManifestPort.GetModule.
func (m *UoWManifest) GetModule() string {
	return m.Module
}

// GetComponent implements UoWManifestPort.GetComponent.
func (m *UoWManifest) GetComponent() string {
	return m.Component
}

// GetTool implements UoWManifestPort.GetTool.
func (m *UoWManifest) GetTool() string {
	return m.Tool
}

// GetExitCode implements UoWManifestPort.GetExitCode.
func (m *UoWManifest) GetExitCode() int {
	return m.ExitCode
}

// GetInputHash implements UoWManifestPort.GetInputHash.
func (m *UoWManifest) GetInputHash() string {
	return m.InputHash
}

// GetExecutedAt implements UoWManifestPort.GetExecutedAt.
func (m *UoWManifest) GetExecutedAt() time.Time {
	return m.ExecutedAt
}

// GetDuration implements UoWManifestPort.GetDuration.
func (m *UoWManifest) GetDuration() time.Duration {
	return m.Duration
}

// GetArtifacts implements UoWManifestPort.GetArtifacts.
func (m *UoWManifest) GetArtifacts() []core.OutputArtifactPort {
	result := make([]core.OutputArtifactPort, len(m.Artifacts))
	for i := range m.Artifacts {
		result[i] = &m.Artifacts[i]
	}
	return result
}

// GetOutputHash implements UoWManifestPort.GetOutputHash.
func (m *UoWManifest) GetOutputHash() string {
	return m.OutputHash
}

// =============================================================================
// Artifact port implementation - implements OutputArtifactPort directly
// =============================================================================

// Ensure Artifact implements OutputArtifactPort.
var _ core.OutputArtifactPort = (*Artifact)(nil)

// GetID implements OutputArtifactPort.GetID.
func (a *Artifact) GetID() string {
	return a.ID
}

// GetPath implements OutputArtifactPort.GetPath.
func (a *Artifact) GetPath() string {
	return a.Path
}

// GetSHA256 implements OutputArtifactPort.GetSHA256.
func (a *Artifact) GetSHA256() string {
	return a.SHA256
}

// GetSize implements OutputArtifactPort.GetSize.
func (a *Artifact) GetSize() int64 {
	return a.Size
}

// GetType implements OutputArtifactPort.GetType.
func (a *Artifact) GetType() string {
	return a.Type
}

// =============================================================================
// ModuleView port implementation - implements ModuleViewPort directly
// =============================================================================

// Ensure ModuleView implements ModuleViewPort.
var _ core.ModuleViewPort = (*ModuleView)(nil)

// GetModule implements ModuleViewPort.GetModule.
func (v *ModuleView) GetModule() string {
	return v.Module
}

// GetStatus implements ModuleViewPort.GetStatus.
func (v *ModuleView) GetStatus() string {
	return string(v.Status)
}

// GetComponents implements ModuleViewPort.GetComponents.
func (v *ModuleView) GetComponents() []core.ComponentViewPort {
	result := make([]core.ComponentViewPort, len(v.Components))
	for i := range v.Components {
		result[i] = &v.Components[i]
	}
	return result
}

// GetTotalSize implements ModuleViewPort.GetTotalSize.
func (v *ModuleView) GetTotalSize() int64 {
	return v.TotalSize
}

// =============================================================================
// ComponentView port implementation - implements ComponentViewPort directly
// =============================================================================

// Ensure ComponentView implements ComponentViewPort.
var _ core.ComponentViewPort = (*ComponentView)(nil)

// GetModule implements ComponentViewPort.GetModule.
func (v *ComponentView) GetModule() string {
	return v.Module
}

// GetComponent implements ComponentViewPort.GetComponent.
func (v *ComponentView) GetComponent() string {
	return v.Component
}

// GetStatus implements ComponentViewPort.GetStatus.
func (v *ComponentView) GetStatus() string {
	return string(v.Status)
}

// GetUoWs implements ComponentViewPort.GetUoWs.
func (v *ComponentView) GetUoWs() []core.UoWManifestPort {
	result := make([]core.UoWManifestPort, len(v.UoWs))
	for i := range v.UoWs {
		result[i] = &v.UoWs[i]
	}
	return result
}

// GetTotalSize implements ComponentViewPort.GetTotalSize.
func (v *ComponentView) GetTotalSize() int64 {
	return v.TotalSize
}

// =============================================================================
// ValidationResult port implementation - implements ValidationResultPort directly
// =============================================================================

// Ensure ValidationResult implements ValidationResultPort.
var _ core.ValidationResultPort = (*ValidationResult)(nil)

// IsValid implements ValidationResultPort.IsValid.
func (r *ValidationResult) IsValid() bool {
	return r.Valid
}

// HasManifest implements ValidationResultPort.HasManifest.
func (r *ValidationResult) HasManifest() bool {
	return r.ManifestExists
}

// IsManifestValid implements ValidationResultPort.IsManifestValid.
func (r *ValidationResult) IsManifestValid() bool {
	return r.ManifestValid
}

// AreArtifactsValid implements ValidationResultPort.AreArtifactsValid.
func (r *ValidationResult) AreArtifactsValid() bool {
	return r.ArtifactsValid
}

// GetMissingArtifacts implements ValidationResultPort.GetMissingArtifacts.
func (r *ValidationResult) GetMissingArtifacts() []string {
	return r.MissingArtifacts
}

// GetCorruptArtifacts implements ValidationResultPort.GetCorruptArtifacts.
func (r *ValidationResult) GetCorruptArtifacts() []string {
	return r.CorruptArtifacts
}

// GetError implements ValidationResultPort.GetError.
func (r *ValidationResult) GetError() error {
	return r.Error
}
