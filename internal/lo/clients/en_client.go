package lo

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/nats-io/nats.go"
)

// NatsActuator implements Actuator and talks to EN over NATS
type NatsActuator struct {
    nc      *nats.Conn
    subject string // NATS subject for EN operations
    timeout time.Duration
}

// NewNatsActuator creates a new NatsActuator
func NewNatsActuator(nc *nats.Conn, subject string, timeout time.Duration) *NatsActuator {
    return &NatsActuator{
        nc:      nc,
        subject: subject,
        timeout: timeout,
    }
}

// Execute sends the operation to EN via NATS
func (a *NatsActuator) Execute(op Operation) error {
    ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
    defer cancel()

    req := ENRequest{
        NodeID:       op.TargetNode,
        DeploymentID: op.DeploymentID,
        Component:    op.Component,
        Action:       string(op.Type),
    }

    b, err := json.Marshal(req)
    if err != nil {
        return fmt.Errorf("failed to marshal ENRequest: %w", err)
    }

    // NATS request-reply
    msg, err := a.nc.RequestWithContext(ctx, a.subject, b)
    if err != nil {
        return fmt.Errorf("nats request failed: %w", err)
    }

    var resp ENResponse
    if err := json.Unmarshal(msg.Data, &resp); err != nil {
        return fmt.Errorf("failed to unmarshal ENResponse: %w", err)
    }

    if !resp.Success {
        return fmt.Errorf("EN operation failed: %s", resp.Message)
    }

    fmt.Printf("[NatsActuator] Node: %s | Deployment: %s | Component: %s | Action: %s\n",
        op.TargetNode, op.DeploymentID, op.Component.Name, op.Type)

    return nil
}
