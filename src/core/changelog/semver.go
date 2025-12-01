package changelog

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// BumpType indicates the type of version bump
type BumpType int

const (
	// BumpNone indicates no version bump needed
	BumpNone BumpType = iota
	// BumpPatch indicates a patch version bump (x.y.Z)
	BumpPatch
	// BumpMinor indicates a minor version bump (x.Y.0)
	BumpMinor
	// BumpMajor indicates a major version bump (X.0.0)
	BumpMajor
)

// String returns the string representation of BumpType
func (b BumpType) String() string {
	switch b {
	case BumpNone:
		return "none"
	case BumpPatch:
		return "patch"
	case BumpMinor:
		return "minor"
	case BumpMajor:
		return "major"
	default:
		return "unknown"
	}
}

// semverRegex matches semantic version format
var semverRegex = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)

// calverRegex matches calendar version format
var calverRegex = regexp.MustCompile(`^(\d{4})\.(\d{2})\.(\d{2})(?:\.(\d+))?$`)

// ParseSemver parses a semantic version string into major, minor, patch
func ParseSemver(version string) (major, minor, patch int, err error) {
	matches := semverRegex.FindStringSubmatch(version)
	if matches == nil {
		return 0, 0, 0, fmt.Errorf("invalid semver format: %s", version)
	}

	major, _ = strconv.Atoi(matches[1])
	minor, _ = strconv.Atoi(matches[2])
	patch, _ = strconv.Atoi(matches[3])

	return major, minor, patch, nil
}

// BumpSemver applies a bump to a semantic version
func BumpSemver(version string, bump BumpType) (string, error) {
	major, minor, patch, err := ParseSemver(version)
	if err != nil {
		return "", err
	}

	switch bump {
	case BumpMajor:
		major++
		minor = 0
		patch = 0
	case BumpMinor:
		minor++
		patch = 0
	case BumpPatch:
		patch++
	case BumpNone:
		// No change
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}

// NextCalver generates the next calendar version
// If existingVersions contains today's date, it adds a suffix (.1, .2, etc.)
func NextCalver(date time.Time, existingVersions []string) string {
	baseVersion := date.Format("2006.01.02")

	// Check for existing versions with today's date
	maxSuffix := -1
	for _, v := range existingVersions {
		if v == baseVersion {
			maxSuffix = 0
		} else if strings.HasPrefix(v, baseVersion+".") {
			suffix := strings.TrimPrefix(v, baseVersion+".")
			if n, err := strconv.Atoi(suffix); err == nil && n > maxSuffix {
				maxSuffix = n
			}
		}
	}

	if maxSuffix >= 0 {
		return fmt.Sprintf("%s.%d", baseVersion, maxSuffix+1)
	}

	return baseVersion
}

// CommitTypeToBump maps conventional commit types to version bumps
func CommitTypeToBump(commitType string, breaking bool) BumpType {
	if breaking {
		return BumpMajor
	}

	switch commitType {
	case "feat":
		return BumpMinor
	case "fix", "perf", "refactor", "style":
		return BumpPatch
	case "docs", "chore", "test", "ci", "build":
		return BumpNone
	default:
		return BumpNone
	}
}

// MaxBump returns the higher of two bump types
func MaxBump(a, b BumpType) BumpType {
	if a > b {
		return a
	}
	return b
}

// CalculateBump determines the appropriate version bump from entries
func CalculateBump(entries []Entry) BumpType {
	maxBump := BumpNone

	for _, e := range entries {
		bump := CommitTypeToBump(e.CommitType, e.Breaking)
		maxBump = MaxBump(maxBump, bump)

		// Short-circuit if we've already hit major
		if maxBump == BumpMajor {
			break
		}
	}

	return maxBump
}

// CalculateNextVersion determines the next version based on current and entries
func CalculateNextVersion(current string, versionType VersionType, entries []Entry, now time.Time, existingVersions []string) (string, error) {
	// Legacy function - assumes file changes if entries exist
	hasFileChanges := len(entries) > 0
	return CalculateNextVersionConstrained(current, versionType, entries, now, existingVersions, BumpMajor, hasFileChanges)
}

// CalculateNextVersionConstrained determines the next version with a maximum bump constraint
// maxBump limits the highest bump type allowed (e.g., BumpPatch means only patch bumps are allowed)
// hasFileChanges indicates if module files changed (determined by file ownership, not commit message)
func CalculateNextVersionConstrained(current string, versionType VersionType, entries []Entry, now time.Time, existingVersions []string, maxBump BumpType, hasFileChanges bool) (string, error) {
	if versionType == Calver {
		// Calver always generates new version if there are file changes
		if hasFileChanges {
			return NextCalver(now, existingVersions), nil
		}
		return current, nil
	}

	// Semver
	// Default to 0.0.1 so first release is 0.0.2 (allows 0.0.1 for bootstrap/initial setup)
	if current == "" {
		current = "0.0.1"
	}

	bump := CalculateBump(entries)

	// File changes always warrant at least a patch bump, regardless of commit message types
	// This ensures that docs/chore/test/ci/build commits still trigger releases when files change
	if hasFileChanges && bump == BumpNone {
		bump = BumpPatch
	}

	if bump == BumpNone {
		return current, nil
	}

	// Apply constraint: limit bump to maxBump
	if bump > maxBump {
		bump = maxBump
	}

	return BumpSemver(current, bump)
}

// CommitTypeToChangeType maps conventional commit types to changelog categories
func CommitTypeToChangeType(commitType string) ChangeType {
	switch commitType {
	case "feat":
		return Added
	case "fix":
		return Fixed
	case "refactor", "perf", "style":
		return Changed
	case "docs":
		return Changed
	case "chore", "test", "ci", "build":
		return Changed
	default:
		return Changed
	}
}
