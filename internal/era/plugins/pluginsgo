package plugins

import (
	"github.com/balaji-balu/margo-hello-world/internal/era/plugins/mock"
	"github.com/balaji-balu/margo-hello-world/internal/era/plugins/wasm"
)

// Select which NewRuntimePlugin to call based on build tag
func NewRuntimePlugin() lifecycle.RuntimePlugin {
	// This function is only used to re-export the correct plugin
	// The correct build tag file (mock or wasm) will be compiled
	return selectPlugin()
}
