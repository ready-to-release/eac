package output

import (
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// ValidateUoW checks if a UoW's output is valid using a UnitID.
func (r *DiskOutputReader) ValidateUoW(id workunit.UnitID) ValidationResult {
	result := ValidationResult{
		MissingArtifacts: []string{},
		CorruptArtifacts: []string{},
	}

	// Check manifest exists and is valid
	manifestPath := r.uowManifestPath(id)
	manifest, err := Load(manifestPath)
	if err != nil {
		result.Valid = false
		result.ManifestExists = false
		result.ManifestValid = false
		result.ArtifactsValid = false
		result.Error = err
		return result
	}

	result.ManifestExists = true
	result.ManifestValid = true

	// Validate artifacts
	if len(manifest.Artifacts) == 0 {
		result.Valid = true
		result.ArtifactsValid = true
		return result
	}

	uowDirPath := r.uowDir(id)
	artifactResult := ValidateArtifacts(uowDirPath, manifest.Artifacts)

	result.Valid = artifactResult.Valid
	result.ArtifactsValid = artifactResult.ArtifactsValid
	result.MissingArtifacts = artifactResult.MissingArtifacts
	result.CorruptArtifacts = artifactResult.CorruptArtifacts
	result.Error = artifactResult.Error

	return result
}

// ValidateModule checks if all expected UoWs for a module are valid.
func (r *DiskOutputReader) ValidateModule(ctx core.ActionType, module string, expectedUoWs []workunit.UnitID) ValidationResult {
	result := ValidationResult{
		Valid:            true,
		ManifestExists:   true,
		ManifestValid:    true,
		ArtifactsValid:   true,
		MissingArtifacts: []string{},
		CorruptArtifacts: []string{},
	}

	// If no expected UoWs, validation passes
	if len(expectedUoWs) == 0 {
		return result
	}

	// Check each expected UoW
	for _, id := range expectedUoWs {
		uowResult := r.ValidateUoW(id)

		if !uowResult.Valid {
			result.Valid = false

			if !uowResult.ManifestExists {
				result.ManifestExists = false
				result.MissingArtifacts = append(result.MissingArtifacts, id.DirName()+"/uow.manifest.json")
			}
			if !uowResult.ManifestValid {
				result.ManifestValid = false
			}
			if !uowResult.ArtifactsValid {
				result.ArtifactsValid = false
				result.MissingArtifacts = append(result.MissingArtifacts, uowResult.MissingArtifacts...)
				result.CorruptArtifacts = append(result.CorruptArtifacts, uowResult.CorruptArtifacts...)
			}
			if uowResult.Error != nil && result.Error == nil {
				result.Error = uowResult.Error
			}
		}
	}

	return result
}

// VerifyModuleIntegrity checks all UoW artifacts for a module.
// Returns nil if all artifacts pass hash verification.
// Returns error describing first failed artifact otherwise.
// This replaces legacy ModuleManifest.VerifyArtifactsIntegrity().
func (r *DiskOutputReader) VerifyModuleIntegrity(ctx core.ActionType, module string) error {
	manifests, err := r.ListUoWs(ctx, module)
	if err != nil {
		return err
	}

	if len(manifests) == 0 {
		return nil // No manifests = nothing to verify
	}

	for _, manifest := range manifests {
		uowDir := filepath.Join(r.workspaceRoot, "out", string(ctx), module, manifest.DirName())
		result := ValidateArtifacts(uowDir, manifest.Artifacts)
		if !result.Valid {
			uowName := manifest.DirName()
			if len(result.MissingArtifacts) > 0 {
				return &IntegrityError{
					UoW:     uowName,
					Message: "missing artifacts: " + strings.Join(result.MissingArtifacts, ", "),
				}
			}
			if len(result.CorruptArtifacts) > 0 {
				return &IntegrityError{
					UoW:     uowName,
					Message: "hash mismatch: " + strings.Join(result.CorruptArtifacts, ", "),
				}
			}
			if result.Error != nil {
				return &IntegrityError{
					UoW:     uowName,
					Message: result.Error.Error(),
				}
			}
		}
	}

	return nil
}

// IntegrityError represents an artifact integrity validation failure.
type IntegrityError struct {
	UoW     string
	Message string
}

func (e *IntegrityError) Error() string {
	return "UoW " + e.UoW + ": " + e.Message
}
