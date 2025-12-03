package reporter

import (
    //"github.com/balaji-balu/margo-hello-world/internal/era/plugins"
    "github.com/balaji-balu/margo-hello-world/internal/era/lifecycle"
    "github.com/balaji-balu/margo-hello-world/pkg/era"
)
type StatusReporter struct {
    plugin lifecycle.RuntimePlugin
}

func NewStatusReporter(p lifecycle.RuntimePlugin) *StatusReporter {
    return &StatusReporter{plugin: p}
}

func (sr *StatusReporter) Status(name string) era.ComponentStatus {
    return sr.plugin.Status(name)
}
