# ADR 0004: Split Product and Runtime State Authority

- Status: accepted
- Date: 2026-08-27

## Context

The public product needs transactional users, projects, uploads, retention, and query state, while
Kubernetes reconciliation needs declarative desired/observed runtime state. Treating both databases
as co-equal authorities produces split-brain and partial-creation failures.

## Decision

PostgreSQL owns product/public state and commits creation with a transactional outbox. The namespaced
`Analysis` CR owns runtime reconciliation and resource inventory. Idempotent dispatch creates the CR;
status events project back into PostgreSQL. Repair loops reconcile missing/stale projections.

## Consequences

The implementation needs explicit invariants, idempotency keys, outbox/consumer checkpoints, and
repair tooling. The API does not directly create VMs, and the operator does not write application
tables.

