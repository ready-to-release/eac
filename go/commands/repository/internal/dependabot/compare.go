package dependabot

// Compare produces a ComparisonReport from declared (dependabot.yml) and discovered (filesystem) entries.
func Compare(declared []UpdateEntry, discovered []EcosystemEntry) *ComparisonReport {
	report := &ComparisonReport{
		Declared:   declared,
		Discovered: discovered,
	}

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
