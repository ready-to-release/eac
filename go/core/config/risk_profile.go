package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	security "github.com/ready-to-release/eac/contracts/security/0.1.0/interfaces"
)

// ProfileWrapper wraps an OSCAL profile to implement ProfilePort.
// It provides a simplified interface to access profile data.
type ProfileWrapper struct {
	profile *oscalTypes.Profile
}

// Verify ProfileWrapper implements ProfilePort.
var _ security.ProfilePort = (*ProfileWrapper)(nil)

// LoadProfileWrapper loads an OSCAL profile from a JSON file.
// Returns an error if the file doesn't exist or doesn't contain a profile.
func LoadProfileWrapper(path string) (*ProfileWrapper, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading profile file: %w", err)
	}

	var doc oscalTypes.OscalModels
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing profile JSON: %w", err)
	}

	if doc.Profile == nil {
		return nil, fmt.Errorf("file does not contain a profile: %s", path)
	}

	return &ProfileWrapper{profile: doc.Profile}, nil
}

// NewProfileWrapper creates a ProfileWrapper from an existing OSCAL profile.
func NewProfileWrapper(profile *oscalTypes.Profile) *ProfileWrapper {
	return &ProfileWrapper{profile: profile}
}

// ControlIDs returns all control IDs selected by this profile.
// Returns nil if the profile has no imports or no included controls.
func (p *ProfileWrapper) ControlIDs() []string {
	if p.profile == nil || len(p.profile.Imports) == 0 {
		return nil
	}

	var ids []string
	seen := make(map[string]bool)

	for _, imp := range p.profile.Imports {
		if imp.IncludeControls == nil {
			continue
		}
		for _, inc := range *imp.IncludeControls {
			if inc.WithIds == nil {
				continue
			}
			for _, id := range *inc.WithIds {
				if !seen[id] {
					ids = append(ids, id)
					seen[id] = true
				}
			}
		}
	}

	return ids
}

// HasControl checks if a control ID is in the profile (case-insensitive).
func (p *ProfileWrapper) HasControl(controlID string) bool {
	for _, id := range p.ControlIDs() {
		if strings.EqualFold(id, controlID) {
			return true
		}
	}
	return false
}

// Title returns the profile title from metadata.
func (p *ProfileWrapper) Title() string {
	if p.profile == nil {
		return ""
	}
	return p.profile.Metadata.Title
}

// Version returns the profile version from metadata.
func (p *ProfileWrapper) Version() string {
	if p.profile == nil {
		return ""
	}
	return p.profile.Metadata.Version
}

// CatalogHref returns the catalog URL from the first import.
func (p *ProfileWrapper) CatalogHref() string {
	if p.profile == nil || len(p.profile.Imports) == 0 {
		return ""
	}
	return p.profile.Imports[0].Href
}
