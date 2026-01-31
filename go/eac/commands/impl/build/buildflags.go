package build

import (
	"fmt"
	"strconv"
	"strings"
)

// BuildSpecificFlags holds flags that are specific to the build command.
// These are not shared with other commands (test, lint, scan).
type BuildSpecificFlags struct {
	TidyFirst       bool   // --tidy-first: Run go mod tidy before building
	NoTidy          bool   // --no-tidy: Skip go mod tidy
	UseExistingDepm bool   // --use-existing-depm: Skip building deps if artifacts exist
	LayeredBuild    bool   // --layered-build: Execute layers sequentially
	Version         string // --version: Version string for binary
	Reproducible    string // --reproducible: MkDocs reproducibility mode
	AcceptWarnings  bool   // --accept-warnings: Don't fail on MkDocs warnings
	ListArtifacts   bool   // --list-artifacts: List artifacts without building
	BuildAll        bool   // --all: Build all artifacts variants
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
	case "--tidy-first":
		flags.TidyFirst = true
		return true, 0, nil
	case "--no-tidy":
		flags.NoTidy = true
		return true, 0, nil
	case "--use-existing-depm":
		flags.UseExistingDepm = true
		return true, 0, nil
	case "--layered-build":
		flags.LayeredBuild = true
		return true, 0, nil
	case "--accept-warnings":
		flags.AcceptWarnings = true
		return true, 0, nil
	case "--list-artifacts":
		flags.ListArtifacts = true
		return true, 0, nil
	case "--all":
		flags.BuildAll = true
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

	return false, 0, nil
}

// ParseIntArg parses a string argument as an integer.
// Exported for use by other packages.
func ParseIntArg(s string) (int, error) {
	return strconv.Atoi(s)
}
