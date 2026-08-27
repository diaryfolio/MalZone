# ADR 0001: Disposable KubeVirt Clones

- Status: accepted
- Date: 2026-08-27

## Context

Windows malware can obtain full guest control. Reusing or “cleaning” an executed VM makes clean
state dependent on detecting every persistence mechanism.

## Decision

Create one fresh writable KubeVirt VM/disk clone from an immutable approved offline snapshot for
each analysis. Never return a detonated VM or writable disk to a pool. Destroy the clone after
collection; pool only clean source snapshots or pre-created clones that have never booted a sample.

## Consequences

CSI snapshot/clone performance and cleanup become core capacity concerns. Golden-image promotion,
clone provenance, residue tests, and storage isolation are mandatory. Boot latency is accepted in
exchange for a clean-state invariant.

