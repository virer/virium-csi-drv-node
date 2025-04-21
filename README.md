# Virim-iSCSI-csi-driver for Kubernetes

### Overview

This is a repository for Virium CSI driver.

This iscsi driver extends this repository https://github.com/kubernetes-csi/csi-driver-iscsi capability.

CSI plugin name: `virium.csi.virer.net`. 

This driver requires existing and already configured Virium server.
And could dynamically attach/mount, detach/unmount based on CSI GRPC calls, goal of this project is to be able to create, resize and delete volume, create snapshot and add clone capabilities

### Project status
Project status: ultra early alpha

### Install driver on a Kubernetes cluster

- Check "docs" directory
