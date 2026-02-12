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
//
// The unsafe.Sizeof and unsafe.Pointer usage below is the standard Go pattern
// for calling Windows APIs via syscall. The MEMORYSTATUSEX struct requires its
// Length field to be set to the struct size before the call, and the struct must
// be passed as a pointer to the syscall. There is no safe alternative for this
// Windows API interop pattern. See: https://learn.microsoft.com/en-us/windows/win32/api/sysinfoapi/nf-sysinfoapi-globalmemorystatusex
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
