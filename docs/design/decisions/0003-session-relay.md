# ADR 0003: Analysis-Scoped Session Relay

- Status: accepted
- Date: 2026-08-27

## Context

Giving a compromised guest object-store, queue, database, or Kubernetes credentials makes the guest
an infrastructure client and expands the attack surface.

## Decision

Create a relay for each analysis. The Windows agent initiates mTLS to a small, versioned,
size/rate-limited protocol. The relay brokers one sample read, event batches, and writes to one
artifact prefix. It never routes IP, exposes generic URLs/proxying, or holds cross-analysis access.

## Consequences

Relay parser security, fuzzing, backpressure, identity rotation, and per-session deployment cost are
first-class concerns. Compromise has a smaller permission and lifetime boundary than a shared relay.

