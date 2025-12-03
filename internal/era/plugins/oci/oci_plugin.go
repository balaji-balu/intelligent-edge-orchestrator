//go:build oci

package plugins

import "github.com/balaji-balu/margo-hello-world/pkg/model"

type Plugin struct {}

func NewPlugin() *Plugin {
    return &Plugin{}
}

// containerd / podman logic here...
