package terminal

import (
	"testing"
)

func TestGetWidth(t *testing.T) {
	width := GetWidth()
	if width <= 0 {
		t.Errorf("GetWidth() returned invalid width: %d", width)
	}
	
	// Should return at least the default value
	if width < 80 {
		t.Errorf("GetWidth() returned width less than default: %d", width)
	}
}

func TestGetSize(t *testing.T) {
	width, height, err := GetSize()
	
	// In CI or non-terminal environments, this might fail
	if err != nil {
		t.Skipf("GetSize() failed (possibly in CI): %v", err)
		return
	}
	
	if width <= 0 || height <= 0 {
		t.Errorf("GetSize() returned invalid dimensions: width=%d, height=%d", width, height)
	}
}

func TestIsTerminal(t *testing.T) {
	// This test might fail in CI environments
	result := IsTerminal()
	t.Logf("IsTerminal() returned: %v", result)
}