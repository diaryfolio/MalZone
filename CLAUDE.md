# MalZone Repository Instructions

These instructions apply to every human or AI-assisted change in this repository.

## Mission

MalZone is an entirely self-hosted, Kubernetes/KubeVirt-based interactive malware-analysis
platform. Preserve disposable execution, guest-to-cluster isolation, per-analysis identity and
networks, hostile-content quarantine, evidence provenance, deterministic cleanup, local/air-gapped
operation, and honest implementation status in every change.

## Mandatory design-sync workflow

Before editing code, APIs, contracts, deployment assets, VM images, network policy, security policy,
dependencies, persistence, or runtime configuration:

1. Read `docs/prompts/governance/major-change-policy.md` and classify the change as `minor` or
   `major`. If uncertain, use `major`.
2. Read `docs/design/high-level/00-implementation-conformance.md` and every affected design document.
3. Record the affected contracts, runtime edges, trust boundaries, state/data owners, Kubernetes or
   KubeVirt resources, credentials, cleanup obligations, SLOs, failure modes, and verification paths.
4. For a major change, create a change record from
   `docs/prompts/governance/major-change-template.md` in the PR description or review artifact.

In the same change set:

1. Update affected high-level design and ADRs before declaring implementation complete.
2. Update `docs/design/high-level/00-implementation-conformance.md` whenever a service, route,
   event type/schema, runtime edge, state owner, CRD field/state, trust boundary, readiness label,
   deployment resource, or verification path changes.
3. Update OpenAPI, CRD, event, artifact-manifest, profile, and compatibility contracts when behavior
   changes. Never allow prose, generated contracts, and runtime behavior to disagree.
4. Update the threat model, NetworkPolicy/RBAC/secret requirements, artifact handling, audit events,
   and required negative tests when data flow, privileges, interaction, or reachability changes.
5. Update deployment topology, resource controls, probes, rollout/rollback, backup/restore,
   observability, residue cleanup, and runbooks when runtime behavior changes.
6. Add or update unit, contract, integration, fault-injection, cleanup, artifact-safety, and negative
   isolation tests proportional to risk.
7. Update the roadmap and conformance status honestly. `designed`, `configuration-ready`, and
   `implemented` are not interchangeable.
8. Keep `README.md`, the high-level design index, ADR index, governance index, and published
   navigation synchronized when documents are added, moved, or renamed.

## Architecture rules

- Use versioned APIs/events between components. Never read another service's database tables or
  object prefixes directly.
- PostgreSQL owns public product state; the `Analysis` CR owns runtime reconciliation. Preserve the
  transactional-outbox/projection boundary.
- The API does not create raw VMs. The operator does not write application tables.
- One approved immutable golden snapshot produces one fresh writable clone per analysis. Never
  return a detonated VM or disk to a clean pool.
- The Windows guest has no Kubernetes pod-network interface, service-account token, host path,
  shared infrastructure credential, or direct database/queue/object-store access.
- Preserve separate per-analysis management and detonation networks. Do not assume Kubernetes
  `NetworkPolicy` controls secondary Multus networks; require provider-specific enforcement and
  real negative tests.
- The session relay terminates a narrow, authenticated protocol. It is never an IP router, generic
  proxy, remote shell, or credential bridge.
- The production network gateway is a no-pod-network appliance VM or equivalent external sandbox
  appliance, not a privileged dual-homed pod bridging hostile traffic into the pod network.
- Network modes are `offline`, `simulated`, and `controlled`. Do not add unrestricted egress.
- Cancellation, timeout, or failure never bypasses collection/finalization and cleanup. Do not mark
  an analysis terminal until the session resource inventory is empty and credentials are revoked.
- Kubernetes is the canonical packaging target. A single-node development cluster cannot prove
  host, CNI, storage, failure-domain, or production isolation.
- All required product paths must work locally and air-gapped. Optional external adapters are
  disabled by default and cannot receive samples or evidence implicitly.

## Security and evidence rules

- Treat the Windows kernel, agent, guest clock, telemetry, filenames, MIME types, screenshots,
  PCAPs, archives, documents, reports, and dropped files as hostile.
- Never render active artifact content in the main application origin. Preserve quarantine,
  bounded parsing, isolated preview, attachment download, and audit controls.
- Never log or trace sample/artifact bytes, tokens, certificates, object grants, clipboard content,
  full URLs, behavioral payloads, or secrets. Operational logs are structured JSON Lines with safe
  opaque correlation IDs and W3C trace context.
- No host networking, host PID/IPC, hostPath, device passthrough, service-account disk, privileged
  pod, added capability, or public egress may be introduced without a major-change record, threat
  model update, least-privilege rationale, and negative tests.
- Preserve one-use, analysis/session/audience-bound bootstrap and workload identities. The guest
  never receives NATS, PostgreSQL, S3, Kubernetes, signing, or infrastructure credentials.
- Security claims require enforcement evidence. Rendering YAML, linting Helm, or creating a
  `NetworkPolicy` object does not prove isolation. Test forbidden paths from the actual guest,
  relay, gateway, and relevant workloads with the selected CNI/CSI/KubeVirt versions.
- Never weaken containment to improve feature parity, startup latency, or developer convenience.

## Required validation

Always run:

```bash
make design-check
```

As implementation targets are introduced, the same change must add stable Make targets and CI for:

- source lint/format and unit tests;
- OpenAPI/CRD/event/schema compatibility tests;
- Helm/Kustomize rendering and policy validation;
- controller envtest and lifecycle/finalizer crash tests;
- local integration tests for database/outbox/NATS/object storage;
- real-cluster KubeVirt clone/boot/console/stop/cleanup tests;
- guest/relay/gateway negative connectivity and authorization tests;
- artifact parser/UI hostile-content tests;
- backup/restore and upgrade/rollback tests for production-impacting changes.

Do not add a validation requirement only in prose: add the executable target or record the gap as
`designed` in the conformance map.

## Required completion block

Every completed change must include:

```text
Design Sync Report
- Change Classification: <minor|major>
- Design Docs Updated: <list or none with reason>
- Contracts Updated: <list or none>
- Code/Deployment Areas Updated: <list or none>
- Architecture Delta: <summary or none>
- Threat/Trust Boundary Delta: <summary or none>
- Tests/Evidence: <list>
- Known Production Gaps: <list or none>
- Sync Status: <PASS|FAIL>
```

`Sync Status` is `FAIL` when implementation, contracts, security controls, deployment, operations,
tests, or design documentation disagree. Do not merge or claim completion with failed design sync.

