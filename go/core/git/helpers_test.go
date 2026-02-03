//go:build L1 && ov
// +build L1,ov

package git

// Test helper functions that provide backward-compatible Init/Open signatures
// for tests. These wrap RepositoryManager with a nil logger.

// Init creates a new git repository at the given path.
// This is a test helper that wraps RepositoryManager.Init with a nil logger.
func Init(path string) (*Repository, error) {
	mgr := NewManager(nil)
	return mgr.Init(path)
}

// Open opens an existing git repository at the given path.
// This is a test helper that wraps RepositoryManager.Open with a nil logger.
func Open(path string) (*Repository, error) {
	mgr := NewManager(nil)
	return mgr.Open(path)
}
