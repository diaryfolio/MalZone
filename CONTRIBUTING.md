# Contributing to MalZone

Every change must follow the repository policy in [`CLAUDE.md`](CLAUDE.md), including human-authored
changes. Before implementation:

1. classify the change with the [major-change policy](docs/prompts/governance/major-change-policy.md);
2. read the implementation conformance map and affected design/ADRs;
3. update design, contracts, security, deployment, operations and tests in the same change set;
4. run `make design-check` plus all affected implementation validation;
5. complete the pull-request Design Sync Report and checklist.

Do not merge a major change without conformance and affected design updates. Do not claim security
or maturity beyond executable evidence. A rendered manifest is not proof of guest, CNI, storage,
artifact, identity, or cleanup isolation.

