// Package changelog provides parsing and generation of Keep a Changelog format files.
// It supports both semantic versioning (semver) and calendar versioning (calver).
package changelog

import (
	"time"
)

// VersionType indicates the versioning scheme used.
type VersionType int

const (
	// Semver indicates semantic versioning (x.y.z).
	Semver VersionType = iota
	// Calver indicates calendar versioning (YYYY.MM.DD or YYYY.MM.DD.N).
	Calver
)

// String returns the string representation of VersionType.
func (v VersionType) String() string {
	switch v {
	case Semver:
		return "semver"
	case Calver:
		return "calver"
	default:
		return "unknown"
	}
}

// ChangeType represents the category of a changelog entry.
type ChangeType string

// ChangeType constants for standard changelog categories.
const (
	Added      ChangeType = "Added"
	Changed    ChangeType = "Changed"
	Deprecated ChangeType = "Deprecated"
	Removed    ChangeType = "Removed"
	Fixed      ChangeType = "Fixed"
	Security   ChangeType = "Security"
)

// Entry represents a single changelog item within a category.
type Entry struct {
	// Description is the changelog entry text
	Description string

	// CommitType is the conventional commit type (feat, fix, etc.)
	CommitType string

	// Scope is the conventional commit scope
	Scope string

	// CommitSHA is the short commit hash for reference
	CommitSHA string

	// Breaking indicates if this is a breaking change
	Breaking bool
}

// Version represents a single version entry in the changelog.
type Version struct {
	// Number is the version string (e.g., "1.2.0" or "2025.12.01")
	Number string

	// Date is the release date
	Date time.Time

	// Added contains new feature entries
	Added []Entry

	// Changed contains modification entries
	Changed []Entry

	// Deprecated contains deprecation entries
	Deprecated []Entry

	// Removed contains removal entries
	Removed []Entry

	// Fixed contains bug fix entries
	Fixed []Entry

	// Security contains security fix entries
	Security []Entry

	// Yanked indicates if this version was yanked/withdrawn
	Yanked bool
}

// HasEntries returns true if the version has any changelog entries.
func (v *Version) HasEntries() bool {
	return len(v.Added) > 0 ||
		len(v.Changed) > 0 ||
		len(v.Deprecated) > 0 ||
		len(v.Removed) > 0 ||
		len(v.Fixed) > 0 ||
		len(v.Security) > 0
}

// AllEntries returns all entries across all categories.
func (v *Version) AllEntries() []Entry {
	all := make([]Entry, 0, len(v.Added)+len(v.Changed)+len(v.Deprecated)+len(v.Removed)+len(v.Fixed)+len(v.Security))
	all = append(all, v.Added...)
	all = append(all, v.Changed...)
	all = append(all, v.Deprecated...)
	all = append(all, v.Removed...)
	all = append(all, v.Fixed...)
	all = append(all, v.Security...)
	return all
}

// HasBreakingChanges returns true if any entry is a breaking change.
func (v *Version) HasBreakingChanges() bool {
	for _, e := range v.AllEntries() {
		if e.Breaking {
			return true
		}
	}
	return false
}

// Changelog represents the complete parsed changelog file.
type Changelog struct {
	// Module is the module moniker this changelog belongs to
	Module string

	// Title is the changelog title (default: "Changelog")
	Title string

	// Description is optional introductory text
	Description string

	// VersionType indicates semver or calver
	VersionType VersionType

	// Unreleased contains pending changes not yet released
	Unreleased *Version

	// Versions is the list of released versions, ordered newest-first
	Versions []Version

	// RepoURL is the base repository URL for generating comparison links
	RepoURL string
}

// LatestVersion returns the most recent released version, or nil if none.
func (c *Changelog) LatestVersion() *Version {
	if len(c.Versions) == 0 {
		return nil
	}
	return &c.Versions[0]
}

// LatestVersionNumber returns the latest version string, or empty if none.
func (c *Changelog) LatestVersionNumber() string {
	if v := c.LatestVersion(); v != nil {
		return v.Number
	}
	return ""
}

// GetVersion returns a specific version by number, or nil if not found.
func (c *Changelog) GetVersion(number string) *Version {
	for i := range c.Versions {
		if c.Versions[i].Number == number {
			return &c.Versions[i]
		}
	}
	return nil
}

// AddVersion prepends a new version to the changelog.
func (c *Changelog) AddVersion(v *Version) {
	c.Versions = append([]Version{*v}, c.Versions...)
}

// PromoteUnreleased moves unreleased entries to a new version.
func (c *Changelog) PromoteUnreleased(versionNumber string, date time.Time) *Version {
	if c.Unreleased == nil || !c.Unreleased.HasEntries() {
		return nil
	}

	newVersion := Version{
		Number:     versionNumber,
		Date:       date,
		Added:      c.Unreleased.Added,
		Changed:    c.Unreleased.Changed,
		Deprecated: c.Unreleased.Deprecated,
		Removed:    c.Unreleased.Removed,
		Fixed:      c.Unreleased.Fixed,
		Security:   c.Unreleased.Security,
	}

	c.AddVersion(&newVersion)

	// Reset unreleased
	c.Unreleased = &Version{}

	return &c.Versions[0]
}
