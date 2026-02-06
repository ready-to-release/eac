// Package envconsts provides environment variable constants for the r2r CLI.
//
// This package exists to maintain go/cli/r2r's architectural isolation requirement.
// The r2r module must remain fully isolated with no local dependencies on go/core
// or other modules, allowing it to be lightweight and independently distributable.
//
// These constants duplicate values from go/core/environments/constants.go by design,
// as required by the module isolation architecture constraint.
package envconsts

// Environment variable names used by the r2r CLI.
const (
	// Application configuration and behavior.
	EnvR2RConfig         = "R2R_CONFIG"
	EnvR2RConfigPath     = "R2R_CONFIG_PATH"
	EnvR2RFixedRedirect  = "R2R_FIXED_REDIRECT"
	EnvR2RSkipPinWarning = "R2R_SKIP_PIN_WARNING"
	EnvR2RNoUpdateCheck  = "R2R_NO_UPDATE_CHECK"

	// Debug and logging configuration.
	EnvR2RDebug      = "R2R_DEBUG"
	EnvR2RLogLevel   = "R2R_LOG_LEVEL"
	EnvR2RVerboseLog = "R2R_VERBOSE_LOG"

	// Testing infrastructure.
	EnvR2RTesting      = "R2R_TESTING"
	EnvR2RCheckTags    = "R2R_CHECK_TAGS"
	EnvR2ROriginalArgs = "R2R_ORIGINAL_ARGS"
	EnvR2RFilteredArgs = "R2R_FILTERED_ARGS"

	// Docker and container runtime configuration.
	EnvR2RHostRepoRoot       = "R2R_HOST_REPOROOT"       // Host's repository root path (DinD mode)
	EnvR2RContainerRepoRoot  = "R2R_CONTAINER_REPOROOT"  // Container's repository root path (typically /var/task)
	EnvR2RDockerMode         = "R2R_DOCKER_MODE"         // Explicitly set when launching containers
	EnvR2RDockerHost         = "R2R_DOCKER_HOST"         // Custom Docker daemon host override
	EnvR2RHostGOOS           = "R2R_HOST_GOOS"           // Host operating system (propagated to containers)
	EnvR2RHostGOARCH         = "R2R_HOST_GOARCH"         // Host architecture (propagated to containers)
	EnvR2RTerminalDetection  = "R2R_TERMINAL_DETECTION"  // Terminal size detection method (auto/default)
)
