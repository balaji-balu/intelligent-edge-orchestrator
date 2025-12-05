# 📘 **ADR-002: Plugin Registry Architecture**

**Title:** Plugin Registry Architecture for ERA Runtime Agent
**Date:** 2025-12-04
**Status:** Proposed
**Authors:** balaji b, ERA Architecture Team

---

# 1. **Context**

ERA (Edge Runtime Agent) must support multiple execution backends:

* OCI (runc or go-containerregistry)
* WASM (wasmtime, wamr)
* containerd-based execution
* embedded k3s / remote k8s / Talos-based k8s
* native/bare-metal binary execution
* future optional runtimes

To allow **extensible runtime selection**, we need:

1. A **central registry** where all runtime plugins register.
2. A mechanism to automatically load plugins during agent startup.
3. A way for the Runtime Manager to select the correct runtime plugin based on deployment spec.
4. A design that supports:

   * unit tests with fake plugins
   * build tags for minimal agents
   * pluggable new runtimes without modifying core ERA code

The Go standard library provides a `.so` plugin system (`plugin` package) but it is:

* unstable across Go versions
* not supported on Windows
* CGO-dependent
* incompatible with static builds or Talos/Linux distros
* incompatible with cross-compilation

**Therefore we must avoid Go `.so` plugins.**

We adopt a **static plugin registration system** using `init()` + registry.

This pattern is widely used in:

* Kubernetes cloud providers
* Terraform providers
* Docker storage drivers
* containerd runtime shims
* Prometheus collector registries

---

# 2. **Decision**

ERA will use a **static plugin registry system**:

### ✔ **Each plugin registers itself using `init()`**

When imported, each plugin executes:

```go
func init() {
    runtime.Register(&SomePlugin{})
}
```

### ✔ **Plugins use a unified interface (`RuntimePlugin`)**

Defined in `pkg/runtime`:

```go
type RuntimePlugin interface {
    Name() string
    Capabilities() []string

    Install(*AppSpec) error
    Start(*AppSpec) error
    Stop(string) error
    Delete(string) error
    Status(string) (AppStatus, error)
}
```

### ✔ **Plugins register themselves into a global registry**

```go
var registry = map[string]RuntimePlugin{}

func Register(p RuntimePlugin) {
    registry[p.Name()] = p
}
```

### ✔ **ERA imports plugins via blank imports**

In `cmd/era/main.go`:

```go
import (
    _ "github.com/era/plugins/wasm"
    _ "github.com/era/plugins/oci"
    _ "github.com/era/plugins/containerd"
    _ "github.com/era/plugins/k3s"
    _ "github.com/era/plugins/talos"
)
```

This ensures each plugin’s `init()` runs and auto-registers itself.

### ✔ **Plugins can be enabled/disabled via build tags**

Example:

```go
//go:build wasm
import _ "github.com/era/plugins/wasm"
```

This supports minimal footprint builds for IoT/Edge.

### ✔ **Runtime Manager selects the plugin at runtime**

```go
plugin := runtime.Get(spec.Runtime)
```

If `spec.Runtime = "wasm"`, the wasm plugin is selected automatically.

---

# 3. **Architecture Overview**

```
         +---------------------------+
         |    ERA Runtime Manager    |
         +---------------------------+
                     |
                     |  Get("wasm")
                     v
         +---------------------------+
         |     Plugin Registry       |
         +---------------------------+
            |        |        |
            v        v        v
     +---------+ +---------+ +----------+
     |  wasm   | |   oci   | | containerd |
     | plugin  | | plugin  | |  plugin    |
     +---------+ +---------+ +----------+
            ^        ^        ^
            |        |        |
  import _  | import _ | import _
            |        |        |
+------------ init() ------------+
| each plugin registers itself   |
+--------------------------------+
```

---

# 4. **Consequences**

## ✔ Pros

* **Extremely simple**: no dynamic loading, no CGO.
* **Safe & stable**: avoids `.so` plugin ABI issues.
* **Static & type-safe**: problems caught at compile time.
* **Modular**: new plugins don’t require changes to core logic.
* **Performance**: zero runtime overhead.
* **Minimal agent builds**: build tags allow stripping unused plugins.
* **Testing-friendly**: tests can register fake plugins easily.

## ✘ Cons

* Plugins must be linked at compile time (not dynamically loaded).
* Requires blank imports in main.
* Plugins must follow interface exactly.

Given ERA’s deployment model (immutable edge agent binaries), **static registration is ideal**.

---

# 5. **Alternatives Considered**

### ❌ **Go plugin system (`plugin/*.so`)**

Rejected because:

* Version dependent
* OS dependent
* CGO required
* Not portable / not stable

### ❌ **Manually wiring plugins in main()**

Rejected because:

* Tight coupling
* Adding plugins requires editing core code
* Harder to test

### ❌ **IoC container / reflection**

Rejected because:

* Overkill
* Not idiomatic in Go
* Adds complexity with no benefit

---

# 6. **Example Implementation**

## Registry

```go
package runtime

var plugins = map[string]RuntimePlugin{}

func Register(p RuntimePlugin) {
    plugins[p.Name()] = p
}

func Get(name string) RuntimePlugin {
    return plugins[name]
}
```

## Plugin Auto-Registering Itself

```go
package wasm

import "github.com/era/runtime"

func init() {
    runtime.Register(&WasmPlugin{})
}
```

## Blank Imports in main()

```go
import _ "github.com/era/plugins/wasm"
```

---

# 7. **Status**

**Approved**
This architecture becomes the default plugin discovery mechanism for ERA v0.1.

