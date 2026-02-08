package drawio

import "github.com/ready-to-release/eac/go/clibase/registry"

func init() {
	registry.RegisterAll(
		DrawioCreate,
		DrawioDecode,
		DrawioEmbed,
		DrawioEncode,
		DrawioInfo,
		DrawioRender,
	)
}
