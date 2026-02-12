package output

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// unitIDFromPort converts a UnitIDPort to a concrete workunit.UnitID,
// including all Extra fields for proper uniqueness handling.
func unitIDFromPort(p core.UnitIDPort) workunit.UnitID {
	return workunit.UnitID{
		Action:        core.ActionType(p.GetAction()),
		Module:        p.GetModule(),
		ComponentType: p.GetComponent(),
		ComponentName: p.GetComponent(),
		Tool:          p.GetTool(),
		Spec:          p.GetSpec(),
		Extra:         p.GetExtra(),
	}
}

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
var _ core.OutputReaderPort = (*OutputReaderAdapter)(nil)

// GetUoW implements OutputReaderPort.GetUoW.
func (a *OutputReaderAdapter) GetUoW(context, module, component, tool string) (core.UoWManifestPort, error) {
	id := workunit.UnitID{
		Action:        core.ActionType(context),
		Module:        module,
		ComponentType: component,
		ComponentName: component,
		Tool:          tool,
		// Note: Extra is not available through this interface signature.
		// Use GetUoWByID for UoWs with Extra fields.
	}
	manifest, err := a.reader.GetUoW(id)
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

// GetUoWByID implements OutputReaderPort with full UnitID support including Extra.
func (a *OutputReaderAdapter) GetUoWByID(unitID core.UnitIDPort) (core.UoWManifestPort, error) {
	id := unitIDFromPort(unitID)
	manifest, err := a.reader.GetUoW(id)
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

// GetModule implements OutputReaderPort.GetModule.
func (a *OutputReaderAdapter) GetModule(context, module string) (core.ModuleViewPort, error) {
	view, err := a.reader.GetModule(core.ActionType(context), module)
	if err != nil {
		return nil, err
	}
	return view, nil
}

// ListUoWs implements OutputReaderPort.ListUoWs.
func (a *OutputReaderAdapter) ListUoWs(context, module string) ([]core.UoWManifestPort, error) {
	manifests, err := a.reader.ListUoWs(core.ActionType(context), module)
	if err != nil {
		return nil, err
	}
	result := make([]core.UoWManifestPort, len(manifests))
	for i, m := range manifests {
		result[i] = m
	}
	return result, nil
}

// ValidateUoW implements OutputReaderPort.ValidateUoW.
func (a *OutputReaderAdapter) ValidateUoW(context, module, component, tool string) core.ValidationResultPort {
	id := workunit.UnitID{
		Action:        core.ActionType(context),
		Module:        module,
		ComponentType: component,
		ComponentName: component,
		Tool:          tool,
		// Note: Extra is not available through this interface signature.
		// Use ValidateUoWByID for UoWs with Extra fields.
	}
	result := a.reader.ValidateUoW(id)
	return &result
}

// ValidateUoWByID implements validation with full UnitID support including Extra.
func (a *OutputReaderAdapter) ValidateUoWByID(unitID core.UnitIDPort) core.ValidationResultPort {
	id := unitIDFromPort(unitID)
	result := a.reader.ValidateUoW(id)
	return &result
}

// ValidateModule implements OutputReaderPort.ValidateModule.
func (a *OutputReaderAdapter) ValidateModule(context, module string, expectedUoWs []core.UnitIDPort) core.ValidationResultPort {
	// Convert UnitIDPort to workunit.UnitID
	ids := make([]workunit.UnitID, len(expectedUoWs))
	for i, id := range expectedUoWs {
		ids[i] = unitIDFromPort(id)
	}
	result := a.reader.ValidateModule(core.ActionType(context), module, ids)
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
var _ core.UoWTrackerPort = (*UoWTrackerAdapter)(nil)

// RecordStart implements UoWTrackerPort.RecordStart.
func (a *UoWTrackerAdapter) RecordStart(unitID core.UnitIDPort) error {
	return a.tracker.RecordStart(unitIDFromPort(unitID))
}

// RecordComplete implements UoWTrackerPort.RecordComplete.
func (a *UoWTrackerAdapter) RecordComplete(unitID core.UnitIDPort, manifest core.UoWManifestPort) error {
	id := unitIDFromPort(unitID)

	// Convert manifest port to concrete type
	uowManifest := &UoWManifest{
		Action:     core.ActionType(manifest.GetAction()),
		Module:     manifest.GetModule(),
		Component:  manifest.GetComponent(),
		Tool:       manifest.GetTool(),
		Extra:      manifest.GetExtra(),
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
func (a *UoWTrackerAdapter) RecordCacheHit(unitID core.UnitIDPort) (core.UoWManifestPort, error) {
	manifest, err := a.tracker.RecordCacheHit(unitIDFromPort(unitID))
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

