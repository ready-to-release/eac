package cacheclear

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/go/core/cache"
)

// clearDockerCache clears Docker caches using docker commands.
func clearDockerCache(cacheType cache.Type, dryRun, verbose bool) (int, int64, []string) {
	var cmd string
	var args []string
	var description string

	switch cacheType {
	case cache.TypeRegistry:
		cmd = "docker"
		args = []string{"image", "prune", "-f"}
		description = "Docker image cache"
	case cache.TypeLayer:
		cmd = "docker"
		args = []string{"builder", "prune", "-f"}
		description = "Docker builder cache"
	default:
		return 0, 0, nil
	}

	if dryRun {
		return 1, 0, []string{fmt.Sprintf("Would run: %s %s (%s)", cmd, strings.Join(args, " "), description)}
	}

	// Execute docker command
	execCmd := exec.Command(cmd, args...)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return 0, 0, []string{fmt.Sprintf("Failed to run %s: %v", cmd, err)}
	}

	// Try to parse reclaimed space from output
	// Docker outputs: "Total reclaimed space: 1.234GB"
	var reclaimedBytes int64
	outputStr := string(output)
	if strings.Contains(outputStr, "reclaimed space:") {
		// Parse the output (this is best-effort)
		reclaimedBytes = parseDockerReclaimedSpace(outputStr)
	}

	return 1, reclaimedBytes, []string{fmt.Sprintf("%s %s (%s)", cmd, strings.Join(args, " "), description)}
}

// parseDockerReclaimedSpace tries to parse reclaimed space from docker prune output.
// Docker outputs lines like "Total reclaimed space: 1.234GB" or "Total reclaimed space: 0B".
// The format can vary across Docker versions, so this is best-effort.
func parseDockerReclaimedSpace(output string) int64 {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "reclaimed space") {
			// Find the last colon to handle lines with multiple colons
			idx := strings.LastIndex(line, ":")
			if idx >= 0 && idx+1 < len(line) {
				sizeStr := strings.TrimSpace(line[idx+1:])
				return parseSizeString(sizeStr)
			}
		}
	}
	return 0
}

// parseSizeString parses a size string like "1.234GB", "500kB", or "0B" into bytes.
// Handles both Docker SI-style suffixes (kB, MB, GB, TB) and IEC-style (KiB, MiB, GiB, TiB).
func parseSizeString(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "0B" || s == "0" {
		return 0
	}

	// Split into numeric part and suffix by scanning from the end
	numEnd := -1
	for i := len(s) - 1; i >= 0; i-- {
		if (s[i] >= '0' && s[i] <= '9') || s[i] == '.' {
			numEnd = i
			break
		}
	}

	if numEnd < 0 {
		return 0 // no numeric content found
	}

	numStr := s[:numEnd+1]
	suffix := strings.TrimSpace(s[numEnd+1:])

	var multiplier float64 = 1
	switch strings.ToUpper(suffix) {
	case "B", "":
		multiplier = 1
	case "KB", "KIB":
		multiplier = 1024
	case "MB", "MIB":
		multiplier = 1024 * 1024
	case "GB", "GIB":
		multiplier = 1024 * 1024 * 1024
	case "TB", "TIB":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0 // unrecognized suffix
	}

	var value float64
	if _, err := fmt.Sscanf(numStr, "%f", &value); err != nil {
		return 0
	}
	return int64(value * multiplier)
}
