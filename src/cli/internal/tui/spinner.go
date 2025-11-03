package tui

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	img "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/ready-to-release/eac/src/cli/internal/conf"
	"github.com/ready-to-release/eac/src/cli/internal/docker"
)

type errMsg error

type dockerPullMsg struct {
	err error
}

type model struct {
	spinner  spinner.Model
	quitting bool
	err      error
	image    string
	name     string
	done     bool
}

func Auth() string {
	// Use centralized authentication function from docker package
	_, authStr, err := docker.CreateGitHubAuthConfig()
	if err != nil {
		log.Fatal(err)
	}
	return authStr
}

func NewPullModel(ext conf.Extension) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{spinner: s, image: ext.Image, name: ext.Name}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, pullDockerImageCmd(m.image))
}

func pullDockerImageCmd(image string) tea.Cmd {
	return func() tea.Msg {
		// Perform the docker pull here
		return dockerPullMsg{err: pullImage(image)}
	}
}

func pullImage(image string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()
	reader, err := cli.ImagePull(ctx, image, img.PullOptions{RegistryAuth: Auth()})
	if err != nil {
		return err
	}
	defer reader.Close()
	_, err = io.Copy(io.Discard, reader)
	return err
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}

	case dockerPullMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	if m.done {
		if m.err != nil {
			return fmt.Sprintf("\n\n   Error: %v\n\n", m.err)
		}
		return fmt.Sprintf("Installed '%s' successfully!\n", m.name)
	}
	str := fmt.Sprintf("\n\n   %s Installing %s... (press q to quit)\n", m.spinner.View(), m.name)
	if m.quitting {
		return str + "\n"
	}
	return str
}
