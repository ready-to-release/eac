package test

import "strings"

// BDDPackagePath is the typed representation of the colon-delimited package
// path protocol used by all BDD runners: "featureName:testRoot:featurePath".
//
//   - FeatureName: human display label and UoW spec name (e.g., "build-module")
//   - TestRoot:    execution directory where the runner is invoked (e.g., "go/commands/test")
//   - FeaturePath: specific .feature file to filter BDD execution
//     (e.g., "specs/eac/build-module/specification.feature")
type BDDPackagePath struct {
	FeatureName string
	TestRoot    string
	FeaturePath string
}

// ParseBDDPackagePath parses a colon-delimited package path into its components.
// For BDD paths (3 parts): "featureName:testRoot:featurePath"
// For unit test paths (1 part): returns FeatureName="" and TestRoot=raw.
func ParseBDDPackagePath(raw string) BDDPackagePath {
	parts := strings.SplitN(raw, ":", 3)
	switch len(parts) {
	case 3:
		return BDDPackagePath{
			FeatureName: parts[0],
			TestRoot:    parts[1],
			FeaturePath: parts[2],
		}
	default:
		return BDDPackagePath{
			TestRoot: raw,
		}
	}
}

// IsBDD returns true if this represents a BDD package path (has all 3 components).
func (p BDDPackagePath) IsBDD() bool {
	return p.FeatureName != "" && p.FeaturePath != ""
}

// String returns the canonical colon-delimited representation.
func (p BDDPackagePath) String() string {
	if p.IsBDD() {
		return p.FeatureName + ":" + p.TestRoot + ":" + p.FeaturePath
	}
	return p.TestRoot
}

// DisplayName returns the human-readable display name.
func (p BDDPackagePath) DisplayName() string {
	if p.IsBDD() {
		return p.FeatureName + ":" + p.TestRoot
	}
	return p.TestRoot
}
