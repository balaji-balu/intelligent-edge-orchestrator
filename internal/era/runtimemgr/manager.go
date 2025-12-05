package runtimemgr

import (
    "fmt"
    "go.uber.org/zap"

    "github.com/balaji-balu/margo-hello-world/pkg/model"
    "github.com/balaji-balu/margo-hello-world/pkg/era/edgeruntime"
    "github.com/balaji-balu/margo-hello-world/internal/natsbroker"
    "github.com/balaji-balu/margo-hello-world/internal/era/lifecycle"
    "github.com/balaji-balu/margo-hello-world/internal/era/reporter"
)

type RuntimeManager struct {
    lifecycle   *lifecycle.LifecycleController
    reporter    *reporter.StatusReporter
    log         *zap.SugaredLogger
    nb          *natsbroker.Broker
}

func NewRuntimeManager(runtime string, nb *natsbroker.Broker, log *zap.SugaredLogger) *RuntimeManager {
    return &RuntimeManager{
        lifecycle: lifecycle.NewLifecycleController(runtime, log),
        reporter:  reporter.NewStatusReporter(runtime, log),
        log: log,
        nb: nb,
    }
}

func (rm *RuntimeManager) Deploy(c edgeruntime.ComponentSpec) error {
    rm.log.Infow("RuntimeManager: Deploy")
    return rm.lifecycle.Apply(c)
}

func (rm *RuntimeManager) GetStatus(name string) edgeruntime.ComponentStatus {
    return rm.reporter.Status(name)
}

func (rm *RuntimeManager) Stop(name string) error {
    return rm.lifecycle.Stop(name)
}

func (rm *RuntimeManager) Delete(name string) error {
    return rm.lifecycle.Delete(name)
}

func (rm *RuntimeManager) LoActionDispatcher(siteID, hostID string){
    go func() {
        subj := fmt.Sprintf("site.%s.deploy.%s", siteID, hostID)
        rm.nb.Subscribe3(subj, func(req model.DiffOp) {
            rm.log.Infow("req received:", "req", req)
            //rm.log.Infow("deploy request received", hostID)

            rm.log.Infow("Received", "Deployment type", req.App.DepType)
            
            //TBD: runtime must be "containerd". rest "not implemented"
            rm.lifecycle.HandleAction(req)
        })
    }()
}
