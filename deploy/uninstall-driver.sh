#!/bin/bash

echo "Uninstalling virium.csi.virer.net CSI driver ..."
kubectl delete -f csi-iscsi-driverinfo.yaml
kubectl delete -f csi-iscsi-node.yaml
echo 'virium.csi.virer.net CSI driver uninstalled successfully.'
