# 📘 **ADR-001: ERA Architecture & Plugin Execution Model**

**Title:** ERA Architecture – Modular Runtime Manager with Plugin-Based Execution Engines
**Date:** 2025-12-04
**Status:** Proposed
**Context:**
ERA (Edge Runtime Agent) must support multiple execution backends:

* OCI containers (runc/Containerd)
* WASM runtimes
* Remote/Local Kubernetes (k3s, k8s, Talos-based k8s)
* Native binary execution
* Future custom runtimes

Constraints:

1. Execution technologies are evolving; ERA must be extensible.
2. To support LO/CO orchestration, ERA must standardize lifecycle + status reporting.
3. Each backend supports different capabilities (logs, metrics, sandboxing, networking, volumes).
4. We want **no branching/ifs in lifecycle logic** — plugins must abstract away differences.

---

## **Decision**

ERA will adopt:

### **A. Three-Core-Module Architecture**

1. **Runtime Manager** – orchestrates deployments and chooses plugins.
2. **Lifecycle Controller** – enforces app lifecycle phases uniformly.
3. **Plugin System** – adapters for execution engines.

### **B. Standard Plugin Interface**

All runtime engines must implement:

```go
type RuntimePlugin interface {
    Name() string
    Capabilities() []string

    Install(app *AppSpec) error
    Start(app *AppSpec) error
    Stop(appID string) error
    Delete(appID string) error

    Status(appID string) (AppStatus, error)
}
```

Capabilities may include:

* `"logs"`
* `"metrics"`
* `"sandbox"`
* `"volume"`
* `"network"`
* `"namespaces"`
* `"k8s"`
* `"wasm"`

Plugins register with:

```go
func RegisterRuntimePlugin(plugin RuntimePlugin)
```

---

### **C. Lifecycle Phases (global, plugin-agnostic)**

```
Validate → Fetch → Prepare → Install → Activate → HealthCheck
→ (Running State)
→ Update / Rollback → Uninstall → Delete
```

A unified state machine (below) will drive this.

Plugins only perform “execution operations”, not lifecycle logic.

---

### **D. ERA State Store**

ERA stores:

* deployments
* versions
* component runtime states
* drift data
* logs/events
* pending operations

Recommended storage: Badger/Pebble.

---

### **E. ERA API (Unix + HTTP/gRPC)**

Provides:

```
/deployments
/deployments/{id}/status
/plugins
/logs
/health
```

---

### **Consequences**

**Pros**

* Unified lifecycle across all runtimes
* Plugins remain small & focused
* Easy to add new runtimes
* ERA remains lightweight
* CO/LO get consistent status model
* Enables e2e testing with fake plugins

**Cons**

* Lifecycle controller must remain backward compatible
* Plugins may need capability mapping for advanced runtimes
* Two layers of abstraction may add code complexity

---

### **Rejected Alternatives**

❌ Hardcode execution methods for each runtime type

* Leads to branching code, hard to maintain.
  ❌ Single giant plugin with flags
* No extensibility.
  ❌ Directly embed containerd/K3s/Wasm code inside lifecycle
* couples lifecycle with execution engine.

---

### **Status**

Approved for implementation in ERA v0.1.

---



