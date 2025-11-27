// Package docs provides Docker client implementation for docs command.
package docs

// ContainerInfo contains information about a running Docker container
type ContainerInfo struct {
	Name string
	URL  string
	Port int
}
