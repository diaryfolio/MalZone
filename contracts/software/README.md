# Software Catalog Contracts

These contracts define the immutable inputs for composing MalZone Windows images:

- [`SoftwarePackageVersion`](../schemas/software-package-version-v1alpha1.schema.json) describes one
  exact installer version, hash, silent-install behavior, detection, licensing, and security review.
- [`WindowsImageRecipe`](../schemas/windows-image-recipe-v1alpha1.schema.json) pins a base image,
  mandatory MalZone platform bundle, exact software versions, OS settings, and validation suites.

Examples use fictional software and placeholder hashes; they are structurally representative but
must never be imported as real catalog entries:

- [Example browser package](examples/example-browser-1.0.0.json)
- [Example Windows image recipe](examples/windows-11-browser-lab.json)

Published manifests are immutable. A metadata or installer change creates a new `revision`; a vendor
version change creates a new version. Approved recipes never contain `latest`, ranges, mutable URLs,
credentials, inline scripts, or shell command strings. Artifact and hook references resolve only to
locally mirrored, hash-verified objects.

The complete lifecycle, client customization model, build security boundary, and promotion rules are
in [Software Catalog and Windows Image Composition](../../docs/design/high-level/10-overall/05-software-catalog-image-composition.md).

