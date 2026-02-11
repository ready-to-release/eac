package oscal

import (
	"strings"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
)

// ProfileHasControl checks if a profile includes a specific control ID.
func ProfileHasControl(profile *oscalTypes.Profile, controlID string) bool {
	controlIDs := GetProfileControlIDs(profile)
	normalizedTarget := strings.ToLower(controlID)

	for _, id := range controlIDs {
		if strings.EqualFold(id, normalizedTarget) {
			return true
		}
	}

	return false
}

// GetProfileTitle returns the profile's title from metadata.
func GetProfileTitle(profile *oscalTypes.Profile) string {
	if profile == nil {
		return ""
	}
	return profile.Metadata.Title
}

// GetProfileVersion returns the profile's version from metadata.
func GetProfileVersion(profile *oscalTypes.Profile) string {
	if profile == nil {
		return ""
	}
	return profile.Metadata.Version
}
