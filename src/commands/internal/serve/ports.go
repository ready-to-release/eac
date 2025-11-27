// Package serve provides shared utilities for serving web content in Docker containers.
// It supports both direct host execution and Docker-in-Docker (DinD) environments.
package serve

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	// DefaultPortRangeStart is the default first port in the allocation range
	DefaultPortRangeStart = 9000
	// DefaultPortRangeEnd is the default last port in the allocation range (inclusive)
	DefaultPortRangeEnd = 9999
)

var (
	// PortRangeStart is the first port in the allocation range (configurable via EAC_PORT_RANGE_START)
	PortRangeStart = getPortRangeStart()
	// PortRangeEnd is the last port in the allocation range (configurable via EAC_PORT_RANGE_END)
	PortRangeEnd = getPortRangeEnd()
)

// getPortRangeStart returns the configured port range start or the default
func getPortRangeStart() int {
	if val := os.Getenv("EAC_PORT_RANGE_START"); val != "" {
		if port, err := strconv.Atoi(val); err == nil && port > 0 && port < 65536 {
			return port
		}
	}
	return DefaultPortRangeStart
}

// getPortRangeEnd returns the configured port range end or the default
func getPortRangeEnd() int {
	if val := os.Getenv("EAC_PORT_RANGE_END"); val != "" {
		if port, err := strconv.Atoi(val); err == nil && port > 0 && port < 65536 {
			return port
		}
	}
	return DefaultPortRangeEnd
}

// SetPortRange allows programmatic configuration of the port range (useful for testing)
func SetPortRange(start, end int) {
	if start > 0 && end >= start && end < 65536 {
		PortRangeStart = start
		PortRangeEnd = end
	}
}

// ResetPortRange resets the port range to defaults
func ResetPortRange() {
	PortRangeStart = getPortRangeStart()
	PortRangeEnd = getPortRangeEnd()
}

// FindAvailablePort finds an unused port in the configured port range (default: 9000-9999).
// The range can be customized via EAC_PORT_RANGE_START and EAC_PORT_RANGE_END environment variables.
// It uses a random starting point and scans circularly for better distribution
// and reduced collision probability when multiple containers start simultaneously.
// Returns the first available port or an error if no ports are available.
func FindAvailablePort() (int, error) {
	// Initialize random seed (safe to call multiple times)
	rand.Seed(time.Now().UnixNano())

	// Calculate port range size
	portRange := PortRangeEnd - PortRangeStart + 1

	// Start at a random position
	startOffset := rand.Intn(portRange)

	// Scan circularly from random start
	for offset := 0; offset < portRange; offset++ {
		// Calculate port with circular wrap-around
		port := PortRangeStart + ((startOffset + offset) % portRange)

		if IsPortAvailable(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available port in range %d-%d", PortRangeStart, PortRangeEnd)
}

// IsPortAvailable checks if a specific port is available for binding.
// Returns true if the port is free, false if it's in use or cannot be bound.
func IsPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// FindAvailablePortOrUse returns the specified port if it's available,
// otherwise finds a random available port in the default range.
// If preferredPort is 0, it always auto-allocates.
func FindAvailablePortOrUse(preferredPort int) (int, error) {
	if preferredPort == 0 {
		return FindAvailablePort()
	}
	if IsPortAvailable(preferredPort) {
		return preferredPort, nil
	}
	// Preferred port is not available, fall back to auto-allocation
	return FindAvailablePort()
}
