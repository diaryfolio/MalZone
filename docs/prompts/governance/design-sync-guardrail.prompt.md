# MalZone Design-Sync Guardrail

Use this workflow before making code, contract, deployment, security, VM-image, or runtime changes.

## Mandatory workflow

1. Read `CLAUDE.md` and classify the requested change with
   `docs/prompts/governance/major-change-policy.md`.
2. Read the conformance map and affected high-level design/ADRs before implementation.
3. Identify changes to contracts, runtime edges, trust boundaries, data owners, privileges,
   credentials, cleanup obligations, deployment resources, SLOs, and failure modes.
4. Update design/contracts and implementation in the same change set.
5. Add tests that prove positive behavior and relevant denial/cleanup behavior.
6. Run `make design-check` plus all implementation-specific validation.
7. Summarize material design impact, validation evidence, and known gaps. Refuse to claim
   completion when design, contracts, implementation, deployment, or evidence disagree.

## Hard rules

- Never merge major behavior without matching high-level design and conformance updates.
- Never promote a capability beyond its executable evidence.
- Never claim network isolation from manifest rendering alone.
- Never bypass cleanup, quarantine, evidence provenance, or air-gapped defaults for convenience.
- If unsure whether a change is major or whether a control is enforced, treat it as major and the
  control as unproven.
