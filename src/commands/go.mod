module github.com/ready-to-release/eac/src/commands

go 1.24.4

require (
	github.com/jedib0t/go-pretty/v6 v6.6.9
	github.com/ready-to-release/eac/src/contracts/reports v0.0.0-20251103160328-9372bc55b01c
	github.com/ready-to-release/eac/src/repository/reports v0.0.0-20251103160328-9372bc55b01c
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/ready-to-release/eac/src/contracts v0.0.0 // indirect
	github.com/ready-to-release/eac/src/repository v0.0.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)

replace github.com/ready-to-release/eac/src/contracts => ../contracts

replace github.com/ready-to-release/eac/src/repository => ../repository

replace github.com/ready-to-release/eac/src/contracts/reports => ../contracts/reports

replace github.com/ready-to-release/eac/src/repository/reports => ../repository/reports
