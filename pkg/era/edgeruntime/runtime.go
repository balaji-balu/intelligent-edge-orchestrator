package edgeruntime
import (
	//"github.com/balaji-balu/margo-hello-world/pkg/era"
)

type ComponentSpec struct {
    Name      string
    Version   string
    Runtime     string
    Image     string
    WasmFile  string
    Artifact  string
    Args      []string
}

type ComponentStatus struct {
    Name       string `json:"name"`
    Version    string `json:"version"`
    State      string `json:"state"`
    Message    string `json:"message,omitempty"`
    Timestamp  int64  `json:"timestamp"`
}

type RuntimePlugin interface {
    Name() string
    Capabilities() []string

    Install(ComponentSpec) error
    Start(ComponentSpec) error
    Stop(string) error
    Delete(string) error
    Status(string) (ComponentStatus, error)
}

// type RuntimePlugin interface {
//     Install(c era.ComponentSpec) error
//     Start(c era.ComponentSpec) error
//     Stop(name string) error
//     Remove(name string) error
//     Status(name string) era.ComponentStatus
// }