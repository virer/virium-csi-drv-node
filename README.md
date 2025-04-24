# Virim iSCSI CSI Driver for Kubernetes

## 🚀 Overview

This repository contains the **Virim CSI driver (node part)** — a Kubernetes CSI-compatible plugin that extends the functionality of the [official iSCSI CSI driver](https://github.com/kubernetes-csi/csi-driver-iscsi) to support **dynamic provisioning**.

**CSI plugin name**: `virium.csi.virer.net`

This driver works in conjunction with a running and properly configured [**Viriumd** API server](https://github.com/virer/viriumd), which handles the underlying LVM and iSCSI operations.

---

## 🔧 Features

- Dynamic volume provisioning and deletion
- Multi-node support
- Attach, mount, detach, unmount workflows via CSI
- Kubernetes native PersistentVolume (PV) lifecycle integration
- Volume resizing support
- Snapshot creation
- Volume cloning

## ⚠️ Requirements

- A running instance of **Viriumd**, configured and reachable by the driver
- iSCSI initiator configured on all cluster nodes
- Kubernetes v1.20+ (CSI-compatible)

## 📚 References

- [Viriumd API Server](https://github.com/virer/viriumd)
- [Virium CSI Driver Controller](https://github.com/virer/virium-csi-drv-controller)
- [Kubernetes CSI Documentation](https://kubernetes-csi.github.io/docs/)
- [Upstream csi-driver-iscsi](https://github.com/kubernetes-csi/csi-driver-iscsi)

---

## 🤝 Contributions

This project is open for testing, feedback, and contributions

# Installation and documentation

Please check the content of [Virium CSI Driver Controller](https://github.com/virer/virium-csi-drv-controller) repository for documentation and Helm Charts.
