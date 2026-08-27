# ADR 0006: No Unrestricted Analysis Egress

- Status: accepted
- Date: 2026-08-27

## Context

Malware often needs network behavior to reveal itself, but arbitrary Internet access can harm third
parties, expose infrastructure, and turn the platform into an abuse source.

## Decision

Profiles offer only `offline`, `simulated`, and `controlled` networking. Controlled mode crosses a
separate sandbox-egress zone with destination-class, protocol, rate, byte, and time controls plus
capture and an independent emergency kill switch. There is no unrestricted mode.

## Consequences

Some malware will behave differently and analysts must see the active mode in results. The platform
needs local simulation services, SSRF/rebinding controls, abuse operations, and controlled-egress
risk acceptance.

