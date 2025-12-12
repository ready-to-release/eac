package initsummary

import (
	"fmt"
	"strings"
)

// FormatCompact formats the summary in a compact format for TUI Init pane.
// Space-efficient, single-line items, suitable for limited display area.
func FormatCompact(s *Summary) string {
	var b strings.Builder

	// Context line
	b.WriteString(fmt.Sprintf("Context: %s\n", s.ExecutionContext))

	// Modules summary
	if s.HasDepmAdded() {
		b.WriteString(fmt.Sprintf("Modules: %d requested → %d total (+%d depm)\n",
			len(s.RequestedModules),
			len(s.CalculatedModules),
			len(s.AddedDepm)))
	} else if len(s.CalculatedModules) > 0 {
		b.WriteString(fmt.Sprintf("Modules: %d\n", len(s.CalculatedModules)))
	}

	// Layers (only if multiple)
	if s.LayerCount > 1 {
		b.WriteString(fmt.Sprintf("Layers: %d (%s)\n", s.LayerCount, formatLayerSizes(s.LayerSizes())))
	}

	// Test suite info
	if s.HasTestInfo() {
		if len(s.Test.SuitesIncluded) > 0 {
			b.WriteString(fmt.Sprintf("Suite: %s (%s)\n", s.Test.SuiteName, strings.Join(s.Test.SuitesIncluded, ", ")))
		} else {
			b.WriteString(fmt.Sprintf("Suite: %s\n", s.Test.SuiteName))
		}
		b.WriteString(fmt.Sprintf("Tests: %d selected (of %d discovered)\n",
			s.Test.Selected, s.Test.TotalDiscovered))
		// Module summary for tests
		withTests := len(s.Test.ModulesInScope)
		noTests := len(s.Test.ModulesNoTests)
		total := withTests + noTests
		if total > 0 {
			b.WriteString(fmt.Sprintf("Modules: %d/%d with tests, %d/%d no tests\n",
				withTests, total, noTests, total))
		}
	}

	// Key flags (only show non-default)
	flagSummary := formatFlagsCompact(s.Flags)
	if flagSummary != "" {
		b.WriteString(fmt.Sprintf("Flags: %s\n", flagSummary))
	}

	// Depm status
	if s.DepmStatus.Skipped {
		b.WriteString("Depm: ⏭️  skipped (--skip-depm)\n")
	} else if s.DepmStatus.Verified && len(s.DepmStatus.Missing) > 0 {
		b.WriteString(fmt.Sprintf("Depm: ❌ %d/%d resolved (%d missing)\n",
			len(s.DepmStatus.Resolved),
			s.DepmStatus.Total,
			len(s.DepmStatus.Missing)))
	} else if s.DepmStatus.Verified && s.DepmStatus.Total > 0 {
		b.WriteString(fmt.Sprintf("Depm: ✅ %d/%d resolved\n",
			len(s.DepmStatus.Resolved),
			s.DepmStatus.Total))
	}

	// Deps status
	if s.DepsStatus.Skipped {
		b.WriteString("Deps: ⏭️  skipped (--skip-deps-verification)\n")
	} else if s.DepsStatus.Verified && len(s.DepsStatus.Missing) > 0 {
		b.WriteString(fmt.Sprintf("Deps: ❌ %d/%d available (%s missing)\n",
			len(s.DepsStatus.Available)-len(s.DepsStatus.Missing),
			len(s.DepsStatus.Required),
			strings.Join(s.DepsStatus.Missing, ", ")))
	} else if s.DepsStatus.Verified && len(s.DepsStatus.Required) > 0 {
		b.WriteString(fmt.Sprintf("Deps: ✅ %d/%d available\n",
			len(s.DepsStatus.Available),
			len(s.DepsStatus.Required)))
	}

	// Artifact validation (test only)
	if s.HasArtifactValidation() {
		av := s.ArtifactValidation
		if av.AllValid() {
			b.WriteString(fmt.Sprintf("Artifacts: ✅ %d module(s) validated\n", len(av.ModulesChecked)))
		} else if !av.AllPresent {
			b.WriteString(fmt.Sprintf("Artifacts: ❌ missing from %s\n", strings.Join(av.MissingFrom, ", ")))
		} else if !av.AllCurrent {
			b.WriteString(fmt.Sprintf("Artifacts: ⚠️ stale builds: %s\n", strings.Join(av.StaleModules, ", ")))
		}
	}

	// Incremental (build only)
	if s.HasIncremental() {
		inc := s.Incremental
		if inc.FreshBuild {
			b.WriteString("Incremental: fresh build (no prior state)\n")
		} else if inc.Enabled {
			total := len(inc.Changed) + len(inc.UpToDate)
			if len(inc.UpToDate) > 0 {
				// Some modules skipped - show ratio
				b.WriteString(fmt.Sprintf("Incremental: %d of %d changed (%v)\n",
					len(inc.Changed), total, inc.DetectionTime.Round(1e6)))
			} else {
				// All modules need building - just show detection time
				b.WriteString(fmt.Sprintf("Incremental: all %d changed (%v)\n",
					total, inc.DetectionTime.Round(1e6)))
			}
		}
	}

	// Output directory
	if s.OutputDir != "" {
		b.WriteString(fmt.Sprintf("Output: %s\n", s.OutputDir))
	}

	return b.String()
}

