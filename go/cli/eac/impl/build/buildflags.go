package build

import (
	"fmt"
	"strconv"
	"strings"
)

// BuildSpecificFlags holds flags that are specific to the build command.
// These are not shared with other commands (test, lint, scan).
type BuildSpecificFlags struct {
	TidyFirst       bool   // --tidy-first, --with-tidy: Run go mod tidy before building
	NoTidy          bool   // --no-tidy: Skip go mod tidy
	UseExistingDepm bool   // --use-existing-depm: Skip building deps if artifacts exist
	Version         string // --version: Version string for binary
	Reproducible    string // --reproducible: MkDocs reproducibility mode
	AcceptWarnings  bool   // --accept-warnings: Don't fail on MkDocs warnings
	ListArtifacts   bool   // --list-artifacts: List artifacts without building
	Artifacts       string // --artifacts: Artifact scope mode (all, reduced). --all is alias for --artifacts all

	// Declarative tracking field
	TidyExplicit bool // True if --with-tidy, --tidy-first, or --no-tidy was used
}

// ParseBuildSpecificFlags parses build-specific flags from remaining args.
// Returns the flags and any unknown/unprocessed args.
func ParseBuildSpecificFlags(args []string) (*BuildSpecificFlags, []string, error) {
	flags := &BuildSpecificFlags{
		Reproducible: "auto", // Default
	}
	var remaining []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		consumed, advance, err := parseBuildFlag(arg, args, i, flags)
		if err != nil {
			return nil, nil, err
		}
		if consumed {
			i += advance
		} else {
			remaining = append(remaining, arg)
		}
	}

	return flags, remaining, nil
}

func parseBuildFlag(arg string, args []string, i int, flags *BuildSpecificFlags) (consumed bool, advance int, err error) {
	switch arg {
	// Declarative tidy flag
	case "--with-tidy":
		flags.TidyFirst = true
		flags.TidyExplicit = true
		return true, 0, nil

	// Legacy tidy flag (backward compat)
	case "--tidy-first":
		flags.TidyFirst = true
		flags.TidyExplicit = true
		return true, 0, nil
	case "--no-tidy":
		flags.NoTidy = true
		flags.TidyExplicit = true
		return true, 0, nil
	case "--use-existing-depm":
		flags.UseExistingDepm = true
		return true, 0, nil
	case "--accept-warnings":
		flags.AcceptWarnings = true
		return true, 0, nil
	case "--list-artifacts":
		flags.ListArtifacts = true
		return true, 0, nil
	case "--all":
		// --all is an alias for --artifacts all
		flags.Artifacts = "all"
		return true, 0, nil
	case "--version":
		if i+1 >= len(args) {
			return true, 0, fmt.Errorf("--version requires a value")
		}
		flags.Version = args[i+1]
		return true, 1, nil
	case "--reproducible":
		if i+1 >= len(args) {
			return true, 0, fmt.Errorf("--reproducible requires a value (auto, true, false)")
		}
		val := args[i+1]
		if !isValidReproducible(val) {
			return true, 0, fmt.Errorf("--reproducible must be 'auto', 'true', or 'false'")
		}
		flags.Reproducible = val
		return true, 1, nil
	case "--artifacts":
		if i+1 >= len(args) {
			return true, 0, fmt.Errorf("--artifacts requires a value (all, reduced)")
		}
		val := args[i+1]
		if !isValidArtifactsMode(val) {
			return true, 0, fmt.Errorf("--artifacts must be 'all' or 'reduced'")
		}
		flags.Artifacts = val
		return true, 1, nil
	}

	// Handle --key=value syntax
	if strings.HasPrefix(arg, "--version=") {
		flags.Version = strings.TrimPrefix(arg, "--version=")
		return true, 0, nil
	}
	if strings.HasPrefix(arg, "--reproducible=") {
		val := strings.TrimPrefix(arg, "--reproducible=")
		if !isValidReproducible(val) {
			return true, 0, fmt.Errorf("--reproducible must be 'auto', 'true', or 'false'")
		}
		flags.Reproducible = val
		return true, 0, nil
	}
	if strings.HasPrefix(arg, "--artifacts=") {
		val := strings.TrimPrefix(arg, "--artifacts=")
		if !isValidArtifactsMode(val) {
			return true, 0, fmt.Errorf("--artifacts must be 'all' or 'reduced'")
		}
		flags.Artifacts = val
		return true, 0, nil
	}

	return false, 0, nil
}

// isValidArtifactsMode checks if an artifacts mode flag value is valid.
func isValidArtifactsMode(value string) bool {
	return value == "all" || value == "reduced"
}

// ParseIntArg parses a string argument as an integer.
// Exported for use by other packages.
func ParseIntArg(s string) (int, error) {
	return strconv.Atoi(s)
}
