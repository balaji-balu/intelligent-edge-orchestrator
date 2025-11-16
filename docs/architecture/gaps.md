features:
| Connectivity resilience | Local autonomy when network link to CO fails |
| Security | Mutual TLS between CO, LO, and EN; Role-based access (Keycloak) |
| Multi-tenancy | Logical isolation of users and sites |

capture success metrics:
- Deployment success rate > 99%
- End-to-end latency (CO→EN) < 2s under normal network
- Offline operation capability at least 24h
- Seamless recovery after network reconnection

---

# ⭐ **Missing Components / Gaps**

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

