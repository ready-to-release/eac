//go:build L1 && ov
// +build L1,ov

package design

import (
	"testing"
)

// Note: detectPortConflict requires Docker and is tested via integration tests
// Unit tests focus on helper functions that don't require external dependencies

func TestParseContainerInfo(t *testing.T) {
	tests := []struct {
		name          string
		dockerPSLine  string
		wantName      string
		wantImage     string
		wantPorts     string
		wantErr       bool
	}{
		{
			name:         "valid container info",
			dockerPSLine: "structurizr-lite-src-cli\tstructurizr/lite:latest\t0.0.0.0:8080->8080/tcp",
			wantName:     "structurizr-lite-src-cli",
			wantImage:    "structurizr/lite:latest",
			wantPorts:    "0.0.0.0:8080->8080/tcp",
			wantErr:      false,
		},
		{
			name:         "container with multiple ports",
			dockerPSLine: "nginx\tnginx:latest\t0.0.0.0:8080->80/tcp, 0.0.0.0:8443->443/tcp",
			wantName:     "nginx",
			wantImage:    "nginx:latest",
			wantPorts:    "0.0.0.0:8080->80/tcp, 0.0.0.0:8443->443/tcp",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotImage, gotPorts, err := parseContainerInfo(tt.dockerPSLine)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseContainerInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotName != tt.wantName {
				t.Errorf("parseContainerInfo() gotName = %v, want %v", gotName, tt.wantName)
			}

			if gotImage != tt.wantImage {
				t.Errorf("parseContainerInfo() gotImage = %v, want %v", gotImage, tt.wantImage)
			}

			if gotPorts != tt.wantPorts {
				t.Errorf("parseContainerInfo() gotPorts = %v, want %v", gotPorts, tt.wantPorts)
			}
		})
	}
}

func TestIsPortInUse(t *testing.T) {
	tests := []struct {
		name      string
		portLine  string
		targetPort string
		want      bool
	}{
		{
			name:      "port 8080 in use - exact match",
			portLine:  "0.0.0.0:8080->8080/tcp",
			targetPort: "8080",
			want:      true,
		},
		{
			name:      "port 8080 in use - with localhost",
			portLine:  "127.0.0.1:8080->8080/tcp",
			targetPort: "8080",
			want:      true,
		},
		{
			name:      "port 8080 not in use - different port",
			portLine:  "0.0.0.0:3000->3000/tcp",
			targetPort: "8080",
			want:      false,
		},
		{
			name:      "port 8080 in use - multiple ports",
			portLine:  "0.0.0.0:3000->3000/tcp, 0.0.0.0:8080->80/tcp",
			targetPort: "8080",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPortInUse(tt.portLine, tt.targetPort); got != tt.want {
				t.Errorf("isPortInUse() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Note: stopConflictingContainer requires Docker and is tested via integration tests
