1. End-to-End Flow (Clean Sequence)

A. App Registration & Deployment Request
- CLI: co add app
- CLI pushes the app package → App Store (git).
- CLI sends Deploy Command → Central Orchestrator (CO).

B. CO Processes Deployment Request

- CO fetches margo.yaml from app registry (git).
- CO generates a DesiredState YAML for the selected sites.
- CO writes the DesiredState YAML into Git (infra repo).

C. Local Orchestrator (LO) reacts

- LO’s GitWatcher detects repo change for its site.
- GitManager reads desiredstate.yaml → stores DesiredState object locally.
- LO triggers Reconciler (GitPolled event).

D. Reconciler Loop

- Reconciler loads:
  - DesiredState (latest)
  - ActualState (local reality)
- ComputeDiff:
  - add app
  - remove app
  - noop
  - add component
  - update component
  - remove component
  - replace component (delete + add)
- Reconciler sends Operations → Actuator.
- Actuator sends commands to Edge Nodes (EN).

E. Execution and State Updates

- EN reports status → LO.
- If successful ("installed"): LO updates ActualState.
- LO sends deployment status to CO.
- CO updates desired-state version/event in git.

F. LO sees new desired-state event

- GitWatcher reads updated desiredstate.yaml again.
- Reconciler runs again → ComputeDiff → operations → actuator.
- This continues until:
    - Actual == Desired, → NOOP.

Failure Handling

- If any operation fails:
    - ActualState is NOT updated
    - Desired and Actual diverge

- Next reconcile → diff still exists → op is retried
- Retry will be idempotent (install again/update again)
- Rollback = simply create NEW desired state (previous version). Reconciler will converge to it.





- cli add app
- app goes to app store
- cli sends deploy command to co
- co reads the margo.yaml from the path mentioned in the cli command in the mentioned app registry (git)
- co creates desiredstate yaml for the targeted sites(selected by the user). 
- writes to the git

- lo monitoring/watching the git for changes for this site
- when there is a desiredstate appears in git for the site, git manager reads the desiredstate.yaml
- Git Watcher + git manager
- Gitpolled event, Reconciler triggered
- Computediff is called
- Send operation (add app, remove app, noop, add component, update comp remove comp) to actuator
- Actuator sends operation to the target nodes
- Target node responds with the deployment status. If "installed " update actual.  Now desired and actual are same. 
- Co triggers desired state event in a git. 
- Git watcher receives the event. Verifies the event is for this site. If the event is for this event, process the event. Reads the desiredstate.yaml. parses it. Store the desired event. 
- Call reconciler. Reconciler starts the process. Reads the desired event and actual event from the store. Computes the diff.(between desired and actual). It generates operations. Send to the target nodes.  
- Once the operations are completed by EN asynchronously, actual is updated for the host. Now actual for this host matches the desired state. Lo will generate deployment status for this host and sends it to co. Same thing happen for other hosts.
- En sends in progress status to lo. Right now it is ignored. In the future it is forwarded to co .
- If deployment is failed, no update of actual for this host. There will be diff between desired and actual.
- Rollback may be done as a new desired state. 
- When a new version is to be installed. Create a new desired state
- When a component has to added, when a component has to be deleted, when a component is upgraded
- when a component has to be replaced, 
- Noop when desired and actual are same.
- When a desired before current one is in progress. what to be done? 
- Version change = update
- Component replacement = remove old + add new
- App replacement = entire new desired
