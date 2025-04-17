TAG=$( cat VERSION )
echo ${TAG}
sed -i "s/TAG=.*/TAG=${TAG}/g" Makefile
sed -i "s/var version =.*/var version = \"${TAG}\"/g" cmd/virium-iscsiplugin/driver.go
sed -i "s#docker.io/scaps/virium-csi-driver-iscsi.*#docker.io/scaps/virium-csi-driver-iscsi:${TAG}#g" deploy/csi-iscsi-daemonset.yaml