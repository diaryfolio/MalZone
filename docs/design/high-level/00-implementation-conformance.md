# MalZone Implementation Conformance

This is the truth map between the target architecture and this repository. Status labels are:

- **implemented:** working code/configuration and an automated verification path exist;
- **configuration-ready:** a working boundary exists but requires environment-owned infrastructure
  or credentials and has a documented conformance test;
- **designed:** the behavior and acceptance evidence are specified, but executable code is absent;
- **not started:** only a product idea or placeholder exists.

Documentation is never sufficient to mark a capability implemented.

## Repository state on 2026-08-27

```mermaid
flowchart LR
    Prompt["Initial prompt"] --> Design["Canonical high-level design"]
    Design --> Governance["design-sync governance + CI gate"]
    Design --> SoftwareContracts["software package + image recipe schemas"]
    Governance & SoftwareContracts --> POC["harmless Linux lifecycle POC"]
    POC --> AgentPOC["observe-only action + ECS metadata POC"]
    Design --> Pages["allow-listed static documentation site"]
    POC -. "does not prove" .-> Production["Windows + KubeVirt production runtime"]
    Production -. "next" .-> Cluster["negative isolation + lifecycle evidence"]
```

| Capability | Status | Current evidence | Promotion evidence required |
|---|---|---|---|
| Product and architecture definition | implemented | `docs/design/high-level/` | Keep all design links and checks passing |
| Repository design-sync governance | implemented | `CLAUDE.md`, `AGENTS.md`, governance pack, reviewer-oriented PR template, automatic changed-file gate and automated tests | Keep CI required and extend synchronized path coverage with each implementation language/toolchain |
| Software package/image recipe contracts | implemented | two JSON schemas, fictional examples and contract tests | Add compatibility tooling and generated types when service implementation starts |
| Harmless lifecycle POC API | implemented | Go HTTP service, versioned POC OpenAPI, unit tests and k3d E2E script | Add authentication, durable product state, idempotency and production `/api/v1` contract before accepting samples |
| POC `Analysis` CRD and Job operator | implemented | namespace-scoped CRD, Go reconciler, context-aware cancellation, lifecycle/cleanup unit tests and k3d E2E evidence | Replace Job backend with the designed KubeVirt/relay/gateway lifecycle and finalizer/reaper fault tests |
| POC Helm packaging and runner isolation | implemented | Helm lint/render, least-privilege RBAC, restricted pod security, tokenless runner and deny-all runner NetworkPolicy | Validate production CNI/CSI/KubeVirt policies and forbidden paths on supported x86_64 virtualization nodes |
| POC bounded agent observation | implemented | observe-only action schema/API, CRD fields, operator result, shell/input denial unit tests and k3d E2E | Production OIDC/machine scope, durable action state, evidence cursor, policy service, session relay, Windows console and prompt-injection/lease/budget/approval tests |
| POC ECS lifecycle export | implemented | read-only adapter service account, deterministic ECS event ID, metadata-only mapping, in-memory deduplicating sink, disclosure unit test and k3d E2E | Canonical event subscription, durable checkpoint/retry/DLQ, disclosure engine, connector secret/TLS/egress controls and supported-SIEM conformance |
| GitHub Pages documentation portal | implemented | allow-listed static source builder, secret-pattern check, local Jekyll render/link validation and manual least-privilege Pages workflow | Enable repository Pages environment and successfully review a deployed artifact; keep deployment-private material outside the public allow-list |
| Software catalog and image-build runtime | designed | catalog/build/promotion design and ADR | API, local mirror, resolver, isolated builder, promotion and negative tests |
| Production Analysis REST/WebSocket API | designed | `/api/v1` contract in design; POC API is explicitly non-production | OpenAPI, OIDC/RBAC, persistence, implementation and unit/contract tests |
| Upload and hash verification | designed | data flow and security requirements | streamed upload tests, hash mismatch and quota negatives |
| Production `Analysis` CRD and KubeVirt operator | designed | lifecycle and CRD shape; POC CRD has only canary/timeout/cancel fields | production CRD, generated types, admission, envtest, idempotency/finalizer and real-cluster tests |
| Windows golden image pipeline | designed | deployment and security requirements | signed manifest, promotion records, clean-room rebuild test |
| Disposable KubeVirt clone | designed | clone/snapshot lifecycle | real CSI clone/boot/delete test and residue scan |
| Analysis-network isolation | designed | two-network topology | CNI-specific manifests plus prohibited-path packet tests |
| Session relay and identity | designed | protocol boundary | implementation, fuzzing, mTLS negatives, rate/size tests |
| Windows collection agent | designed | collector contract | signed binary, collector tests, tamper/degradation behavior |
| Network gateway/simulation/PCAP | designed | three network modes | offline/simulated/controlled egress tests and PCAP evidence |
| Metadata database and outbox | designed | ownership model | migrations, crash/replay tests, backup/restore test |
| Event stream and projections | designed | event envelope and ordering | schemas, compatibility tests, replay/deduplication test |
| Artifact storage and quarantine | designed | bucket/prefix policy | object policy tests, malicious preview/download tests |
| Web interface and console proxy | designed | UX/API boundary | browser tests, authorization and hostile-content tests |
| OIDC/RBAC/audit | designed | security model | IdP integration and role/cross-project negative tests |
| Report/export API and workflow integrations | designed | API/identity/integration design | OpenAPI, export jobs, webhook dispatcher, signed delivery and adapter tests |
| Production AI/playbook interaction | designed | bounded action/model architecture and ADR 0010; POC has no VM desktop | `/api/v1` schemas, OIDC/scopes, durable policy/audit, session relay, Windows action effects and prompt-injection/stale-cursor/budget/approval negatives |
| Production SIEM/OCSF/STIX integration | designed | canonical event/disclosure/adapter architecture and ADR 0010; POC sink is not a supported SIEM | durable event adapter, ECS/OCSF/STIX schemas, connector credentials, egress policy, backpressure/replay/DLQ and destination conformance |
| Metrics/logs/traces/SLOs | designed | operations baseline | dashboards, alerts, load and fault-injection evidence |
| HA, DR, and production upgrade | not started | roadmap requirements only | restore drill, rollback test, failure-domain test |

