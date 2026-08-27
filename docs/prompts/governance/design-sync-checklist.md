# MalZone Design-Sync Checklist

Use this checklist for every PR that changes code, contracts, infrastructure, policy, or behavior.

## Classification and impact

- [ ] Change classified `minor` or `major` using `major-change-policy.md`; uncertainty became major.
- [ ] Affected routes/events/contracts, runtime edges, state/data owners, trust boundaries,
      privileges, credentials, cleanup resources, SLOs, and failure modes are listed.
- [ ] A major change includes a completed major-change record.

## Design and contracts

- [ ] Implementation conformance reflects exact shipped maturity and evidence.
- [ ] Architecture index/flow and affected overall documents are updated.
- [ ] Deployment, security, operations, roadmap, and ADR documents are updated where affected.
- [ ] OpenAPI, CRD, event, profile, artifact-manifest, and compatibility contracts agree with code.
- [ ] README and navigation/indexes resolve and remain current.

## Containment and evidence

- [ ] Guest/pod/secondary-network reachability remains deny-by-default and is tested at runtime.
- [ ] Service accounts, RBAC, secrets, identities, object prefixes, and NATS subjects are least privilege.
- [ ] Cancellation/failure/timeout and crash paths preserve collection, revocation, cleanup, and residue proof.
- [ ] Hostile samples, telemetry, filenames, archives, previews, reports, and downloads remain quarantined/bounded.
- [ ] Audit, provenance, collector degradation, and known evidence gaps remain visible.

## Operations and validation

- [ ] Resource limits, probes, telemetry, alerts, capacity, backup/restore, rollout/rollback, and runbooks are synchronized.
- [ ] Unit, contract, integration, fault, negative security, artifact-safety, and cleanup tests match risk.
- [ ] `make design-check` and all affected executable validation passed.
- [ ] Security claims cite real enforcement evidence, not only generated/rendered configuration.

## Completion

- [ ] Material design impact, validation evidence, and known production gaps are summarized.
- [ ] Known production gaps are explicit.
- [ ] The automatic changed-file design-sync gate passes.

If any required box is unchecked, do not merge.
