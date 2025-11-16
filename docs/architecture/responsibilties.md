Here is a **clean, structured, complete list of CO / LO / EN responsibilities** based on your description, plus the **missing pieces** that are usually needed in an orchestrator like this.

---

# ✅ **CO (Central Orchestrator) Responsibilities**

### **Core Functions**

1. **Health Checks**

   * Expose `/healthz`
   * Read `/healthz` from LO and EN as needed (optional)

2. **App Management**

   * Read app-registry from GitHub (remote or local)
   * Read `margo.yaml`
   * YAML unmarshal → Go structs
   * Write to Postgres:

     * Application Descriptor
     * Deployment Profiles
     * Components

3. **List Apps**

   * Query DB for apps, components, profiles

4. **Deploy App**

   * Read app descriptor + deployment profile from DB
   * Create deployments for each site → Write deployment YAMLs into Git repo (`deployments/`)
   * Commit + push (write service)
   * Broadcast deployment events (stream to UI/CLI)
   * Receive status events from LO and proxy to clients
   * Maintain deployment state machine in DB:

     * Pending → InProgress → Installing → Installed / Failed

---

# ✅ **LO (Local Orchestrator) Responsibilities**

### **Core Functions**

1. **Health Check**

   * Expose `/healthz`

2. **Watch Deployment Repo**

   * Wait for new deployments for its site
   * Pull latest Git revision
   * Compare desired-state vs current-state

3. **Node-Level Deployment Scheduling**

   * Identify connected EN nodes
   * Send deployment requests to each EN
   * Support:

     * Install
     * Update
     * Rollback
     * Uninstall

4. **Collect Deployment Status**

   * Receive status events from EN:

     * pending
     * installing
     * installed
     * failed
   * Forward these events to CO

5. **Heartbeat / Node Management**

   * Track EN Heartbeats (“active nodes” logic)
   * Maintain node registry locally
   * Report connected nodes to CO (optional)

---

# ✅ **EN (Execution Node) Responsibilities**

### **Core Functions**

1. **Health Check**

   * Expose `/healthz`

2. **Container Deployment Engine**

   * Receive deployment request from LO
   * Install app (containerd/podman/docker)
   * Start container
   * Validate health/readiness

3. **Send Status Updates to LO**

   * pending
   * installing
   * installed
   * failed

4. **Log / Metrics / Trace Integration (optional)**

   * Promtail: logs → Loki
   * Tempo: traces
   * Prometheus exporter (deployment metrics per node)

---

# ⭐ **Missing Components / Gaps You Should Add**

Here are the **important missing pieces** to make the system robust:

---

## **1. App Versioning**

* Each app should have:

  * app version
  * deployment profile version
  * component version
* Store versions in DB
* Do version-based rollout

---

## **2. Rollbacks**

Deployment should support:

* auto rollback on failure
* manual rollback via CLI/UI

---

## **3. Deployment Garbage Collection**

* Remove old revisions from repo
* Keep N versions per site
* Cleanup old container images on EN

---

## **4. Node Inventory / Capability Check**

Before deploying:

* LO or CO should check node capabilities:

  * CPU, memory, accelerators
  * labels (gpu, edge type)
  * connectivity
* Match deployment requirements → schedule only on compatible EN

---

## **5. Heartbeat & Failure Handling**

LO should:

* Track EN heartbeat
* Mark nodes "offline"
* Reconcile deployments when node comes back

---

## **6. Desired-State Reconciliation Loop**

Every orchestration system needs a **reconciler**:

* LO continuously checks:

  * desired-state (from Git)
  * observed-state (from EN)
* If mismatch → send new requests to EN

---

## **7. UI/CLI Streaming Events (already mentioned but formalize)**

* SSE/WebSocket endpoint
* Events:

  * deployment created
  * deployment updated
  * node joined/left
  * logs tailing (optional)

---

## **8. Authentication & Authorization**

If CO deals with multiple sites and UI:

* JWT tokens
* Roles:

  * admin
  * operator
  * read-only

---

## **9. App Removal / Undeploy**

Missing in your description:

* LO should detect deletion in Git repo
* Trigger EN to remove running containers

---

## **10. Observability Integration**

Already partly mentioned but needs explicit functions:

### **CO**

* metrics: request duration, DB queries, git operations

### **LO**

* metrics: reconciler duration, EN response latency

### **EN**

* metrics: container lifecycle, runtime resource usage

---

# ✅ **Final Combined Architecture Overview**

### **CO**

* API + UI/CLI → manage apps
* Git write service for deployments
* Global state database
* Event broadcast
* Versioning, rollback, auth

### **LO**

* Git read + watch loop
* Node registry
* Deployment scheduler
* Reconciler
* Forward deployment status → CO

### **EN**

* Container operations
* Health checks
* Log/metric collectors
* Heartbeat + status events → LO

---