## Implemented POC runtime edges

There are no executable **production** runtime services in the repository yet. The following edges
exist only in the harmless namespace-scoped POC; it accepts no files, malware, commands, URLs, image
references, desktop input, or arbitrary agent tools and must not be used for malware. Undocumented
implemented edges fail design conformance.

| Source | Destination | Protocol/authentication | Data | Bound and denial evidence |
|---|---|---|---|---|
| local developer | POC API through `kubectl port-forward` | HTTP; no application authentication | bounded canary request/status only | ClusterIP only; 4 KiB body; arbitrary sample kinds return 422 |
| POC API | Kubernetes API | HTTPS; namespace-scoped `malzone-api` service account | create/get/list/patch POC `Analysis` CRs | 10-second client timeout; RBAC denies Jobs, Pods, Secrets and other namespaces |
| POC operator | Kubernetes API | HTTPS; namespace-scoped `malzone-operator` service account | list/status CRs; create/get/delete Jobs; list/read runner logs | 10-second client timeout; no Secrets or cross-namespace access |
| POC runner | Kubernetes API and all network destinations | TCP attempted only as a denial canary; no token | no application data | deny-all ingress/egress NetworkPolicy and `automountServiceAccountToken: false`; result records both denials |
| POC operator | runner Pod log subresource | HTTPS; namespace-scoped service account | at most 16 KiB structured canary output | RBAC limited to pods/log; arbitrary output is not executed or rendered |
| local developer/agent | POC action API through `kubectl port-forward` | HTTP; no application authentication | observe rationale/expectation only | Running phase only; 2 KiB body; 20-action CRD budget; shell/click/type/launch and unknown fields return 422/400 |
| POC API | Kubernetes API | HTTPS; namespace-scoped `malzone-api` service account | append observe-only action to one POC `Analysis` spec | resourceVersion conflict prevents lost update; CRD enum and bounds reject other action types |
| POC operator | Kubernetes API | HTTPS; namespace-scoped `malzone-operator` service account | append safe runner-state observation to `Analysis` status | no desktop/sample access; one result per action ID; status subresource only |
| POC SIEM adapter | Kubernetes API | HTTPS; namespace-scoped `malzone-siem-adapter` service account | list terminal `Analysis` lifecycle metadata and SHA-256 | get/list only; Secrets, Jobs, Pods, status writes and other namespaces denied |
| POC SIEM adapter | development SIEM sink | HTTP/ClusterIP | no destination credential in POC | exact `/events`; 5-second timeout; terminal cleanup-verified events only; no sample content, summary, rationale or observation detail |
| local developer | development SIEM sink through `kubectl port-forward` | HTTP; no application authentication | list in-memory POC ECS events | ClusterIP only; test sink service account has no token/RBAC; never expose or use as a production SIEM |

## Target production runtime edges