// FormatDetailed formats the summary with full details for non-TUI/console output.
// Complete information, multi-line sections, suitable for logging.
func FormatDetailed(s *Summary) string {
	var b strings.Builder

	// Header
	commandTitle := strings.ToUpper(s.Command[:1]) + s.Command[1:]
	b.WriteString(fmt.Sprintf("═══ %s Initialization ═══\n\n", commandTitle))

	// Execution context
	b.WriteString(fmt.Sprintf("Execution Context: %s\n\n", s.ExecutionContext))

	// Test suite info (if test command)
	if s.HasTestInfo() {
		b.WriteString("── Suite ──\n")
		b.WriteString(fmt.Sprintf("  Name: %s\n", s.Test.SuiteName))
		if s.Test.SuiteDescription != "" {
			b.WriteString(fmt.Sprintf("  Description: %s\n", s.Test.SuiteDescription))
		}
		if len(s.Test.SuitesIncluded) > 0 {
			b.WriteString(fmt.Sprintf("  Includes: %s\n", strings.Join(s.Test.SuitesIncluded, ", ")))
		}
		b.WriteString("\n")

		b.WriteString("── Test Discovery ──\n")
		b.WriteString(fmt.Sprintf("  Discovered: %d tests\n", s.Test.TotalDiscovered))
		if s.Test.Skipped > 0 {
			b.WriteString(fmt.Sprintf("  Skipped (@skip:*): %d\n", s.Test.Skipped))
		}
		if s.Test.NotMatchingSuite > 0 {
			b.WriteString(fmt.Sprintf("  Not matching suite: %d\n", s.Test.NotMatchingSuite))
		}
		if s.Test.OSFiltered > 0 {
			b.WriteString(fmt.Sprintf("  OS incompatible: %d\n", s.Test.OSFiltered))
		}
		b.WriteString(fmt.Sprintf("  Selected: %d\n", s.Test.Selected))
		if s.Test.InferenceRulesApplied > 0 {
			b.WriteString(fmt.Sprintf("  Inference rules applied: %d\n", s.Test.InferenceRulesApplied))
		}
		b.WriteString("\n")

		// Module filtering (test only)
		b.WriteString("── Modules ──\n")
		if len(s.Test.ModulesRequested) > 0 {
			b.WriteString(fmt.Sprintf("  Requested: %d (%s)\n", len(s.Test.ModulesRequested), strings.Join(s.Test.ModulesRequested, ", ")))
		} else {
			b.WriteString("  Requested: all\n")
		}
		b.WriteString(fmt.Sprintf("  In scope: %d (%s)\n", len(s.Test.ModulesInScope), strings.Join(s.Test.ModulesInScope, ", ")))
		if len(s.Test.ModulesNoTests) > 0 {
			b.WriteString(fmt.Sprintf("  No tests: %d (%s)\n", len(s.Test.ModulesNoTests), strings.Join(s.Test.ModulesNoTests, ", ")))
		}
		b.WriteString("\n")
	}

	// Modules section (build only - tests have their own module section above)
	if !s.HasTestInfo() {
		b.WriteString("── Modules ──\n")
		b.WriteString(fmt.Sprintf("  Requested: %d%s\n",
			len(s.RequestedModules),
			formatModuleListInline(s.RequestedModules, 60)))

		if s.HasDepmAdded() {
			b.WriteString(fmt.Sprintf("  Added depm: %d%s\n",
				len(s.AddedDepm),
				formatModuleListInline(s.AddedDepm, 60)))
			b.WriteString(fmt.Sprintf("  Total: %d modules\n", len(s.CalculatedModules)))
		}
		b.WriteString("\n")
	}

	// Execution plan
	if s.LayerCount > 0 {
		b.WriteString("── Execution Plan ──\n")
		b.WriteString(fmt.Sprintf("  Layers: %d\n", s.LayerCount))
		for i, layer := range s.ExecutionLayers {
			b.WriteString(fmt.Sprintf("  Layer %d: %s\n", i+1, strings.Join(layer, ", ")))
		}
		b.WriteString("\n")
	}

	// Flags section
	b.WriteString("── Flags ──\n")
	b.WriteString(formatFlagsDetailed(s.Flags, s.Command))
	b.WriteString("\n")

	// Module dependencies (depm)
	b.WriteString("── Module Dependencies (depm) ──\n")
	if s.DepmStatus.Skipped {
		b.WriteString("  ⏭️  Skipped (--skip-depm)\n")
	} else if !s.DepmStatus.Verified {
		b.WriteString("  Not verified\n")
	} else if s.DepmStatus.Total == 0 {
		b.WriteString("  ✅ No module dependencies\n")
	} else if len(s.DepmStatus.Missing) > 0 {
		b.WriteString(fmt.Sprintf("  ❌ %d/%d resolved\n", len(s.DepmStatus.Resolved), s.DepmStatus.Total))
		b.WriteString(fmt.Sprintf("  Missing: %s\n", strings.Join(s.DepmStatus.Missing, ", ")))
	} else {
		b.WriteString(fmt.Sprintf("  ✅ %d/%d resolved\n", len(s.DepmStatus.Resolved), s.DepmStatus.Total))
		if len(s.DepmStatus.Resolved) > 0 {
			b.WriteString(fmt.Sprintf("  Modules: %s\n", strings.Join(s.DepmStatus.Resolved, ", ")))
		}
	}
	b.WriteString("\n")

	// System dependencies (deps)
	b.WriteString("── System Dependencies (deps) ──\n")
	if s.DepsStatus.Skipped {
		b.WriteString("  ⏭️  Skipped (--skip-deps-verification)\n")
	} else if !s.DepsStatus.Verified {
		b.WriteString("  Not verified\n")
	} else if len(s.DepsStatus.Required) == 0 {
		b.WriteString("  ✅ No system dependencies required\n")
	} else {
		for _, dep := range s.DepsStatus.Available {
			if dep.Available {
				if dep.Version != "" {
					b.WriteString(fmt.Sprintf("  ✅ %s (%s)\n", dep.Name, dep.Version))
				} else {
					b.WriteString(fmt.Sprintf("  ✅ %s\n", dep.Name))
				}
			} else {
				b.WriteString(fmt.Sprintf("  ❌ %s - not available\n", dep.Name))
			}
		}
	}
	b.WriteString("\n")

	// Artifact validation (test only)
	if s.HasArtifactValidation() {
		b.WriteString("── Build Artifacts ──\n")
		av := s.ArtifactValidation
		if av.AllValid() {
			b.WriteString(fmt.Sprintf("  ✅ Validated %d module(s)\n", len(av.ModulesChecked)))
			if len(av.ModulesChecked) > 0 {
				b.WriteString(fmt.Sprintf("  Modules: %s\n", strings.Join(av.ModulesChecked, ", ")))
			}
		} else {
			if !av.AllPresent {
				b.WriteString(fmt.Sprintf("  ❌ Missing artifacts from: %s\n", strings.Join(av.MissingFrom, ", ")))
				// Show top 5 missing artifacts
				if len(av.MissingArtifactDetails) > 0 {
					b.WriteString("  Top missing:\n")
					count := 0
					for module, artifacts := range av.MissingArtifactDetails {
						for _, artifact := range artifacts {
							if count >= 5 {
								remaining := countTotalMissingArtifacts(av.MissingArtifactDetails) - 5
								if remaining > 0 {
									b.WriteString(fmt.Sprintf("    ... and %d more\n", remaining))
								}
								break
							}
							b.WriteString(fmt.Sprintf("    - %s: %s\n", module, artifact))
							count++
						}
						if count >= 5 {
							break
						}
					}
				}
			}
			if !av.AllCurrent && len(av.StaleModules) > 0 {
				b.WriteString(fmt.Sprintf("  ⚠️  Stale builds: %s\n", strings.Join(av.StaleModules, ", ")))
				// Show top 5 stale reasons
				b.WriteString("  Reasons:\n")
				count := 0
				for _, module := range av.StaleModules {
					if count >= 5 {
						remaining := len(av.StaleModules) - 5
						if remaining > 0 {
							b.WriteString(fmt.Sprintf("    ... and %d more\n", remaining))
						}
						break
					}
					reason := av.StaleReasons[module]
					b.WriteString(fmt.Sprintf("    - %s: %s\n", module, reason))
					count++
				}
			}
		}
		b.WriteString("\n")
	}

	// Incremental build (build only)
	if s.HasIncremental() {
		b.WriteString("── Incremental Build ──\n")
		inc := s.Incremental
		if !inc.Enabled {
			b.WriteString("  ⏭️  Disabled (CI or --rebuild)\n")
		} else if inc.FreshBuild {
			b.WriteString("  🆕 Fresh build (no prior state)\n")
		} else {
			b.WriteString(fmt.Sprintf("  ✅ Enabled (detection: %v)\n", inc.DetectionTime.Round(1e6)))
			b.WriteString(fmt.Sprintf("  Building: %d modules\n", len(inc.Changed)))
			b.WriteString(fmt.Sprintf("  Skipped: %d modules\n", len(inc.UpToDate)))
		}
		b.WriteString("\n")
	}

	// Output directory
	if s.OutputDir != "" {
		b.WriteString("── Output ──\n")
		b.WriteString(fmt.Sprintf("  Directory: %s\n", s.OutputDir))
		b.WriteString("\n")
	}

	b.WriteString("═══════════════════════════════\n")

	return b.String()
}

