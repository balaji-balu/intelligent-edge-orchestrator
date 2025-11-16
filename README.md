<p align="center">
  <img src="docs/logo.png" alt="Margo Logo" width="150"/>
</p>

# Intelligent Edge Orch

<p align="center">
  <em>Lightweight Edge Orchestration Platform (CO / LO / EN)</em>
</p>

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)

**Intelligent Edge Orch** is a lightweight **edge orchestration platform** demonstrating CO (Central Orchestrator), LO (Local Orchestrator), and EN (Edge Node) services for IoT, retail, and industrial edge computing environments.


## Key Features

- **Security-First Architecture** – Enterprise-grade protection for all edge deployments.
- **Deployment automation** - Ensures that software can be delivered to thousands of distributed edge nodes (far edge or device edge) with minimal manual intervention.  
- **Deploy AI Workloads at the Edge & IoT Devices** – Run applications closer to the data for faster insights.
- **Intelligent Edge Node Selection** – Automatically pick the best devices for deployment.
- **Offline-First Support** – Works smoothly even with intermittent or unreliable network connections.
- **App & Device Stores** –
    - App developers can register applications in the App Store.
    - Device developers can register edge devices in the Device Store.
- **Customer Access via CLI and Web Portal** – Manage and monitor deployments through command-line tools or a web interface.
- **Remote Workloads Lifecycle Management** – Add, update, and remove workloads on edge devices remotely. Air-gap environment supported
- **Interoperable Runtime Support** – Works with different runtime environments, including WASM/Ocre, container-based, and Kubernetes-based runtimes. Supports a range of devices from **small microcontrollers (MCP)** to **more powerful microprocessors (MPC)**.
- **Industry Focus** – Designed for industrial settings, manufacturing shop floors, and retail environments. 

## Quick Start

### 1. Clone the repository

```sh
git clone https://github.com/balaji-balu/margo-hello-world.git
cd margo-hello-world
```

### 2. Build and run services
1. start the services
```
podman-compose --env-file .env.staging up -d

docker ps
```
CO – Central orchestrator service
LO – Local orchestrator
EN – Edge node runtime

2. Open web portal
```
npm run dev
```
Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

3. open CLI
go to edgectl directory
```
go run main.go co status
```
4. Add the sample app to app store
```
$ go run main.go co add app --name "com-northstartida-digitron-orchestrator" --artifact https://github.com/edge-orchestration-platform/app-registry
Adding new application ...
✅ Application added successfully!

$ go run main.go co list apps
Listing apps ...
✅ Available applications:
 1. Digitron orchestrator 1.2.1      http://www.northstar-ida.com com-northstartida-digitron-orchestrator
```
5. deploy sample app
```
go run main.go co deploy --app "Digitron orchestrator" --site 3e5c21bc-2fef-4fd7-a2d0-60fc6b3260ad --deploytype compose 
```

### 3. Run tests
go test ./...

Optional: run smoke tests via GitHub Actions or local scripts.

## Architecture

Three-tier orchestration system:
- CO – Central Orchestrator  
- LO – Local Orchestrator (per site)  
- EN – Edge Node (per host)  
Users interact via Web Portal or CLI through CO API.

```
            +----------------------+
            | Web Portal / CLI     |
            +----------+-----------+
                       |
                       v
            +----------------------+
            | Central Orchestrator |
            |  (CO)                |
            +----------+-----------+
                       | Margo API
    -------------------------------------------------
    |                    |                         |
+----------------+   +----------------+     +----------------+
| Local Orchestr.|   | Local Orchestr.|     | Local Orchestr.|
|   (LO - Site A)|   |   (LO - Site B)| ... |   (LO - Site N)|
+--------+-------+   +--------+-------+     +--------+-------+
         |                  |                        |
   Overlay or Local Net  Overlay or Local Net  Overlay or Local Net
         |                  |                        |
   +-----------+       +-----------+           +-----------+
   | Host1     |       | Host4     |           | Host7     |
   | (EN1)     |       | (EN4)     |           | (EN7)     |
   | K8s/OCI/WASM|     | K8s/OCI/WASM|         | K8s/OCI/WASM|
   +-----------+       +-----------+           +-----------+
   | Host2     |       | Host5     |           | Host8     |
   | (EN2)     |       | (EN5)     |           | (EN8)     |
   | K8s/OCI/WASM|     | K8s/OCI/WASM|         | K8s/OCI/WASM|
   +-----------+       +-----------+           +-----------+
   | Host3     |       | Host6     |           | Host9     |
   | (EN3)     |       | (EN6)     |           | (EN9)     |
   | K8s/OCI/WASM|     | K8s/OCI/WASM|         | K8s/OCI/WASM|
   +-----------+       +-----------+           +-----------+
```        

- **ENs (Edge Nodes)** now explicitly show that each node can run workloads in multiple runtimes:
    - **K8s** – Kubernetes orchestrated containers.
    - **OCI / containerd** – Standard OCI containers.
    - **WASM** – Lightweight WebAssembly workloads.
- Each **LO** manages a **site** (e.g., factory floor, retail store) and connects to its **ENs** via overlay or local network.
- CO (Central Orchestrator) remains the control plane, accessible through **Web Portal / CLI**, using **Margo API**.

## Roadmap
- runtime: wasm, k3s/k8, Talos linux based k8s
- rich intelligent profile selection at CO
- margo based security between CO and LO
- otel  based observability
- CLI, web portal: OAuth2 for portal/API using keycloak
- web portal


## 💬 Contributing

See [CONTRIBUTING.md](contributing.md)
 for guidelines.
We welcome contributions in all forms:
- Bug reports
- Feature requests
- Code fixes
- Tests & CI/CD improvements
- Documentation improvements

### Folders structure
cmd
- cmd
	- co
	- lo
	- en
- ent                                <-- persistable data models
	- schema                    <-- this is where you create data model
	- migrate/migrations <-- atlas based migrations
- internal
	- orchestrator
	- gitobserver
	- gitwriter
	- gitfetcher
	- ocifetch
	- api
	- edgenodes 
- pkg
	- model
- atlas.hcl                          <--- db migration configuration
- docker-compose.yaml   <--- spins nats, postgres, co, lo, en containers
- dockerfile.co                  <--- co container
- dockerfile.lo                   <--- lo container
- dockerfile.en                  <--- en container 
- go.mod                           

### Tech stack
co, lo, en:
- messaging: NATS
- db: ent, atlas, postgres, boltz db
- go-git (neutralized git access)
- golang
- metrics: prometheus
- logger: zap 

web portal: 
- next.js, shadcn, tailwind.css 

cli: 
- cobra

## Contributors
Thanks to all contributors!

<a href="https://github.com/balaji-balu/intelligent-edge-orchestrator/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=balaji-balu/intelligent-edge-orchestrator" />
</a>

Made with [contrib.rocks](https://contrib.rocks).
## 📄 License
This project is licensed under the MIT License – see [LICENSE](license.md)
 for details.