| Source | Destination | Target protocol | Authentication | Intended data |
|---|---|---|---|---|
| browser/client | edge/API | HTTPS + WebSocket | OIDC access token | commands, metadata, live event envelopes |
| API | PostgreSQL | PostgreSQL/TLS | workload DB identity | product metadata and transactional outbox |
| dispatcher/operator | Kubernetes API | HTTPS | dedicated workload identity + RBAC | MalZone/KubeVirt resources only |
| relay | NATS | TLS | analysis-relay workload identity | bounded telemetry batches |
| relay | object broker/storage | HTTPS | analysis-scoped grant | one sample read; analysis-prefix writes |
| Windows agent | session relay secondary NIC | HTTPS/mTLS | one-use analysis certificate | heartbeat, sample request, events, artifacts |
| Windows VM | network gateway | IP | network placement; no trust | all detonation traffic |
| network gateway | approved public targets | proxied TCP/UDP | policy/session identity as applicable | controlled malware traffic |
| projector | PostgreSQL/object storage | TLS | dedicated workload identities | query projections and event chunks |
| image-build agent | per-build relay | HTTPS/mTLS | one-use build/session identity | package results, safe logs, provenance, candidate output |
| per-build relay | local installer/artifact broker | HTTPS | build-scoped grants | exact hash-bound installer reads and candidate writes |
| report/export worker | object storage and public API | TLS | workload identity + project policy | deterministic reports and bounded exports |
| webhook dispatcher/adapter | approved integration endpoint | HTTPS/mTLS as configured | connector-owned credential and signed delivery | safe event/report metadata allowed by disclosure policy |
| AI agent/planner | action and observation API | HTTPS | scoped machine identity, analysis interaction scope and optional approval | closed action proposals, sanitized evidence projection and cursor; never shell/VNC/Kubernetes credentials |
| interaction policy | session relay | authenticated internal protocol | workload identity plus analysis/session/controller lease | admitted normalized keyboard/pointer/wait/sample-launch events only |
| canonical event adapter | local SIEM/TIP | HTTPS/mTLS or destination protocol | adapter-owned connector credential | disclosure-controlled ECS/OCSF metadata or STIX IOC/report exports |

## Implemented API surface

The implemented routes below are development-only and served by a ClusterIP service. They are not
the production `/api/v1` product contract.

| Service | Method and path | Authentication/scope | State owner | Contract and verification |
|---|---|---|---|---|
| POC API | `GET /healthz` | none; in-cluster/port-forward only | process | POC OpenAPI, Go handler test, Kubernetes liveness probe |
| POC API | `GET /readyz` | none; in-cluster/port-forward only | Kubernetes API reachability | POC OpenAPI, Go handler path, Kubernetes readiness probe |
| POC API | `GET /api/v1alpha1/analyses` | none; one installation namespace only | POC `Analysis` CRs | POC OpenAPI and k3d E2E |
| POC API | `POST /api/v1alpha1/analyses` | none; one installation namespace only | POC `Analysis` CRs | 4 KiB JSON, canary-only schema, Go negative tests and k3d E2E |
| POC API | `GET /api/v1alpha1/analyses/{name}` | none; one installation namespace only | POC `Analysis` status | POC OpenAPI and k3d E2E terminal/cleanup assertions |
| POC API | `POST /api/v1alpha1/analyses/{name}/cancel` | none; one installation namespace only | `spec.cancelRequested` | POC OpenAPI and operator cancellation unit test |
| POC API | `POST /api/v1alpha1/analyses/{name}/actions` | none; one installation namespace only | observe-only entry in `spec.interactions` | POC OpenAPI/CRD, Go allow/deny tests and k3d E2E |
| development SIEM sink | `GET /healthz` | none; in-cluster/port-forward only | process | Go handler and Kubernetes probes |
| development SIEM sink | `POST /events` | none; ClusterIP accepts adapter traffic only by deployment convention | bounded in-memory map keyed by deterministic `event.id` | 16 KiB limit, strict ECS POC JSON, deduplication unit test and k3d E2E |
| development SIEM sink | `GET /events` | none; in-cluster/port-forward only | bounded in-memory event view | Go unit test and k3d E2E disclosure assertion |

`tests/test_design_alignment.py` scans supported server route registrations and rejects implemented
literal routes absent from this conformance document.

## Mandatory synchronization

Update this file in the same change whenever a service, route, event type, state owner, runtime
edge, trust boundary, readiness status, or verification path changes. An implementation is not
complete if this map claims more or less than the repository proves.
