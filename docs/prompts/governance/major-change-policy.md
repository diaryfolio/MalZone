# MalZone Major Change Policy

This policy classifies repository changes for mandatory design synchronization.

## Major-change criteria

A change is `major` if any of the following is true:

1. It adds, removes, replaces, or splits a service, worker, collector, adapter, gateway, datastore,
   queue, operator, CRD, VM profile, image pipeline, or external integration.
2. It changes request/event flow, runtime edges, state authority, data ownership, object/bucket
   access, message subjects, or synchronization semantics.
3. It changes an API, CRD, event, artifact-manifest, profile, identity, or compatibility contract in
   a breaking way, or introduces a new version.
4. It changes a trust boundary or security control: OIDC/RBAC, workload/session identity, secrets,
   encryption, audit, quarantine, parser/preview/download behavior, console/clipboard/file transfer,
   or analyst/project isolation.
5. It changes guest reachability, Multus/CNI policy, NetworkPolicy, DNS, simulated services,
   controlled egress, packet capture, gateway placement, host/node access, or any privileged
   capability.
6. It changes VM cloning, snapshot/golden-image provenance, KubeVirt/CDI/CSI behavior, scheduling,
   resource limits, timeout, cancellation, finalizers, cleanup inventory, reaping, or residue proof.
7. It changes deployment topology, namespaces, service accounts, Kubernetes RBAC/admission, storage
   classes, production profiles, cluster/node separation, HA, DR, backup/restore, rollout/rollback,
   or compatibility support.
8. It changes SLOs, capacity/admission, logging, metrics, tracing, alerts, audit export, operational
   ownership, failure handling, or incident response.
9. It changes a capability's status between `not started`, `designed`, `configuration-ready`, and
   `implemented`, or changes the evidence used to justify that status.
10. It adds a dependency or feature that can execute untrusted content, open external connections,
    receive sensitive evidence, access host resources, or hold infrastructure credentials.

If none apply, classify the change as `minor`. If uncertain, use `major`.

## Minor changes

Minor changes include typo/clarity fixes with no semantic effect, test refactoring with identical
coverage, and internal code cleanup that changes no contract, route, runtime edge, state owner,
trust boundary, deployment resource, operational behavior, or maturity claim.

Minor does not mean “no design check.” Update the conformance map whenever its enumerated truth
changes, even if the implementation diff is small.

## Required updates by impact

| Impact | Required design synchronization |
|---|---|
| any route/service/event/state/maturity/evidence change | `docs/design/high-level/00-implementation-conformance.md` |
| system boundary or primary flow | `design_01.md` and relevant `10-overall/` documents |
| Kubernetes/KubeVirt/CNI/CSI/image/topology | `20-deployment/01-kubernetes-kubevirt.md` |
| trust/reachability/identity/artifact/console/audit | `30-security/01-threat-model-zero-trust.md` |
| SLO/telemetry/failure/backup/upgrade/runbook | `40-operations/01-day2-sre.md` |
| phase scope, acceptance evidence, maturity | `50-roadmap/01-delivery-roadmap.md` |
| durable architectural choice or reversal | new/superseding ADR under `docs/design/decisions/` |
| public documentation/navigation | `README.md` and design/governance indexes as applicable |

## Merge gate

A major change is not ready unless:

1. the PR/review record says `Change Classification: major`;
2. conformance and all affected design/contracts are updated in the same change set;
3. the architecture and trust-boundary delta is explicit;
4. tests/evidence and known production gaps are listed;
5. the mandatory Design Sync Report says `Sync Status: PASS`;
6. CI design/governance checks pass.

