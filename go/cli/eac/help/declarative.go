package help

import (
	"fmt"
	"io"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// BehaviorGroup represents a grouped set of declarative flags.
type BehaviorGroup struct {
	Behavior    string
	EnableFlag  *core.FlagSpec
	DisableFlag *core.FlagSpec
}

// CategorizeFlags separates declarative behavior flags from regular flags.
// Returns regular flags and behavior groups (each group contains enable and disable flags).
// Behavior flags are detected by convention: --with-X and --no-X pairs.
func CategorizeFlags(flags []core.FlagSpec) (regular []core.FlagSpec, groups []BehaviorGroup) {
	// Detect behavior flag pairs by naming convention (--with-X / --no-X)
	behaviorMap := make(map[string]*BehaviorGroup)
	var behaviorOrder []string
	behaviorFlags := make(map[int]bool) // track indices that are behavior flags

	for i := range flags {
		flag := &flags[i]
		var behavior string
		var isEnable bool

		if strings.HasPrefix(flag.Name, "with-") {
			behavior = strings.TrimPrefix(flag.Name, "with-")
			isEnable = true
		} else if strings.HasPrefix(flag.Name, "no-") {
			behavior = strings.TrimPrefix(flag.Name, "no-")
			isEnable = false
		}

		if behavior == "" {
			continue
		}

		// Check if the pair exists
		pairName := "no-" + behavior
		if !isEnable {
			pairName = "with-" + behavior
		}
		hasPair := false
		for j := range flags {
			if flags[j].Name == pairName {
				hasPair = true
				break
			}
		}

		if !hasPair {
			continue
		}

		behaviorFlags[i] = true

		group, exists := behaviorMap[behavior]
		if !exists {
			group = &BehaviorGroup{Behavior: behavior}
			behaviorMap[behavior] = group
			behaviorOrder = append(behaviorOrder, behavior)
		}

		if isEnable {
			group.EnableFlag = flag
		} else {
			group.DisableFlag = flag
		}
	}

	// Collect regular (non-behavior) flags
	for i := range flags {
		if !behaviorFlags[i] {
			regular = append(regular, flags[i])
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

// FormatDefault formats the default value for a flag.
func FormatDefault(flag *core.FlagSpec) string {
	if flag.DefaultValue == "true" {
		return "ON"
	}
	return "OFF"
}
