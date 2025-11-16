features:
| Connectivity resilience | Local autonomy when network link to CO fails |
| Security | Mutual TLS between CO, LO, and EN; Role-based access (Keycloak) |
| Multi-tenancy | Logical isolation of users and sites |

capture success metrics:
- Deployment success rate > 99%
- End-to-end latency (CO→EN) < 2s under normal network
- Offline operation capability at least 24h
- Seamless recovery after network reconnection