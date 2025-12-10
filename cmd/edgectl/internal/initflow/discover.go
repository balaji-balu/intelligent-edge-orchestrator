package initflow

import (
	"fmt"
	"runtime"
)

func DiscoverEnvironment(ctx *Context) error {
	fmt.Println("✔ Discovering environment")

	ctx.OS = runtime.GOOS
	ctx.Arch = runtime.GOARCH

	// systemd detection can be added later
	ctx.HasSystemd = true

	return nil
}
