# ADR 0008: Immutable Software Recipes

- Status: accepted
- Date: 2026-08-27

## Context

Clients need selectable browsers, runtimes, document tools, licensed applications, locales, and
research utilities. Installing from the Internet or installing packages during detonation is
non-reproducible and exposes credentials/package infrastructure to a compromised guest.

## Decision

Represent each exact package version/revision with a hash-bound manifest. Compose exact pins into a
`WindowsImageRecipe`, then build, validate, sign, and promote an immutable golden image in an
isolated no-pod-network builder zone using a local package mirror. Analysts choose promoted images;
they do not mutate software at analysis start.

## Consequences

MalZone needs catalog/mirror/resolver/build/promotion services, protected builder nodes, license
handling, provenance/SBOMs, compatibility tests, and image lifecycle capacity. Custom images are
asynchronous but reproducible and air-gapped.

