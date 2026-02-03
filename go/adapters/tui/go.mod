module github.com/ready-to-release/eac/go/adapters/tui

go 1.24.4

require (
	github.com/atotto/clipboard v0.1.4
	github.com/charmbracelet/bubbles v0.21.0
	github.com/charmbracelet/bubbletea v1.3.10
	github.com/charmbracelet/lipgloss v1.1.0
	github.com/lrstanley/bubblezone v1.0.0
	github.com/mattn/go-runewidth v0.0.19
	github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces v0.0.0
	github.com/ready-to-release/eac/contracts/tui-adapter/0.1.0/interfaces v0.0.0
	github.com/ready-to-release/eac/go/core v0.0.0-00010101000000-000000000000
	github.com/shahar3/bubble-grid v1.0.0
	github.com/shirou/gopsutil/v3 v3.24.5
	golang.org/x/term v0.39.0
)

require (
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/charmbracelet/colorprofile v0.3.1 // indirect
	github.com/charmbracelet/x/ansi v0.10.1 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.13 // indirect
	github.com/charmbracelet/x/term v0.2.1 // indirect
	github.com/clipperhouse/uax29/v2 v2.2.0 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces => ../../../contracts/core/0.1.0/interfaces
	github.com/ready-to-release/eac/contracts/docker-adapter/0.1.0/interfaces => ../../../contracts/docker-adapter/0.1.0/interfaces
	github.com/ready-to-release/eac/contracts/security/0.1.0/interfaces => ../../../contracts/security/0.1.0/interfaces
	github.com/ready-to-release/eac/contracts/testing/0.1.0/interfaces => ../../../contracts/testing/0.1.0/interfaces
	github.com/ready-to-release/eac/contracts/tools/0.1.0/interfaces => ../../../contracts/tools/0.1.0/interfaces
	github.com/ready-to-release/eac/contracts/tui-adapter/0.1.0/interfaces => ../../../contracts/tui-adapter/0.1.0/interfaces
	github.com/ready-to-release/eac/go/core => ../../core
	github.com/ready-to-release/eac/go/godog => ../../godog
)
