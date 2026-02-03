// Package cache provides a 2D taxonomy (Level x Type) for cache control.
// This enables fine-grained control over which caches to skip via --skip-cache=<spec>.
package cache

import (
	"fmt"
	"strings"
)

// Level represents where a cache lives.
type Level string

const (
	LevelLocal  Level = "local"  // On the local machine
	LevelRemote Level = "remote" // On the network
	LevelAll    Level = "all"    // Both levels (wildcard)
)

// Type represents what kind of cache it is.
type Type string

const (
	TypeRegistry Type = "registry" // Container/package registries
	TypeState    Type = "state"    // Incremental build state (state.json)
	TypeAsset    Type = "asset"    // Rendered assets (mermaid, structurizr)
	TypeLayer    Type = "layer"    // Build layer caches (BuildKit)
	TypeWork     Type = "work"     // Ephemeral work directories
	TypeAll      Type = "all"      // All types (wildcard)
)

// AllLevels returns all concrete Level values (excluding LevelAll).
func AllLevels() []Level {
	return []Level{LevelLocal, LevelRemote}
}

// AllTypes returns all concrete Type values (excluding TypeAll).
func AllTypes() []Type {
	return []Type{TypeRegistry, TypeState, TypeAsset, TypeLayer, TypeWork}
}

// Spec represents a cache specification (level:type pair).
type Spec struct {
	Level Level
	Type  Type
}

// String returns the canonical string representation.
func (s Spec) String() string {
	if s.Level == LevelAll && s.Type == TypeAll {
		return "all"
	}
	if s.Type == TypeAll {
		return string(s.Level)
	}
	if s.Level == LevelAll {
		return string(s.Type)
	}
	return fmt.Sprintf("%s:%s", s.Level, s.Type)
}

// Matches returns true if this spec matches the given level/type combination.
func (s Spec) Matches(level Level, typ Type) bool {
	levelMatch := s.Level == LevelAll || s.Level == level
	typeMatch := s.Type == TypeAll || s.Type == typ
	return levelMatch && typeMatch
}

// Validate checks for invalid combinations.
func (s Spec) Validate() error {
	// work type is local-only
	if s.Type == TypeWork && s.Level == LevelRemote {
		return fmt.Errorf("work cache type is local-only, cannot use remote:work")
	}
	return nil
}

// ParseSpec parses a cache specification string.
func ParseSpec(s string) (Spec, error) {
	s = strings.TrimSpace(strings.ToLower(s))

	if s == "all" {
		return Spec{Level: LevelAll, Type: TypeAll}, nil
	}

	parts := strings.Split(s, ":")
	switch len(parts) {
	case 1:
		// Single value: determine if it's a level or type
		if isLevel(parts[0]) {
			return Spec{Level: Level(parts[0]), Type: TypeAll}, nil
		}
		if isType(parts[0]) {
			return Spec{Level: LevelAll, Type: Type(parts[0])}, nil
		}
		return Spec{}, fmt.Errorf("unknown cache spec: %s", s)
	case 2:
		level := Level(parts[0])
		typ := Type(parts[1])
		if !isLevel(string(level)) {
			return Spec{}, fmt.Errorf("unknown cache level: %s", parts[0])
		}
		if !isType(string(typ)) {
			return Spec{}, fmt.Errorf("unknown cache type: %s", parts[1])
		}
		spec := Spec{Level: level, Type: typ}
		if err := spec.Validate(); err != nil {
			return Spec{}, err
		}
		return spec, nil
	default:
		return Spec{}, fmt.Errorf("invalid cache spec format: %s", s)
	}
}

func isLevel(s string) bool {
	switch Level(s) {
	case LevelLocal, LevelRemote, LevelAll:
		return true
	}
	return false
}

func isType(s string) bool {
	switch Type(s) {
	case TypeRegistry, TypeState, TypeAsset, TypeLayer, TypeWork, TypeAll:
		return true
	}
	return false
}

// ParseSpecs parses a comma-separated list of cache specs.
func ParseSpecs(s string) ([]Spec, error) {
	var specs []Spec
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		spec, err := ParseSpec(part)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}
	return specs, nil
}