// FormatConsole formats the summary for console output (legacy, medium detail).
// Kept for backwards compatibility.
func FormatConsole(s *Summary) string {
	return FormatDetailed(s)
}

// formatModuleListInline formats a list of modules for inline display.
func formatModuleListInline(modules []string, maxLen int) string {
	if len(modules) == 0 {
		return ""
	}
	list := truncateList(modules, maxLen)
	return " (" + list + ")"
}

// truncateList joins items with ", " and truncates with "..." if too long.
func truncateList(items []string, maxLen int) string {
	if len(items) == 0 {
		return ""
	}

	result := strings.Join(items, ", ")
	if len(result) <= maxLen {
		return result
	}

	// Truncate and add "..."
	for i := len(items) - 1; i >= 1; i-- {
		result = strings.Join(items[:i], ", ") + ", ..."
		if len(result) <= maxLen {
			return result
		}
	}
	return items[0][:min(len(items[0]), maxLen-3)] + "..."
}

// formatLayerSizes formats layer sizes as "2 → 1 → 3".
func formatLayerSizes(sizes []int) string {
	if len(sizes) == 0 {
		return "none"
	}

	parts := make([]string, len(sizes))
	for i, size := range sizes {
		parts[i] = fmt.Sprintf("%d", size)
	}
	return strings.Join(parts, " → ")
}

