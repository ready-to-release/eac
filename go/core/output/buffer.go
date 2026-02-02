// Package output provides controlled stdout/stderr capture for TUI mode.
package output

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
)

// Ensure outputBuffer implements OutputBufferPort.
var _ interfaces.OutputBufferPort = (*outputBuffer)(nil)

// outputBuffer captures stdout/stderr at the OS level using pipes.
// When active, all writes to os.Stdout/os.Stderr are buffered until Stop().
type outputBuffer struct {
	mu     sync.Mutex
	active bool

	// Original file descriptors
	origStdout *os.File
	origStderr *os.File

	// Pipes for capture
	stdoutReader *os.File
	stdoutWriter *os.File
	stderrReader *os.File
	stderrWriter *os.File

	// Buffers
	stdoutBuf bytes.Buffer
	stderrBuf bytes.Buffer

	// Pump goroutine synchronization
	wg sync.WaitGroup
}

// NewBuffer creates a new output buffer for TUI mode.
// Call Start() to begin capturing stdout/stderr.
func NewBuffer() interfaces.OutputBufferPort {
	return &outputBuffer{}
}

// Start begins capturing stdout/stderr.
func (b *outputBuffer) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.active {
		return nil
	}

	// Save original stdout/stderr
	b.origStdout = os.Stdout
	b.origStderr = os.Stderr

	// Create pipes
	var err error
	b.stdoutReader, b.stdoutWriter, err = os.Pipe()
	if err != nil {
		return fmt.Errorf("create stdout pipe: %w", err)
	}
	b.stderrReader, b.stderrWriter, err = os.Pipe()
	if err != nil {
		b.stdoutReader.Close()
		b.stdoutWriter.Close()
		return fmt.Errorf("create stderr pipe: %w", err)
	}

	// Redirect stdout/stderr to pipes
	os.Stdout = b.stdoutWriter
	os.Stderr = b.stderrWriter

	// Start pump goroutines with WaitGroup
	b.wg.Add(2)
	go b.pump(b.stdoutReader, &b.stdoutBuf)
	go b.pump(b.stderrReader, &b.stderrBuf)

	b.active = true
	return nil
}

// pump reads from reader and writes to buffer until reader is closed.
func (b *outputBuffer) pump(reader *os.File, buf *bytes.Buffer) {
	defer b.wg.Done()
	tmp := make([]byte, 4096)
	for {
		n, err := reader.Read(tmp)
		if n > 0 {
			b.mu.Lock()
			buf.Write(tmp[:n])
			b.mu.Unlock()
		}
		if err != nil {
			return
		}
	}
}

// Stop ends capture and restores original stdout/stderr.
// Buffered content is flushed to real stdout/stderr.
func (b *outputBuffer) Stop() {
	b.mu.Lock()
	if !b.active {
		b.mu.Unlock()
		return
	}

	// Close write ends to signal pump goroutines to exit
	b.stdoutWriter.Close()
	b.stderrWriter.Close()
	b.mu.Unlock()

	// Wait for pumps to drain all data
	b.wg.Wait()

	b.mu.Lock()
	defer b.mu.Unlock()

	// Restore original stdout/stderr
	os.Stdout = b.origStdout
	os.Stderr = b.origStderr

	// Flush buffered content to real stdout/stderr
	if b.stdoutBuf.Len() > 0 {
		b.origStdout.Write(b.stdoutBuf.Bytes())
	}
	if b.stderrBuf.Len() > 0 {
		b.origStderr.Write(b.stderrBuf.Bytes())
	}

	// Clean up
	b.stdoutReader.Close()
	b.stderrReader.Close()
	b.stdoutBuf.Reset()
	b.stderrBuf.Reset()

	b.active = false
}

// Flush writes buffered content to real stdout/stderr without stopping.
func (b *outputBuffer) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.active {
		return
	}

	// Write buffered content to original file descriptors
	if b.stdoutBuf.Len() > 0 {
		b.origStdout.Write(b.stdoutBuf.Bytes())
		b.stdoutBuf.Reset()
	}
	if b.stderrBuf.Len() > 0 {
		b.origStderr.Write(b.stderrBuf.Bytes())
		b.stderrBuf.Reset()
	}
}

// IsActive returns true if capture is active.
func (b *outputBuffer) IsActive() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.active
}

// GetBuffered returns current buffered content without flushing.
func (b *outputBuffer) GetBuffered() (stdout, stderr []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Return copies to avoid data races
	if b.stdoutBuf.Len() > 0 {
		stdout = make([]byte, b.stdoutBuf.Len())
		copy(stdout, b.stdoutBuf.Bytes())
	}
	if b.stderrBuf.Len() > 0 {
		stderr = make([]byte, b.stderrBuf.Len())
		copy(stderr, b.stderrBuf.Bytes())
	}
	return stdout, stderr
}
