package environments

// CI/CD platform detection and GitHub Actions metadata constants.
const (
	// CI environment detection.
	EnvCI            = "CI"
	EnvGitHubActions = "GITHUB_ACTIONS"
	EnvGitLabCI      = "GITLAB_CI"

	// GitHub CI metadata.
	EnvGitHubSHA        = "GITHUB_SHA"
	EnvGitHubToken      = "GITHUB_TOKEN"
	EnvGitHubUsername   = "GITHUB_USERNAME"
	EnvGitHubActor      = "GITHUB_ACTOR"
	EnvGitHubRunID      = "GITHUB_RUN_ID"
	EnvGitHubRepository = "GITHUB_REPOSITORY"
	EnvGitHubEnv        = "GITHUB_ENV"
	EnvGHRepo           = "GH_REPO"
)
