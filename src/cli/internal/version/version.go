package version

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	versionMutex   sync.RWMutex
	versionPattern = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-[\w.-]+)?(?:\+[\w.-]+)?$`)
	minVersion     = "v0.0.0"
	maxVersion     = "v99.99.99"
)

// Info holds version information for the application
type Info struct {
	Version   string
	Timestamp string
	Commit    string
	BuildTime string
	Modified  string
}

// Default version info (can be set via build flags)
var (
	Version   = "undefined"
	Timestamp = time.Now().Format(time.RFC3339)
	Commit    = ""
	BuildTime = "no build time"
	Modified  = "no information about modification"
)

// GetInfo returns the current version information
func GetInfo() Info {
	versionMutex.RLock()
	defer versionMutex.RUnlock()

	return Info{
		Version:   Version,
		Timestamp: Timestamp,
		Commit:    Commit,
		BuildTime: BuildTime,
		Modified:  Modified,
	}
}

// SetVersion sets the version information (thread-safe)
func SetVersion(v, ts, c, bt, m string) {
	versionMutex.Lock()
	defer versionMutex.Unlock()

	if v != "" {
		Version = v
	}
	if ts != "" {
		Timestamp = ts
	}
	if c != "" {
		Commit = c
	}
	if bt != "" {
		BuildTime = bt
	}
	if m != "" {
		Modified = m
	}
}

// ResetToDefaults resets all version information to default values (useful for testing)
func ResetToDefaults() {
	versionMutex.Lock()
	defer versionMutex.Unlock()

	Version = "undefined"
	Timestamp = time.Now().Format(time.RFC3339)
	Commit = ""
	BuildTime = "no build time"
	Modified = "no information about modification"
}

// EnsurePrefix ensures a version string has the specified prefix
func EnsurePrefix(version, prefix string) string {
	if version == "" {
		return prefix
	}
	if !strings.HasPrefix(version, prefix) {
		return prefix + version
	}
	return version
}

// SplitVersion splits a version string into main version and prerelease parts
func SplitVersion(version string) (main, prerelease string) {
	// First, remove build metadata (e.g., "1.2.3+build" or "1.2.3-alpha+build")
	buildIdx := strings.Index(version, "+")
	if buildIdx != -1 {
		version = version[:buildIdx]
	}

	// Handle prerelease (e.g., "1.2.3-alpha")
	if idx := strings.Index(version, "-"); idx != -1 {
		return version[:idx], version[idx+1:]
	}

	return version, ""
}

// CompareVersions compares two semantic version strings
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func CompareVersions(v1, v2 string) int {
	// Remove v prefix if present
	v1Clean := strings.TrimPrefix(v1, "v")
	v2Clean := strings.TrimPrefix(v2, "v")

	// Split version and prerelease parts
	v1Main, v1Pre := SplitVersion(v1Clean)
	v2Main, v2Pre := SplitVersion(v2Clean)

	// Compare main version parts (x.y.z)
	v1Parts := strings.Split(v1Main, ".")
	v2Parts := strings.Split(v2Main, ".")

	// Ensure we have exactly 3 parts
	if len(v1Parts) != 3 || len(v2Parts) != 3 {
		return 0
	}

	// Compare each numeric part
	for i := 0; i < 3; i++ {
		n1, err1 := strconv.Atoi(v1Parts[i])
		n2, err2 := strconv.Atoi(v2Parts[i])
		if err1 != nil || err2 != nil {
			return 0
		}
		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	// Main versions are equal, compare prerelease
	// No prerelease > prerelease
	if v1Pre == "" && v2Pre != "" {
		return 1
	}
	if v1Pre != "" && v2Pre == "" {
		return -1
	}
	// Both have prerelease or both don't - consider equal for our purposes
	return 0
}

// Validate validates the current version against format and range constraints
func Validate(forceUpdate bool) error {
	versionMutex.RLock()
	defer versionMutex.RUnlock()

	// Skip validation for undefined or development versions
	if Version == "" || Version == "undefined" || Version == "dev" {
		log.Debug().Str("version", Version).Msg("Undefined or development version detected, skipping validation")
		return nil
	}

	// Check R2R_NO_UPDATE_CHECK environment variable
	if os.Getenv("R2R_NO_UPDATE_CHECK") == "true" {
		log.Debug().Msg("Update checks disabled by R2R_NO_UPDATE_CHECK environment variable")
		return nil
	}

	// Check if force-update flag is used
	if forceUpdate {
		log.Warn().Msg("Force update flag used - bypassing version validation")
		return nil
	}

	// Validate version format
	if !versionPattern.MatchString(strings.TrimPrefix(Version, "v")) {
		return fmt.Errorf("invalid version format %q (must follow semver.org specification)", Version)
	}

	// Check version range
	ver := EnsurePrefix(Version, "v")
	min := EnsurePrefix(minVersion, "v")
	max := EnsurePrefix(maxVersion, "v")

	if CompareVersions(ver, min) < 0 || CompareVersions(ver, max) > 0 {
		return fmt.Errorf("version %q is outside allowed range (%s - %s)", Version, minVersion, maxVersion)
	}

	log.Debug().Str("version", Version).Msg("Version validation passed")
	return nil
}

// IsValid checks if a version string is valid without validating against constraints
func IsValid(v string) bool {
	return versionPattern.MatchString(strings.TrimPrefix(v, "v"))
}

// GetRange returns the allowed version range
func GetRange() (min, max string) {
	return minVersion, maxVersion
}

// SetRange sets the allowed version range (for testing purposes)
func SetRange(min, max string) {
	versionMutex.Lock()
	defer versionMutex.Unlock()

	if min != "" {
		minVersion = min
	}
	if max != "" {
		maxVersion = max
	}
}
