//go:build !windows

package fileutil

import "os"

// RemoveAllWithRetry removes path. On non-Windows platforms,
// this delegates directly to os.RemoveAll.
func RemoveAllWithRetry(path string) error {
	return os.RemoveAll(path)
}
