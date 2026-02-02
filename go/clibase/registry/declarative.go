package registry

import (
	"strconv"
	"strings"
)

// DeclarativeFlagDefLike defines the minimal interface for declarative flag conversion.
// This mirrors flags.DeclarativeFlagDef to avoid circular imports between registry and flags.
type DeclarativeFlagDefLike struct {
	Behavior    string
	EnableFlag  string
	DisableFlag string
	DefaultOn   bool
	EnvAware    bool
	EnvDefaults map[string]bool
	Description string
}

// BuildDeclarativeMetadata converts a declarative flag definition to FlagMetadata entries.
// It generates metadata for the enable flag and disable flag.
func BuildDeclarativeMetadata(def DeclarativeFlagDefLike) []FlagMetadata {
	result := make([]FlagMetadata, 0, 2)

	enableName := stripFlagPrefix(def.EnableFlag)
	disableName := stripFlagPrefix(def.DisableFlag)

	// Enable flag (e.g., --with-cache)
	result = append(result, FlagMetadata{
		Name:         enableName,
		Type:         "bool",
		DefaultValue: boolToString(def.DefaultOn),
		Usage:        def.Description,
		Behavior:     def.Behavior,
		IsEnableFlag: true,
		PairFlagName: disableName,
		EnvAware:     def.EnvAware,
		EnvDefaults:  def.EnvDefaults,
	})

	// Disable flag (e.g., --no-cache)
	result = append(result, FlagMetadata{
		Name:         disableName,
		Type:         "bool",
		DefaultValue: "false",
		Usage:        "Disable: " + def.Description,
		Behavior:     def.Behavior,
		IsEnableFlag: false,
		PairFlagName: enableName,
	})

	return result
}

// stripFlagPrefix removes the -- prefix from a flag name.
func stripFlagPrefix(flag string) string {
	return strings.TrimPrefix(flag, "--")
}

// boolToString converts a boolean to its string representation.
func boolToString(b bool) string {
	return strconv.FormatBool(b)
}
