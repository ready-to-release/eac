// Memory detection utilities for dynamic resource allocation.
package environments

// GetSystemMemoryBytes returns the total system memory in bytes.
// Returns 0 if memory detection fails.
func GetSystemMemoryBytes() uint64 {
	return getSystemMemoryBytes()
}

// GetPDFExportConcurrency returns the recommended number of concurrent PDF exports
// based on available system memory.
// - < 8GB: 1 concurrent export (memory constrained)
// - 8-16GB: 2 concurrent exports
// - >= 16GB: 3 concurrent exports (max to prevent resource exhaustion).
func GetPDFExportConcurrency() int {
	return CalculatePDFConcurrency(GetSystemMemoryBytes())
}

// CalculatePDFConcurrency calculates the PDF export concurrency for a given memory size.
// This is exported for testing with specific memory values.
func CalculatePDFConcurrency(memoryBytes uint64) int {
	const (
		gb8  = 8 * 1024 * 1024 * 1024
		gb16 = 16 * 1024 * 1024 * 1024
	)

	switch {
	case memoryBytes >= gb16:
		return 3
	case memoryBytes >= gb8:
		return 2
	default:
		return 1
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
