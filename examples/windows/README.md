# Windows 11 KubeVirt Starter

The companion `windows11-enterprise-template.yaml.example` is a non-runnable design skeleton for a
future dedicated x86_64 KubeVirt cluster. It is deliberately halted and contains unresolved PVC and
Multus network placeholders. Do not apply it to the current k3d cluster.

The recommended starting point is Windows 11 Enterprise on x86_64 using customer-supplied licensed
media. Microsoft provides a 90-day Enterprise evaluation for professional evaluation, but MalZone
must not redistribute it or silently download it. The image pipeline must verify the published
SHA-256, attach reviewed VirtIO storage/network drivers during installation, install the signed
MalZone agent, disable automatic updates, seal provenance, and promote an immutable golden source.

At analysis time the operator must replace every placeholder with one per-analysis writable disk
clone and two per-analysis Multus networks. The guest has no pod network. The management network
permits outbound agent traffic only to its relay; the detonation network permits traffic only to
the session gateway. The VM remains halted until those resources, identities, and policies are
ready.

The current ARM64 k3d cluster is not the Windows target. KubeVirt documents important Arm64 limits,
including no CDI support in its Arm64 operations matrix; the cluster also lacks KubeVirt, Multus,
and snapshot-capable CSI storage. Software emulation would test syntax, not production isolation or
performance.

Official starting references:

- [Windows 11 Enterprise Evaluation](https://www.microsoft.com/en-us/evalcenter/evaluate-windows-11-enterprise)
- [KubeVirt Windows VirtIO drivers](https://kubevirt.io/user-guide/user_workloads/windows_virtio_drivers/)
- [KubeVirt persistent TPM and UEFI state](https://kubevirt.io/user-guide/compute/persistent_tpm_and_uefi_state/)
- [KubeVirt Arm64 VM limitations](https://kubevirt.io/user-guide/cluster_admin/virtual_machines_on_Arm64/)
- [KubeVirt Arm64 operation limitations](https://kubevirt.io/user-guide/cluster_admin/operations_on_Arm64/)

Before this skeleton becomes deployable, pin and test KubeVirt/CDI/Multus/CSI versions, choose a
supported storage/clone strategy for Windows 11 TPM/UEFI state, replace all placeholders, add
admission policy, and pass real guest forbidden-path and zero-residue tests.
