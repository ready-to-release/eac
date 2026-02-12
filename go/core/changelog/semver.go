package changelog

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// BumpType indicates the type of version bump.
type BumpType int

const (
	// BumpNone indicates no version bump needed.
	BumpNone BumpType = iota
	// BumpPatch indicates a patch version bump (x.y.Z).
	BumpPatch
	// BumpMinor indicates a minor version bump (x.Y.0).
	BumpMinor
	// BumpMajor indicates a major version bump (X.0.0).
	BumpMajor
)

// String returns the string representation of BumpType.
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

// semverRegex matches semantic version format.
var semverRegex = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)

// ParseSemver parses a semantic version string into major, minor, patch.
func ParseSemver(version string) (major, minor, patch int, err error) {
	matches := semverRegex.FindStringSubmatch(version)
	if matches == nil {
		return 0, 0, 0, fmt.Errorf("invalid semver format: %s", version)
	}

	// Regex only captures digits, so Atoi errors indicate overflow
	major, err = strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %w", err)
	}
	minor, err = strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %w", err)
	}
	patch, err = strconv.Atoi(matches[3])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version: %w", err)
	}

	return major, minor, patch, nil
}

// BumpSemver applies a bump to a semantic version.
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

// CommitTypeToBump maps conventional commit types to version bumps.
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

// MaxBump returns the higher of two bump types.
func MaxBump(a, b BumpType) BumpType {
	if a > b {
		return a
	}
	return b
}

// CalculateBump determines the appropriate version bump from entries.
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

// VersionCalcOptions holds the parameters for CalculateNextVersionConstrained.
type VersionCalcOptions struct {
	// Current is the current version string (e.g., "1.2.3" or "2025.01.15").
	// If empty for semver, defaults to "0.0.1".
	Current string

	// VersionType selects the versioning scheme (Semver or Calver).
	VersionType VersionType

	// Entries are the changelog entries used to determine the bump level.
	Entries []Entry

	// Now is the current time, used for calver date generation.
	Now time.Time

	// ExistingVersions is the list of already-released versions, used
	// by calver to generate collision suffixes (e.g., 2025.01.15.1).
	ExistingVersions []string

	// MaxBump limits the highest bump type allowed (e.g., BumpPatch means
	// only patch bumps are permitted even if entries warrant a minor/major bump).
	MaxBump BumpType

	// HasFileChanges indicates whether module-owned files changed on disk
	// (determined by file ownership, not commit message). When true and the
	// calculated bump is BumpNone, the bump is elevated to BumpPatch so that
	// docs/chore/test/ci/build commits still trigger releases.
	HasFileChanges bool
}

// CalculateNextVersionConstrained determines the next version with a maximum bump constraint.
func CalculateNextVersionConstrained(current string, versionType VersionType, entries []Entry, now time.Time, existingVersions []string, maxBump BumpType, hasFileChanges bool) (string, error) {
	return CalculateNextVersion(VersionCalcOptions{
		Current:          current,
		VersionType:      versionType,
		Entries:          entries,
		Now:              now,
		ExistingVersions: existingVersions,
		MaxBump:          maxBump,
		HasFileChanges:   hasFileChanges,
	})
}

// CalculateNextVersion determines the next version using the provided options.
// This is the options-struct variant of CalculateNextVersionConstrained.
func CalculateNextVersion(opts VersionCalcOptions) (string, error) {
	current := opts.Current
	if opts.VersionType == Calver {
		// Calver always generates new version if there are file changes
		if opts.HasFileChanges {
			return NextCalver(opts.Now, opts.ExistingVersions), nil
		}
		return current, nil
	}

	// Semver
	// Default to 0.0.1 so first release is 0.0.2 (allows 0.0.1 for bootstrap/initial setup)
	if current == "" {
		current = "0.0.1"
	}

	bump := CalculateBump(opts.Entries)

	// File changes always warrant at least a patch bump, regardless of commit message types
	// This ensures that docs/chore/test/ci/build commits still trigger releases when files change
	if opts.HasFileChanges && bump == BumpNone {
		bump = BumpPatch
	}

	if bump == BumpNone {
		return current, nil
	}

	// Apply constraint: limit bump to maxBump
	if bump > opts.MaxBump {
		bump = opts.MaxBump
	}

	return BumpSemver(current, bump)
}

// CommitTypeToChangeType maps conventional commit types to changelog categories.
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
