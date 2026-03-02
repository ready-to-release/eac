package flags

// RebuildUnconsumedArgs reconstructs remaining and positional args in their
// original order. The shared parser separates unknown flags (remaining) from
// positional args (positional), but this loses ordering information needed for
// value-taking command-specific flags like --component, --version.
//
// By restoring original order, command-specific parsers can correctly pair
// flags with their values (e.g., "--component site" stays together).
func RebuildUnconsumedArgs(originalArgs, remaining, positional []string) []string {
	// Count occurrences of each unconsumed arg
	unconsumed := make(map[string]int)
	for _, r := range remaining {
		unconsumed[r]++
	}
	for _, p := range positional {
		unconsumed[p]++
	}

	// Scan original args, keeping only unconsumed ones in original order
	var result []string
	for _, arg := range originalArgs {
		if unconsumed[arg] > 0 {
			result = append(result, arg)
			unconsumed[arg]--
		}
	}
	return result
}
