package reconciler 

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/balaji-balu/margo-hello-world/internal/lo/logger"
)

type Reconciler struct {
	store    Store
	actuator Actuator
	reporter Reporter
	mu       sync.Mutex
}

func NewReconciler(s Store, a Actuator, r Reporter) *Reconciler {
	return &Reconciler{store: s, actuator: a, reporter: r}
}

func HashComponentSpec(spec interface{}) (string, error) {
	b, err := json.Marshal(spec)
	if err != nil { return "", err }
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// diffMulti compares desiredHashes vs actualHashes and returns operations per host
func diffMulti(
	components []ComponentSpec,
	deploymentID string,
	desiredHashes map[string]string,
	actual map[string]map[string]string,
	hosts []*Host,
) map[string][]Operation {

	result := make(map[string][]Operation)

	for _, host := range hosts {
		if host.Status != "alive" {
			logger.Debug("Skipping host because not alive", "hostID", host.ID, "status", host.Status)
			continue
		}
		hostID := host.ID
		actualForHost := map[string]string{}
		if actual != nil {
			if a, ok := actual[hostID]; ok {
				actualForHost = a
			}
		}

		for _, comp := range components {
			if len(comp.Properties.NodeSelector) > 0 && !nodeMatches(comp.Properties.NodeSelector, host.Labels) {
				logger.Debug("Skipping component due to node selector mismatch", "component", comp.Name, "hostID", hostID)
				continue
			}

			dHash := desiredHashes[comp.Name]
			aHash := ""
			if actualForHost != nil {
				aHash = actualForHost[comp.Name]
			}

			if aHash == "" {
				op := Operation{
					Type:         OpInstall,
					DeploymentID: deploymentID,
					Component:    comp,
					DesiredHash:  dHash,
					Actual:       nil,
					TargetNode:   hostID,
				}
				result[hostID] = append(result[hostID], op)
				continue
			}

			if aHash != dHash {
				op := Operation{
					Type:         OpUpdate,
					DeploymentID: deploymentID,
					Component:    comp,
					DesiredHash:  dHash,
					Actual: &DeploymentComponentStatus{
						DeploymentID:  deploymentID,
						HostID:        hostID,
						ComponentName: comp.Name,
						ActualHash:    aHash,
						DesiredHash:   dHash,
					},
					TargetNode:   hostID,
				}
				result[hostID] = append(result[hostID], op)
			}
		}
	}

	return result
}

func nodeMatches(selector map[string]string, labels map[string]string) bool {
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}

// ReconcileMulti updated to use SetDesired + GetActualHashes
func (r *Reconciler) ReconcileMulti(siteID string, 
	deploymentID string, dep ApplicationDeployment) error {
    
	ctx := context.Background()
    logger.Info("ReconcileMulti: enter", "siteID", siteID, "deploymentID", deploymentID)

    r.mu.Lock()
    defer func() {
        r.mu.Unlock()
        logger.Info("ReconcileMulti: exit", "siteID", siteID, "deploymentID", deploymentID)
    }()

	components := dep.Spec.DeploymentProfile.Components
	//deploymentID := desired.Metadata.Annotations.ID

    // 1. READ desired from store instead of computing
    desiredHashes, err := r.store.GetDesiredHashes(ctx, deploymentID)
    if err != nil {
        logger.Error("Failed to get desired hashes", "err", err)
        return fmt.Errorf("get desired hashes: %w", err)
    }

    // 2. fetch nodes
    nodes, err := r.store.GetNodes(ctx, siteID)
    if err != nil {
        logger.Error("Failed to get nodes", "err", err, "siteID", siteID)
        return fmt.Errorf("get nodes: %w", err)
    }

    // 3. fetch actual hashes (per-host)
    actualHashes, err := r.store.GetActualHashes(ctx, deploymentID)
    if err != nil {
        logger.Error("Failed to get actual hashes", "err", err)
        return fmt.Errorf("get actual hashes: %w", err)
    }

    // 4. compute operations
    opMap := diffMulti(components, deploymentID, desiredHashes, actualHashes, nodes)

    // 5. execute operations per node
    for nodeID, ops := range opMap {
        for _, op := range ops {
            op.TargetNode = nodeID
            logger.Info("Executing operation", "deploymentID", op.DeploymentID, "nodeID", nodeID, "component", op.Component.Name, "opType", op.Type)

            if err := r.actuator.Execute(op); err != nil {
                logger.Error("Operation execution failed", "deploymentID", op.DeploymentID, "nodeID", nodeID, "component", op.Component.Name, "opType", op.Type, "err", err)
            }
        }
    }

    // 6. fetch updated actual for reporting
    updatedActual, err := r.store.GetActual(ctx, deploymentID)
    if err != nil {
        logger.Error("Failed to fetch updated actual", "err", err)
        return fmt.Errorf("fetch updated actual: %w", err)
    }

    if r.reporter != nil {
        if err := r.reporter.ReportState(deploymentID, updatedActual); err != nil {
            logger.Error("Reporter failed", "err", err, "deploymentID", deploymentID)
        }
    }

    return nil
}
