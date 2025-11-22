package reconciler

// --------------------
// Models (kept same)
// --------------------
import (
	"context"
	"time"
	"github.com/google/uuid"
)
type ComponentSpec struct {
	Name       string `yaml:"name"`
	Properties struct {
		PackageURL string `yaml:"packageLocation,omitempty"`
		KeyURL     string `yaml:"keyLocation,omitempty"`
		Repository string `yaml:"repository,omitempty"`
		Revision   string `yaml:"revision,omitempty"`
		Wait       *bool  `yaml:"wait,omitempty"`
		Timeout    string `yaml:"timeout,omitempty"`
		NodeSelector map[string]string `yaml:"nodeSelector,omitempty"`
	} `yaml:"properties"`
}

type ApplicationDeployment struct {
	Metadata struct {
		Annotations struct {
			ID string `yaml:"id"`
		} `yaml:"annotations"`
	} `yaml:"metadata"`
	Spec struct {
		DeploymentProfile struct {
			Type       string          `yaml:"type"`
			Components []ComponentSpec `yaml:"components"`
		} `yaml:"deploymentProfile"`
	} `yaml:"spec"`
}

type DeploymentDesiredComponent struct {
	DeploymentID string
	Component    ComponentSpec
	SpecHash     string
}

type DeploymentComponentStatus struct {
	ID            uuid.UUID
	DeploymentID  string
	HostID        string
	ComponentName string
	DesiredHash   string
	ActualHash    string
	Status        string
	Message       string
	LastUpdate    time.Time
}

type Host struct {
	ID     string
	Labels map[string]string
	Status string // "alive" or "offline"
}

type OpType string

const (
	OpInstall OpType = "install"
	OpUpdate  OpType = "update"
	OpRemove  OpType = "remove"
)

type Operation struct {
	Type         OpType
	DeploymentID string
	Component    ComponentSpec
	DesiredHash  string
	Actual       *DeploymentComponentStatus // may be nil for install
	TargetNode   string
}

// --------------------
// Store interface — updated for hash-based API
// --------------------

type Store interface {
	// Hash-focused APIs
	// desired: desired[deploymentID][component] = hash
	SetDesired(ctx context.Context, depID string, comp ComponentSpec, specHash string) error
	GetDesiredHashes(ctx context.Context, depID string) (map[string]string, error)

	// actual: actual[deploymentID][hostID][component] = hash
	SetActualHash(ctx context.Context, depID string, hostID string, compName string, hash string) error
	GetActualHashes(ctx context.Context, depID string) (map[string]map[string]string, error)

	// helpers kept for reporting/compat
	GetActual(ctx context.Context, deploymentID string) ([]*DeploymentComponentStatus, error)
	GetNodes(ctx context.Context, siteID string) ([]*Host, error)
}

type Actuator interface {
	Execute(op Operation) error
}

type Reporter interface {
	ReportState(deploymentID string, actual []*DeploymentComponentStatus) error
}