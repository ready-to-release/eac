package oscal

import (
	"strings"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
)

// GetControlIDsFromProfile extracts all control IDs referenced in a profile's imports
func GetControlIDsFromProfile(profile *oscalTypes.Profile) []string {
	if profile == nil || len(profile.Imports) == 0 {
		return []string{}
	}

	var controlIDs []string
	seen := make(map[string]bool)

	for _, imp := range profile.Imports {
		// Get control IDs from include-controls
		if imp.IncludeControls != nil {
			for _, include := range *imp.IncludeControls {
				// Add WithIds controls
				if include.WithIds != nil {
					for _, id := range *include.WithIds {
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

// ProfileHasControl checks if a profile includes a specific control ID
func ProfileHasControl(profile *oscalTypes.Profile, controlID string) bool {
	controlIDs := GetControlIDsFromProfile(profile)
	normalizedTarget := strings.ToLower(controlID)

	for _, id := range controlIDs {
		if strings.ToLower(id) == normalizedTarget {
			return true
		}
	}

	return false
}

// GetProfileTitle returns the profile's title from metadata
func GetProfileTitle(profile *oscalTypes.Profile) string {
	if profile == nil {
		return ""
	}
	return profile.Metadata.Title
}

// GetProfileVersion returns the profile's version from metadata
func GetProfileVersion(profile *oscalTypes.Profile) string {
	if profile == nil {
		return ""
	}
	return profile.Metadata.Version
}
