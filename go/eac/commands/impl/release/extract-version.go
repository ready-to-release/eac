// Command: release extract-version
// Description: Extract and validate release version from tag or input
// Short: Extract and validate release version from tag or input
// Flag.module: type=string, usage=Module prefix (e.g., r2r-cli, docs)
// Flag.type: type=string, default=semver, usage=Version type (semver or calver)
// Flag.ref: type=string, usage=Git ref (e.g., refs/tags/r2r-cli/1.0.0)
// Flag.version: type=string, usage=Explicit version (for workflow_dispatch)
// Flag.format: type=string, default=shell, usage=Output format (shell, json, yaml)
// Long: The release extract-version command extracts version information from a git tag ref
// Long: or explicit version input, validates the format (semver or calver), and outputs
// Long: structured data for use in release workflows.
// Long:
// Long: This command replaces the extract-release-version GitHub Action with pure Go logic,
// Long: making it testable and usable locally.
// Long:
// Long: Expected Output (--format shell):
// Long:   VERSION="1.0.0"
// Long:   TAG_NAME="r2r-cli/1.0.0"
// Long:   IS_VALID="true"
// Long:
// Long: Example:
// Long:   release extract-version --module r2r-cli --ref refs/tags/r2r-cli/1.0.0
// Long:   release extract-version --module docs --type calver --version ""
// Long:   eval $(release extract-version --module r2r-cli --ref "$GITHUB_REF" --format shell)
package release

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"gopkg.in/yaml.v3"
)

func init() {
	registry.Register(ExtractVersion)
}

// ExtractVersionOutput contains the extracted version information
type ExtractVersionOutput struct {
	Version string `json:"version" yaml:"version"`
	TagName string `json:"tag_name" yaml:"tag_name"`
	IsValid bool   `json:"is_valid" yaml:"is_valid"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

// ExtractVersion extracts and validates a release version
func ExtractVersion() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "release", and "extract-version"

	module := ""
	versionType := "semver"
	ref := ""
	version := ""
	format := "shell"

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case strings.HasPrefix(arg, "--module="):
			module = strings.TrimPrefix(arg, "--module=")
		case strings.HasPrefix(arg, "--type="):
			versionType = strings.TrimPrefix(arg, "--type=")
		case strings.HasPrefix(arg, "--ref="):
			ref = strings.TrimPrefix(arg, "--ref=")
		case strings.HasPrefix(arg, "--version="):
			version = strings.TrimPrefix(arg, "--version=")
		case strings.HasPrefix(arg, "--format="):
			format = strings.TrimPrefix(arg, "--format=")
		}
	}

	if module == "" {
		fmt.Fprintln(os.Stderr, "Error: --module is required")
		return 1
	}

	output := extractVersionLogic(module, versionType, ref, version)

	// Output in requested format
	switch format {
	case "shell":
		fmt.Printf("VERSION=%q\n", output.Version)
		fmt.Printf("TAG_NAME=%q\n", output.TagName)
		fmt.Printf("IS_VALID=%q\n", fmt.Sprintf("%t", output.IsValid))
		if output.Message != "" {
			fmt.Printf("MESSAGE=%q\n", output.Message)
		}
	case "json":
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	case "yaml":
		data, _ := yaml.Marshal(output)
		fmt.Print(string(data))
	default:
		// Simple output for direct use
		fmt.Println(output.Version)
	}

	return 0
}

func extractVersionLogic(module, versionType, ref, explicitVersion string) ExtractVersionOutput {
	output := ExtractVersionOutput{}

	// Determine version based on inputs
	if ref != "" && strings.HasPrefix(ref, "refs/tags/") {
		// Extract from tag ref
		tagName := strings.TrimPrefix(ref, "refs/tags/")
		output.TagName = tagName
		output.Version = strings.TrimPrefix(tagName, module+"/")
	} else if explicitVersion != "" {
		// Use explicit version
		output.Version = explicitVersion
		output.TagName = module + "/" + explicitVersion
	} else if versionType == "calver" {
		// Auto-generate calver
		output.Version = generateCalver()
		output.TagName = module + "/" + output.Version
		output.Message = "Auto-generated calver"
	} else {
		// No version available
		output.IsValid = false
		output.Message = "No version provided (--ref or --version required)"
		return output
	}

	// Validate version format
	if versionType == "calver" {
		output.IsValid = isValidCalver(output.Version)
		if !output.IsValid {
			output.Message = fmt.Sprintf("Invalid calver format: %s (expected YYYY.MMDD.HHMM)", output.Version)
		}
	} else {
		output.IsValid = isValidSemver(output.Version)
		if !output.IsValid {
			output.Message = fmt.Sprintf("Invalid semver format: %s (expected x.y.z)", output.Version)
		}
	}

	return output
}

// generateCalver generates a calver timestamp in YYYY.MMDD.HHMM format
func generateCalver() string {
	now := time.Now().UTC()
	return fmt.Sprintf("%d.%02d%02d.%02d%02d",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute())
}

// isValidCalver validates a calver string (YYYY.MMDD.HHMM)
func isValidCalver(version string) bool {
	calverRegex := regexp.MustCompile(`^\d{4}\.\d{4}\.\d{4}$`)
	return calverRegex.MatchString(version)
}

// isValidSemver validates a semver string (x.y.z without v prefix)
func isValidSemver(version string) bool {
	// No v prefix allowed
	if strings.HasPrefix(version, "v") || strings.HasPrefix(version, "V") {
		return false
	}
	semverRegex := regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)
	return semverRegex.MatchString(version)
}
