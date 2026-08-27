# MalZone Contracts

Contracts are the machine-readable boundary between MalZone components and clients. They are
versioned independently from implementations and are committed before a capability is promoted.

## Present contracts

- [Software catalog contracts](software/README.md)
  - [`SoftwarePackageVersion v1alpha1`](schemas/software-package-version-v1alpha1.schema.json)
  - [`WindowsImageRecipe v1alpha1`](schemas/windows-image-recipe-v1alpha1.schema.json)

## Planned contract groups

- `openapi/`: public and internal HTTP/WebSocket APIs;
- `crd/`: `Analysis`, image-build and related Kubernetes resource schemas;
- `events/`: analysis, collector, detection, build, report and webhook event schemas;
- `artifacts/`: result/provenance manifests and export descriptors;
- `profiles/`: analysis, network, image and integration adapter profiles.

CI validates committed examples and compatibility. A route, event, CRD field, profile, package,
artifact, export, or integration behavior is not implemented merely because prose describes it;
the [implementation conformance map](../docs/design/high-level/00-implementation-conformance.md)
records executable maturity.

