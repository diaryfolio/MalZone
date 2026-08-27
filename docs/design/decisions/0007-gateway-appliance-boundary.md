# ADR 0007: Gateway Appliance Boundary

- Status: accepted
- Date: 2026-08-27

## Context

DHCP, simulation, NAT, packet capture, and filtering need elevated network privileges. A privileged,
dual-homed pod that joins the hostile guest network to the Kubernetes pod network would turn a
gateway flaw into a cluster path.

## Decision

The production per-analysis gateway is a hardened disposable Linux KubeVirt appliance with no pod
network, or an environment-owned sandbox network appliance providing equivalent per-session VRF/VNI
isolation. It uses separate detonation, relay-management, and optional sandbox-egress networks. It
streams bounded capture output to the session relay with its own one-use identity.

## Consequences

Each run has additional boot/capacity cost and the gateway image needs the same signed promotion and
cleanup discipline as Windows. A gateway pod is allowed only in a development profile that makes no
production isolation claim.

