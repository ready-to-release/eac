module github.com/ready-to-release/eac/go/specs/repository

go 1.24.4

require (
	github.com/bmatcuk/doublestar/v4 v4.10.0
	github.com/cucumber/godog v0.15.1
	github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces v0.0.0-00010101000000-000000000000
	github.com/ready-to-release/eac/go/core v0.0.0
	github.com/ready-to-release/eac/go/godog v0.0.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.1.6 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/cucumber/gherkin/go/v26 v26.2.0 // indirect
	github.com/cucumber/messages/go/v21 v21.0.1 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/defenseunicorns/go-oscal v0.7.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.2 // indirect
	github.com/go-git/go-git/v5 v5.16.4 // indirect
	github.com/gofrs/uuid v4.3.1+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-memdb v1.3.4 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/pjbgf/sha1cd v0.3.2 // indirect
	github.com/ready-to-release/eac/contracts v0.0.0-20260129143239-c3ff637b8ca9 // indirect
	github.com/ready-to-release/eac/contracts/security/0.1.0/interfaces v0.0.0-00010101000000-000000000000 // indirect
	github.com/ready-to-release/eac/contracts/testing/0.1.0/interfaces v0.0.0-00010101000000-000000000000 // indirect
	github.com/ready-to-release/eac/contracts/tui-adapter/0.1.0/interfaces v0.0.0-20260202144134-d0b6883b9564 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/skeema/knownhosts v1.3.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)

replace (
	github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces => ../../../contracts/core/0.1.0/interfaces
	github.com/ready-to-release/eac/contracts/docker-adapter/0.1.0/interfaces => ../../../contracts/docker-adapter/0.1.0/interfaces
	github.com/ready-to-release/eac/contracts/security/0.1.0/interfaces => ../../../contracts/security/0.1.0/interfaces
	github.com/ready-to-release/eac/contracts/testing/0.1.0/interfaces => ../../../contracts/testing/0.1.0/interfaces
	github.com/ready-to-release/eac/contracts/tools/0.1.0/interfaces => ../../../contracts/tools/0.1.0/interfaces
	github.com/ready-to-release/eac/go/cli/eac => ../../cli/eac
	github.com/ready-to-release/eac/go/core => ../../core
	github.com/ready-to-release/eac/go/godog => ../../godog
)
