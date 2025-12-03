package lifecycle

import (
    "log"
    //"github.com/balaji-balu/margo-hello-world/internal/era/runtimemanager"
    //"github.com/balaji-balu/margo-hello-world/internal/era/plugins"
    "github.com/balaji-balu/margo-hello-world/pkg/era"
)

type RuntimePlugin interface {
    Install(c era.ComponentSpec) error
    Start(c era.ComponentSpec) error
    Stop(name string) error
    Remove(name string) error
    Status(name string) era.ComponentStatus
}

type LifecycleController struct {
    plugin RuntimePlugin
}

func NewLifecycleController(plugin RuntimePlugin) *LifecycleController {
    return &LifecycleController{plugin: plugin}
}

func (lc *LifecycleController) Apply(c era.ComponentSpec) error {
    log.Println("LifecycleController: Apply")
    if err := lc.plugin.Install(c); err != nil {
        return err
    }
    return lc.plugin.Start(c)
}

func (lc *LifecycleController) Stop(name string) error {
    return lc.plugin.Stop(name)
}
