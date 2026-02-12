// Package environments provides shared environment variable constants
// used across commands, specs, and core infrastructure.
//
// All application-specific environment variables should be defined here
// to provide a single source of truth and eliminate typo risks from
// hardcoded string literals.
//
// Constants are organised into domain-scoped files:
//   - constants.go: Repository paths and git configuration
//   - constants_app.go: Application configuration and behavior
//   - constants_build.go: Build, execution, and system configuration
//   - constants_ci.go: CI/CD platform detection and GitHub Actions metadata
//   - constants_debug.go: Debug flags and logging configuration
//   - constants_docker.go: Docker and container runtime configuration
//   - constants_testing.go: Test infrastructure, BDD framework, mock config
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

// Repository path environment variables.
const (
	EnvCLIEPWD           = "CLIE_PWD"
	EnvCLIERepoRoot      = "CLIE_REPO_ROOT"
	EnvCLIEContainerRoot = "CLIE_CONTAINER_ROOT"
	EnvCLIEDockerMode    = "CLIE_DOCKER_MODE"
)

// Git configuration.
const (
	EnvGitAuthorName  = "GIT_AUTHOR_NAME"
	EnvGitAuthorEmail = "GIT_AUTHOR_EMAIL"
)
