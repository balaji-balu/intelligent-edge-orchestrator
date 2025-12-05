package reporter

import (
    "go.uber.org/zap"
    
    "github.com/balaji-balu/margo-hello-world/internal/era/plugins"
    //"github.com/balaji-balu/margo-hello-world/internal/era/lifecycle"
    "github.com/balaji-balu/margo-hello-world/pkg/era/edgeruntime"
)
type StatusReporter struct {
    log *zap.SugaredLogger
    plugin edgeruntime.RuntimePlugin
}

func NewStatusReporter(runtime string, log *zap.SugaredLogger ) *StatusReporter {
    return &StatusReporter{
        plugin: plugins.Get(runtime),
        log: log,
    }
}

func (sr *StatusReporter) Status(name string) edgeruntime.ComponentStatus {
    status, _ := sr.plugin.Status(name)
    return status
}
