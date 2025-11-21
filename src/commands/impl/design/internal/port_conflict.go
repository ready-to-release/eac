package design

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

// ContainerInfo holds information about a Docker container
type ContainerInfo struct {
	Name  string
	Image string
	Ports string
}

// detectPortConflict checks if the specified port is in use by a Docker container
// Returns: (hasConflict bool, containerName string, error)
func detectPortConflict(port string) (bool, string, error) {
	// Run docker ps to get all running containers with their ports
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}\t{{.Image}}\t{{.Ports}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If docker is not available or fails, we can't detect conflicts
		return false, "", fmt.Errorf("failed to query Docker containers: %w", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		// No containers running
		return false, "", nil
	}

	// Parse each line to find containers using the target port
	scanner := bufio.NewScanner(strings.NewReader(outputStr))
	for scanner.Scan() {
		line := scanner.Text()
		name, _, ports, err := parseContainerInfo(line)
		if err != nil {
			continue // Skip malformed lines
		}

		if isPortInUse(ports, port) {
			return true, name, nil
		}
	}

	return false, "", nil
}

// parseContainerInfo parses a line from docker ps output
// Format: NAME\tIMAGE\tPORTS
func parseContainerInfo(line string) (name, image, ports string, err error) {
	parts := strings.Split(line, "\t")
	if len(parts) < 3 {
		return "", "", "", fmt.Errorf("invalid docker ps output format")
	}

	return parts[0], parts[1], parts[2], nil
}

// isPortInUse checks if a specific port is mentioned in the ports string
// Port format examples: "0.0.0.0:8080->8080/tcp", "127.0.0.1:8080->8080/tcp"
func isPortInUse(portLine, targetPort string) bool {
	// Check for patterns like "0.0.0.0:8080->" or "127.0.0.1:8080->"
	patterns := []string{
		"0.0.0.0:" + targetPort + "->",
		"127.0.0.1:" + targetPort + "->",
		":::" + targetPort + "->", // IPv6
	}

	for _, pattern := range patterns {
		if strings.Contains(portLine, pattern) {
			return true
		}
	}

	return false
}

// stopConflictingContainer stops and removes the specified container
func stopConflictingContainer(containerName string) error {
	// Stop the container
	stopCmd := exec.Command("docker", "stop", containerName)
	if err := stopCmd.Run(); err != nil {
		return fmt.Errorf("failed to stop container %s: %w", containerName, err)
	}

	// Remove the container
	rmCmd := exec.Command("docker", "rm", containerName)
	if err := rmCmd.Run(); err != nil {
		return fmt.Errorf("failed to remove container %s: %w", containerName, err)
	}

	return nil
}

// handlePortConflict manages port conflict resolution
// Returns true if we should continue starting the new container, false to abort
func handlePortConflict(containerName, newModuleName string, autoStop bool) (bool, error) {
	fmt.Printf("\n⚠️  Port Conflict Detected\n")
	fmt.Printf("   Port 8080 is already in use by container: %s\n\n", containerName)

	// If auto-stop flag is set, stop the container automatically
	if autoStop {
		fmt.Printf("🔄 Automatically stopping conflicting container...\n")
		if err := stopConflictingContainer(containerName); err != nil {
			return false, fmt.Errorf("failed to stop conflicting container: %w", err)
		}
		fmt.Printf("✅ Conflicting container stopped\n\n")
		return true, nil
	}

	// Interactive mode - prompt user
	fmt.Println("What would you like to do?")
	fmt.Println("  1) Stop the existing container and start the new one")
	fmt.Println("  2) Keep the existing container running (abort)")
	fmt.Println("  3) Abort operation")
	fmt.Print("\nEnter your choice (1-3): ")

	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		fmt.Printf("\n🔄 Stopping container: %s\n", containerName)
		if err := stopConflictingContainer(containerName); err != nil {
			return false, fmt.Errorf("failed to stop conflicting container: %w", err)
		}
		fmt.Printf("✅ Container stopped successfully\n\n")
		return true, nil

	case "2":
		fmt.Printf("\n✅ Keeping existing container running\n")
		fmt.Printf("💡 You can view it at: http://localhost:8080\n\n")
		return false, nil

	case "3":
		fmt.Printf("\n✅ Operation aborted\n\n")
		return false, nil

	default:
		fmt.Printf("\n❌ Invalid choice: %s\n", choice)
		fmt.Printf("✅ Operation aborted\n\n")
		return false, nil
	}
}
