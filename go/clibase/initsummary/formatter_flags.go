package initsummary

import (
	"fmt"
	"strings"
)

// formatFlagsCompact returns a compact one-line summary of non-default flags.
func formatFlagsCompact(f Flags) string {
	var parts []string

	if f.SkipDepm {
		parts = append(parts, "skip-depm")
	}
	if f.SkipDeps {
		parts = append(parts, "skip-deps")
	}
	if f.ForceRebuild {
		parts = append(parts, "rebuild")
	}
	if f.DryRun {
		parts = append(parts, "dry-run")
	}
	if f.ArtifactsMode == "all" {
		parts = append(parts, "artifacts:all")
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

	if f.SkipDeps {
		lines = append(lines, "  ⚠️  skip-deps: enabled")
	}

	if f.DryRun {
		lines = append(lines, "  🧪 dry-run: enabled")
	}

	if f.ArtifactsMode == "all" {
		lines = append(lines, "  📦 artifacts: all")
	} else if f.ArtifactsMode != "" {
		lines = append(lines, "  📦 artifacts: "+f.ArtifactsMode)
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
