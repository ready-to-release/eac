// Runtime environment detection for CLI execution context.
// This is hardcoded - not configurable via YAML.
package environments

import "os"

// RuntimeEnv represents the CLI execution context.
// Only two values: DevBox (local development) or CI (continuous integration).
type RuntimeEnv string

const (
	// DevBox is the default local development environment.
	DevBox RuntimeEnv = "devbox"

	// CI is the continuous integration environment (GitHub Actions, etc.).
	CI RuntimeEnv = "ci"
)

// DetectRuntime returns the current runtime environment.
// Returns CI if running in a CI system, DevBox otherwise.
func DetectRuntime() RuntimeEnv {
	// Check for CI environment indicators
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		return CI
	}
	return DevBox
}

// IsCI returns true if running in a CI environment.
func IsCI() bool {
	return DetectRuntime() == CI
}

// IsDevBox returns true if running in a local development environment.
func IsDevBox() bool {
	return DetectRuntime() == DevBox
}

// String returns the string representation of the runtime environment.
func (r RuntimeEnv) String() string {
	return string(r)
}
