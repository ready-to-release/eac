package docker

import (
	"os"
	"runtime"
	"strconv"

	"github.com/ready-to-release/eac/go/cli/r2r/internal/conf"
	"github.com/ready-to-release/eac/go/cli/r2r/internal/logging"
	"github.com/ready-to-release/eac/go/cli/r2r/internal/terminal"
	"github.com/ready-to-release/eac/go/cli/r2r/internal/envconsts"
)

// BuildEnvironmentVars creates the environment variable list for a container.
func (ch *ContainerHost) BuildEnvironmentVars(ext *ExtensionConfig) []string {
	// In Docker-in-Docker mode, propagate the ORIGINAL host path to child containers
	// so they can correctly mount volumes for further nested containers.
	hostRepoRoot := ch.rootDir
	if existingHostRoot := os.Getenv(envconsts.EnvR2RHostRepoRoot); existingHostRoot != "" {
		hostRepoRoot = existingHostRoot
		logging.Debugf("Docker-in-Docker: propagating original host path to child: host_root=%s", hostRepoRoot)
	}

	envVars := []string{
		envconsts.EnvR2RDockerMode + "=true",
		envconsts.EnvR2RHostGOOS + "=" + runtime.GOOS,
		envconsts.EnvR2RHostGOARCH + "=" + runtime.GOARCH,
		envconsts.EnvR2RContainerRepoRoot + "=/var/task",
		envconsts.EnvR2RHostRepoRoot + "=" + hostRepoRoot,
	}

	// Add terminal dimensions
	// First try to get from environment (if already set)
	cols := os.Getenv("COLUMNS")
	lines := os.Getenv("LINES")

	// Always try to detect terminal size for better accuracy
	if width, height, err := terminal.GetSize(); err == nil && width > 0 && height > 0 {
		// Successfully detected terminal size
		cols = strconv.Itoa(width)
		lines = strconv.Itoa(height)
		logging.Debugf("Terminal size detected: detected_width=%d detected_height=%d", width, height)
		envVars = append(envVars, "COLUMNS="+cols, "LINES="+lines, envconsts.EnvR2RTerminalDetection+"=auto")
	} else {
		// Failed to detect, use environment or defaults
		if cols == "" {
			cols = "80"
		}
		if lines == "" {
			lines = "24"
		}
		logging.Debugf("Using default terminal size: cols=%s lines=%s", cols, lines)
		envVars = append(envVars, "COLUMNS="+cols, "LINES="+lines, envconsts.EnvR2RTerminalDetection+"=default")
	}

	// 1. CI Environment Detection & Defaults
	if ch.detectCIEnvironment() {
		envVars = append(envVars, ch.getCIDefaults()...)
	} else {
		// 2. Inherit Current Shell Settings (when not in CI)
		envVars = append(envVars, ch.getShellColorSettings()...)
	}

	// 3. Add global environment variables from config
	if conf.Global.Environment != nil {
		for _, env := range conf.Global.Environment.Global {
			envVars = append(envVars, env.Name+"="+env.Value)
		}

		// Add secrets from config (these get values from host environment)
		for _, secret := range conf.Global.Environment.Secrets {
			if value := os.Getenv(secret.Env); value != "" {
				envVars = append(envVars, secret.Name+"="+value)
			}
		}
	}

	// 4. Always ensure GITHUB_USERNAME and GITHUB_TOKEN are available if set in host environment
	// This is critical for extensions that need to access GitHub Container Registry
	if githubUsername := os.Getenv("GITHUB_USERNAME"); githubUsername != "" {
		envVars = append(envVars, "GITHUB_USERNAME="+githubUsername)
	}
	if githubToken := os.Getenv("GITHUB_TOKEN"); githubToken != "" {
		envVars = append(envVars, "GITHUB_TOKEN="+githubToken)
	}

	// 5. Add extension-specific env vars (these can override defaults)
	for _, env := range ext.Env {
		if env.Value != "" {
			// Static value
			envVars = append(envVars, env.Name+"="+env.Value)
		} else {
			// Passthrough from host (value omitted)
			if hostValue := os.Getenv(env.Name); hostValue != "" {
				envVars = append(envVars, env.Name+"="+hostValue)
				logging.Debugf("Passing through environment variable from host: env=%s", env.Name)
			} else if env.Required {
				logging.Errorf("Required environment variable not set on host: env=%s", env.Name)
			}
		}
	}

	return envVars
}

// detectCIEnvironment checks multiple CI indicators beyond just CI=true.
func (ch *ContainerHost) detectCIEnvironment() bool {
	ciIndicators := []string{
		"CI", "CONTINUOUS_INTEGRATION",
		"GITHUB_ACTIONS", "AZUREDEVOPS_URL", "GITLAB_CI",
		"AZURE_HTTP_USER_AGENT", "TF_BUILD", "BUILDKITE",
		"CIRCLECI", "TRAVIS", "DRONE", "SEMAPHORE",
		"APPVEYOR", "CODEBUILD_BUILD_ID", "TEAMCITY_VERSION",
	}

	for _, indicator := range ciIndicators {
		if value := os.Getenv(indicator); value != "" && value != "false" && value != "0" {
			return true
		}
	}
	return false
}

// getCIDefaults returns CI-appropriate environment settings.
func (ch *ContainerHost) getCIDefaults() []string {
	return []string{
		"NO_COLOR=1",    // Disable colors in CI
		"TERM=dumb",     // Simple terminal for CI
		"FORCE_COLOR=0", // Force disable color
		"CI=true",       // Indicate CI environment
	}
}

// getShellColorSettings inherits current shell color capabilities.
func (ch *ContainerHost) getShellColorSettings() []string {
	envVars := []string{}

	// Inherit color support from current shell
	colorEnvVars := []string{
		"TERM", "COLORTERM", "CLICOLOR", "CLICOLOR_FORCE",
		"NO_COLOR", "FORCE_COLOR", "COLOR",
	}

	for _, envVar := range colorEnvVars {
		if value := os.Getenv(envVar); value != "" {
			envVars = append(envVars, envVar+"="+value)
		}
	}

	// Apply sensible defaults if no color settings detected
	if len(envVars) == 0 {
		envVars = append(envVars, ch.getDefaultColorSettings()...)
	}

	return envVars
}

// getDefaultColorSettings provides fallback color settings for environments without explicit settings.
func (ch *ContainerHost) getDefaultColorSettings() []string {
	// Detect terminal capabilities
	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm-256color" // Sensible default for modern terminals
	}

	return []string{
		"TERM=" + term,
		"COLORTERM=truecolor", // Modern terminal default supporting full color
		// Don't set NO_COLOR or FORCE_COLOR - let programs decide based on their logic
	}
}
