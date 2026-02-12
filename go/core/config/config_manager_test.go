//go:build L1 && ov

package config

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigManager_Isolated(t *testing.T) {
	mgr := NewConfigManager()
	assert.NotNil(t, mgr, "NewConfigManager should return a non-nil manager")
}

func TestConfigManager_Global_LoadsConfig(t *testing.T) {
	mgr := NewConfigManager()
	cfg := mgr.Global()
	require.NotNil(t, cfg, "Global() should return a loaded config")
	assert.NotEmpty(t, cfg.RepoRoot, "RepoRoot should be set")
	assert.NotNil(t, cfg.Repository, "Repository should be loaded")
}

func TestConfigManager_GlobalWithError(t *testing.T) {
	mgr := NewConfigManager()
	cfg, err := mgr.GlobalWithError()
	require.NoError(t, err, "GlobalWithError should not return an error")
	require.NotNil(t, cfg, "GlobalWithError should return a config")
	assert.NotEmpty(t, cfg.RepoRoot, "RepoRoot should be set")
}

func TestConfigManager_Global_IsSingleton(t *testing.T) {
	mgr := NewConfigManager()
	cfg1 := mgr.Global()
	cfg2 := mgr.Global()
	assert.Same(t, cfg1, cfg2, "Global() should return the same instance on repeated calls")
}

func TestConfigManager_Isolation(t *testing.T) {
	// Two independent managers should not share state
	mgr1 := NewConfigManager()
	mgr2 := NewConfigManager()

	// Set a mock config on mgr1
	mockCfg := &EACConfig{RepoRoot: "/mock/root"}
	mgr1.SetGlobalForTesting(mockCfg)

	// mgr2 should still load real config, not the mock
	cfg2 := mgr2.Global()
	assert.NotEqual(t, "/mock/root", cfg2.RepoRoot,
		"Isolated manager should not see mock from another manager")

	// mgr1 should return the mock
	cfg1 := mgr1.Global()
	assert.Equal(t, "/mock/root", cfg1.RepoRoot,
		"Manager with mock should return the mock config")
}

func TestConfigManager_SetGlobalForTesting(t *testing.T) {
	mgr := NewConfigManager()

	mockCfg := &EACConfig{RepoRoot: "/test/root"}
	mgr.SetGlobalForTesting(mockCfg)

	cfg := mgr.Global()
	assert.Equal(t, "/test/root", cfg.RepoRoot, "Should return the injected mock config")

	cfg, err := mgr.GlobalWithError()
	assert.NoError(t, err, "Error should be nil after SetGlobalForTesting")
	assert.Equal(t, "/test/root", cfg.RepoRoot)
}

func TestConfigManager_ResetGlobalForTesting(t *testing.T) {
	mgr := NewConfigManager()

	// Set mock, then reset
	mockCfg := &EACConfig{RepoRoot: "/test/root"}
	mgr.SetGlobalForTesting(mockCfg)
	mgr.ResetGlobalForTesting()

	// After reset, Global() should load real config
	cfg := mgr.Global()
	assert.NotEqual(t, "/test/root", cfg.RepoRoot,
		"After reset, Global() should load real config, not the mock")
	assert.NotEmpty(t, cfg.RepoRoot, "RepoRoot should be set from real loading")
}

func TestConfigManager_Cache(t *testing.T) {
	mgr := NewConfigManager()

	cache := mgr.Cache()
	require.NotNil(t, cache, "Cache() should return a non-nil cache")

	// Cache should be the same instance on repeated calls
	cache2 := mgr.Cache()
	assert.Same(t, cache, cache2, "Cache() should return the same instance")
}

func TestConfigManager_ClearCache(t *testing.T) {
	mgr := NewConfigManager()

	// Populate cache
	cache := mgr.Cache()
	testCfg := &EACConfig{RepoRoot: "/cached"}
	cache.Set("/cached", true, testCfg)

	// Verify it's cached
	got, ok := cache.Get("/cached", true)
	require.True(t, ok, "Config should be in cache")
	assert.Same(t, testCfg, got)

	// Clear and verify
	mgr.ClearCache()
	_, ok = cache.Get("/cached", true)
	assert.False(t, ok, "Cache should be empty after ClearCache")
}

