/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	klog "k8s.io/klog/v2"
)

type driver struct {
	name          string
	nodeID        string
	version       string
	endpoint      string
	initiatorName string
	cap           []*csi.VolumeCapability_AccessMode
	nscap         []*csi.NodeServiceCapability
}

const (
	driverName = "virium.csi.virer.net"
)

var version = "v0.2.3.4"

func NewDriver(nodeID, endpoint, initiatorName string) *driver {
	klog.V(1).Infof("driver: %s version: %s nodeID: %s endpoint: %s initiator: %s", driverName, version, nodeID, endpoint, initiatorName)

	d := &driver{
		name:          driverName,
		version:       version,
		nodeID:        nodeID,
		endpoint:      endpoint,
		initiatorName: initiatorName,
	}

	if err := os.MkdirAll(fmt.Sprintf("/var/run/%s", driverName), 0o755); err != nil {
		panic(err)
	}
	d.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER})

	return d
}

func NewNodeServer(d *driver) *nodeServer {
	return &nodeServer{
		Driver: d,
	}
}

func (d *driver) Run() {
	s := NewNonBlockingGRPCServer()
	s.Start(d.endpoint,
		NewDefaultIdentityServer(d),
		nil,
		NewNodeServer(d))
	s.Wait()
}

func (d *driver) AddVolumeCapabilityAccessModes(vc []csi.VolumeCapability_AccessMode_Mode) []*csi.VolumeCapability_AccessMode {
	var vca []*csi.VolumeCapability_AccessMode
	for _, c := range vc {
		klog.Infof("enabling volume access mode: %v", c.String())
		vca = append(vca, &csi.VolumeCapability_AccessMode{Mode: c})
	}
	d.cap = vca
	return vca
}
