// Package scanner defines the security scanning contract interfaces.
package scanner

// RiskConfigPort provides access to risk and compliance configuration.
// Implementations load risk profiles and scoring configuration from files.
type RiskConfigPort interface {
	// GetProfile returns the solution-wide OSCAL profile.
	// Returns an error if no profile is configured.
	GetProfile() (ProfilePort, error)

	// GetModuleProfile returns the profile for a specific module.
	// Falls back to solution profile if module has no specific profile.
	GetModuleProfile(moniker string) (ProfilePort, error)

	// GetCatalogURL returns the OSCAL catalog URL for control validation.
	GetCatalogURL() string

	// GetScoring returns the risk scoring configuration.
	GetScoring() RiskScoringPort

	// ListModuleProfiles returns monikers of modules with custom profiles.
	ListModuleProfiles() []string
}

// ProfilePort provides access to an OSCAL profile's data.
// This is a lightweight wrapper around OSCAL profile structures.
type ProfilePort interface {
	// ControlIDs returns all control IDs selected by this profile.
	ControlIDs() []string

	// HasControl checks if a control ID is in the profile (case-insensitive).
	HasControl(controlID string) bool

	// Title returns the profile title from metadata.
	Title() string

	// Version returns the profile version from metadata.
	Version() string

	// CatalogHref returns the catalog URL from the first import.
	CatalogHref() string
}

// RiskScoringPort provides access to risk scoring configuration.
// Used to calculate impact and criticality based on module type.
type RiskScoringPort interface {
	// GetImpact returns the impact rating (1-5) for a module type.
	// Falls back to _default if type is not found.
	GetImpact(moduleType string) int

	// GetCriticality returns the criticality (high/medium/low) for a module type.
	// Falls back to _default if type is not found.
	GetCriticality(moduleType string) string

	// GetSeverityWeight returns the likelihood increment for a severity level.
	// Valid severity levels: critical, high, medium, low.
	GetSeverityWeight(severity string) int
}
