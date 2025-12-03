package runtimemanager

import (
    "log"
    "github.com/balaji-balu/margo-hello-world/pkg/era"
    "github.com/balaji-balu/margo-hello-world/internal/era/lifecycle"
    "github.com/balaji-balu/margo-hello-world/internal/era/reporter"
)

type RuntimeManager struct {
    lifecycle *lifecycle.LifecycleController
    reporter  *reporter.StatusReporter
}

func NewRuntimeManager(p lifecycle.RuntimePlugin) *RuntimeManager {
    return &RuntimeManager{
        lifecycle: lifecycle.NewLifecycleController(p),
        reporter:  reporter.NewStatusReporter(p),
    }
}

func (rm *RuntimeManager) Deploy(c era.ComponentSpec) error {
    log.Println("RuntimeManager: Deploy")
    return rm.lifecycle.Apply(c)
}

func (rm *RuntimeManager) Stop(name string) error {
    return rm.lifecycle.Stop(name)
}

func (rm *RuntimeManager) GetStatus(name string) era.ComponentStatus {
    return rm.reporter.Status(name)
}
