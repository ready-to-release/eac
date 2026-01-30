package workunit

import (
	"fmt"
	"path/filepath"
)

// UnitID uniquely identifies a unit of work.
// Format: context:module:component:tool[:extra]
type UnitID struct {
	Context   Context           // build, test, lint, scan
	Module    string            // module moniker (e.g., "eac-core")
	Component string            // component name (e.g., "go", "gherkin")
	Tool      string            // handler/provider/scanner (e.g., "go", "gotest", "golangci-lint")
	Extra     map[string]string // context-specific (e.g., testset: "unit")
	Spec      string            // Spec name for BDD tests (godog, tscucumber), e.g., "build-module"
}

// Shortname returns display name: module:component, or just spec name for BDD tests.
// Deprecated: Use Path() for stable identifiers or DisplayName() for contextual display.
func (u UnitID) Shortname() string {
	if u.Spec != "" {
		return u.Spec // "build-module"
	}
	return u.Module + ":" + u.Component
}

// Path returns module:component for component-level identification.
// Note: This is NOT a unique work unit identifier - use Longname() for that.
// For unique work unit identification, context, tool, and extras also matter.
func (u UnitID) Path() string {
	return u.Module + ":" + u.Component
}

// ComponentName returns just the component name.
func (u UnitID) ComponentName() string {
	return u.Component
}

// DisplayName returns context-appropriate name.
// For BDD tests, returns spec name (or module:spec:specname if disambiguate is true).
// If disambiguate is true, returns full module:component.
// If disambiguate is false, returns just the component.
func (u UnitID) DisplayName(disambiguate bool) string {
	if u.Spec != "" {
		if disambiguate {
			return u.Module + ":spec:" + u.Spec // "eac-commands:spec:build-module"
		}
		return u.Spec // "build-module"
	}
	if disambiguate {
		return u.Module + ":" + u.Component
	}
	return u.Component
}

// TabLabel returns truncated name for TUI tabs (max width).
// For BDD tests, uses the spec name. Otherwise uses component name.
// If the name exceeds maxWidth, it's truncated with "...".
func (u UnitID) TabLabel(maxWidth int) string {
	name := u.Component
	if u.Spec != "" {
		name = u.Spec
	}
	if maxWidth <= 3 {
		if len(name) > maxWidth {
			return name[:maxWidth]
		}
		return name
	}
	if len(name) > maxWidth {
		return name[:maxWidth-3] + "..."
	}
	return name
}

// Longname returns full ID: context:module:component:tool[:extra]
// For BDD tests (godog, tscucumber), returns "module:spec:specname" format.
// For test context with testset extra, appends the testset value.
func (u UnitID) Longname() string {
	if u.Spec != "" {
		return fmt.Sprintf("%s:spec:%s", u.Module, u.Spec) // "eac-commands:spec:build-module"
	}
	base := fmt.Sprintf("%s:%s:%s:%s", u.Context, u.Module, u.Component, u.Tool)
	if testset, ok := u.Extra["testset"]; ok && testset != "" {
		base += ":" + testset
	}
	return base
}

// String returns Longname.
func (u UnitID) String() string {
	return u.Longname()
}

// OutDir returns the unique output directory for this unit.
// Format: out/<context>/<module>/<component>[/<extra>...]
func (u UnitID) OutDir() string {
	base := filepath.Join("out", string(u.Context), u.Module, u.Component)
	if testset, ok := u.Extra["testset"]; ok && testset != "" {
		base = filepath.Join(base, testset)
	}
	return base
}

// LockFile returns the path to the execution lock.
func (u UnitID) LockFile() string {
	return filepath.Join(u.OutDir(), ".lock")
}

// StateFile returns the path to the cache state.
func (u UnitID) StateFile() string {
	return filepath.Join(u.OutDir(), "state.json")
}

// LogFile returns the path to the execution log.
func (u UnitID) LogFile() string {
	return filepath.Join(u.OutDir(), "execution.log")
}

// ResultsFile returns the path to results (test/lint/scan).
func (u UnitID) ResultsFile() string {
	return filepath.Join(u.OutDir(), "results.json")
}
