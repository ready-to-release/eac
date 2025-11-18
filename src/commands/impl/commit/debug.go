package commit

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// debugWriter handles debug file output with consistent error handling.
//
// Thread-safe: All methods use mutex to protect concurrent file operations.
// This is required for parallel module section generation where multiple
// goroutines may write debug files simultaneously.
type debugWriter struct {
	enabled       bool
	workspaceRoot string
	mu            sync.Mutex // Protects file write operations
}

func newDebugWriter(enabled bool, workspaceRoot string) *debugWriter {
	return &debugWriter{
		enabled:       enabled,
		workspaceRoot: workspaceRoot,
	}
}

func (d *debugWriter) write(filename string, content string) {
	if !d.enabled {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	debugFile := filepath.Join(d.workspaceRoot, "out", filename)
	if err := os.WriteFile(debugFile, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Failed to write debug file %s: %v\n", debugFile, err)
	} else {
		fmt.Fprintf(os.Stderr, "🔍 DEBUG: Saved to %s\n", debugFile)
	}
}

func (d *debugWriter) writef(format string, content string, args ...interface{}) {
	if !d.enabled {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	filename := fmt.Sprintf(format, args...)
	debugFile := filepath.Join(d.workspaceRoot, "out", filename)
	if err := os.WriteFile(debugFile, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Failed to write debug file %s: %v\n", debugFile, err)
	} else {
		fmt.Fprintf(os.Stderr, "🔍 DEBUG: Saved to %s\n", debugFile)
	}
}

func (d *debugWriter) log(format string, args ...interface{}) {
	if !d.enabled {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	fmt.Fprintf(os.Stderr, "🔍 DEBUG: "+format+"\n", args...)
}
