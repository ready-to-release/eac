package oscal

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
)

// GetProfileControlIDs extracts all control IDs from a profile's imports.
func GetProfileControlIDs(profile *oscalTypes.Profile) []string {
	var controlIDs []string
	seen := make(map[string]bool)

	for _, imp := range profile.Imports {
		if imp.IncludeControls != nil {
			for _, includeControl := range *imp.IncludeControls {
				if includeControl.WithIds != nil {
					for _, id := range *includeControl.WithIds {
						if !seen[id] {
							controlIDs = append(controlIDs, id)
							seen[id] = true
						}
					}
				}
			}
		}
	}

	return controlIDs
}

// NewProfileDocument creates a new OSCAL profile document with required metadata.
func NewProfileDocument(title, catalogURL string, controlIDs []string) (*oscalTypes.OscalModels, error) {
	profileUUID := uuid.New().String()
	now := time.Now().UTC()

	// Build include-controls with control IDs
	withIds := make([]string, len(controlIDs))
	copy(withIds, controlIDs)

	includeControls := []oscalTypes.SelectControlById{
		{
			WithIds: &withIds,
		},
	}

	profile := oscalTypes.Profile{
		UUID: profileUUID,
		Metadata: oscalTypes.Metadata{
			Title:        title,
			LastModified: now,
			OscalVersion: oscalTypes.Version, // "1.1.2"
			Version:      "1.0.0",
		},
		Imports: []oscalTypes.Import{
			{
				Href:            catalogURL,
				IncludeControls: &includeControls,
			},
		},
	}

	return &oscalTypes.OscalModels{
		Profile: &profile,
	}, nil
}

// NewProfileDocumentWithInfo creates a new OSCAL profile document with control information in back-matter.
func NewProfileDocumentWithInfo(title, catalogURL string, controlInfo []ControlInfo) (*oscalTypes.OscalModels, error) {
	profileUUID := uuid.New().String()
	now := time.Now().UTC()

	// Extract control IDs
	controlIDs := make([]string, len(controlInfo))
	for i, info := range controlInfo {
		controlIDs[i] = info.ID
	}

	// Build include-controls with control IDs
	includeControls := []oscalTypes.SelectControlById{
		{
			WithIds: &controlIDs,
		},
	}

	// Build back-matter with control descriptions
	var resources []oscalTypes.Resource
	for _, info := range controlInfo {
		resourceUUID := uuid.New().String()

		// Build description text
		desc := fmt.Sprintf("%s: %s", info.ID, info.Title)
		if info.Description != "" {
			desc = fmt.Sprintf("%s: %s - %s", info.ID, info.Title, info.Description)
		}

		resource := oscalTypes.Resource{
			UUID:        resourceUUID,
			Description: desc,
		}

		resources = append(resources, resource)
	}

	backMatter := &oscalTypes.BackMatter{
		Resources: &resources,
	}

	profile := oscalTypes.Profile{
		UUID: profileUUID,
		Metadata: oscalTypes.Metadata{
			Title:        title,
			LastModified: now,
			OscalVersion: oscalTypes.Version,
			Version:      "1.0.0",
		},
		Imports: []oscalTypes.Import{
			{
				Href:            catalogURL,
				IncludeControls: &includeControls,
			},
		},
		BackMatter: backMatter,
	}

	return &oscalTypes.OscalModels{
		Profile: &profile,
	}, nil
}

// NewAssessmentResultsDocument creates a new OSCAL assessment-results document.
func NewAssessmentResultsDocument(title, profileHref string) (*oscalTypes.OscalModels, error) {
	arUUID := uuid.New().String()
	resultUUID := uuid.New().String()
	now := time.Now().UTC()

	assessmentResults := oscalTypes.AssessmentResults{
		UUID: arUUID,
		Metadata: oscalTypes.Metadata{
			Title:        title,
			LastModified: now,
			OscalVersion: oscalTypes.Version, // "1.1.2"
			Version:      "1.0.0",
		},
		ImportAp: oscalTypes.ImportAp{
			Href: profileHref,
		},
		Results: []oscalTypes.Result{
			{
				UUID:        resultUUID,
				Title:       "Assessment Results",
				Description: fmt.Sprintf("Assessment results generated at %s", now.Format(time.RFC3339)),
				Start:       now,
			},
		},
	}

	return &oscalTypes.OscalModels{
		AssessmentResults: &assessmentResults,
	}, nil
}

// UpdateProfileMetadata updates the last-modified timestamp on a profile.
func UpdateProfileMetadata(profile *oscalTypes.Profile) {
	profile.Metadata.LastModified = time.Now().UTC()
}

// UpdateAssessmentResultsMetadata updates the last-modified timestamp on assessment-results.
func UpdateAssessmentResultsMetadata(ar *oscalTypes.AssessmentResults) {
	ar.Metadata.LastModified = time.Now().UTC()
}

// GetProfileCatalogURL extracts the catalog URL from the first import in a profile.
func GetProfileCatalogURL(profile *oscalTypes.Profile) string {
	if len(profile.Imports) > 0 {
		return profile.Imports[0].Href
	}
	return ""
}

// NewAssessmentResults creates a new assessment-results structure with basic metadata.
// Returns the inner AssessmentResults object (not wrapped in OscalModels).
func NewAssessmentResults(arUUID, title, profileHref string) *oscalTypes.AssessmentResults {
	now := time.Now().UTC()

	return &oscalTypes.AssessmentResults{
		UUID: arUUID,
		Metadata: oscalTypes.Metadata{
			Title:        title,
			LastModified: now,
			OscalVersion: oscalTypes.Version,
			Version:      "1.0.0",
		},
		ImportAp: oscalTypes.ImportAp{
			Href: profileHref,
		},
		Results: []oscalTypes.Result{},
	}
}

// AddResult adds a Result to an AssessmentResults.
func AddResult(ar *oscalTypes.AssessmentResults, result oscalTypes.Result) {
	ar.Results = append(ar.Results, result)
}
