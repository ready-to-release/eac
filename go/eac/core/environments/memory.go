// Memory detection utilities for dynamic resource allocation.
package environments

// GetSystemMemoryBytes returns the total system memory in bytes.
// Returns 0 if memory detection fails.
func GetSystemMemoryBytes() uint64 {
	return getSystemMemoryBytes()
}

// GetPDFExportConcurrency returns the recommended number of concurrent PDF exports.
// CI always returns 2. Devbox uses memory-based calculation.
func GetPDFExportConcurrency() int {
	return GetPDFExportConcurrencyWithTurbo(false)
}

// GetPDFExportConcurrencyWithTurbo returns PDF concurrency with optional turbo boost.
// CI always returns 2 (turbo ignored). Devbox uses memory-based calculation.
func GetPDFExportConcurrencyWithTurbo(turbo bool) int {
	// CI: fixed at 2 concurrent PDF exports
	if IsCI() {
		return 2
	}
	// Devbox: memory-based with turbo support
	return CalculatePDFConcurrencyWithTurbo(GetSystemMemoryBytes(), turbo)
}

// CalculatePDFConcurrency calculates the PDF export concurrency for a given memory size.
// This is exported for testing with specific memory values.
// Note: This does NOT check for CI - use GetPDFExportConcurrency for runtime detection.
func CalculatePDFConcurrency(memoryBytes uint64) int {
	return CalculatePDFConcurrencyWithTurbo(memoryBytes, false)
}

// CalculatePDFConcurrencyWithTurbo calculates PDF concurrency for devbox environments.
// For low RAM systems (≤16GB), concurrency is reduced by 1 and turbo is ignored.
// For high RAM systems (>16GB), turbo can increase concurrency to max 4.
//
// PDF concurrency is special: only ever 1, 2, 3, or 4 concurrent exports.
//
// Devbox memory tiers (low RAM penalty for ≤16GB):
//   - >16GB: 3 base, +1 with turbo = 4 (max)
//   - 16GB: 2 (base 3, -1 low RAM penalty, turbo ignored)
//   - 8-16GB: 1 (base 2, -1 low RAM penalty, turbo ignored)
//   - <8GB: 1 (minimum)
func CalculatePDFConcurrencyWithTurbo(memoryBytes uint64, turbo bool) int {
	const (
		gb8  = 8 * 1024 * 1024 * 1024
		gb16 = 16 * 1024 * 1024 * 1024
	)

	// High RAM (>16GB): full concurrency, turbo supported (max 4)
	if memoryBytes > gb16 {
		if turbo {
			return 4
		}
		return 3
	}

	// Low RAM (≤16GB): reduced concurrency (-1 penalty), turbo ignored
	switch {
	case memoryBytes >= gb16:
		return 2 // 16GB: base 3, -1 penalty
	case memoryBytes >= gb8:
		return 1 // 8-16GB: base 2, -1 penalty
	default:
		return 1 // <8GB: minimum
	}
}

// GetContainerMemoryBytes returns recommended memory for a container (half of host).
// - 32GB host = 16GB container
// - 16GB host = 8GB container
// - Minimum 2GB.
func GetContainerMemoryBytes() int64 {
	return CalculateContainerMemory(GetSystemMemoryBytes())
}

// CalculateContainerMemory calculates container memory allocation for a given host memory.
// Returns half of host memory with a minimum of 2GB.
func CalculateContainerMemory(hostMemoryBytes uint64) int64 {
	const minMemory = 2 * 1024 * 1024 * 1024 // 2GB minimum

	containerMem := int64(hostMemoryBytes / 2)
	if containerMem < minMemory {
		return minMemory
	}
	return containerMem
}
