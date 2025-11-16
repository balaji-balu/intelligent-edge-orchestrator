
# ✅ What Is Desired-State Reconciliation?

A **reconciler** continuously checks two states:

1. **Desired State** – what should be running
   (from Git repo deployments/ directory)

2. **Observed State** – what is actually running
   (from EN reports and LO’s current knowledge)

Whenever these two differ, LO takes action to **make reality match the desired state**.

This ensures:

* self-healing
* idempotent deployments
* automatic resync if LO restarts
* no “manual push” required

---

# ✅ High-Level Flow (LO perspective)

```
loop every X seconds:
    desired_state = read_from_git(site)
    observed_state = get_from_nodes(site)

    diff = compare(desired_state, observed_state)

    for each change in diff:
        take_action()
```

---

# 🧠 Detailed Step-by-Step Reconciliation (Your Architecture)

### **1. LO reads desired state from Git**

* Clone/pull the `deployments` repo
* Parse all YAMLs for its site:

  ```
  deployments/site-A/*.yaml
  ```
* Build an in-memory object:

  ```go
  map[deploymentID]DesiredState
  ```

### **2. LO collects observed state**

LO maintains a dynamic view of EN activity:

* Heartbeat messages from EN
* Status events:

  * pending
  * installing
  * installed
  * failed
* EN container runtime query:

  * what is running
  * what version
  * health
  * image hash

→ Build this into:

```go
map[nodeID]ObservedState
```

---

# 🔍 3. LO does a diff: Desired vs Observed

### Case A: **Deployment not running but should be**

→ LO sends an InstallRequest to EN
This is the most common scenario.

### Case B: **Deployment running but version is outdated**

→ LO sends an UpdateRequest to EN

### Case C: **Deployment running but NOT in desired state anymore**

GitOps-style deletion.

→ LO sends DeleteRequest to EN

### Case D: **EN reports Failed**

→ LO either:

* retry (3 times)
* mark failed and report to CO
* or roll back to previous version (if enabled)

### Case E: **EN offline**

* LO waits
* When it comes back:

  * run reconciliation again
  * deploy the desired version immediately

---

# 🔁 4. LO sends action requests to EN

Actions generated from the diff:

| Action  | Trigger                      | EN Response                      |
| ------- | ---------------------------- | -------------------------------- |
| Install | Not running but should       | pending → installing → installed |
| Update  | Running but wrong version    | installing → installed           |
| Delete  | Running but removed from Git | deleted                          |
| Restart | Running but unhealthy        | restarted                        |

---

# 📡 5. EN executes and sends status events back to LO

EN states drive reconciliation:

```
pending → installing → installed  (success)
pending → installing → failed     (failure)
```

LO updates its internal map and forwards events to CO.

---

# ♻️ 6. LO re-runs reconciliation until state matches

Reconciliation is continuous and idempotent.

Meaning:

* If LO crashes → next loop reconciles automatically.
* If EN crashes → when it comes back, reconciliation fixes it.
* If Git changed → next loop applies the change.

This is why GitOps is reliable.

---

# 🌪️ Example Reconciliation Scenarios

## **Scenario 1: New Deployment Pushed to Git**

Git:

```
site-A/app1.yaml
```

LO diff:

* Desired: app1 (v1)
* Observed: nothing → install

---

## **Scenario 2: New Version Released**

Git:

```
app1: version 2
```

LO diff:

* Desired: v2
* Observed: v1 → update

---

## **Scenario 3: App Deleted in Git**

Git:

```
(no app1.yaml anymore)
```

LO diff:

* Desired: {}
* Observed: app1 running → delete

---

## **Scenario 4: EN Crashed / Rebooted**

Observed is empty → EN joins → LO sends install to match desired

---

## **Scenario 5: Container is Unhealthy**

Observed: unhealthy
Desired: healthy
→ LO sends restart

---

# 🎯 Important Design Rules

### **1. Reconciliation must be idempotent**

Sending the same install/update/delete request twice should be safe.

### **2. Reconciliation should never assume previous state**

Always re-derive the world state.

### **3. EN is “dumb”, LO is “smart”**

EN just runs containers.
LO does decision-making and tracking.

### **4. LO must avoid action loops**

Example: install → fail → install → fail → infinity
Use max retries & backoff.

### **5. Git is the source of truth**

Not CO, not LO, not EN.

---

