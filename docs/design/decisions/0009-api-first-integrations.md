# ADR 0009: API-First Identity, Exports, and Integrations

- Status: accepted
- Date: 2026-08-27

## Context

Clients need SSO, automation, monitoring, report extraction, SIEM/SOAR/TIP workflows, and a rich UI.
UI-only behavior or direct access to databases, queues, storage, Kubernetes, or VMs would create
privileged hidden contracts and vendor coupling.

## Decision

The UI, CLI, SDKs, machine clients, exports, and workflow adapters use committed versioned public
APIs. Human identity is OIDC; machines use narrow short-lived scopes. Integrations receive signed
metadata events or request authorized exports through isolated credential-owning adapters.
Observability uses local JSON/OpenMetrics/OpenTelemetry contracts.

## Consequences

MalZone needs committed OpenAPI, compatibility policy, machine identity, resumable streams,
asynchronous exports, signed webhook delivery, adapter descriptors/conformance, and strict
disclosure/audit controls. Core services remain provider-neutral and air-gapped by default.

