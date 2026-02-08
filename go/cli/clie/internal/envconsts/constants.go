// Package envconsts provides environment variable constants for the clie CLI.
//
// This package exists to maintain go/cli/clie's architectural isolation requirement.
// The clie module must remain fully isolated with no local dependencies on go/core
// or other modules, allowing it to be lightweight and independently distributable.
//
// These constants duplicate values from go/core/environments/constants.go by design,
// as required by the module isolation architecture constraint.
package envconsts

// Environment variable names used by the clie CLI.
const (
	// Application configuration and behavior.
	EnvCLIEConfig         = "CLIE_CONFIG"
	EnvCLIEConfigPath     = "CLIE_CONFIG_PATH"
	EnvCLIEFixedRedirect  = "CLIE_FIXED_REDIRECT"
	EnvCLIESkipPinWarning = "CLIE_SKIP_PIN_WARNING"
	EnvCLIENoUpdateCheck  = "CLIE_NO_UPDATE_CHECK"

	// Debug and logging configuration.
	EnvCLIEDebug      = "CLIE_DEBUG"
	EnvCLIELogLevel   = "CLIE_LOG_LEVEL"
	EnvCLIEVerboseLog = "CLIE_VERBOSE_LOG"

	// Testing infrastructure.
	EnvCLIETesting      = "CLIE_TESTING"
	EnvCLIECheckTags    = "CLIE_CHECK_TAGS"
	EnvCLIEOriginalArgs = "CLIE_ORIGINAL_ARGS"
	EnvCLIEFilteredArgs = "CLIE_FILTERED_ARGS"

	// Docker and container runtime configuration.
	EnvCLIEHostRepoRoot       = "CLIE_HOST_REPOROOT"       // Host's repository root path (DinD mode)
	EnvCLIEContainerRepoRoot  = "CLIE_CONTAINER_REPOROOT"  // Container's repository root path (typically /var/task)
	EnvCLIEDockerMode         = "CLIE_DOCKER_MODE"         // Explicitly set when launching containers
	EnvCLIEDockerHost         = "CLIE_DOCKER_HOST"         // Custom Docker daemon host override
	EnvCLIEHostGOOS           = "CLIE_HOST_GOOS"           // Host operating system (propagated to containers)
	EnvCLIEHostGOARCH         = "CLIE_HOST_GOARCH"         // Host architecture (propagated to containers)
	EnvCLIETerminalDetection  = "CLIE_TERMINAL_DETECTION"  // Terminal size detection method (auto/default)
)
