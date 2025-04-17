#!/bin/bash

echo "Uninstalling virium-iscsi.csi.k8s.io CSI driver ..."
kubectl delete -f csi-iscsi-driverinfo.yaml
kubectl delete -f csi-iscsi-node.yaml
echo 'virium-iscsi.csi.k8s.io CSI driver uninstalled successfully.'
