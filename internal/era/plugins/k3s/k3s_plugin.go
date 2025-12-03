//go:build k3s

package plugins

import "github.com/balaji-balu/margo-hello-world/pkg/model"

type Plugin struct {}

func NewPlugin() *Plugin {
    return &Plugin{}
}

// k3s / k8s logic here...
