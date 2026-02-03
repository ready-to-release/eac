package output

import (
	"time"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// =============================================================================
// OutputReaderAdapter - wraps DiskOutputReader to implement OutputReaderPort
// =============================================================================

// OutputReaderAdapter wraps DiskOutputReader to implement OutputReaderPort.
type OutputReaderAdapter struct {
	reader *DiskOutputReader
}

// NewOutputReaderAdapter creates an OutputReaderPort from a DiskOutputReader.
func NewOutputReaderAdapter(reader *DiskOutputReader) *OutputReaderAdapter {
	return &OutputReaderAdapter{reader: reader}
}

// Ensure OutputReaderAdapter implements OutputReaderPort.
var _ interfaces.OutputReaderPort = (*OutputReaderAdapter)(nil)

// GetUoW implements OutputReaderPort.GetUoW.
func (a *OutputReaderAdapter) GetUoW(context, module, component, tool string) (interfaces.UoWManifestPort, error) {
	manifest, err := a.reader.GetUoW(workunit.Context(context), module, component, tool)
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

// GetModule implements OutputReaderPort.GetModule.
func (a *OutputReaderAdapter) GetModule(context, module string) (interfaces.ModuleViewPort, error) {
	view, err := a.reader.GetModule(workunit.Context(context), module)
	if err != nil {
		return nil, err
	}
	return view, nil
}

// ListUoWs implements OutputReaderPort.ListUoWs.
func (a *OutputReaderAdapter) ListUoWs(context, module string) ([]interfaces.UoWManifestPort, error) {
	manifests, err := a.reader.ListUoWs(workunit.Context(context), module)
	if err != nil {
		return nil, err
	}
	result := make([]interfaces.UoWManifestPort, len(manifests))
	for i, m := range manifests {
		result[i] = m
	}
	return result, nil
}

// ValidateUoW implements OutputReaderPort.ValidateUoW.
func (a *OutputReaderAdapter) ValidateUoW(context, module, component, tool string) interfaces.ValidationResultPort {
	result := a.reader.ValidateUoW(workunit.Context(context), module, component, tool)
	return &result
}

// ValidateModule implements OutputReaderPort.ValidateModule.
func (a *OutputReaderAdapter) ValidateModule(context, module string, expectedUoWs []interfaces.UnitIDPort) interfaces.ValidationResultPort {
	// Convert UnitIDPort to workunit.UnitID
	ids := make([]workunit.UnitID, len(expectedUoWs))
	for i, id := range expectedUoWs {
		ids[i] = workunit.UnitID{
			Context:   workunit.Context(id.GetContext()),
			Module:    id.GetModule(),
			Component: id.GetComponent(),
			Tool:      id.GetTool(),
		}
	}
	result := a.reader.ValidateModule(workunit.Context(context), module, ids)
	return &result
}

// =============================================================================
// UoWTrackerAdapter - wraps InMemoryTracker to implement UoWTrackerPort
// =============================================================================

// UoWTrackerAdapter wraps InMemoryTracker to implement UoWTrackerPort.
type UoWTrackerAdapter struct {
	tracker *InMemoryTracker
}

// NewUoWTrackerAdapter creates a UoWTrackerPort from an InMemoryTracker.
func NewUoWTrackerAdapter(tracker *InMemoryTracker) *UoWTrackerAdapter {
	return &UoWTrackerAdapter{tracker: tracker}
}

// Ensure UoWTrackerAdapter implements UoWTrackerPort.
var _ interfaces.UoWTrackerPort = (*UoWTrackerAdapter)(nil)

// RecordStart implements UoWTrackerPort.RecordStart.
func (a *UoWTrackerAdapter) RecordStart(unitID interfaces.UnitIDPort) error {
	id := workunit.UnitID{
		Context:   workunit.Context(unitID.GetContext()),
		Module:    unitID.GetModule(),
		Component: unitID.GetComponent(),
		Tool:      unitID.GetTool(),
	}
	return a.tracker.RecordStart(id)
}

// RecordComplete implements UoWTrackerPort.RecordComplete.
func (a *UoWTrackerAdapter) RecordComplete(unitID interfaces.UnitIDPort, manifest interfaces.UoWManifestPort) error {
	id := workunit.UnitID{
		Context:   workunit.Context(unitID.GetContext()),
		Module:    unitID.GetModule(),
		Component: unitID.GetComponent(),
		Tool:      unitID.GetTool(),
	}

	// Convert manifest port to concrete type
	uowManifest := &UoWManifest{
		Context:    workunit.Context(manifest.GetContext()),
		Module:     manifest.GetModule(),
		Component:  manifest.GetComponent(),
		Tool:       manifest.GetTool(),
		ExitCode:   manifest.GetExitCode(),
		InputHash:  manifest.GetInputHash(),
		ExecutedAt: manifest.GetExecutedAt(),
		Duration:   manifest.GetDuration(),
		OutputHash: manifest.GetOutputHash(),
	}

	// Convert artifacts
	for _, art := range manifest.GetArtifacts() {
		uowManifest.Artifacts = append(uowManifest.Artifacts, Artifact{
			ID:     art.GetID(),
			Path:   art.GetPath(),
			SHA256: art.GetSHA256(),
			Size:   art.GetSize(),
			Type:   art.GetType(),
		})
	}

	return a.tracker.RecordComplete(id, uowManifest)
}

// RecordCacheHit implements UoWTrackerPort.RecordCacheHit.
func (a *UoWTrackerAdapter) RecordCacheHit(unitID interfaces.UnitIDPort) (interfaces.UoWManifestPort, error) {
	id := workunit.UnitID{
		Context:   workunit.Context(unitID.GetContext()),
		Module:    unitID.GetModule(),
		Component: unitID.GetComponent(),
		Tool:      unitID.GetTool(),
	}
	manifest, err := a.tracker.RecordCacheHit(id)
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

// =============================================================================
// UoWManifest port implementation - implements UoWManifestPort directly
// =============================================================================

// Ensure UoWManifest implements UoWManifestPort.
var _ interfaces.UoWManifestPort = (*UoWManifest)(nil)

// GetContext implements UoWManifestPort.GetContext.
func (m *UoWManifest) GetContext() string {
	return string(m.Context)
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
func (m *UoWManifest) GetArtifacts() []interfaces.OutputArtifactPort {
	result := make([]interfaces.OutputArtifactPort, len(m.Artifacts))
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
var _ interfaces.OutputArtifactPort = (*Artifact)(nil)

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
var _ interfaces.ModuleViewPort = (*ModuleView)(nil)

// GetModule implements ModuleViewPort.GetModule.
func (v *ModuleView) GetModule() string {
	return v.Module
}

// GetStatus implements ModuleViewPort.GetStatus.
func (v *ModuleView) GetStatus() string {
	return string(v.Status)
}

// GetComponents implements ModuleViewPort.GetComponents.
func (v *ModuleView) GetComponents() []interfaces.ComponentViewPort {
	result := make([]interfaces.ComponentViewPort, len(v.Components))
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
var _ interfaces.ComponentViewPort = (*ComponentView)(nil)

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
func (v *ComponentView) GetUoWs() []interfaces.UoWManifestPort {
	result := make([]interfaces.UoWManifestPort, len(v.UoWs))
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
var _ interfaces.ValidationResultPort = (*ValidationResult)(nil)

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
