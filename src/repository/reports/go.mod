module github.com/ready-to-release/eac/src/repository/reports

go 1.24.4

require github.com/ready-to-release/eac/src/repository v0.0.0

require (
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/ready-to-release/eac/src/contracts v0.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/ready-to-release/eac/src/repository => ../

replace github.com/ready-to-release/eac/src/contracts => ../../contracts
