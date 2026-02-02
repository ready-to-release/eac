//go:build darwin
// +build darwin

package environments

import (
	"syscall"
	"unsafe"
)

// getSystemMemoryBytes returns total system memory on macOS.
func getSystemMemoryBytes() uint64 {
	// Use sysctl to get hw.memsize
	val, err := syscall.Sysctl("hw.memsize")
	if err != nil {
		return 0
	}

	// hw.memsize returns an 8-byte little-endian integer as a string
	if len(val) < 8 {
		return 0
	}

	// Convert bytes to uint64 (little-endian)
	return *(*uint64)(unsafe.Pointer(&[]byte(val)[0]))
}
