# ADR 0005: Local and Air-Gapped Baseline

- Status: accepted
- Date: 2026-08-27

## Context

The product goal is an interactive sandbox completely under the operator's control. Hidden cloud
dependencies would disclose samples/telemetry or make disconnected operation incomplete.

## Decision

All required control, identity, storage, queue, telemetry, detection, network simulation, console,
and reporting capabilities have self-hosted implementations and work without external routes.
Threat intelligence, public reputation, SIEM, or model services are optional adapters disabled by
the air-gapped profile.

## Consequences

Installation is larger and operators own updates, rules, capacity, keys, backups, and availability.
CI and release tests scan for runtime CDNs, fonts, analytics, license checks, update calls, and other
undeclared external dependencies.

