# ADR 0002: Guest Network Isolation

- Status: accepted
- Date: 2026-08-27

## Context

A KubeVirt pod-network interface can make Kubernetes DNS, Services, nodes, and control-plane routes
reachable from a compromised guest. Kubernetes `NetworkPolicy` also cannot be assumed to govern
every secondary-network implementation.

## Decision

Do not attach the pod network to Windows. Explicitly disable automatic attachment and give the guest
two per-analysis Multus networks: a no-default-route management link to its session relay and a
detonation link to its gateway. Require CNI-specific secondary-network enforcement and tests.

## Consequences

The deployment needs Multus and an environment-selected, supported isolation provider. Network
inventory/cleanup and real-guest negative tests are release gates. Single-node bridge networking is
development evidence only.

