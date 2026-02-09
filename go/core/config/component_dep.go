package config

import (
	"fmt"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// ParsedComponentDep represents a parsed inter-module component dependency.
// Format: "module:componentName[:componentType[:tool]]" (2-4 parts).
type ParsedComponentDep struct {
	Module        string // Target module moniker (required)
	ComponentName string // Target component name (required)
	ComponentType string // Target component type (optional, 3+ parts)
	Tool          string // Target tool (optional, 4 parts)
	Raw           string // Original unparsed string
}

// ParseComponentDep splits a component_deps entry on ":" and validates 2-4 parts.
func ParseComponentDep(dep string) (ParsedComponentDep, error) {
	parts := strings.Split(dep, ":")
	if len(parts) < 2 || len(parts) > 4 {
		return ParsedComponentDep{}, fmt.Errorf("component_deps entry %q: expected 2-4 colon-separated parts, got %d", dep, len(parts))
	}
	for i, p := range parts {
		if p == "" {
			return ParsedComponentDep{}, fmt.Errorf("component_deps entry %q: part %d is empty", dep, i+1)
		}
	}

	parsed := ParsedComponentDep{
		Module:        parts[0],
		ComponentName: parts[1],
		Raw:           dep,
	}
	if len(parts) >= 3 {
		parsed.ComponentType = parts[2]
	}
	if len(parts) == 4 {
		parsed.Tool = parts[3]
	}
	return parsed, nil
}

// MatchesUnitID returns true if this parsed dep matches the given UnitID.
// Matching is progressive by the number of parts specified:
//   - 2-part (module:name): matches if Module and ComponentName match
//   - 3-part (module:name:type): also requires ComponentType match
//   - 4-part (module:name:type:tool): also requires Tool match
func (p ParsedComponentDep) MatchesUnitID(uid core.UnitID) bool {
	if uid.Module != p.Module || uid.ComponentName != p.ComponentName {
		return false
	}
	if p.ComponentType != "" && uid.ComponentType != p.ComponentType {
		return false
	}
	if p.Tool != "" && uid.Tool != p.Tool {
		return false
	}
	return true
}
