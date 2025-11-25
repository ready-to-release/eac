// Package serve provides shared utilities for serving web content in Docker containers.
// It supports both direct host execution and Docker-in-Docker (DinD) environments.
package serve

import (
	"fmt"
	"net"
)

const (
	// PortRangeStart is the first port in the allocation range
	PortRangeStart = 9000
	// PortRangeEnd is the last port in the allocation range (inclusive)
	PortRangeEnd = 9999
)

// FindAvailablePort finds an unused port in the 9000-9999 range.
// It attempts to bind to each port sequentially until one is available.
// Returns the first available port or an error if no ports are available.
func FindAvailablePort() (int, error) {
	for port := PortRangeStart; port <= PortRangeEnd; port++ {
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