// formatFlagsCompact returns a compact one-line summary of non-default flags.
func formatFlagsCompact(f Flags) string {
	var parts []string

	if f.SkipDepm {
		parts = append(parts, "skip-depm")
	}
	if f.SkipDepsVerification {
		parts = append(parts, "skip-deps")
	}
	if f.ForceRebuild {
		parts = append(parts, "rebuild")
	}
	if f.DryRun {
		parts = append(parts, "dry-run")
	}
	if f.BuildAll {
		parts = append(parts, "all")
	}
	if f.UseExistingDepm {
		parts = append(parts, "use-existing-depm")
	}
	if f.ListOnly {
		parts = append(parts, "list-only")
	}
	if f.Version != "" {
		parts = append(parts, "version="+f.Version)
	}

	return strings.Join(parts, ", ")
}

// formatFlagsDetailed formats flags with full detail for console output.
func formatFlagsDetailed(f Flags, command string) string {
	var lines []string

	// Build-specific flags
	if command == "build" {
		// Tidy mode
		if f.TidyFirst {
			if f.TidyExplicit {
				lines = append(lines, "  ✅ tidy-first: enabled (explicit)")
			} else {
				lines = append(lines, "  ✅ tidy-first: enabled (default for local)")
			}
		} else {
			if f.TidyExplicit {
				lines = append(lines, "  ⏭️  tidy-first: disabled (explicit)")
			} else {
				lines = append(lines, "  ⏭️  tidy-first: disabled (default for CI)")
			}
		}

		if f.ForceRebuild {
			lines = append(lines, "  🔄 rebuild: enabled")
		}
	}

	// Common flags
	if f.SkipDepm {
		lines = append(lines, "  ⚠️  skip-depm: enabled")
	}

	if f.SkipDepsVerification {
		lines = append(lines, "  ⚠️  skip-deps-verification: enabled")
	}

	if f.DryRun {
		lines = append(lines, "  🧪 dry-run: enabled")
	}

	if f.BuildAll {
		lines = append(lines, "  📦 all: enabled")
	}

	if f.UseExistingDepm {
		lines = append(lines, "  ⏭️  use-existing-depm: enabled")
	}

	if f.ListOnly {
		lines = append(lines, "  📋 list-only: enabled")
	}

	if f.Version != "" {
		lines = append(lines, fmt.Sprintf("  🔧 version: %s", f.Version))
	}

	if f.ShowTimings {
		lines = append(lines, "  ⏱️  timings: enabled")
	}

	if f.DebugMode {
		lines = append(lines, "  🐛 debug: enabled")
	}

	if len(lines) == 0 {
		lines = append(lines, "  (defaults)")
	}

	return strings.Join(lines, "\n") + "\n"
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// countTotalMissingArtifacts counts total number of missing artifacts across all modules.
func countTotalMissingArtifacts(details map[string][]string) int {
	total := 0
	for _, artifacts := range details {
		total += len(artifacts)
	}
	return total
}
