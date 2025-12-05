package plugins

import (
    //"log"
    "github.com/balaji-balu/margo-hello-world/pkg/era/edgeruntime"
    //"github.com/balaji-balu/margo-hello-world/pkg/logx"
)
type MapRegistry struct{}

var plugins = map[string]edgeruntime.RuntimePlugin{}
//var log = logx.New("era.registry")

func Register(p edgeruntime.RuntimePlugin) {
    //log.Infow("registry.Register", p.Name())
    plugins[p.Name()] = p
}

func Get(name string) edgeruntime.RuntimePlugin {
    //log.Infow("registry.Get")
    return plugins[name]
}

// func (r MapRegistry) Get(name string) RuntimePlugin {
//     return plugins[name]
// }