#!/bin/bash

echo "Installing virium-iscsi.csi.k8s.io CSI driver..."
kubectl apply -f csi-iscsi-driverinfo.yaml
kubectl apply -f rbac.yaml
kubectl apply -f csi-iscsi-node.yaml
echo 'virium-iscsi.csi.k8s.io CSI driver installed successfully.'
