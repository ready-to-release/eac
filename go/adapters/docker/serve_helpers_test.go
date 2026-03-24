package docker

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
)

func TestMatchesContainerName(t *testing.T) {
	tests := []struct {
		name          string
		containerName string
		pattern       string
		want          bool
	}{
		{
			name:          "exact match without slash prefix",
			containerName: "myapp",
			pattern:       "myapp",
			want:          true,
		},
		{
			name:          "exact match with slash prefix",
			containerName: "/myapp",
			pattern:       "myapp",
			want:          true,
		},
		{
			name:          "port suffix match with slash prefix",
			containerName: "/myapp-9801",
			pattern:       "myapp",
			want:          true,
		},
		{
			name:          "port suffix match without slash prefix",
			containerName: "myapp-9801",
			pattern:       "myapp",
			want:          true,
		},
		{
			name:          "no match different name",
			containerName: "/other-app",
			pattern:       "myapp",
			want:          false,
		},
		{
			name:          "partial name overlap not at dash boundary",
			containerName: "/myapp2",
			pattern:       "myapp",
			want:          false,
		},
		{
			// Empty pattern: cleanName == "" is false, and HasPrefix(cleanName, "-") is
			// false because "myapp" does not start with "-". The function returns false.
			name:          "empty pattern does not match non-empty name",
			containerName: "/myapp",
			pattern:       "",
			want:          false,
		},
		{
			name:          "pattern is prefix but separated by additional dash segment",
			containerName: "/myapp-extra-9801",
			pattern:       "myapp",
			want:          true,
		},
		{
			name:          "multi-word name exact match",
			containerName: "/cli-serve-docs",
			pattern:       "cli-serve-docs",
			want:          true,
		},
		{
			name:          "multi-word name with port suffix",
			containerName: "/cli-serve-docs-9801",
			pattern:       "cli-serve-docs",
			want:          true,
		},
		{
			name:          "multi-word name no match",
			containerName: "/cli-serve-design-9801",
			pattern:       "cli-serve-docs",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesContainerName(tt.containerName, tt.pattern)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerToServeResult(t *testing.T) {
	tests := []struct {
		name      string
		container types.Container
		wantID    string
		wantName  string
		wantPort  int
		wantURL   string
	}{
		{
			name: "single port and single name",
			container: types.Container{
				ID:    "abc123",
				Names: []string{"/myapp-9801"},
				Ports: []types.Port{
					{PublicPort: 9801, Type: "tcp"},
				},
			},
			wantID:   "abc123",
			wantName: "myapp-9801",
			wantPort: 9801,
			wantURL:  "http://localhost:9801",
		},
		{
			name: "multiple ports uses first non-zero public port",
			container: types.Container{
				ID:    "def456",
				Names: []string{"/myapp-9802"},
				Ports: []types.Port{
					{PublicPort: 0, Type: "tcp"},
					{PublicPort: 9802, Type: "tcp"},
					{PublicPort: 9803, Type: "tcp"},
				},
			},
			wantID:   "def456",
			wantName: "myapp-9802",
			wantPort: 9802,
			wantURL:  "http://localhost:9802",
		},
		{
			name: "no public ports yields zero host port",
			container: types.Container{
				ID:    "ghi789",
				Names: []string{"/myapp-no-port"},
				Ports: []types.Port{
					{PublicPort: 0, Type: "tcp"},
				},
			},
			wantID:   "ghi789",
			wantName: "myapp-no-port",
			wantPort: 0,
			wantURL:  "http://localhost:0",
		},
		{
			name: "empty ports slice yields zero host port",
			container: types.Container{
				ID:    "jkl000",
				Names: []string{"/myapp-empty-ports"},
				Ports: []types.Port{},
			},
			wantID:   "jkl000",
			wantName: "myapp-empty-ports",
			wantPort: 0,
			wantURL:  "http://localhost:0",
		},
		{
			name: "no names yields empty container name",
			container: types.Container{
				ID:    "mno111",
				Names: []string{},
				Ports: []types.Port{
					{PublicPort: 9900, Type: "tcp"},
				},
			},
			wantID:   "mno111",
			wantName: "",
			wantPort: 9900,
			wantURL:  "http://localhost:9900",
		},
		{
			name: "slash-prefixed name is stripped",
			container: types.Container{
				ID:    "pqr222",
				Names: []string{"/prefixed-name-9801"},
				Ports: []types.Port{
					{PublicPort: 9801, Type: "tcp"},
				},
			},
			wantID:   "pqr222",
			wantName: "prefixed-name-9801",
			wantPort: 9801,
			wantURL:  "http://localhost:9801",
		},
		{
			name: "multiple names uses first name",
			container: types.Container{
				ID:    "stu333",
				Names: []string{"/primary-name-9801", "/alias-name"},
				Ports: []types.Port{
					{PublicPort: 9801, Type: "tcp"},
				},
			},
			wantID:   "stu333",
			wantName: "primary-name-9801",
			wantPort: 9801,
			wantURL:  "http://localhost:9801",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containerToServeResult(tt.container)

			assert.NotNil(t, got)
			assert.Equal(t, tt.wantID, got.ContainerID)
			assert.Equal(t, tt.wantName, got.ContainerName)
			assert.Equal(t, tt.wantPort, got.HostPort)
			assert.Equal(t, tt.wantURL, got.URL)
		})
	}
}
