package actuators

import (
    "context"
    //"encoding/json"
    "fmt"
    "log"
    "time"

    //"github.com/nats-io/nats.go"

    "github.com/balaji-balu/margo-hello-world/internal/natsbroker"
	"github.com/balaji-balu/margo-hello-world/internal/lo/reconciler"
	"github.com/balaji-balu/margo-hello-world/pkg/model"
)

// NatsActuator implements Actuator and talks to EN over NATS
type NatsActuator struct {
    nc      *natsbroker.Broker
    subject string // NATS subject for EN operations
    timeout time.Duration
    siteId  string
}

// NewNatsActuator creates a new NatsActuator
func NewNatsActuator(nc *natsbroker.Broker, siteId string, timeout time.Duration) *NatsActuator {
    //nc := natsbroker.New()
    return &NatsActuator{
        nc:      nc,
        //subject: subject,
        siteId: siteId,
        timeout: timeout,
    }
}

// Execute sends the operation to EN via NATS
func (a *NatsActuator) Execute(op reconciler.Operation) error {
    _, cancel := context.WithTimeout(context.Background(), a.timeout)
    defer cancel()

    log.Println("NatsActuator.Execute enter")

    log.Println("NatsActuator.Execute. operation: ", op)

    req := model.HostDeployRequest{
        HostID:       op.TargetNode,
        DeploymentID: op.DeploymentID,
        Component: model.ComponentProperties {
            Name     : op.Component.Name,   
            Repository: op.Component.Properties.Repository,
            Revision: op.Component.Properties.Revision,
            PackageLocation:op.Component.Properties.PackageURL,
            KeyLocation: op.Component.Properties.KeyURL,
        },
        Action:       string(op.Type),
    }

    subject := fmt.Sprintf("site.%s.deploy.%s", a.siteId, op.TargetNode)
    if err := a.nc.Publish(subject, req); err != nil {
        fmt.Errorf("NatsActuator: publish error", err)
    }
    // b, err := json.Marshal(req)
    // if err != nil {
    //     return fmt.Errorf("failed to marshal ENRequest: %w", err)
    // }

    // // NATS request-reply
    // msg, err := a.nc.RequestWithContext(ctx, a.subject, b)
    // if err != nil {
    //     return fmt.Errorf("nats request failed: %w", err)
    // }

    // var resp ENResponse
    // if err := json.Unmarshal(msg.Data, &resp); err != nil {
    //     return fmt.Errorf("failed to unmarshal ENResponse: %w", err)
    // }

    // if !resp.Success {
    //     return fmt.Errorf("EN operation failed: %s", resp.Message)
    // }

    log.Printf("[NatsActuator] Node: %s | Deployment: %s | Component: %s | Action: %s\n",
        op.TargetNode, op.DeploymentID, op.Component.Name, op.Type)

    log.Println("NatsActuator.Execute exit")    
    return nil
}
