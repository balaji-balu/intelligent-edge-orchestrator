lo.start
    in loop
        - networkmodechange (given to dispatcher) -->
          - push mode
          - pull mode  
            - watcher (+ gitmanager) : observes the git for desired changes for the this site
            - any change, triggers gitpolled event(dispatcher call gitpolled event)
          - offline mode
    nats : subscribe health receive from en 
        - host lifecycle. health monitoring. receives health from en. heartbeat monitor. 
    nats : subscribe status receive from en 
        - update the actual hash when status is success

gitpolled event
    - calculate hash for each component and store
    - call reconcile, operations output
      - call diff
    - for each host in the selected list
      - actuator execute is called
        - send the request to en via nats 

Schema: 
```
siteinfo/
    site-id → { site metadata }

desired/
    site-id/
        app-id/
            version
            components/
                comp-name → { version, content }

hosts/
    host-id → { Alive: true/false }

actual/
    site-id/
        host-id/
            app-id/
                version
                components/
                    comp-name → { status, last_updated, hash }

ops/
    site-id/
        host-id/
            op-uuid → { action, app, comp, version, timestamp }

site_state/
    site-id → { last_desired_sync, last_actual_sync }


```
| Scenario                     | Condition                                | Action                        |
| ---------------------------- | ---------------------------------------- | ----------------------------- |
| App version differs          | desired vs actual                        | Full app update               |
| Component version differs    | app version same, component hash differs | Component-level update op     |
| App removed                  | exists in actual, not in desired         | Remove app and all components |
| Component removed            | exists in actual, not in desired         | Remove component              |
| App added                    | exists in desired, not in actual         | Install app + all components  |
| Component added              | exists in desired, not in actual         | Install component             |
| App and components identical | versions & hashes match                  | Do nothing                    |

```