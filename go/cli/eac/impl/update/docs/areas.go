// Package docs provides the update docs command implementation.
package docs

import "fmt"

// Area represents a documentation update area.
type Area string

const (
	// AreaMermaid scans markdown for mermaid blocks and renders missing diagrams to cache.
	AreaMermaid Area = "mermaid"
	// AreaDrawio scans for drawio.png images, optimizes and caches them.
	AreaDrawio Area = "drawio"
	// AreaCommandRefs syncs command reference docs (creates missing, removes orphans).
	AreaCommandRefs Area = "command-refs"
	// AreaAll runs all areas (default when --area not specified).
	AreaAll Area = "all"
)

// ValidAreas returns the list of valid area values.
func ValidAreas() []Area {
	return []Area{AreaMermaid, AreaDrawio, AreaCommandRefs, AreaAll}
}

// ValidAreaStrings returns the list of valid area values as strings.
func ValidAreaStrings() []string {
	return []string{"mermaid", "drawio", "command-refs", "all"}
}

// ParseArea validates and parses an area string.
func ParseArea(s string) (Area, error) {
	switch s {
	case "mermaid":
		return AreaMermaid, nil
	case "drawio":
		return AreaDrawio, nil
	case "command-refs":
		return AreaCommandRefs, nil
	case "all":
		return AreaAll, nil
	default:
		return "", fmt.Errorf("invalid area: %s (valid: mermaid, drawio, command-refs, all)", s)
	}
}

// AreaSet manages selected areas.
type AreaSet struct {
	areas map[Area]bool
}

// NewAreaSet creates an area set from a list of areas.
// An empty list or a list containing "all" means all areas.
func NewAreaSet(areas []Area) *AreaSet {
	s := &AreaSet{areas: make(map[Area]bool)}
	for _, a := range areas {
		if a == AreaAll {
			return &AreaSet{areas: make(map[Area]bool)}
		}
		s.areas[a] = true
	}
	return s
}

// Includes returns true if the area should run.
// An empty area set means all areas are included.
func (s *AreaSet) Includes(area Area) bool {
	if len(s.areas) == 0 {
		return true
	}
	return s.areas[area]
}

// IsAll returns true if all areas are selected.
func (s *AreaSet) IsAll() bool {
	return len(s.areas) == 0
}

// Selected returns the list of explicitly selected areas.
// Returns nil if all areas are selected.
func (s *AreaSet) Selected() []Area {
	if len(s.areas) == 0 {
		return nil
	}
	result := make([]Area, 0, len(s.areas))
	for a := range s.areas {
		result = append(result, a)
	}
	return result
}
