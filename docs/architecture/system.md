### Components

- **CO (Central Orchestrator):** API server that manages profile selection, app registry, deployment registry, and git push operations. Handles all communications with Local Orchestrators (LO) and Edge Nodes (EN). Implements the server side of the API and provides access to both the CLI and the Web Portal. Supports GitOps workflows. 
- **LO (Local Orchestrator):** Observes git repositories (pull-based) and pushes changes to Edge Nodes. Deploys workloads to multiple edges in parallel. Adaptive to intermittent network conditions - changes to Git pull, push or NATS based event driven communication with CO
- **EN (Edge Node):** Agent implementing the client side of the API. Responsible for fetching OCI artifacts, supporting Helm-based deployments on Kubernetes, container-based runtimes, and WASM-based workloads.
- **Deployment Status & Node Capabilities:** Tracks deployment states and capabilities of each edge device.
- **edgectl:** CLI tool to access the CO API.
- **Web Portal:** Provides a graphical interface for managing deployments, apps, and devices.

### Data modeling
[Margo](margo.org) specification defined these data models
- *[application description](https://specification.margo.org/specification/application-package/application-description/):* defines the app specification, artifacts location, application resources requirements. 
- *[application deployment](https://specification.margo.org/specification/margo-management-interface/desired-state/)*: used for deploying workloads
- *[deployment status](https://specification.margo.org/specification/margo-management-interface/deployment-status/):* used for communicating the workload deployment status
- *[device capabilities](https://specification.margo.org/specification/margo-management-interface/device-capabilities/)*: edge node capabilities like cpu, memory, gpu etc
- *sites, hosts, orchestrator, tenants*