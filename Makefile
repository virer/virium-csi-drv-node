TAG=v0.1.5


clean:
	rm -f bin/virium-iscsiplugin

mod-check:
	go mod verify && [ "$(shell sha512sum go.mod)" = "`sha512sum go.mod`" ] || ( echo "ERROR: go.mod was modified by 'go mod verify'" && false )

newversion:
	echo ${TAG}
	sed -i "s/var version =.*/var version = \"${TAG}\"/g" cmd/virium-iscsiplugin/driver.go
	sed -i "s#docker.io/scaps/virium-csi-driver-iscsi.*#docker.io/scaps/virium-csi-driver-iscsi:${TAG}#g" deploy/csi-iscsi-daemonset.yaml

all:
	rm -f bin/virium-iscsiplugin
	cd cmd/virium-iscsiplugin; CGO_ENABLED=0 GOOS=linux go build -o ../../bin/virium-iscsiplugin 
	podman build -t docker.io/scaps/virium-csi-driver-iscsi:${TAG} . && podman push --authfile=${HOME}/.docker/dockerconfig docker.io/scaps/virium-csi-driver-iscsi:${TAG}

push:
	podman push --authfile=${HOME}/.docker/dockerconfig docker.io/scaps/virium-csi-driver-iscsi:${TAG}