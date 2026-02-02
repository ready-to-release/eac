package selector

import (
	"github.com/ready-to-release/eac/go/adapters/tui"
)

func init() {
	tui.RegisterSelector(Factory())
}
