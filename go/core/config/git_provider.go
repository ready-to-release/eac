package config

// GitRemoteProvider is a function that returns the remote URL for a given
// repository root and remote name. This allows the config package to resolve
// git remote URLs without importing os/exec.
//
// Set via SetGitRemoteProvider() during CLI bootstrap, before config loading.
// Falls back to directory-name inference if not set.
type GitRemoteProvider func(repoRoot, remoteName string) (string, error)

var gitRemoteProvider GitRemoteProvider

// SetGitRemoteProvider configures the function used to resolve git remote URLs.
// Must be called before LoadAll(). Typically called during CLI bootstrap:
//
//	config.SetGitRemoteProvider(func(root, remote string) (string, error) {
//	    mgr := git.NewManager(nil)
//	    repo, err := mgr.Open(root)
//	    if err != nil {
//	        return "", err
//	    }
//	    return repo.RemoteURL(remote)
//	})
func SetGitRemoteProvider(provider GitRemoteProvider) {
	gitRemoteProvider = provider
}

// resolveGitRemoteURL returns the remote URL using the configured provider.
// Returns empty string if no provider is set (caller handles fallback).
func resolveGitRemoteURL(repoRoot, remoteName string) (string, error) {
	if gitRemoteProvider == nil {
		return "", nil
	}
	return gitRemoteProvider(repoRoot, remoteName)
}
