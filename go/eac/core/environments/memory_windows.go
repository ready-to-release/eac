//go:build windows
// +build windows

package environments

import (
	"syscall"
	"unsafe"
)

// memoryStatusEx is the Windows MEMORYSTATUSEX structure.
type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

// getSystemMemoryBytes returns total system memory on Windows.
func getSystemMemoryBytes() uint64 {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	globalMemoryStatusEx := kernel32.NewProc("GlobalMemoryStatusEx")

	var memInfo memoryStatusEx
	memInfo.Length = uint32(unsafe.Sizeof(memInfo))

	ret, _, _ := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memInfo))) //nolint:errcheck // Windows syscall returns success via ret value, not error
	if ret == 0 {
		return 0
	}

	return memInfo.TotalPhys
}
