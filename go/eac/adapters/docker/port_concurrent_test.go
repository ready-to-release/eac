//go:build L1

package docker

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestConcurrentPortReservation tests that multiple goroutines can reserve ports without collisions
func TestConcurrentPortReservation(t *testing.T) {
	const numGoroutines = 20
	const portsPerGoroutine = 5

	// Channel to collect reserved ports
	portsChan := make(chan int, numGoroutines*portsPerGoroutine)
	var wg sync.WaitGroup

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < portsPerGoroutine; j++ {
				port, err := FindAndReservePort()
				if err != nil {
					t.Errorf("Failed to reserve port: %v", err)
					return
				}
				portsChan <- port
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(portsChan)

	// Collect all reserved ports
	localReservedPorts := make(map[int]int) // port -> count
	for port := range portsChan {
		localReservedPorts[port]++
	}

	// Assert no port was reserved more than once
	for port, count := range localReservedPorts {
		assert.Equal(t, 1, count, "Port %d was reserved %d times (should be 1)", port, count)
	}

	// Assert we got the expected number of unique ports
	assert.Equal(t, numGoroutines*portsPerGoroutine, len(localReservedPorts),
		"Expected %d unique ports, got %d", numGoroutines*portsPerGoroutine, len(localReservedPorts))

	// Clean up
	for port := range localReservedPorts {
		ReleasePort(port)
	}
}

// TestConcurrentReserveAndRelease tests concurrent reserve/release cycles
func TestConcurrentReserveAndRelease(t *testing.T) {
	const numGoroutines = 10
	const cyclesPerGoroutine = 10

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < cyclesPerGoroutine; j++ {
				// Reserve port
				port, err := FindAndReservePort()
				if err != nil {
					t.Errorf("Failed to reserve port: %v", err)
					return
				}

				// Verify port is reserved
				if ReservePort(port) {
					t.Errorf("Port %d should already be reserved", port)
				}

				// Release port
				ReleasePort(port)

				// Verify port can be reserved again
				if !ReservePort(port) {
					t.Errorf("Port %d should be available after release", port)
				}
				ReleasePort(port)
			}
		}()
	}

	wg.Wait()
}

// TestPortReservationWithPreferred tests concurrent use of preferred ports
func TestPortReservationWithPreferred(t *testing.T) {
	const numGoroutines = 10

	var wg sync.WaitGroup
	ports := make(chan int, numGoroutines)

	// All goroutines try to use the same preferred port
	preferredPort := PortRangeStart + 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			port, err := FindAndReservePortOrUse(preferredPort)
			if err != nil {
				t.Errorf("Failed to allocate port: %v", err)
				return
			}
			ports <- port
		}()
	}

	wg.Wait()
	close(ports)

	// Collect all allocated ports
	allocatedPorts := make(map[int]int)
	preferredCount := 0
	for port := range ports {
		allocatedPorts[port]++
		if port == preferredPort {
			preferredCount++
		}
	}

	// Assert exactly one goroutine got the preferred port
	assert.Equal(t, 1, preferredCount, "Expected exactly 1 goroutine to get preferred port")

	// Assert all ports are unique
	for port, count := range allocatedPorts {
		assert.Equal(t, 1, count, "Port %d was allocated %d times", port, count)
	}

	// Assert we got the expected number of ports
	assert.Equal(t, numGoroutines, len(allocatedPorts))

	// Clean up
	for port := range allocatedPorts {
		ReleasePort(port)
	}
}

// TestPortReservationStress tests port reservation under high stress
func TestPortReservationStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	const numGoroutines = 50
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*operationsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Reserve
				port, err := FindAndReservePort()
				if err != nil {
					errors <- err
					continue
				}

				// Immediately release
				ReleasePort(port)
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Logf("Error during stress test: %v", err)
		errorCount++
	}

	// Allow some errors in stress test, but not too many
	maxAllowedErrors := (numGoroutines * operationsPerGoroutine) / 100 // 1%
	assert.LessOrEqual(t, errorCount, maxAllowedErrors,
		"Too many errors (%d) during stress test (max allowed: %d)", errorCount, maxAllowedErrors)
}

// TestConcurrentPortAvailabilityCheck tests concurrent port availability checks
func TestConcurrentPortAvailabilityCheck(t *testing.T) {
	const numGoroutines = 20

	// Reserve a port first
	port, err := FindAndReservePort()
	assert.NoError(t, err)
	defer ReleasePort(port)

	var wg sync.WaitGroup
	results := make(chan bool, numGoroutines)

	// Multiple goroutines check if the port is available
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Try to reserve the already-reserved port
			reserved := ReservePort(port)
			results <- reserved
		}()
	}

	wg.Wait()
	close(results)

	// All attempts should fail (return false) since port is already reserved
	successCount := 0
	for success := range results {
		if success {
			successCount++
		}
	}

	assert.Equal(t, 0, successCount, "No goroutine should succeed in reserving an already-reserved port")
}

// TestGetReservedPorts tests the GetReservedPorts function under concurrent access
func TestGetReservedPortsConcurrent(t *testing.T) {
	// Clean state
	portMutex.Lock()
	reservedPorts = make(map[int]time.Time)
	portMutex.Unlock()

	const numPorts = 10

	// Reserve some ports
	ports := make([]int, numPorts)
	for i := 0; i < numPorts; i++ {
		port, err := FindAndReservePort()
		assert.NoError(t, err)
		ports[i] = port
	}

	// Get reserved ports
	reserved := GetReservedPorts()

	// Verify count
	assert.Equal(t, numPorts, len(reserved), "Expected %d reserved ports", numPorts)

	// Verify all our ports are in the reserved list
	for _, port := range ports {
		_, exists := reserved[port]
		assert.True(t, exists, "Port %d should be in reserved list", port)
	}

	// Clean up
	for _, port := range ports {
		ReleasePort(port)
	}

	// Verify cleanup
	reserved = GetReservedPorts()
	assert.Equal(t, 0, len(reserved), "All ports should be released")
}

// TestPortReservationRaceDetector tests for race conditions (run with -race flag)
func TestPortReservationRaceDetector(t *testing.T) {
	const numGoroutines = 20

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Reserve
			port, err := FindAndReservePort()
			if err != nil {
				return
			}

			// Check reserved ports
			_ = GetReservedPorts()

			// Release
			ReleasePort(port)
		}()
	}

	wg.Wait()
}
