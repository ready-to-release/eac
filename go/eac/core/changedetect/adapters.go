package changedetect

import (
	"context"

	"github.com/ready-to-release/eac/go/eac/core/git"
	"github.com/ready-to-release/eac/go/eac/core/hash"
)

// GitRepositoryAdapter adapts git.GitRepository to GitStateProvider.
type GitRepositoryAdapter struct {
	repo git.GitRepository
}

// NewGitRepositoryAdapter creates a GitStateProvider from a GitRepository.
func NewGitRepositoryAdapter(repo git.GitRepository) *GitRepositoryAdapter {
	return &GitRepositoryAdapter{repo: repo}
}

// HeadCommit returns the full SHA of HEAD.
func (a *GitRepositoryAdapter) HeadCommit(ctx context.Context) (string, error) {
	return a.repo.HeadCommit()
}

// UncommittedFiles returns paths of files with uncommitted changes.
func (a *GitRepositoryAdapter) UncommittedFiles(ctx context.Context) ([]string, error) {
	return a.repo.UncommittedFiles()
}

// HashAdapter adapts the hash package to FileHasher interface.
type HashAdapter struct{}

// NewHashAdapter creates a FileHasher using the hash package functions.
func NewHashAdapter() *HashAdapter {
	return &HashAdapter{}
}

// HashFiles computes a deterministic hash of file contents.
func (a *HashAdapter) HashFiles(ctx context.Context, workspaceRoot string, files []string) (string, error) {
	return hash.Files(workspaceRoot, files)
}

// HashUncommittedState computes a hash representing dirty working tree state.
func (a *HashAdapter) HashUncommittedState(ctx context.Context, workspaceRoot string, files []string) (string, error) {
	// The hash package's UncommittedState returns string, not error
	return hash.UncommittedState(workspaceRoot, files), nil
}

// GlobExpanderAdapter adapts the hash.ExpandGlobPatterns to ModuleFileResolver.
// This is a simple adapter that expands glob patterns directly.
// For more sophisticated module resolution, use the contracts/modules package.
type GlobExpanderAdapter struct {
	workspaceRoot string
}

// NewGlobExpanderAdapter creates a ModuleFileResolver that expands glob patterns.
func NewGlobExpanderAdapter(workspaceRoot string) *GlobExpanderAdapter {
	return &GlobExpanderAdapter{workspaceRoot: workspaceRoot}
}

// GetModuleFiles expands glob patterns. For this simple adapter, the moniker IS the glob pattern.
// For real module resolution, use a more sophisticated adapter with the contracts/modules package.
func (a *GlobExpanderAdapter) GetModuleFiles(ctx context.Context, workspaceRoot string, moniker string) ([]string, error) {
	// Simple adapter: treat moniker as a glob pattern
	return hash.ExpandGlobPatterns(workspaceRoot, []string{moniker})
}

// ModuleContractProvider provides access to module glob patterns by moniker.
// This interface allows the changedetect package to remain decoupled from the contracts package.
type ModuleContractProvider interface {
	// GetGlobPatterns returns the glob patterns for a module identified by moniker.
	// Returns nil if the module is not found.
	GetGlobPatterns(moniker string) []string
}

// ContractResolverAdapter adapts module contracts to ModuleFileResolver.
// This is the production adapter that looks up modules from the registry
// and expands their glob patterns to actual files.
type ContractResolverAdapter struct {
	provider ModuleContractProvider
}

// NewContractResolverAdapter creates a ModuleFileResolver that uses module contracts.
// The provider should be an adapter that wraps a modules.Registry.
func NewContractResolverAdapter(provider ModuleContractProvider) *ContractResolverAdapter {
	return &ContractResolverAdapter{provider: provider}
}

// GetModuleFiles resolves glob patterns from the module contract and expands them to files.
func (a *ContractResolverAdapter) GetModuleFiles(ctx context.Context, workspaceRoot string, moniker string) ([]string, error) {
	patterns := a.provider.GetGlobPatterns(moniker)
	if len(patterns) == 0 {
		// Module not found or has no patterns - return empty list
		return nil, nil
	}
	return hash.ExpandGlobPatterns(workspaceRoot, patterns)
}

// RegistryAdapter wraps a modules.Registry to implement ModuleContractProvider.
// This adapter bridges the contracts/modules package with changedetect.
type RegistryAdapter struct {
	getContract func(moniker string) (interface{ GetGlobPatterns() []string }, bool)
}

// NewRegistryAdapterFunc creates a RegistryAdapter from a function that gets contracts.
// This allows flexible integration with different registry implementations.
//
// Example usage:
//
//	adapter := NewRegistryAdapterFunc(func(m string) (interface{ GetGlobPatterns() []string }, bool) {
//	    return registry.Get(m)
//	})
func NewRegistryAdapterFunc(getContract func(moniker string) (interface{ GetGlobPatterns() []string }, bool)) *RegistryAdapter {
	return &RegistryAdapter{getContract: getContract}
}

// GetGlobPatterns returns glob patterns for the given module moniker.
func (a *RegistryAdapter) GetGlobPatterns(moniker string) []string {
	contract, exists := a.getContract(moniker)
	if !exists || contract == nil {
		return nil
	}
	return contract.GetGlobPatterns()
}
