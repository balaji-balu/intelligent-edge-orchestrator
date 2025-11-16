
# 🔥 **Deployment State Machine (CO-LO-EN)**

```
 ┌────────────┐
 │   PENDING  │  (CO created deployment)
 └──────┬─────┘
        │ LO schedules deployment to EN
        ▼
 ┌────────────┐
 │ INSTALLING │  (EN pulling image, creating container)
 └──────┬─────┘
        │ success
        ▼
 ┌────────────┐
 │ INSTALLED  │  (running + healthy)
 └──────┬─────┘
        │ Git changed / new version
        ▼
 ┌────────────┐
 │ UPDATING   │
 └──────┬─────┘
        │ success
        ▼
 ┌────────────┐
 │ INSTALLED  │
 └────────────┘


      (failure paths)
        ▲
        │ error during install/update
 ┌──────┴─────┐
 │   FAILED   │
 └──────┬─────┘
        │ LO may retry or rollback
        ▼
 ┌────────────┐
 │ ROLLINGBACK│
 └──────┬─────┘
        │ rollback success
        ▼
 ┌────────────┐
 │ INSTALLED  │  (previous version)
 └────────────┘


       (delete path)
         ▲
         │ Git removed deployment
 ┌───────┴──────┐
 │  DELETING     │ (EN stopping container)
 └───────┬──────┘
         │ success
         ▼
 ┌────────────┐
 │  DELETED   │
 └────────────┘
```

---

# 🧠 **State Definitions**

### **1. PENDING (CO → LO)**

* CO created deployment entry
* LO has not delivered to EN yet
  Triggers:
* LO detects new desired-state in Git
* LO schedules deployment

---

### **2. INSTALLING (LO → EN)**

EN is:

* pulling image
* creating container
* setting up networks
* writing configs

---

### **3. INSTALLED (EN → LO → CO)**

* Container is running AND healthy
* Health check passed
* LO reports installed → CO updates DB
* This is a stable state until a change happens

---

### **4. UPDATING**

Triggered when:

* margo.yaml version changed
* deployment profile changed
* component changed
* config changed
  EN executes:
* pull new image
* stop old container
* create new container

---

### **5. FAILED**

EN sends error:

* image pull failed
* container creation failed
* health check failed
* node offline during install
  LO receives failure and can:
* retry (backoff)
* mark FAILED
* OR transition to ROLLINGBACK

---

### **6. ROLLINGBACK**

Used when:

* update fails
* LO has rollback policy enabled

EN:

* stops new version
* reverts to old version
* starts old container

---

### **7. DELETING**

Triggered when:

* Deployment YAML removed from Git
* Desired state no longer includes this deployment

EN:

* stops container
* removes image if required

---

### **8. DELETED**

Final state after EN confirms:

```
deleted: true
exit_code: 0
```

LO removes deployment entry from its registry
CO may archive it in DB

---

# 🚦 **State Transitions Table (Very Useful)**

| From        | To          | Trigger                   | Actor |
| ----------- | ----------- | ------------------------- | ----- |
| PENDING     | INSTALLING  | LO schedules install      | LO    |
| INSTALLING  | INSTALLED   | install success           | EN    |
| INSTALLING  | FAILED      | install error             | EN    |
| INSTALLED   | UPDATING    | Git changed / new version | LO    |
| UPDATING    | INSTALLED   | update success            | EN    |
| UPDATING    | FAILED      | update error              | EN    |
| FAILED      | ROLLINGBACK | rollback enabled          | LO    |
| ROLLINGBACK | INSTALLED   | rollback success          | EN    |
| ANY         | DELETING    | Git removed               | LO    |
| DELETING    | DELETED     | delete success            | EN    |

---

# 🏁 **Final Architecture Note**

Your CO-LO-EN state machine is **GitOps compliant**, **edge-friendly**, and **failure aware**.

This is exactly what we see in:

* Kubernetes reconciliation loops
* ArgoCD GitOps state machines
* Hashicorp Nomad’s deployment lifecycle
* AWS IoT Greengrass component lifecycles

Well designed!

---

