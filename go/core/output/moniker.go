package output

import "strings"

// Moniker represents a fully qualified unit identifier.
// Format: [context:]module[:component[:tool]]
// Examples:
//   - "build:core:go:go" (full 4-part)
//   - "core:go:gotest" (3-part, no context)
//   - "eac:docker" (2-part)
//   - "core" (1-part, module only)
type Moniker string

// ParsedMoniker contains the parsed components of a moniker.
type ParsedMoniker struct {
	Action    string // e.g., "build", "test", "lint", "scan"
	Module    string // e.g., "core", "eac"
	Component string // e.g., "go", "docker"
	Tool      string // e.g., "go", "gotest", "trivy-vuln"
	Full      string // Original full moniker string
}

// Parse breaks down a moniker into its components.
// For 4-part monikers: context:module:component:tool
// For 3-part monikers: module:component:tool (context empty)
// For 2-part monikers: module:component (tool empty)
// For 1-part monikers: module only (component and tool empty)
func (m Moniker) Parse() ParsedMoniker {
	parts := strings.Split(string(m), ":")
	pm := ParsedMoniker{Full: string(m)}

	switch len(parts) {
	case 4:
		pm.Action = parts[0]
		pm.Module = parts[1]
		pm.Component = parts[2]
		pm.Tool = parts[3]
	case 3:
		pm.Module = parts[0]
		pm.Component = parts[1]
		pm.Tool = parts[2]
	case 2:
		pm.Module = parts[0]
		pm.Component = parts[1]
	case 1:
		pm.Module = parts[0]
	}
	return pm
}

// ModuleName extracts just the first part before any colon.
// For "build:core:go:go" returns "build".
// For "core:go" returns "core".
// For "core" returns "core".
func (m Moniker) ModuleName() string {
	s := string(m)
	if idx := strings.Index(s, ":"); idx >= 0 {
		return s[:idx]
	}
	return s
}

// Shortname returns a display-friendly short form (module:component).
// For "build:core:go:go" returns "core:go".
// For "core:go:gotest" returns "core:go".
// For "eac:docker" returns "eac:docker".
// For "core" returns "core".
func (m Moniker) Shortname() string {
	p := m.Parse()
	if p.Module == "" {
		return ""
	}
	if p.Component == "" {
		return p.Module
	}
	return p.Module + ":" + p.Component
}

// String returns the moniker as a string.
func (m Moniker) String() string {
	return string(m)
}
