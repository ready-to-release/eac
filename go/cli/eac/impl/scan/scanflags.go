package scan

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/scan/internal"
)

// ScanSpecificFlags holds flags that are specific to the scan command.
// These are not shared with other commands (build, test, lint).
type ScanSpecificFlags struct {
	Scanners           []internal.ScannerType // --scanner: Scanner types to run
	SBOMFormat         string                 // --format: SBOM format
	VulnSeverities     []internal.Severity    // --severity: Severity filter
	SemgrepConfig      string                 // --config: SAST config
	ComplianceStandard string                 // --compliance: Compliance standard
}

// ParseScanSpecificFlags parses scan-specific flags from remaining args.
// Returns the flags and any unknown/unprocessed args.
func ParseScanSpecificFlags(args []string) (*ScanSpecificFlags, []string, error) {
	flags := &ScanSpecificFlags{
		SBOMFormat:         "cyclonedx",
		SemgrepConfig:      "auto",
		ComplianceStandard: "k8s-cis",
	}
	var remaining []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		consumed, advance, err := parseScanFlag(arg, args, i, flags)
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

func parseScanFlag(arg string, args []string, i int, flags *ScanSpecificFlags) (consumed bool, advance int, err error) {
	switch arg {
	case "--scanner":
		if i+1 >= len(args) {
			return true, 0, fmt.Errorf("--scanner requires a value")
		}
		scanners, err := parseScannerList(args[i+1])
		if err != nil {
			return true, 0, err
		}
		flags.Scanners = scanners
		return true, 1, nil
	case "--format":
		if i+1 >= len(args) {
			return true, 0, fmt.Errorf("--format requires a value")
		}
		flags.SBOMFormat = args[i+1]
		return true, 1, nil
	case "--severity":
		if i+1 >= len(args) {
			return true, 0, fmt.Errorf("--severity requires a value")
		}
		severities, err := parseSeverityList(args[i+1])
		if err != nil {
			return true, 0, err
		}
		flags.VulnSeverities = severities
		return true, 1, nil
	case "--config":
		if i+1 >= len(args) {
			return true, 0, fmt.Errorf("--config requires a value")
		}
		flags.SemgrepConfig = args[i+1]
		return true, 1, nil
	case "--compliance":
		if i+1 >= len(args) {
			return true, 0, fmt.Errorf("--compliance requires a value")
		}
		flags.ComplianceStandard = args[i+1]
		return true, 1, nil
	}

	// Handle --flag=value syntax
	if strings.HasPrefix(arg, "--scanner=") {
		scanners, err := parseScannerList(strings.TrimPrefix(arg, "--scanner="))
		if err != nil {
			return true, 0, err
		}
		flags.Scanners = scanners
		return true, 0, nil
	}
	if strings.HasPrefix(arg, "--format=") {
		flags.SBOMFormat = strings.TrimPrefix(arg, "--format=")
		return true, 0, nil
	}
	if strings.HasPrefix(arg, "--severity=") {
		severities, err := parseSeverityList(strings.TrimPrefix(arg, "--severity="))
		if err != nil {
			return true, 0, err
		}
		flags.VulnSeverities = severities
		return true, 0, nil
	}
	if strings.HasPrefix(arg, "--config=") {
		flags.SemgrepConfig = strings.TrimPrefix(arg, "--config=")
		return true, 0, nil
	}
	if strings.HasPrefix(arg, "--compliance=") {
		flags.ComplianceStandard = strings.TrimPrefix(arg, "--compliance=")
		return true, 0, nil
	}

	return false, 0, nil
}
