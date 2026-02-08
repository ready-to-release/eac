// Package core defines the interface contracts for eac-core.
//
// This is a self-contained contract package: Go port interfaces + embedded config
// (schemas, YAML defaults) in one Go module. It has zero implementation dependencies
// — only stdlib or other contracts.
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
//	import core "github.com/ready-to-release/eac/contracts/core/0.1.0"
//
//	func MyCommand(cfg core.ConfigPort, log core.LoggerPort) error {
//	    // Use ports, not concrete types
//	}
package core
