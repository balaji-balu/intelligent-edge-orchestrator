# Installation

Edge-Orch is composed of three independent components:

| Component | Role                 |
| --------- | -------------------- |
| **CO**    | Central Orchestrator |
| **LO**    | Local Orchestrator   |
| **ERA**   | Edge Runtime Agent   |

In production, these typically run on **different hosts**.
For demos and development, they may be co-located on a single host.

---

## Design Principles

* ✅ One binary per component
* ✅ One install root per host
* ✅ No hardcoded filesystem paths
* ✅ Same binaries for demo and production
* ✅ Lifecycle managed by **systemd *or* container runtime**

---

## Install Layout (per host)

Each host has a single install root (example: `/opt/edge-orch`).

```text
/opt/edge-orch/
├── bin/
│   ├── co
│   ├── lo
│   └── era
├── config/
│   └── edge-orch/
│       ├── co/co.yaml
│       ├── lo/lo.yaml
│       └── era/era.yaml
├── data/
│   ├── co/
│   ├── lo/
│   └── era/
├── logs/
│   ├── co/
│   ├── lo/
│   └── era/
└── install.yaml
```

💡 A host only contains the components installed on it.

---

## Download

```bash
curl -LO https://github.com/<org>/edge-orch/releases/download/v0.1.0/edge-orch_0.1.0_linux_amd64.tar.gz
tar xzf edge-orch_0.1.0_linux_amd64.tar.gz
cd edge-orch_0.1.0_linux_amd64
```

---

## Quick Start (Demo – single host)

Install all components on one machine.

```bash
sudo ./edgectl init \
  --root-dir /opt/edge-orch \
  --units co,lo,era
```

This:

* creates the install layout
* copies binaries
* installs default configs
* generates systemd services
* enables and starts services

### Verify

```bash
edgectl status
```

```text
UNIT  STATE
co    running
lo    running
era   running
```

---

## Production Install (recommended)

### CO host

```bash
sudo ./edgectl init --unit co --root-dir /opt/edge-orch
```

### LO host

```bash
sudo ./edgectl init --unit lo --root-dir /opt/edge-orch
```

### ERA host(s)

```bash
sudo ./edgectl init --unit era --root-dir /opt/edge-orch
```

✅ Same binaries
✅ Same layout
✅ Different hosts

---

## Service Management (systemd)

On systemd-based hosts:

```bash
edgectl start era
edgectl stop era
edgectl restart era
edgectl status
```

Under the hood, `edgectl` delegates to `systemctl`.

---

## Containerized Environments

Edge-Orch components may also run as containers.

In this mode:

* lifecycle is managed by Docker / Kubernetes
* systemd and `edgectl start|stop` are **not used**
* filesystem layout is provided via a mounted volume

### Example (Docker)

```bash
docker run -d \
  --name era \
  -e EDGE_ORCH_ROOT=/edge-orch \
  -v /data/edge-orch:/edge-orch \
  edge-orch/era:latest
```

The same layout applies inside the container:

```text
/edge-orch/config/edge-orch/era/era.yaml
```

---

## Configuration

Each component reads its own config file:

```text
<ROOT>/config/edge-orch/<unit>/<unit>.yaml
```

Override if needed:

```bash
era --config /custom/path/era.yaml
```

---

## Environment Variables

| Variable          | Purpose                   |
| ----------------- | ------------------------- |
| `EDGE_ORCH_ROOT`  | Install root (required)   |
| `APP_CONFIG_FILE` | Explicit config file      |
| `APP_CONFIG_DIR`  | Explicit config directory |

---

## Uninstall

```bash
sudo edgectl stop co lo era
sudo rm -rf /opt/edge-orch
```

---

## Development Mode (no install)

```bash
go run ./cmd/era --config configs/era/dev.yaml
```

---

## Key Rule

> **A component is managed by exactly one supervisor per host:
> systemd *or* a container runtime — never both.**

---

## Architectural Note

> **Co-location is a demo convenience, not a production assumption.**
> All components communicate over the network, even when running on the same host.

