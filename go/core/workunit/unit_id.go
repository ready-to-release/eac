package workunit

import (
	"fmt"
	"path/filepath"
	"sort"
)

// UnitID uniquely identifies a unit of work.
// Format: context:module:component:tool[:extra]
type UnitID struct {
	Context   Context           // build, test, lint, scan
	Module    string            // module moniker (e.g., "core")
	Component string            // component name (e.g., "go", "gherkin")
	Tool      string            // handler/provider/scanner (e.g., "go", "gotest", "golangci-lint")
	Extra     map[string]string // context-specific (e.g., testset: "unit")
	Spec      string            // Spec name for BDD tests (godog, tscucumber), e.g., "build-module"
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

// DisplayName returns a context-aware compact display name.
// This is the primary method for TUI tabs and progress indicators.
//
// Format by context:
//   - Test (BDD):  "<spec-name>: <tool>"       e.g., "build-module: godog"
//   - Test (unit): "<testname>: unit"          e.g., "impl-build: unit"
//   - Build:       "<module>: <component>"     e.g., "core: go"
//                  "<module>: <comp>: <tool>"  e.g., "docs: site: mkdocs"
//   - Lint:        "lint:<component>:<tool>"   e.g., "lint:go:golangci-lint"
//   - Scan:        "scan:<component>:<category>" e.g., "scan:go:vuln"
func (u UnitID) DisplayName() string {
	switch u.Context {
	case ContextTest:
		if u.Spec != "" {
			return u.Spec + ": " + u.Tool
		}
		testname := u.Extra["testname"]
		if testname == "" {
			testname = u.Component
		}
		return testname + ": unit"

	case ContextBuild:
		if u.Component == u.Tool || u.Tool == "" {
			return u.Module + ": " + u.Component
		}
		return u.Module + ": " + u.Component + ": " + u.Tool

	case ContextLint:
		return "lint:" + u.Component + ":" + u.Tool

	case ContextScan:
		category := u.Extra["category"]
		if category == "" {
			category = u.Tool
		}
		return "scan:" + u.Component + ":" + category

	default:
		return u.Component
	}
}

// NormalDisplay returns a human-readable sentence describing the work unit.
// For logs, summaries, and headers where full context is helpful.
//
// Format: Human-readable sentence with all relevant information
//   - Build:       "Building <component> in <module>"
//                  "Building <component> in <module> with <tool>"
//   - Test (BDD):  "Testing spec <spec-name> in <module>"
//   - Test (unit): "Testing <testname> in <module>"
//   - Lint:        "Linting <component> in <module> with <tool>"
//   - Scan:        "Scanning <component> in <module> for <category>"
//
// Examples:
//   - "Building go in core"
//   - "Building site in docs with mkdocs"
//   - "Testing spec build-module in eac-cli"
//   - "Testing impl-build in core"
//   - "Linting go in core with golangci-lint"
//   - "Scanning go in core for vulnerabilities"
func (u UnitID) NormalDisplay() string {
	switch u.Context {
	case ContextBuild:
		if u.Tool == "" || u.Tool == u.Component {
			return "Building " + u.Component + " in " + u.Module
		}
		return "Building " + u.Component + " in " + u.Module + " with " + u.Tool

	case ContextTest:
		if u.Spec != "" {
			return "Testing spec " + u.Spec + " in " + u.Module
		}
		testname := u.Extra["testname"]
		if testname == "" {
			testname = u.Component
		}
		return "Testing " + testname + " in " + u.Module

	case ContextLint:
		return "Linting " + u.Component + " in " + u.Module + " with " + u.Tool

	case ContextScan:
		category := u.Extra["category"]
		if category == "" {
			category = u.Tool
		}
		return "Scanning " + u.Component + " in " + u.Module + " for " + category

	default:
		return string(u.Context) + " " + u.Component + " in " + u.Module
	}
}

// DisplayKey returns the key used for disambiguation checking.
// For test context, uses spec or testname; otherwise uses component.
func (u UnitID) DisplayKey() string {
	if u.Context == ContextTest {
		if u.Spec != "" {
			return u.Spec
		}
		if testname := u.Extra["testname"]; testname != "" {
			return testname
		}
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

// Longname returns full ID: context:module:component:tool[:extra...]
// All Extra field values are appended in sorted key order for uniqueness.
func (u UnitID) Longname() string {
	base := fmt.Sprintf("%s:%s:%s:%s", u.Context, u.Module, u.Component, u.Tool)

	// Append all Extra values in sorted key order for consistency
	if len(u.Extra) > 0 {
		keys := make([]string, 0, len(u.Extra))
		for k := range u.Extra {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if v := u.Extra[k]; v != "" {
				base += ":" + v
			}
		}
	}
	return base
}

// String returns Longname.
func (u UnitID) String() string {
	return u.Longname()
}

// DirName returns the unique directory name for this unit.
// Format: component-tool[-extraVal1][-extraVal2]...
// Tool and extra values are appended with dashes in sorted key order for uniqueness.
// This is used by OutDir() and the manifest system for consistent path generation.
func (u UnitID) DirName() string {
	dirName := u.Component
	if u.Tool != "" {
		dirName += "-" + u.Tool
	}

	// Append all Extra values with dashes in sorted key order
	if len(u.Extra) > 0 {
		keys := make([]string, 0, len(u.Extra))
		for k := range u.Extra {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if v := u.Extra[k]; v != "" {
				dirName += "-" + v
			}
		}
	}
	return dirName
}

// OutDir returns the unique output directory for this unit.
// Format: out/<context>/<module>/<dirname>
// Where dirname = component[-extraVal1][-extraVal2]... (flat, same depth for all UoWs)
func (u UnitID) OutDir() string {
	return filepath.Join("out", string(u.Context), u.Module, u.DirName())
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

// ============================================================================
// UnitIDPort interface implementation
// ============================================================================

// GetContext returns the operation type: "build", "test", "lint", "scan".
// Implements interfaces.UnitIDPort.
func (u UnitID) GetContext() string {
	return string(u.Context)
}

// GetModule returns the module moniker (e.g., "core").
// Implements interfaces.UnitIDPort.
func (u UnitID) GetModule() string {
	return u.Module
}

// GetComponent returns the component name (e.g., "go", "gherkin").
// Implements interfaces.UnitIDPort.
func (u UnitID) GetComponent() string {
	return u.Component
}

// GetTool returns the handler/provider/scanner (e.g., "go", "gotest", "golangci-lint").
// Implements interfaces.UnitIDPort.
func (u UnitID) GetTool() string {
	return u.Tool
}

// GetSpec returns the spec name for BDD tests (e.g., "build-module").
// Implements interfaces.UnitIDPort.
func (u UnitID) GetSpec() string {
	return u.Spec
}

// GetExtra returns the Extra map containing context-specific discriminators.
// Implements interfaces.UnitIDPort.
func (u UnitID) GetExtra() map[string]string {
	return u.Extra
}
