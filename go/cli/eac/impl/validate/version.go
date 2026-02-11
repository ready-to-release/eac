package validate

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

type validateVersionCommand struct{}

var _ core.SimpleCommandPort = (*validateVersionCommand)(nil)

func (c *validateVersionCommand) Name() string { return "validate version" }

func (c *validateVersionCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "validate-version",
		Short:         "Validate version format",
		Long:          "Validates that a version string matches the expected format.\n\nSupported types:\n  semver: x.y.z (e.g., 1.0.0, 2.3.1)\n  calver: YYYY.MMDD.HHMM (e.g., 2025.0116.1430)\n\nExit codes:\n  0 - Valid version\n  1 - Invalid version or error\n\nOutput formats:\n  default: Human readable message\n  --format shell: VALID=\"true\" TYPE=\"semver\" VERSION=\"1.0.0\"\n\nExample:\n  validate version 1.0.0 --type semver\n  validate version 2025.0116.1430 --type calver\n  eval $(validate version 1.0.0 --type semver --format shell)",
		Flags: []core.FlagSpec{
			{Name: "type", Type: "string", Usage: "Version type: semver or calver (required)"},
			{Name: "format", Type: "string", Usage: "Output format (default, shell)"},
		},
	}
}

func (c *validateVersionCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ValidateVersion()
}

// CalVer: YYYY.MMDD.HHMM.
var calverRegex = regexp.MustCompile(`^\d{4}\.\d{4}\.\d{4}$`)

func ValidateVersion() int {
	// Parse arguments
	version := ""
	versionType := ""
	format := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--help" || arg == "-h":
			// Help is handled by the framework via command comments
			return 0
		case arg == "--type" && i+1 < len(os.Args):
			versionType = os.Args[i+1]
			i++
		case arg == "--format" && i+1 < len(os.Args):
			format = os.Args[i+1]
			i++
		case !strings.HasPrefix(arg, "--") && version == "":
			version = arg
		}
	}

	if version == "" {
		fmt.Fprintln(os.Stderr, "Error: version required")
		fmt.Fprintln(os.Stderr, "Usage: validate version <version> --type <semver|calver>")
		return 1
	}

	if versionType == "" {
		fmt.Fprintln(os.Stderr, "Error: --type is required")
		fmt.Fprintln(os.Stderr, "Usage: validate version <version> --type <semver|calver>")
		return 1
	}

	// Strip leading 'v' if present for semver
	cleanVersion := version
	if versionType == "semver" && (strings.HasPrefix(version, "v") || strings.HasPrefix(version, "V")) {
		cleanVersion = version[1:]
	}

	// Validate based on type
	var valid bool
	var errorMsg string

	switch versionType {
	case "semver":
		valid = semverRegex.MatchString(cleanVersion)
		if !valid {
			errorMsg = "Expected format: x.y.z (e.g., 1.0.0)"
		}
	case "calver":
		valid = calverRegex.MatchString(cleanVersion)
		if !valid {
			errorMsg = "Expected format: YYYY.MMDD.HHMM (e.g., 2025.0116.1430)"
		}
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown version type '%s'\n", versionType)
		fmt.Fprintln(os.Stderr, "Supported types: semver, calver")
		return 1
	}

	// Output based on format
	if format == "shell" {
		fmt.Printf("VALID=\"%t\"\n", valid)
		fmt.Printf("TYPE=\"%s\"\n", versionType)
		fmt.Printf("VERSION=\"%s\"\n", cleanVersion)
		if !valid {
			fmt.Printf("ERROR=\"%s\"\n", errorMsg)
		}
	} else {
		if valid {
			fmt.Printf("Version %s is valid %s\n", cleanVersion, versionType)
		} else {
			fmt.Fprintf(os.Stderr, "Invalid %s version: %s\n", versionType, version)
			fmt.Fprintln(os.Stderr, errorMsg)
		}
	}

	if valid {
		return 0
	}
	return 1
}