func TestConfigManager_ClearCache_NilCache(t *testing.T) {
	mgr := NewConfigManager()
	// ClearCache on a manager whose cache hasn't been initialized should not panic
	mgr.ClearCache()
}

func TestConfigManager_GitRemoteProvider(t *testing.T) {
	mgr := NewConfigManager()

	// Initially nil
	assert.Nil(t, mgr.GetGitRemoteProvider(), "GitRemoteProvider should be nil initially")

	// Set a provider
	called := false
	provider := func(repoRoot, remoteName string) (string, error) {
		called = true
		return "https://github.com/test/repo.git", nil
	}
	mgr.SetGitRemoteProvider(provider)

	// Get and invoke
	got := mgr.GetGitRemoteProvider()
	require.NotNil(t, got, "GetGitRemoteProvider should return the set provider")

	url, err := got("root", "origin")
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/test/repo.git", url)
	assert.True(t, called, "Provider should have been called")
}

func TestConfigManager_GitRemoteProvider_Isolation(t *testing.T) {
	mgr1 := NewConfigManager()
	mgr2 := NewConfigManager()

	mgr1.SetGitRemoteProvider(func(root, remote string) (string, error) {
		return "mgr1-url", nil
	})

	// mgr2 should not see mgr1's provider
	assert.Nil(t, mgr2.GetGitRemoteProvider(),
		"Isolated manager should not inherit provider from another manager")

	// mgr1 should have its provider
	p := mgr1.GetGitRemoteProvider()
	require.NotNil(t, p)
	url, _ := p("", "")
	assert.Equal(t, "mgr1-url", url)
}

func TestConfigManager_ResolveGitRemoteURL_NoProvider(t *testing.T) {
	mgr := NewConfigManager()

	url, err := mgr.resolveGitRemoteURL("/root", "origin")
	assert.NoError(t, err, "Should not error when no provider is set")
	assert.Empty(t, url, "Should return empty string when no provider is set")
}

func TestConfigManager_ResolveGitRemoteURL_WithProvider(t *testing.T) {
	mgr := NewConfigManager()

	mgr.SetGitRemoteProvider(func(repoRoot, remoteName string) (string, error) {
		return "https://github.com/" + remoteName, nil
	})

	url, err := mgr.resolveGitRemoteURL("/root", "origin")
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/origin", url)
}

func TestConfigManager_ConcurrentAccess(t *testing.T) {
	mgr := NewConfigManager()

	var wg sync.WaitGroup
	const goroutines = 10

	// Concurrent Global() calls should all succeed without races
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			cfg := mgr.Global()
			assert.NotNil(t, cfg)
		}()
	}
	wg.Wait()

	// Concurrent GitRemoteProvider set/get should not race
	wg.Add(goroutines * 2)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			mgr.SetGitRemoteProvider(func(root, remote string) (string, error) {
				return "url", nil
			})
		}()
		go func() {
			defer wg.Done()
			_ = mgr.GetGitRemoteProvider()
		}()
	}
	wg.Wait()
}

// TestPackageLevelFunctions_DelegateToDefaultManager verifies that the
// package-level functions still work correctly (backward compatibility).
func TestPackageLevelFunctions_DelegateToDefaultManager(t *testing.T) {
	// ClearCache should not panic
	ClearCache()

	// Global should work
	cfg := Global()
	require.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.RepoRoot)

	// GlobalWithError should work
	cfg2, err := GlobalWithError()
	require.NoError(t, err)
	assert.Same(t, cfg, cfg2, "Global and GlobalWithError should return same instance")

	// SetGitRemoteProvider + GetGitRemoteProvider round-trip
	originalProvider := GetGitRemoteProvider()
	testProvider := func(root, remote string) (string, error) {
		return "test-url", nil
	}
	SetGitRemoteProvider(testProvider)
	defer SetGitRemoteProvider(originalProvider) // Restore

	got := GetGitRemoteProvider()
	require.NotNil(t, got)
	url, err := got("", "")
	assert.NoError(t, err)
	assert.Equal(t, "test-url", url)
}
