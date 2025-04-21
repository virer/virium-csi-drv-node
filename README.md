# Virim-iSCSI-csi-driver for Kubernetes

### Overview

This is a repository for Virium CSI driver.

This iscsi driver extends this repository https://github.com/kubernetes-csi/csi-driver-iscsi capability.

CSI plugin name: `virium.csi.virer.net`. 

This driver requires existing and already configured Viriumd API server.
Virum CSI driver can dynamically create and delete volume, attach/mount, detach/unmount based on CSI GRPC calls.
The goal of this project is to be able to also add the ability to resize volumes, create snapshot and add clone capabilities.

### Project status
Project status: beta 
Persistent Volume creation and deletion are working on a multinode cluster

### Install driver on a Kubernetes cluster

- Check the README.md file in the "deploy" directory
