#!/bin/bash

echo "Installing virium.csi.virer.net CSI driver..."
kubectl apply -f csi-virium-driverinfo.yaml
kubectl apply -f csi-virium-rbac.yaml
kubectl apply -f csi-virium-daemonset.yaml
echo 'virium.csi.virer.net CSI driver installed successfully.'
