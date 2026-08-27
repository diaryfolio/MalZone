# MalZone

MalZone is a planned, entirely self-hosted interactive malware-analysis platform. The product goal
is an analyst experience inspired by [ANY.RUN](https://any.run/)—live desktop interaction, process
tree and timeline, network/DNS/HTTP activity, files and registry, detections, artifacts, PCAP, and
reports—implemented on infrastructure the operator controls.

MalZone is not affiliated with ANY.RUN and does not depend on its service. The required control,
identity, virtualization, telemetry, storage, detection, and reporting paths are designed to work
locally and air-gapped. Optional public-Internet access is a separately governed sandbox-egress
mode, never a requirement or default.

## Current status

The repository is in the architecture/design stage. No executable sandbox is claimed yet. The
[implementation-conformance map](docs/design/high-level/00-implementation-conformance.md) is the
authoritative record of what is implemented, configuration-ready, designed, or not started.

The original idea is preserved at [docs/prompts/v1.md](docs/prompts/v1.md). The canonical design
derived from it starts at [docs/design/high-level/design_01.md](docs/design/high-level/design_01.md).

## Architecture in one view

```mermaid
flowchart LR
    Analyst["analyst browser/API"] --> Control["local edge + control plane"]
    Control --> KubeVirt["Kubernetes + KubeVirt operator"]
    KubeVirt --> VM["fresh disposable Windows VM"]
    VM --> Relay["analysis-scoped session relay"]
    VM --> Gateway["isolated network gateway + capture"]
    Relay --> Evidence["local NATS + PostgreSQL + S3 storage"]
    Gateway --> Sim["offline/simulated services"]
    Gateway -. "explicit controlled mode" .-> Internet["public Internet"]
```

The Windows guest receives no Kubernetes pod network, service-account token, host mount, shared
credential, or direct storage access. Each analysis receives a fresh snapshot clone, unique
networks and identity, bounded relay, and cleanup inventory. A terminal result is not published
until all session resources are gone.

## Design documentation

- [High-level design index](docs/design/high-level/README.md)
- [Objectives and product capability map](docs/design/high-level/10-overall/01-objectives-principles.md)
- [Runtime lifecycle and `Analysis` CRD](docs/design/high-level/10-overall/02-runtime-topology-lifecycle.md)
- [APIs, events, and data ownership](docs/design/high-level/10-overall/03-contracts-data.md)
- [Components and technology decisions](docs/design/high-level/10-overall/04-components-technology.md)
- [Kubernetes/KubeVirt deployment](docs/design/high-level/20-deployment/01-kubernetes-kubevirt.md)
- [Threat model and zero-trust controls](docs/design/high-level/30-security/01-threat-model-zero-trust.md)
- [Day-2 operations and SRE](docs/design/high-level/40-operations/01-day2-sre.md)
- [Delivery roadmap and acceptance gates](docs/design/high-level/50-roadmap/01-delivery-roadmap.md)
- [Architecture decisions](docs/design/decisions/README.md)

## Safety

MalZone is intended only for authorized defensive research, DFIR, and SOC work. Treat samples,
guest telemetry, screenshots, packet captures, dropped files, archives, reports, and filenames as
hostile. Do not connect an analysis network to corporate systems or enable controlled egress until
the real-cluster negative isolation suite passes.

## Design validation

The documentation checks use only the Python standard library:

```bash
make design-check
```
