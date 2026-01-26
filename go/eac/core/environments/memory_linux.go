//go:build linux
// +build linux

package environments

import "syscall"

// getSystemMemoryBytes returns total system memory on Linux.
func getSystemMemoryBytes() uint64 {
	var info syscall.Sysinfo_t
	if err := syscall.Sysinfo(&info); err != nil {
		return 0
	}
	// Total RAM in bytes
	return uint64(info.Totalram) * uint64(info.Unit)
}
