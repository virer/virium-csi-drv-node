# Installation:

For installation please use helm charts (files located here are for dev usage)

Create a values.yaml 
```
virium:
  virium:
    image:
      repository: docker.io/scaps/virium-csi-driver-iscsi
      tag: v0.2.1.5
  nodeSelector:
    kubernetes.io/os: linux
viriumConfig:
  apiUsername: "virium_api_username"
  apiPassword: "virium_api_password"
  apiurl: "http://192.168.0.147:8787"
  initiator: "iqn.2025-04.net.virer.virium:target1"
  debug: "2"
```

Then configure the helm repository and deploy the charts
```
helm repo add virium https://virer.github.io/virium-helm-repo/charts/
helm repo update
helm search repo virium
helm install a1 virium --namespace=virium --create-namespace -f values.yaml 
```

### Tricks for RHOCP
```
kubectl patch --type merge -p '{"spec": {"claimPropertySets": [{"accessModes": ["ReadWriteOnce"]}]}}' StorageProfile virium
```