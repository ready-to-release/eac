// Package environments provides shared environment variable constants
// used across commands, specs, and core infrastructure.
//
// All application-specific environment variables should be defined here
// to provide a single source of truth and eliminate typo risks from
// hardcoded string literals.
//
// Constants are organised into domain-scoped files:
//   - constants.go: Application config, paths, Docker/container, build/system
//   - constants_ci.go: CI/CD platform detection and GitHub Actions metadata
//   - constants_testing.go: Test infrastructure, BDD framework, mock config
//   - constants_debug.go: Debug and logging configuration
//
// # Excluded Variables
//
// The following variables are intentionally NOT added as constants:
//   - Standard OS variables: COLUMNS, LINES, TERM, HOSTNAME, HOME, USER
//   - Third-party API keys: OPENAI_API_KEY, ANTHROPIC_API_KEY, etc.
//   - Go toolchain internals: GOPATH, GOROOT
//
// # Usage
//
//	import "github.com/ready-to-release/eac/go/core/environments"
//
//	if containerRoot := os.Getenv(environments.EnvCLIEContainerRoot); containerRoot != "" {
//	    // use container root
//	}
package environments

// Application configuration, paths, Docker/container, build, and system constants.
const (
	// Repository path environment variables.
	EnvCLIEPWD           = "CLIE_PWD"
	EnvCLIERepoRoot      = "CLIE_REPO_ROOT"
	EnvCLIEContainerRoot = "CLIE_CONTAINER_ROOT"
	EnvCLIEDockerMode    = "CLIE_DOCKER_MODE"

	// Docker and container runtime configuration.
	EnvCLIEHostRepoRoot      = "CLIE_HOST_REPOROOT"
	EnvCLIEContainerRepoRoot = "CLIE_CONTAINER_REPOROOT"
	EnvCLIEDockerHost        = "CLIE_DOCKER_HOST"
	EnvCLIEHostGOOS          = "CLIE_HOST_GOOS"
	EnvCLIEHostGOARCH        = "CLIE_HOST_GOARCH"
	EnvCLIETerminalDetection = "CLIE_TERMINAL_DETECTION"

	// Application configuration and behavior.
	EnvCLIEConfig         = "CLIE_CONFIG"
	EnvCLIEConfigPath     = "CLIE_CONFIG_PATH"
	EnvCLIEContext        = "CLIE_CONTEXT"
	EnvCLIENoBrowser      = "CLIE_NO_BROWSER"
	EnvCLIENoUpdateCheck  = "CLIE_NO_UPDATE_CHECK"
	EnvCLIESkipPinWarning = "CLIE_SKIP_PIN_WARNING"
	EnvCLIEFixedRedirect  = "CLIE_FIXED_REDIRECT"

	// Build and execution.
	EnvEACUseDirectBinary = "EAC_USE_DIRECT_BINARY"

	// Build and system configuration.
	EnvCommandsPath      = "COMMANDS_PATH"
	EnvGOOS              = "GOOS"
	EnvEACPortRangeStart = "EAC_PORT_RANGE_START"
	EnvEACPortRangeEnd   = "EAC_PORT_RANGE_END"

	// Git configuration.
	EnvGitAuthorName  = "GIT_AUTHOR_NAME"
	EnvGitAuthorEmail = "GIT_AUTHOR_EMAIL"
)
