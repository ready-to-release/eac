package docker

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/core/config"
)

var (
	// reservedPorts tracks ports currently reserved with their reservation time.
	reservedPorts = make(map[int]time.Time)
	// portMutex protects the reservedPorts map.
	portMutex sync.Mutex
)

// ReservePort attempts to reserve a specific port.
// Returns true if reservation successful, false if already reserved.
func ReservePort(port int) bool {
	portMutex.Lock()
	defer portMutex.Unlock()

	// Clean up expired reservations
	cleanupExpiredReservations()

	// Check if port is already reserved
	if _, exists := reservedPorts[port]; exists {
		return false
	}

	// Reserve the port
	reservedPorts[port] = time.Now()
	return true
}

// ReleasePort releases a previously reserved port.
func ReleasePort(port int) {
	portMutex.Lock()
	defer portMutex.Unlock()
	delete(reservedPorts, port)
}

// cleanupExpiredReservations removes reservations older than TTL.
// Must be called with portMutex held.
func cleanupExpiredReservations() {
	now := time.Now()
	for port, reservedAt := range reservedPorts {
		if now.Sub(reservedAt) > config.PortReservationTTL() {
			delete(reservedPorts, port)
		}
	}
}

// FindAndReservePort finds an available port and reserves it atomically.
// This prevents race conditions when multiple containers start simultaneously.
func FindAndReservePort() (int, error) {
	// Calculate port range size
	portRange := PortRangeEnd - PortRangeStart + 1

	// Start at a random position (Go 1.22+ auto-seeds)
	startOffset := rand.IntN(portRange)

	// Scan circularly from random start
	for offset := 0; offset < portRange; offset++ {
		// Calculate port with circular wrap-around
		port := PortRangeStart + ((startOffset + offset) % portRange)

		// Check if port is available and reserve it atomically
		if IsPortAvailable(port) && ReservePort(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available port in range %d-%d", PortRangeStart, PortRangeEnd)
}

// FindAndReservePortOrUse returns the preferred port if available and reserves it,
// otherwise finds and reserves a random available port.
// If preferredPort is 0, it always auto-allocates.
func FindAndReservePortOrUse(preferredPort int) (int, error) {
	if preferredPort == 0 {
		return FindAndReservePort()
	}

	// Check if preferred port is available and reserve it
	if IsPortAvailable(preferredPort) && ReservePort(preferredPort) {
		return preferredPort, nil
	}

	// Preferred port not available, fall back to auto-allocation
	return FindAndReservePort()
}

// GetReservedPorts returns a copy of currently reserved ports for debugging.
// This is useful for monitoring and troubleshooting.
func GetReservedPorts() map[int]time.Time {
	portMutex.Lock()
	defer portMutex.Unlock()

	// Clean up before returning
	cleanupExpiredReservations()

	// Return a copy to avoid external mutations
	result := make(map[int]time.Time)
	for port, reservedAt := range reservedPorts {
		result[port] = reservedAt
	}
	return result
}
