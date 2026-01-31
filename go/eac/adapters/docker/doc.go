// Package docker provides Docker container management capabilities.
//
// This adapter module wraps the Docker SDK to provide:
//   - Container lifecycle management (create, start, stop, remove)
//   - Image building and pulling
//   - Port allocation and reservation
//   - Browser launch integration
//   - Docker-in-Docker (DinD) support for path translation
//
// Usage:
//
//	import "github.com/ready-to-release/eac/go/eac/adapters/docker"
//
//	client, err := docker.NewClient()
//	if err != nil {
//	    return err
//	}
//	defer client.Close()
//
//	result, err := docker.StartServe(ctx, &docker.ServeConfig{
//	    Name:          "my-container",
//	    Image:         "nginx:latest",
//	    ContentPath:   "/path/to/content",
//	    ContainerPath: "/usr/share/nginx/html",
//	    ContainerPort: 80,
//	})
package docker
