package output

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuffer_StartAndStop(t *testing.T) {
	buf := NewBuffer()

	// Initially not active
	assert.False(t, buf.IsActive())

	// Start capture
	err := buf.Start()
	require.NoError(t, err)
	assert.True(t, buf.IsActive())

	// Write to stdout
	fmt.Fprint(os.Stdout, "test output")

	// Give pump goroutine time to read
	time.Sleep(10 * time.Millisecond)

	// Check buffered content
	stdout, stderr := buf.GetBuffered()
	assert.Equal(t, "test output", string(stdout))
	assert.Empty(t, stderr)

	// Stop should flush and restore
	buf.Stop()
	assert.False(t, buf.IsActive())
}

func TestBuffer_StderrCapture(t *testing.T) {
	buf := NewBuffer()

	err := buf.Start()
	require.NoError(t, err)

	// Write to stderr
	fmt.Fprint(os.Stderr, "error output")

	// Give pump goroutine time to read
	time.Sleep(10 * time.Millisecond)

	// Check buffered content
	stdout, stderr := buf.GetBuffered()
	assert.Empty(t, stdout)
	assert.Equal(t, "error output", string(stderr))

	buf.Stop()
}

func TestBuffer_BothStreams(t *testing.T) {
	buf := NewBuffer()

	err := buf.Start()
	require.NoError(t, err)

	// Write to both
	fmt.Fprint(os.Stdout, "stdout msg")
	fmt.Fprint(os.Stderr, "stderr msg")

	// Give pump goroutines time to read
	time.Sleep(10 * time.Millisecond)

	// Check both buffers
	stdout, stderr := buf.GetBuffered()
	assert.Equal(t, "stdout msg", string(stdout))
	assert.Equal(t, "stderr msg", string(stderr))

	buf.Stop()
}

func TestBuffer_Flush(t *testing.T) {
	buf := NewBuffer()

	err := buf.Start()
	require.NoError(t, err)

	// Write some content
	fmt.Fprint(os.Stdout, "before flush")

	// Give pump time to read
	time.Sleep(10 * time.Millisecond)

	// Flush clears buffer
	buf.Flush()

	// Buffer should be empty after flush
	stdout, _ := buf.GetBuffered()
	assert.Empty(t, stdout)

	// Still active after flush
	assert.True(t, buf.IsActive())

	buf.Stop()
}

func TestBuffer_DoubleStart(t *testing.T) {
	buf := NewBuffer()

	err := buf.Start()
	require.NoError(t, err)

	// Second start should be no-op
	err = buf.Start()
	require.NoError(t, err)
	assert.True(t, buf.IsActive())

	buf.Stop()
}

func TestBuffer_DoubleStop(t *testing.T) {
	buf := NewBuffer()

	err := buf.Start()
	require.NoError(t, err)

	// First stop
	buf.Stop()
	assert.False(t, buf.IsActive())

	// Second stop should be safe no-op
	buf.Stop()
	assert.False(t, buf.IsActive())
}

func TestBuffer_ConcurrentWrites(t *testing.T) {
	buf := NewBuffer()

	err := buf.Start()
	require.NoError(t, err)

	// Concurrent writes from multiple goroutines
	var wg sync.WaitGroup
	numWriters := 10
	writesPerWriter := 100

	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < writesPerWriter; j++ {
				fmt.Fprintf(os.Stdout, "w%d-%d ", id, j)
			}
		}(i)
	}

	wg.Wait()

	// Give pump time to read all data
	time.Sleep(50 * time.Millisecond)

	// Should have captured all writes (exact content order may vary)
	stdout, _ := buf.GetBuffered()
	assert.NotEmpty(t, stdout)

	buf.Stop()
}

func TestBuffer_StopWaitsForPumps(t *testing.T) {
	buf := NewBuffer()

	err := buf.Start()
	require.NoError(t, err)

	// Write a lot of data
	for i := 0; i < 1000; i++ {
		fmt.Fprintf(os.Stdout, "line %d\n", i)
	}

	// Stop should wait for pumps to finish (not return prematurely)
	startStop := time.Now()
	buf.Stop()
	stopDuration := time.Since(startStop)

	// Stop should complete (not hang) but should wait for data
	assert.Less(t, stopDuration, 5*time.Second, "Stop took too long")
	assert.False(t, buf.IsActive())
}

func TestPassthrough_NoCapture(t *testing.T) {
	buf := NewPassthrough()

	// Always inactive
	assert.False(t, buf.IsActive())

	// Start is no-op
	err := buf.Start()
	require.NoError(t, err)
	assert.False(t, buf.IsActive())

	// GetBuffered returns nil
	stdout, stderr := buf.GetBuffered()
	assert.Nil(t, stdout)
	assert.Nil(t, stderr)

	// Stop and Flush are safe no-ops
	buf.Flush()
	buf.Stop()
}

func TestBuffer_GetBufferedReturnsClone(t *testing.T) {
	buf := NewBuffer()

	err := buf.Start()
	require.NoError(t, err)

	fmt.Fprint(os.Stdout, "original")
	time.Sleep(10 * time.Millisecond)

	// Get buffered content
	stdout1, _ := buf.GetBuffered()

	// Write more
	fmt.Fprint(os.Stdout, " more")
	time.Sleep(10 * time.Millisecond)

	// Get again - should have both
	stdout2, _ := buf.GetBuffered()

	// First slice should be unchanged (it's a copy)
	assert.Equal(t, "original", string(stdout1))
	assert.Equal(t, "original more", string(stdout2))

	buf.Stop()
}
