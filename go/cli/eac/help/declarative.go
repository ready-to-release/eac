package help

import (
	"fmt"
	"io"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/registry"
)

// BehaviorGroup represents a grouped set of declarative flags.
type BehaviorGroup struct {
	Behavior    string
	EnableFlag  *registry.FlagMetadata
	DisableFlag *registry.FlagMetadata
}

// CategorizeFlags separates declarative behavior flags from regular flags.
// Returns regular flags and behavior groups (each group contains enable and disable flags).
func CategorizeFlags(flags []registry.FlagMetadata) (regular []registry.FlagMetadata, groups []BehaviorGroup) {
	// Map behavior -> group
	behaviorMap := make(map[string]*BehaviorGroup)
	var behaviorOrder []string

	for i := range flags {
		flag := &flags[i]

		if flag.Behavior == "" {
			regular = append(regular, *flag)
			continue
		}

		// Find or create behavior group
		group, exists := behaviorMap[flag.Behavior]
		if !exists {
			group = &BehaviorGroup{Behavior: flag.Behavior}
			behaviorMap[flag.Behavior] = group
			behaviorOrder = append(behaviorOrder, flag.Behavior)
		}

		// Categorize within group
		if flag.IsEnableFlag {
			group.EnableFlag = flag
		} else {
			group.DisableFlag = flag
		}
	}

	// Convert to slice preserving order
	for _, behavior := range behaviorOrder {
		groups = append(groups, *behaviorMap[behavior])
	}

	return regular, groups
}

// PrintBehaviorFlags prints declarative flag groups with a section header.
func PrintBehaviorFlags(w io.Writer, groups []BehaviorGroup) {
	if len(groups) == 0 {
		return
	}

	fmt.Fprintln(w, "Behavior Flags:")
	for _, group := range groups {
		PrintBehaviorGroup(w, group)
	}
	fmt.Fprintln(w)
}

// PrintBehaviorGroup prints a single behavior flag group.
func PrintBehaviorGroup(w io.Writer, group BehaviorGroup) {
	if group.EnableFlag == nil || group.DisableFlag == nil {
		return // Incomplete group
	}

	// Format: --with-X / --no-X    Description (default: ON|OFF|context)
	flagPair := fmt.Sprintf("--%s / --%s",
		group.EnableFlag.Name,
		group.DisableFlag.Name)

	defaultDisplay := FormatDefault(group.EnableFlag)

	fmt.Fprintf(w, "  %-34s %s (default: %s)\n",
		flagPair,
		group.EnableFlag.Usage,
		defaultDisplay)
}

// FormatDefault formats the default value, handling environment-aware flags.
func FormatDefault(flag *registry.FlagMetadata) string {
	if !flag.EnvAware {
		if flag.DefaultValue == "true" {
			return "ON"
		}
		return "OFF"
	}

	// Environment-aware display
	if len(flag.EnvDefaults) == 0 {
		return "context-aware"
	}

	var parts []string
	if on, ok := flag.EnvDefaults["local"]; ok {
		if on {
			parts = append(parts, "ON locally")
		} else {
			parts = append(parts, "OFF locally")
		}
	}

	if on, ok := flag.EnvDefaults["CI"]; ok {
		if on {
			parts = append(parts, "ON in CI")
		} else {
			parts = append(parts, "OFF in CI")
		}
	}

	if len(parts) == 0 {
		return "context-aware"
	}
	return strings.Join(parts, ", ")
}
