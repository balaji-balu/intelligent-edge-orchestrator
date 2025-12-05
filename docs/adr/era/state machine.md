# 🔄 **ERA Deployment State Machine (Final Version)**

This is the state machine that your Runtime Manager + Lifecycle Controller should use.

---

# **1. High-Level Diagram**

```
+-------------------------------+
|        REQUEST RECEIVED       |
+-------------------------------+
                |
                v
+-------------------------------+
|          VALIDATING           |
+-------------------------------+
                |
                v
+-------------------------------+
|       FETCHING ARTIFACT       |
+-------------------------------+
                |
                v
+-------------------------------+
|    PREPARING ENVIRONMENT      |
+-------------------------------+
                |
                v
+-------------------------------+
|           INSTALLING          |
+-------------------------------+
                |
                v
+-------------------------------+
|           ACTIVATING          |
+-------------------------------+
                |
                v
+-------------------------------+
|          HEALTHCHECK          |
+-------------------------------+
         |             |
         | success     | failed
         v             v
+-----------------+  +--------------------+
|     RUNNING     |  |   ROLLBACK NEEDED  |
+-----------------+  +--------------------+
         |                        |
         | update request         |
         v                        v
+-------------------------------+  (from here transitions to)
|            UPDATING           | ------→ ROLLBACK
+-------------------------------+
         |
         v
+-------------------------------+
|            STOPPING           |
+-------------------------------+
         |
         v
+-------------------------------+
|           UNINSTALL           |
+-------------------------------+
         |
         v
+-------------------------------+
|           DELETED             |
+-------------------------------+
```

---

# **2. Full State List**

### **INITIAL STATES**

* `Requested`
* `Validating`

### **SETUP STATES**

* `Fetching`
* `Preparing`
* `Installing`
* `Activating`
* `HealthChecking`

### **PRIMARY STEADY STATE**

* `Running`

### **UPDATE STATES**

* `Updating`
* `RollingBack`

### **TEARDOWN STATES**

* `Stopping`
* `Uninstalling`
* `Deleted`

### **ERROR STATES**

* `Failed`
* `RollbackFailed`

---

# **3. State Transitions**

### **Install Flow**

```
Requested → Validating → Fetching → Preparing → Installing →
Activating → HealthChecking → Running
```

### **Update Flow**

```
Running → Updating → (Installing new version) → Activating → HealthChecking → Running
```

### **Rollback Flow**

```
(any failure before Running) → RollingBack → Running (previous version)
```

### **Delete Flow**

```
Running → Stopping → Uninstalling → Deleted
```

---

# **4. Per-State Responsibilities**

### **Validating**

* schema validation
* plugin availability
* version requirement checks

### **Fetching**

* download OCI/WASM/binary
* verify checksum/signature

### **Preparing**

* setup directories
* env variables
* volumes
* networking prerequisites

### **Installing**

* plugin: `Install()`
* prepare runtime-specific artefacts

### **Activating**

* plugin: `Start()`
* bind ports/namespaces

### **HealthChecking**

* plugin: `Status()` must return `Healthy`
* or use liveness probes

### **Updating**

* store previous version
* stop app → install new → start

### **RollingBack**

* stop failed version
* restore previous version

### **Stopping**

* plugin: `Stop()`

### **Uninstalling**

* plugin: `Delete()`
* cleanup leftover data

