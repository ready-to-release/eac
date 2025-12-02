// +build windows

package terminal

import (
	"syscall"
	"unsafe"
)

// getTerminalSize implements terminal size detection for Windows
func getTerminalSize() (width, height int, err error) {
	// First try the standard Windows API
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleScreenBufferInfo := kernel32.NewProc("GetConsoleScreenBufferInfo")

	type coord struct {
		X int16
		Y int16
	}

	type smallRect struct {
		Left   int16
		Top    int16
		Right  int16
		Bottom int16
	}

	type consoleScreenBufferInfo struct {
		Size              coord
		CursorPosition    coord
		Attributes        uint16
		Window            smallRect
		MaximumWindowSize coord
	}

	var csbi consoleScreenBufferInfo
	handle, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err != nil {
		return 0, 0, err
	}

	ret, _, apiErr := getConsoleScreenBufferInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&csbi)))
	if ret == 0 {
		// If the Windows API fails (e.g., stdout is redirected), try mode con
		if w, h, modeErr := GetConsoleSizeViaMode(); modeErr == nil && w > 0 && h > 0 {
			return w, h, nil
		}
		return 0, 0, apiErr
	}

	width = int(csbi.Window.Right - csbi.Window.Left + 1)
	height = int(csbi.Window.Bottom - csbi.Window.Top + 1)
	return width, height, nil
}
