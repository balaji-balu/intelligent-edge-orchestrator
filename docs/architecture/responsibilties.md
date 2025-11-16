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


