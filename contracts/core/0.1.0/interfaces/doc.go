// Package ports defines the interface contracts for eac-core.
//
// These interfaces define the injectable dependencies, enabling:
//   - Loose coupling between go/cli/eac and go/core
//   - Fast test execution via mock implementations
//   - Controlled cache invalidation (interface changes are intentional API changes)
//
// Core is built into commands. Adapters implement these ports and are
// injected at runtime. Mocks implement these ports for testing.
//
// Implementations:
//   - go/adapters: real implementations
//   - go/mocks: test implementations
//
// Usage:
//
//	import "github.com/ready-to-release/eac/go/core/ports"
//
//	func MyCommand(cfg ports.ConfigPort, log ports.LoggerPort) error {
//	    // Use ports, not concrete types
//	}
package interfaces
