package dependabot

// Compare produces a ComparisonReport from declared (dependabot.yml) and discovered (filesystem) entries.
// Consolidated entries (using directories: rather than directory:) are handled specially:
// their listed directories are considered covered and excluded from the 1:1 comparison.
func Compare(declared []UpdateEntry, discovered []EcosystemEntry) *ComparisonReport {
	// Partition declared entries into singular vs consolidated
	var singular []UpdateEntry
	var consolidated []UpdateEntry
	for _, d := range declared {
		if d.IsConsolidated() {
			consolidated = append(consolidated, d)
		} else {
			singular = append(singular, d)
		}
	}

	// Build set of directories covered by consolidated entries
	coveredDirs := make(map[string]bool)
	for _, c := range consolidated {
		for _, dir := range c.Directories {
			coveredDirs[c.PackageEcosystem+":"+dir] = true
		}
	}

	// Filter discovered: remove entries covered by a consolidated entry
	var uncoveredDiscovered []EcosystemEntry
	for _, e := range discovered {
		if coveredDirs[e.Key()] {
			continue
		}
		uncoveredDiscovered = append(uncoveredDiscovered, e)
	}

	// Run the 1:1 comparison on the remainder
	report := compareExact(singular, uncoveredDiscovered)
	report.Declared = declared
	report.Discovered = discovered
	report.Consolidated = consolidated
	return report
}

// compareExact performs 1:1 key matching between declared and discovered entries.
func compareExact(declared []UpdateEntry, discovered []EcosystemEntry) *ComparisonReport {
	report := &ComparisonReport{}

	declaredSet := make(map[string]UpdateEntry, len(declared))
	for _, d := range declared {
		declaredSet[d.Key()] = d
	}

	discoveredSet := make(map[string]EcosystemEntry, len(discovered))
	for _, e := range discovered {
		discoveredSet[e.Key()] = e
	}

	// Find missing (discovered but not declared) and matched
	for _, e := range discovered {
		if _, found := declaredSet[e.Key()]; !found {
			report.Missing = append(report.Missing, e)
		} else {
			report.Matched = append(report.Matched, e)
		}
	}

	// Find extra (declared but not discovered)
	for _, d := range declared {
		if _, found := discoveredSet[d.Key()]; !found {
			report.Extra = append(report.Extra, d)
		}
	}

	return report
}
