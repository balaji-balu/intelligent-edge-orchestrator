// Package lo: Reconciler + diff implementation + schemas for EN->LO->CO
// File: LO_reconciler_diff_and_schema.go

package lo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// -------------------------
// Types / Schemas
// -------------------------

// ComponentSpec describes how a component should be deployed.
type ComponentSpec struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Env         map[string]string `json:"env,omitempty"`
	Ports       []int             `json:"ports,omitempty"`
	Volumes     []string          `json:"volumes,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	// Optional scheduling constraints / node selector
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Arbitrary config that affects runtime; included in hash
	Config map[string]interface{} `json:"config,omitempty"`
}

// DeploymentSpec is the desired set of components for a deployment id.
type DeploymentSpec struct {
	DeploymentID string          `json:"deployment_id"`
	SiteID       string          `json:"site_id"`
	Components   []ComponentSpec `json:"components"`
	Generation   int64           `json:"generation"`
}

// ComponentState is what EN reports back about a running component.
type ComponentState struct {
	Name      string `json:"name"`
	Image     string `json:"image"`
	SpecHash  string `json:"spec_hash"`
	Status    string `json:"status"` // pending, installing, installed, failed
	UpdatedAt int64  `json:"updated_at"`
	NodeID    string `json:"node_id"`
}

// DeploymentState is LO's view of actual for a deployment across nodes
type DeploymentState struct {
	DeploymentID string           `json:"deployment_id"`
	SiteID       string           `json:"site_id"`
	Components   []ComponentState `json:"components"`
	ObservedGen  int64            `json:"observed_generation"`
}

// NodeState represents each execution node
type NodeState struct {
	NodeID    string            `json:"node_id"`
	Labels    map[string]string `json:"labels"`
	LastHeartbeat int64         `json:"last_heartbeat"`
	Online    bool              `json:"online"`
}

// -------------------------
// Operation (result of diff)
// -------------------------

type OperationType string

const (
	OpInstall OperationType = "install"
	OpUpdate  OperationType = "update"
	OpRemove  OperationType = "remove"
	OpNoOp    OperationType = "noop"
)

// Operation expresses what the reconciler wants to do on a specific node.
type Operation struct {
	Type      OperationType
	DeploymentID string
	Component ComponentSpec
	Actual    *ComponentState // may be nil
	TargetNode string          // node to apply the op
}

// -------------------------
// Hashing utilities
// -------------------------

// hashComponentSpec computes a stable hash for the meaningful part of the spec.
func hashComponentSpec(spec ComponentSpec) (string, error) {
	// We marshal a canonical subset to avoid hashing runtime-only fields.
	b, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// -------------------------
// diff() implementation (single-node view)
// -------------------------

// diff computes operations required to make actual match desired for a single logical deployment
func diff(desired DeploymentSpec, actual DeploymentState) ([]Operation, error) {
	ops := make([]Operation, 0)

	// Build a map of actual components by name
	actualMap := map[string]ComponentState{}
	for _, a := range actual.Components {
		actualMap[a.Name] = a
	}

	for _, d := range desired.Components {
		a, exists := actualMap[d.Name]

		specHash, err := hashComponentSpec(d)
		if err != nil {
			return nil, err
		}

		if !exists {
			ops = append(ops, Operation{
				Type: OpInstall,
				DeploymentID: desired.DeploymentID,
				Component: d,
				Actual: nil,
			})
			continue
		}

		// Compare hashes — if different, schedule update
		if a.SpecHash != specHash {
			ops = append(optsafe(), Operation{
				Type: OpUpdate,
				DeploymentID: desired.DeploymentID,
				Component: d,
				Actual: &a,
			})
		} else {
			ops = append(ops, Operation{
				Type: OpNoOp,
				DeploymentID: desired.DeploymentID,
				Component: d,
				Actual: &a,
			})
		}

		delete(actualMap, d.Name)
	}

	// Anything left in actualMap is not desired -> remove
	for _, leftover := range actualMap {
		ops = append(ops, Operation{
			Type: OpRemove,
			DeploymentID: actual.DeploymentID,
			Component: ComponentSpec{Name: leftover.Name},
			Actual: &leftover,
		})
	}

	return ops, nil
}

// -------------------------
// Multi-node diff: assign operations to nodes
// -------------------------

// Strategy: For simple cases we schedule every component to all active nodes
// or use NodeSelector to pick matching nodes. This function returns per-node ops.
func diffMulti(desired DeploymentSpec, actualPerNode map[string]DeploymentState, nodes map[string]NodeState) (map[string][]Operation, error) {
	// result: nodeID -> ops
	result := map[string][]Operation{}

	// Precompute spec hashes for desired components
	hashes := map[string]string{}
	for _, c := range desired.Components {
		h, err := hashComponentSpec(c)
		if err != nil {
			return nil, err
		}
		hashes[c.Name] = h
	}

	// For each node, decide what to do.
	for nodeID, node := range nodes {
		if !node.Online {
			// skip offline nodes
			continue
		}

		actual, found := actualPerNode[nodeID]
		if !found {
			// no observed state for this node — treat as empty
			actual = DeploymentState{DeploymentID: desired.DeploymentID, SiteID: desired.SiteID}
		}

		// Build actualMap for the node
		actualMap := map[string]ComponentState{}
		for _, a := range actual.Components {
			actualMap[a.Name] = a
		}

		for _, d := range desired.Components {
			// quick nodeSelector match: if spec has nodeSelector, ensure it matches node labels
			if len(d.NodeSelector) > 0 {
				if !nodeMatches(d.NodeSelector, node.Labels) {
					// skip this node for this component
					continue
				}
			}

			a, exists := actualMap[d.Name]
			if !exists {
				op := Operation{Type: OpInstall, DeploymentID: desired.DeploymentID, Component: d, Actual: nil, TargetNode: nodeID}
				result[nodeID] = append(result[nodeID], op)
				continue
			}

			if a.SpecHash != hashes[d.Name] {
				op := Operation{Type: OpUpdate, DeploymentID: desired.DeploymentID, Component: d, Actual: &a, TargetNode: nodeID}
				result[nodeID] = append(result[nodeID], op)
			} else {
				// Noop intentionally not added — optional
			}

			// processed this component for node
			delete(actualMap, d.Name)
		}

		// Anything left in actualMap -> remove on this node
		for _, leftover := range actualMap {
			result[nodeID] = append(result[nodeID], Operation{Type: OpRemove, DeploymentID: desired.DeploymentID, Component: ComponentSpec{Name: leftover.Name}, Actual: &leftover, TargetNode: nodeID})
		}
	}

	return result, nil
}

func nodeMatches(selector, labels map[string]string) bool {
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}

// -------------------------
// Reconciler wiring
// -------------------------

// Store provides access to desired/actual state caches. In production this is backed by DB/cache.
type Store interface {
	GetDesired(deploymentID string) (DeploymentSpec, error)
	GetActual(deploymentID string) (DeploymentState, error)
	GetActualPerNode(deploymentID string) (map[string]DeploymentState, error)
	GetNodes(siteID string) (map[string]NodeState, error)
}

// Actuator executes operations on nodes (talks to EN).
type Actuator interface {
	Execute(op Operation) error
}

// Reporter reports reconciliation results back to CO.
type Reporter interface {
	ReportState(deploymentID string, state DeploymentState) error
}

// Reconciler runs reconciliation for a deployment key.
type Reconciler struct {
	store    Store
	actuator Actuator
	reporter Reporter
	mu       sync.Mutex
}

func NewReconciler(s Store, a Actuator, r Reporter) *Reconciler {
	return &Reconciler{store: s, actuator: a, reporter: r}
}

// Reconcile main entry: single-node (simple) version
func (r *Reconciler) Reconcile(deploymentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	desired, err := r.store.GetDesired(deploymentID)
	if err != nil {
		return err
	}

	actual, err := r.store.GetActual(deploymentID)
	if err != nil {
		return err
	}

	ops, err := diff(desired, actual) 
	if err != nil {
		return err
	}

	// Execute operations sequentially (could be parallel with worker pool)
	for _, op := range ops {
		// choose target node if not set — default to node specified in desired or first node
		if op.TargetNode == "" {
			// best-effort: pick node from actual if available
			if op.Actual != nil && op.Actual.NodeID != "" {
				op.TargetNode = op.Actual.NodeID
			}
		}

		if err := r.actuator.Execute(op); err != nil {
			// Log and continue — real impl should mark status and possibly retry
			fmt.Printf("actuator execute failed: %v\n", err)
		}
	}

	// After executing, fetch updated actual and report to CO
	updatedActual, err := r.store.GetActual(deploymentID)
	if err != nil {
		return err
	}

	if err := r.reporter.ReportState(deploymentID, updatedActual); err != nil {
		return err
	}

	return nil
}

// ReconcileMulti runs multi-node reconciliation and assigns ops per node.
func (r *Reconciler) ReconcileMulti(deploymentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	desired, err := r.store.GetDesired(deploymentID)
	if err != nil {
		return err
	}

	actualPerNode, err := r.store.GetActualPerNode(deploymentID)
	if err != nil {
		return err
	}

	nodes, err := r.store.GetNodes(desired.SiteID)
	if err != nil {
		return err
	}

	opMap, err := diffMulti(desired, actualPerNode, nodes)
	if err != nil {
		return err
	}

	// Execute per-node ops (sequential here; for performance use goroutines and rate-limits)
	for nodeID, ops := range opMap {
		for _, op := range ops {
			op.TargetNode = nodeID
			if err := r.actuator.Execute(op); err != nil {
				fmt.Printf("execute op on node %s failed: %v\n", nodeID, err)
			}
		}
	}
/*
	// After executing, collect per-node updated actual and merge/report as needed
	// For simplicity, fetch a single aggregated actual view and report
	updatedActual, err := r.store.GetActual(deploymentID)
	if err != nil {
		return err
	}

	if err := r.reporter.ReportState(deploymentID, updatedActual); err != nil {
		return err
	}
*/

	return nil
}



